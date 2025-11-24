package service

import (
	"context"
	"errors"
	"time"
	"ttanalytic/internal/models"
)

type Repository interface {
	FindVideoByTikTokID(ctx context.Context, tikTokID string) (*models.Video, error)
	CreateVideo(ctx context.Context, input models.CreateVideoInput) (*models.Video, error)
}
type Logger interface {
	Errorf(format string, args ...any)
	Warnf(format string, args ...any)
	Infof(format string, args ...any)
	Info(args ...any)
}
type Service struct {
	repo   Repository
	logger Logger
}

func NewService(repo Repository, logger Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}
func (s *Service) TrackVideo(ctx context.Context, req models.TrackVideoRequest) (models.TrackVideoResponse, error) {
	start := time.Now()
	s.logger.Infof("Service: TrackVideo start at:%v", start)

	//validate
	tikTokID := req.TikTokID
	if tikTokID == "" {
		return models.TrackVideoResponse{},
			errors.New("tikTokID is empty and URL parsing is not implemented yet")
	}

	video, err := s.repo.FindVideoByTikTokID(ctx, tikTokID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			video = nil
		} else {
			s.logger.Errorf("TrackVideo: FindVideoByTikTokID(%s) error: %v", tikTokID, err)
			return models.TrackVideoResponse{}, err
		}
	}

	//if there is not video -> create new video in db
	if video == nil {
		input := models.CreateVideoInput{
			TikTokID: tikTokID,
			URL:      req.URL,
		}

		video, err = s.repo.CreateVideo(ctx, input)
		if err != nil {
			s.logger.Errorf("TrackVideo: CreateVideo(%s) error: %v", tikTokID, err)
			return models.TrackVideoResponse{}, err
		}
		s.logger.Infof("Created video %s with ID %d", tikTokID, video.ID)
	}

	resp := models.TrackVideoResponse{
		VideoID:        video.ID,
		TikTokID:       video.TikTokID,
		URL:            video.URL,
		Title:          "stub title",
		CurrentViews:   0,
		CurrentEarning: 0,
		Currency:       "USD",
		LastUpdatedAt:  time.Now().UTC().Format(time.RFC3339),
		Status:         "active",
	}

	finish := time.Now()
	s.logger.Infof("Service: TrackVideo finish at:%v", finish)

	return resp, nil
}
