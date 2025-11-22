package service

import "context"

type Repository interface {
	CreateEntity(ctx context.Context, entity models.Entity) (models.EntityID, error)
	GetEntityByID(ctx context.Context, id models.EntityID) (models.Entity, error)
	UpdateEntity(ctx context.Context, entity models.Entity) error
	DeleteEntity(ctx context.Context, id models.EntityID) error
	ListEntities(ctx context.Context, status *models.Status, limit, offset int) ([]models.Entity, int, error)
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
