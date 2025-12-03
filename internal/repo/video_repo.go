package repo

import (
	"context"
	"errors"
	"fmt"
	"time"
	"ttanalytic/internal/infrastructure/dbtx"
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
type dbRunner interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
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

func (r *Repository) getDB(ctx context.Context) dbRunner {
	if tx, ok := dbtx.TxFromContext(ctx); ok {
		return tx
	}
	return r.db
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
	ctx, cancel := context.WithTimeout(ctx, time.Duration(r.timeoutSec)*time.Second)
	defer cancel()

	query := `
    SELECT
        v.id,
        v.tiktok_id,
        v.url,
        v.current_views,
        v.current_earnings,
        v.created_at,
        v.updated_at,
		v.tracking_status,
        v.last_error,
        v.last_error_at
    FROM videos v
    WHERE v.tiktok_id = $1
`

	var v models.Video
	err := r.db.QueryRow(ctx, query, tikTokID).Scan(
		&v.ID,
		&v.TikTokID,
		&v.URL,
		&v.CurrentViews,
		&v.CurrentEarnings,
		&v.CreatedAt,
		&v.UpdatedAt,
		&v.TrackingStatus,
		&v.LastError,
		&v.LastErrorAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrNotFound
		}
		return nil, err
	}

	return &v, nil
}

// CreateVideo creates a new video in the database
func (r *Repository) CreateVideo(ctx context.Context, input models.CreateVideoInput) (*models.Video, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(r.timeoutSec)*time.Second)
	defer cancel()

	db := r.getDB(ctx)

	query := `
    INSERT INTO videos (tiktok_id, url, current_views, current_earnings)
    VALUES ($1, $2, $3, $4)
    RETURNING
        id,
        tiktok_id,
        url,
        current_views,
        current_earnings,
        created_at,
        updated_at
`

	var v models.Video
	err := db.QueryRow(
		ctx,
		query,
		input.TikTokID,
		input.URL,
		input.CurrentViews,
		input.CurrentEarnings,
	).Scan(
		&v.ID,
		&v.TikTokID,
		&v.URL,
		&v.CurrentViews,
		&v.CurrentEarnings,
		&v.CreatedAt,
		&v.UpdatedAt,
	)
	if err != nil {
		r.logger.Errorf("CreateVideo query error: %v", err)
		return nil, err
	}

	return &v, nil
}
func (r *Repository) AppendVideoStats(ctx context.Context, input models.CreateVideoStatsInput) error {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(r.timeoutSec)*time.Second)
	defer cancel()

	db := r.getDB(ctx)

	query := `
        INSERT INTO video_stats  (video_id, views, earnings)
        VALUES ($1, $2, $3)
    `

	_, err := db.Exec(ctx, query,
		input.VideoID,
		input.Views,
		input.Earnings,
	)
	if err != nil {
		r.logger.Errorf("Repository: AppendVideoStats query error: %v", err)
		return err
	}

	return nil
}
func (r *Repository) ListVideosForUpdate(ctx context.Context, minupdateage time.Duration, limit int) ([]models.Video, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(r.timeoutSec)*time.Second)
	defer cancel()

	//control time update video
	cutoff := time.Now().Add(-minupdateage)

	query := `
        SELECT
            id,
            tiktok_id,
            url,
            current_views,
            current_earnings,
            created_at,
            updated_at,
			tracking_status,
    		last_error,
    		last_error_at
        FROM videos
		WHERE tracking_status = 'active'
  			AND updated_at <= $1
        ORDER BY updated_at ASC
        LIMIT $2
    `

	rows, err := r.db.Query(ctx, query, cutoff, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]models.Video, 0, limit)

	for rows.Next() {
		var v models.Video
		if err := rows.Scan(
			&v.ID,
			&v.TikTokID,
			&v.URL,
			&v.CurrentViews,
			&v.CurrentEarnings,
			&v.CreatedAt,
			&v.UpdatedAt,
			&v.TrackingStatus,
			&v.LastError,
			&v.LastErrorAt,
		); err != nil {
			return nil, err
		}
		result = append(result, v)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
func (r *Repository) UpdateVideoAggregates(ctx context.Context, input models.UpdateVideoAggregatesInput) error {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(r.timeoutSec)*time.Second)
	defer cancel()

	db := r.getDB(ctx)

	query := `
        UPDATE videos
        SET
            current_views    = $1,
            current_earnings = $2,
			    updated_at  = NOW()
        WHERE id = $3
    `

	_, err := db.Exec(ctx, query,
		input.Views,
		input.Earnings,
		input.VideoID,
	)
	if err != nil {
		r.logger.Errorf("Repository: UpdateVideoAggregates query error: %v", err)
		return err
	}

	return nil
}
func (r *Repository) GetVideoHistory(ctx context.Context, videoID int64, from, to *time.Time) ([]*models.VideoStatPoint, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(r.timeoutSec)*time.Second)
	defer cancel()

	db := r.getDB(ctx)

	query := `
        SELECT captured_at, views, earnings
        FROM video_stats
        WHERE video_id = $1
    `

	args := []any{videoID}
	argPos := 2

	if from != nil {
		query += fmt.Sprintf(" AND captured_at >= $%d", argPos)
		args = append(args, *from)
		argPos++
	}

	if to != nil {
		query += fmt.Sprintf(" AND captured_at < $%d", argPos)
		args = append(args, *to)
		argPos++
	}

	query += " ORDER BY captured_at ASC"

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*models.VideoStatPoint

	for rows.Next() {
		var v models.VideoStatPoint
		if err := rows.Scan(
			&v.CapturedAt,
			&v.Views,
			&v.Earnings,
		); err != nil {
			return nil, err
		}
		result = append(result, &v)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
func (r *Repository) SetVideoErrorStatus(ctx context.Context, videoID int64, errText string) error {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(r.timeoutSec)*time.Second)
	defer cancel()

	db := r.getDB(ctx)

	query := `
        UPDATE videos
        SET
            tracking_status = $1,
            last_error      = $2,
            last_error_at   = NOW(),
            updated_at      = NOW()
        WHERE id = $3
    `

	_, err := db.Exec(ctx, query,
		models.VideoStatusError,
		errText,
		videoID,
	)
	if err != nil {
		r.logger.Errorf("Repository: SetVideoErrorStatus video_id=%d error: %v", videoID, err)
		return err
	}

	return nil
}
func (r *Repository) SetVideoStoppedStatus(ctx context.Context, videoID int64) error {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(r.timeoutSec)*time.Second)
	defer cancel()

	db := r.getDB(ctx)

	query := `
        UPDATE videos
        SET
            tracking_status = $1,
            updated_at      = NOW()
        WHERE id = $2
    `

	_, err := db.Exec(ctx, query,
		models.VideoStatusStopped,
		videoID,
	)
	if err != nil {
		r.logger.Errorf("Repository: SetVideoStoppedStatus video_id=%d error: %v", videoID, err)
		return err
	}

	return nil
}
