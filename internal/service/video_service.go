package service

import (
	"context"
	"errors"
	"time"
	"ttanalytic/internal/models"
)

const (
	CurrencyUSD = "USD"
)

type Repository interface {
	FindVideoByTikTokID(ctx context.Context, tikTokID string) (*models.Video, error)
	CreateVideo(ctx context.Context, input models.CreateVideoInput) (*models.Video, error)
	AppendVideoStats(ctx context.Context, input models.CreateVideoStatsInput) error
	GetVideoHistory(ctx context.Context, videoID int64, from, to *time.Time) ([]*models.VideoStatPoint, error)
	SetVideoErrorStatus(ctx context.Context, videoID int64, errText string) error
	SetVideoStoppedStatus(ctx context.Context, videoID int64) error
}
type TikTokProvider interface {
	GetVideoStats(ctx context.Context, videoURL string) (*models.VideoStats, error)
}
type Transactor interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
type Logger interface {
	Errorf(format string, args ...any)
	Warnf(format string, args ...any)
	Infof(format string, args ...any)
	Info(args ...any)
}
type EarningsConfig struct {
	Rate float64
	Per  int64
}

func (e EarningsConfig) Calc(views int64) float64 {
	if e.Per == 0 {
		return 0
	}
	return float64(views) / float64(e.Per) * e.Rate
}

type initialVideoState struct {
	Views    int64
	Earnings float64
}
type Service struct {
	repo        Repository
	provider    TikTokProvider
	earningsCfg EarningsConfig
	logger      Logger
	transactor  Transactor
}

func NewService(repo Repository, prov TikTokProvider, earningsCfg EarningsConfig, logger Logger, transactor Transactor) *Service {
	return &Service{
		repo:        repo,
		provider:    prov,
		earningsCfg: earningsCfg,
		logger:      logger,
		transactor:  transactor,
	}
}

func (s *Service) TrackVideo(ctx context.Context, req models.TrackVideoRequest) (models.TrackVideoResponse, error) {
	//try to find existing video
	video, err := s.repo.FindVideoByTikTokID(ctx, req.TikTokID)
	if err != nil {
		if !errors.Is(err, models.ErrNotFound) {
			s.logger.Errorf("TrackVideo: FindVideoByTikTokID(%s) error: %v", req.TikTokID, err)
			return models.TrackVideoResponse{}, err
		}
	}

	//video already exists
	if video != nil {
		s.logger.Infof("TrackVideo: video %s found in DB, not calling provider", req.TikTokID)
		return s.buildTrackVideoResponse(video), nil
	}

	//call the provider
	stats, err := s.provider.GetVideoStats(ctx, req.URL)
	if err != nil {
		s.logger.Errorf("TrackVideo: provider error for %s: %v", req.URL, err)
		return models.TrackVideoResponse{}, err
	}

	//calculate
	initState := s.calculateInitialVideoState(stats)

	var createdVideo *models.Video

	err = s.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		// create video in db
		input := models.CreateVideoInput{
			TikTokID:        req.TikTokID,
			URL:             req.URL,
			CurrentViews:    initState.Views,
			CurrentEarnings: initState.Earnings,
			TrackingStatus:  models.VideoStatusActive,
		}

		video, err := s.repo.CreateVideo(txCtx, input)
		if err != nil {
			s.logger.Errorf("TrackVideo: CreateVideo(%s) error: %v", req.TikTokID, err)
			return err
		}

		createdVideo = video

		// write first point in journal
		statInput := models.CreateVideoStatsInput{
			VideoID:  video.ID,
			Views:    initState.Views,
			Earnings: initState.Earnings,
		}

		if err := s.repo.AppendVideoStats(txCtx, statInput); err != nil {
			s.logger.Errorf("Service: TrackVideo failed to append stats for video_id=%d: %v", video.ID, err)
			return err
		}
		return nil
	})
	if err != nil {
		return models.TrackVideoResponse{}, err
	}

	//build response
	return s.buildTrackVideoResponse(createdVideo), nil
}

func (s *Service) GetVideo(ctx context.Context, tikTokID string) (models.TrackVideoResponse, error) {
	video, err := s.repo.FindVideoByTikTokID(ctx, tikTokID)
	if err != nil {
		s.logger.Errorf("Service: GetVideo repo error: %v", err)
		return models.TrackVideoResponse{}, err
	}

	//build response
	return s.buildTrackVideoResponse(video), nil
}
func (s *Service) GetVideoHistory(ctx context.Context, videoID int64, from, to *time.Time) (models.VideoHistoryResponse, error) {
	points, err := s.repo.GetVideoHistory(ctx, videoID, from, to)
	if err != nil {
		s.logger.Errorf("Service: GetVideoHistory repo error: %v", err)
		return models.VideoHistoryResponse{}, err
	}

	historyVideo := make([]models.VideoStatPoint, 0, len(points))
	for _, p := range points {
		historyVideo = append(historyVideo, *p)
	}

	return models.VideoHistoryResponse{
		VideoID:      videoID,
		HistoryVideo: historyVideo,
	}, nil
}
func (s *Service) StopTracking(ctx context.Context, videoID int64) error {
	if err := s.repo.SetVideoStoppedStatus(ctx, videoID); err != nil {
		s.logger.Errorf("StopTracking: SetVideoStoppedStatus(%d) error: %v", videoID, err)
		return err
	}

	return nil
}

// helpers
func (s *Service) calculateEarnings(views int64) float64 {
	return s.earningsCfg.Calc(views)
}

func (s *Service) calculateInitialVideoState(stats *models.VideoStats) initialVideoState {
	views := stats.Views
	earnings := s.calculateEarnings(views)

	return initialVideoState{
		Views:    views,
		Earnings: earnings,
	}
}

func (s *Service) buildTrackVideoResponse(video *models.Video) models.TrackVideoResponse {
	return models.TrackVideoResponse{
		VideoID:         video.ID,
		TikTokID:        video.TikTokID,
		URL:             video.URL,
		CurrentViews:    video.CurrentViews,
		CurrentEarnings: video.CurrentEarnings,
		Currency:        CurrencyUSD,
		LastUpdatedAt:   video.UpdatedAt.UTC().Format(time.RFC3339),
		CreatedAt:       video.CreatedAt.UTC().Format(time.RFC3339),
		Status:          video.TrackingStatus,
	}
}
