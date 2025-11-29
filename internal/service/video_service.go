package service

import (
	"context"
	"errors"
	"time"
	"ttanalytic/internal/models"
	"ttanalytic/internal/provider"
)

const (
	CurrencyUSD = "USD"
)

type Repository interface {
	FindVideoByTikTokID(ctx context.Context, tikTokID string) (*models.Video, error)
	CreateVideo(ctx context.Context, input models.CreateVideoInput) (*models.Video, error)
	AppendVideoStats(ctx context.Context, input models.CreateVideoStatsInput) error
	GetVideoHistory(ctx context.Context, videoID int64, from, to *time.Time) ([]*models.VideoStatPoint, error)
}
type Logger interface {
	Errorf(format string, args ...any)
	Warnf(format string, args ...any)
	Infof(format string, args ...any)
	Info(args ...any)
}

type Service struct {
	repo     Repository
	provider provider.TikTokProvider
	logger   Logger
}

func NewService(repo Repository, prov provider.TikTokProvider, logger Logger) *Service {
	return &Service{
		repo:     repo,
		provider: prov,
		logger:   logger,
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

		return models.TrackVideoResponse{
			VideoID:         video.ID,
			TikTokID:        video.TikTokID,
			URL:             video.URL,
			CurrentViews:    video.CurrentViews,
			CurrentEarnings: video.CurrentEarnings,
			Currency:        CurrencyUSD,
			LastUpdatedAt:   video.UpdatedAt.UTC().Format(time.RFC3339),
			CreatedAt:       video.CreatedAt.UTC().Format(time.RFC3339),
			Status:          "active",
		}, nil
	}

	//call the provider
	stats, err := s.provider.GetVideoStats(ctx, req.URL)
	if err != nil {
		s.logger.Errorf("TrackVideo: provider error for %s: %v", req.URL, err)
		return models.TrackVideoResponse{}, err
	}

	//calculate
	views := stats.Views
	earnings := s.calculateEarnings(stats.Views)

	//create video in db
	input := models.CreateVideoInput{
		TikTokID:        req.TikTokID,
		URL:             req.URL,
		CurrentViews:    views,
		CurrentEarnings: earnings,
	}

	//create new video in db
	video, err = s.repo.CreateVideo(ctx, input)
	if err != nil {
		s.logger.Errorf("TrackVideo: CreateVideo(%s) error: %v", req.TikTokID, err)
		return models.TrackVideoResponse{}, err
	}

	//write in jurnal
	if err := s.repo.AppendVideoStats(ctx,
		models.CreateVideoStatsInput{
			VideoID:  video.ID,
			Views:    views,
			Earnings: earnings,
		}); err != nil {
		s.logger.Errorf("Service: TrackVideo failed to append stats for video_id=%d: %v", video.ID, err)
		return models.TrackVideoResponse{}, err
	}

	//build response
	return models.TrackVideoResponse{
		VideoID:         video.ID,
		TikTokID:        video.TikTokID,
		URL:             video.URL,
		CurrentViews:    views,
		CurrentEarnings: earnings,
		Currency:        CurrencyUSD,
		LastUpdatedAt:   video.UpdatedAt.UTC().Format(time.RFC3339),
		CreatedAt:       video.CreatedAt.UTC().Format(time.RFC3339),
		Status:          "active",
	}, nil
}
func (s *Service) calculateEarnings(views int64) float64 {
	const c = 0.10
	return float64(views) / 1000.0 * c
}

func (s *Service) GetVideo(ctx context.Context, tikTokID string) (models.TrackVideoResponse, error) {
	video, err := s.repo.FindVideoByTikTokID(ctx, tikTokID)
	if err != nil {
		s.logger.Errorf("Service: GetVideo repo error: %v", err)
		return models.TrackVideoResponse{}, err
	}

	return models.TrackVideoResponse{
		VideoID:         video.ID,
		TikTokID:        video.TikTokID,
		URL:             video.URL,
		CurrentViews:    video.CurrentViews,
		CurrentEarnings: video.CurrentEarnings,
		Currency:        CurrencyUSD,
		LastUpdatedAt:   video.UpdatedAt.UTC().Format(time.RFC3339),
		CreatedAt:       video.CreatedAt.UTC().Format(time.RFC3339),
		Status:          "active",
	}, nil
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
