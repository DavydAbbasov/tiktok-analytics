package repo

import (
	"context"
	"database/sql"
	"errors"
	"ttanalytic/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// PgDriver is the interface for database operations
type PgDriver interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Close()
}

// Logger is the interface for logging
type Logger interface {
	Errorf(format string, args ...any)
	Warnf(format string, args ...any)
	Infof(format string, args ...any)
	Info(args ...any)
}

// Repository handles database operations
type Repository struct {
	db         PgDriver
	logger     Logger
	timeoutSec int
}

// NewRepository creates a new repository instance
func NewRepository(db PgDriver, logger Logger, timeoutSec int) *Repository {
	return &Repository{
		db:         db,
		logger:     logger,
		timeoutSec: timeoutSec,
	}
}
func (r *Repository) FindVideoByTikTokID(ctx context.Context, tikTokID string) (*models.Video, error) {
	query := `
        SELECT id, tiktok_id, url, created_at, updated_at
        FROM videos
        WHERE tiktok_id = $1
    `

	var v models.Video

	err := r.db.QueryRow(ctx, query, tikTokID).
		Scan(
			&v.ID,
			&v.TikTokID,
			&v.URL,
			&v.CreatedAt,
			&v.UpdatedAt,
		)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNotFound
		}
		return nil, err
	}

	return &v, nil
}

// Реализация CreateVideo.
func (r *Repository) CreateVideo(ctx context.Context, input models.CreateVideoInput) (*models.Video, error) {
	query := `
        INSERT INTO videos (tiktok_id, url)
        VALUES ($1, $2)
        RETURNING id, tiktok_id, url, created_at, updated_at
    `

	var v models.Video

	err := r.db.QueryRow(ctx, query, input.TikTokID, input.URL).
		Scan(
			&v.ID,
			&v.TikTokID,
			&v.URL,
			&v.CreatedAt,
			&v.UpdatedAt,
		)
	if err != nil {
		return nil, err
	}

	return &v, nil
}
