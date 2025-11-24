package repo

import (
	"context"
	"errors"
	"time"
	"ttanalytic/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// PgDriver is the interface for database operations
type PgDriver interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row
	Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error)
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

// FindVideoByTikTokID fiend video - returns video with the db
func (r *Repository) FindVideoByTikTokID(ctx context.Context, tikTokID string) (*models.Video, error) {
	start := time.Now()
	r.logger.Infof("Repository: FindVideoByTikTokID start at:%v", start)

	ctx, cancel := context.WithTimeout(ctx, time.Duration(r.timeoutSec)*time.Second)
	defer cancel()

	query := `
        SELECT id, tiktok_id, url, created_at, updated_at
        FROM videos
        WHERE tiktok_id = $1
    `

	var v models.Video
	err := r.db.QueryRow(ctx, query, tikTokID).Scan(
		&v.ID,
		&v.TikTokID,
		&v.URL,
		&v.CreatedAt,
		&v.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrNotFound
		}
		return nil, err
	}

	finish := time.Now()
	r.logger.Infof("Repository: FindVideoByTikTokID finish at:%v", finish)

	return &v, nil
}

// CreateVideo creates a new video in the database
func (r *Repository) CreateVideo(ctx context.Context, input models.CreateVideoInput) (*models.Video, error) {
	start := time.Now()
	r.logger.Infof("Repository: CreateVideo start work:%v", start)

	ctx, cancel := context.WithTimeout(ctx, time.Duration(r.timeoutSec)*time.Second)
	defer cancel()

	query := `
        INSERT INTO videos (tiktok_id, url)
        VALUES ($1, $2)
        RETURNING id, tiktok_id, url, created_at, updated_at
    `

	var v models.Video
	err := r.db.QueryRow(ctx, query, input.TikTokID, input.URL).Scan(
		&v.ID,
		&v.TikTokID,
		&v.URL,
		&v.CreatedAt,
		&v.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrNotFound
		}
		r.logger.Errorf("FindVideoByTikTokID query error: %v", err)
		return nil, err
	}

	finish := time.Now()
	r.logger.Infof("Repository: CreateVideo finish at:%v", finish)

	return &v, nil
}
