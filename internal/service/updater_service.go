package service

import (
	"context"
	"fmt"
	"time"

	"ttanalytic/internal/config"
	"ttanalytic/internal/models"
	"ttanalytic/internal/provider"
)

type UpdaterRepository interface {
	ListVideosForUpdate(ctx context.Context, minupdateage time.Duration, limit int) ([]*models.Video, error)
	AppendVideoStats(ctx context.Context, input models.CreateVideoStatsInput) error
	UpdateVideoAggregates(ctx context.Context, input models.UpdateVideoAggregatesInput) error
}
type UpdaterConfig struct {
	Interval     time.Duration
	BatchSize    int
	MinUpdateAge time.Duration
}
type UpdaterService struct {
	repo        UpdaterRepository
	provider    provider.TikTokProvider
	logger      Logger
	cfg         UpdaterConfig
	earningsCfg config.EarningsConfig
}

func NewUpdaterService(
	repo UpdaterRepository,
	provider provider.TikTokProvider,
	logger Logger,
	cfg UpdaterConfig,
	earningsCfg config.EarningsConfig,
) *UpdaterService {
	return &UpdaterService{
		repo:        repo,
		provider:    provider,
		logger:      logger,
		cfg:         cfg,
		earningsCfg: earningsCfg,
	}
}
func (u *UpdaterService) Run(ctx context.Context) {
	u.logger.Info("Updater: run loop started")

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
	videos, err := u.repo.ListVideosForUpdate(ctx, u.cfg.MinUpdateAge, u.cfg.BatchSize)
	if err != nil {
		return fmt.Errorf("list videos for update: %w", err)
	}

	if len(videos) == 0 {
		u.logger.Info("updater: no videos to update")
		return nil
	}

	for _, v := range videos {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		info, err := u.provider.GetVideoStats(ctx, v.URL)
		if err != nil {
			u.logger.Errorf("updater: get info for video ID%s: URL%v", v.TikTokID, v.URL, err)
			continue
		}

		// //test without provider
		// info := struct{ Views int64 }{
		// 	Views: v.CurrentViews + 10000,
		// }

		oldViews := v.CurrentViews
		newViews := info.Views

		if newViews <= oldViews {
			u.logger.Infof(
				"updater: no new views for video %s (old=%d, new=%d)",
				v.TikTokID, oldViews, newViews,
			)
			continue
		}

		deltaViews := newViews - oldViews

		earningsDelta := (float64(deltaViews) / float64(u.earningsCfg.Per)) * u.earningsCfg.Rate
		newTotalEarnings := v.CurrentEarnings + earningsDelta

		statInput := models.CreateVideoStatsInput{
			VideoID:  v.ID,
			Views:    newViews,
			Earnings: newTotalEarnings,
		}

		if err := u.repo.AppendVideoStats(ctx, statInput); err != nil {
			u.logger.Errorf("updater: create stat for video %d: %v", v.ID, err)
			continue
		}

		updInput := models.UpdateVideoAggregatesInput{
			VideoID:  v.ID,
			Views:    newViews,
			Earnings: newTotalEarnings,
		}

		if err := u.repo.UpdateVideoAggregates(ctx, updInput); err != nil {
			u.logger.Errorf("updater: update aggregates for video %d: %v", v.ID, err)
			continue
		}
	}

	return nil
}
