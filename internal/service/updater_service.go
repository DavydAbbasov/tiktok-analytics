package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"ttanalytic/internal/config"
	"ttanalytic/internal/models"
)

type UpdaterRepository interface {
	ListVideosForUpdate(ctx context.Context, minupdateage time.Duration, limit int) ([]models.Video, error)
	AppendVideoStats(ctx context.Context, input models.CreateVideoStatsInput) error
	UpdateVideoAggregates(ctx context.Context, input models.UpdateVideoAggregatesInput) error
	SetVideoErrorStatus(ctx context.Context, videoID int64, errText string) error
}

type UpdaterConfig struct {
	Interval       time.Duration
	BatchSize      int
	MinUpdateAge   time.Duration
	MaxConcurrency int
}
type UpdaterService struct {
	repo        UpdaterRepository
	provider    TikTokProvider
	logger      Logger
	cfg         UpdaterConfig
	earningsCfg config.EarningsConfig
	transactor  Transactor
}

func NewUpdaterService(
	repo UpdaterRepository,
	provider TikTokProvider,
	logger Logger,
	cfg UpdaterConfig,
	earningsCfg config.EarningsConfig,
	transactor Transactor,
) *UpdaterService {
	return &UpdaterService{
		repo:        repo,
		provider:    provider,
		logger:      logger,
		cfg:         cfg,
		earningsCfg: earningsCfg,
		transactor:  transactor,
	}
}
func (u *UpdaterService) Run(ctx context.Context) {
	ticker := time.NewTicker(u.cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			u.logger.Infof("Updater: shutdown")
			return
		case <-ticker.C:
			if err := u.processBatch(ctx); err != nil {
				u.logger.Errorf("Updater: batch error: %v", err)
			}
		}
	}
}
func (u *UpdaterService) processBatch(ctx context.Context) error {
	for {

		videos, err := u.repo.ListVideosForUpdate(ctx, u.cfg.MinUpdateAge, u.cfg.BatchSize)
		if err != nil {
			return fmt.Errorf("list videos for update: %w", err)
		}

		if len(videos) == 0 {
			u.logger.Info("updater: no videos to update")
			break
		}

		sem := make(chan struct{}, u.cfg.MaxConcurrency)
		var wg sync.WaitGroup

		for _, v := range videos {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			sem <- struct{}{}
			wg.Add(1)

			video := v

			go func(video models.Video) {
				defer wg.Done()
				defer func() { <-sem }()

				if ctx.Err() != nil {
					return
				}

				//provider
				info, err := u.provider.GetVideoStats(ctx, video.URL)
				if err != nil {
					u.logger.Errorf("updater: get info for video ID=%s URL=%s: %v", video.TikTokID, video.URL, err)

					if setErr := u.repo.SetVideoErrorStatus(ctx, video.ID, err.Error()); setErr != nil {
						u.logger.Errorf("updater: failed to set error status for video %d: %v", video.ID, setErr)
					}
					return
				}

				//calculate
				statInput, aggInput, ok := u.prepareVideoUpdate(video, info)
				if !ok {
					return
				}

				//transaction
				if txErr := u.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
					if err := u.repo.AppendVideoStats(txCtx, statInput); err != nil {
						u.logger.Errorf("updater: create stat for video %d: %v", video.ID, err)
						return err
					}

					if err := u.repo.UpdateVideoAggregates(txCtx, aggInput); err != nil {
						u.logger.Errorf("updater: update aggregates for video %d: %v", video.ID, err)
						return err
					}

					return nil
				}); txErr != nil {
					u.logger.Errorf("updater: transaction failed for video %d: %v", video.ID, err)
					return
				}

			}(video)
		}

		wg.Wait()

		if ctx.Err() != nil {
			return ctx.Err()
		}

	}
	return nil
}
func (u *UpdaterService) prepareVideoUpdate(video models.Video, stats *models.VideoStats) (statInput models.CreateVideoStatsInput, aggInput models.UpdateVideoAggregatesInput, ok bool) {
	oldViews := video.CurrentViews
	newViews := stats.Views

	if newViews <= oldViews {
		u.logger.Infof(
			"updater: no new views for video %s (old=%d, new=%d)",
			video.TikTokID, oldViews, newViews,
		)
		return models.CreateVideoStatsInput{}, models.UpdateVideoAggregatesInput{}, false
	}

	deltaViews := newViews - oldViews
	earningsDelta := (float64(deltaViews) / float64(u.earningsCfg.Per)) * u.earningsCfg.Rate
	newTotalEarnings := video.CurrentEarnings + earningsDelta

	statInput = models.CreateVideoStatsInput{
		VideoID:  video.ID,
		Views:    newViews,
		Earnings: newTotalEarnings,
	}

	aggInput = models.UpdateVideoAggregatesInput{
		VideoID:  video.ID,
		Views:    newViews,
		Earnings: newTotalEarnings,
	}

	return statInput, aggInput, true
}
