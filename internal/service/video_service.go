package service

import (
	"context"
	"errors"
	"time"
	"ttanalytic/internal/models"
	"ttanalytic/internal/provider"
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
	start := time.Now()
	s.logger.Infof("Service: TrackVideo start at:%v", start)

	//validation
	tikTokID := req.TikTokID
	if req.URL == "" && req.TikTokID == "" {
		return models.TrackVideoResponse{}, errors.New("either url or tiktok_id must be provided")
	}

	//try to find existing video
	video, err := s.repo.FindVideoByTikTokID(ctx, tikTokID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			video = nil
		} else {
			s.logger.Errorf("TrackVideo: FindVideoByTikTokID(%s) error: %v", tikTokID, err)
			return models.TrackVideoResponse{}, err
		}
	}

	//video already exists
	if video != nil {
		s.logger.Infof("TrackVideo: video %s found in DB, not calling provider", tikTokID)

		resp := models.TrackVideoResponse{
			VideoID:         video.ID,
			TikTokID:        video.TikTokID,
			URL:             video.URL,
			Title:           "stub title",
			CurrentViews:    video.CurrentViews,
			CurrentEarnings: video.CurrentEarnings,
			Currency:        "USD",
			LastUpdatedAt:   video.UpdatedAt.UTC().Format(time.RFC3339),
			CreatedAt:       video.CreatedAt.UTC().Format(time.RFC3339),
			Status:          "active",
		}

		s.logger.Infof("Service: TrackVideo finish (from DB) at:%v", time.Now())
		return resp, nil
	}

	videoURL := req.URL

	//call the provider
	stats, err := s.provider.GetVideoStats(ctx, videoURL)
	if err != nil {
		s.logger.Errorf("TrackVideo: provider error for %s: %v", videoURL, err)
		return models.TrackVideoResponse{}, err
	}
	s.logger.Infof("Provider stats for %s: views=%d", videoURL, stats.Views)

	//calculate
	views := stats.Views
	earnings := s.calculateEarnings(stats.Views)

	//create video in db
	input := models.CreateVideoInput{
		TikTokID:        tikTokID,
		URL:             req.URL,
		CurrentViews:    views,
		CurrentEarnings: earnings,
	}

	//create new video in db
	video, err = s.repo.CreateVideo(ctx, input)
	if err != nil {
		s.logger.Errorf("TrackVideo: CreateVideo(%s) error: %v", tikTokID, err)
		return models.TrackVideoResponse{}, err
	}
	s.logger.Infof("Created video %s with ID %d", tikTokID, video.ID)

	//write in jurnal
	if err := s.repo.AppendVideoStats(ctx, models.CreateVideoStatsInput{
		VideoID:  video.ID,
		Views:    views,
		Earnings: earnings,
	}); err != nil {
		s.logger.Errorf("Service: TrackVideo failed to append stats for video_id=%d: %v", video.ID, err)
		return models.TrackVideoResponse{}, err
	}

	//build response
	resp := models.TrackVideoResponse{
		VideoID:         video.ID,
		TikTokID:        video.TikTokID,
		URL:             video.URL,
		Title:           "stub title",
		CurrentViews:    views,
		CurrentEarnings: earnings,
		Currency:        "USD",
		LastUpdatedAt:   video.UpdatedAt.UTC().Format(time.RFC3339),
		CreatedAt:       video.CreatedAt.UTC().Format(time.RFC3339),
		Status:          "active",
	}

	s.logger.Infof("Service: TrackVideo finish (created new) at:%v", time.Now())

	return resp, nil
}
func (s *Service) calculateEarnings(views int64) float64 {

	const c = 0.10
	return float64(views) / 1000.0 * c
}

func (s *Service) GetVideo(ctx context.Context, tikTokID string) (models.TrackVideoResponse, error) {
	start := time.Now()
	s.logger.Infof("Service: GetVideo start at:%v, tiktok_id=%s", start, tikTokID)

	v, err := s.repo.FindVideoByTikTokID(ctx, tikTokID)
	if err != nil {
		s.logger.Errorf("Service: GetVideo repo error: %v", err)
		return models.TrackVideoResponse{}, err
	}

	resp := models.TrackVideoResponse{
		VideoID:         v.ID,
		TikTokID:        v.TikTokID,
		URL:             v.URL,
		Title:           "stub title",
		CurrentViews:    v.CurrentViews,
		CurrentEarnings: v.CurrentEarnings,
		Currency:        "USD",
		LastUpdatedAt:   v.UpdatedAt.UTC().Format(time.RFC3339),
		CreatedAt:       v.CreatedAt.UTC().Format(time.RFC3339),
		Status:          "active",
	}

	s.logger.Infof("Service: GetVideo finish in %s", time.Since(start))
	return resp, nil
}
func (s *Service) GetVideoHistory(ctx context.Context, videoID int64, from, to *time.Time) (models.VideoHistoryResponse, error) {
	s.logger.Infof("Service: GetVideoHistory start video_id=%d", videoID)

	points, err := s.repo.GetVideoHistory(ctx, videoID, from, to)
	if err != nil {
		s.logger.Errorf("Service: GetVideoHistory repo error: %v", err)
		return models.VideoHistoryResponse{}, err
	}

	historyV := make([]models.VideoStatPoint, 0, len(points))
	for _, p := range points {
		historyV = append(historyV, *p)
	}
	resp := models.VideoHistoryResponse{
		VideoID:      videoID,
		HistoryVideo: historyV,
	}
	s.logger.Infof("Service: GetVideoHistory success video_id=%d count=%d", videoID, len(historyV))
	return resp, nil
}
