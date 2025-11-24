package service

import (
	"context"
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
	// 1. Нормализуем вход (URL / TikTokID)
	tikTokID := req.TikTokID
	if tikTokID == "" {
		// потом сюда добавим парсинг из URL
		tikTokID = "parsed-from-url"
	}

	// 2. Ищем видео в БД
	video, err := s.repo.FindVideoByTikTokID(ctx, tikTokID)
	if err != nil {
		// тут будет разбор ошибок: not found / db error
		// псевдо:
		// if errors.Is(err, repository.ErrNotFound) { ... }
	}

	// 3. Если не нашли — создаём
	if video == nil {
		input := models.CreateVideoInput{
			TikTokID: tikTokID,
			URL:      req.URL,
		}

		video, err = s.repo.CreateVideo(ctx, input)
		if err != nil {
			return models.TrackVideoResponse{}, err
		}
		s.logger.Infof("Created video %s with ID %d", tikTokID, video.ID)
	}

	// 4. Позже здесь будет вызов провайдера (ensemble),
	//    сохранение снапшота в video_stats через другой репозиторий и расчёт earnings.

	// 5. Пока возвращаем заглушку на основе video
	resp := models.TrackVideoResponse{
		VideoID:        int64(video.ID),
		TikTokID:       video.TikTokID,
		URL:            video.URL,
		Title:          "stub title",
		CurrentViews:   0,
		CurrentEarning: 0,
		Currency:       "USD",
		LastUpdatedAt:  "2025-11-24T01:30:00Z",
		Status:         "active",
	}

	return resp, nil
}
