package repo

import (
	"context"
	"errors"
	"fmt"
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
    SELECT
        v.id,
        v.tiktok_id,
        v.url,
        v.current_views,
        v.current_earnings,
        v.created_at,
        v.updated_at
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
	err := r.db.QueryRow(
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

	r.logger.Infof("Repository: CreateVideo finish at:%v", time.Now())

	return &v, nil
}
func (r *Repository) AppendVideoStats(ctx context.Context, input models.CreateVideoStatsInput) error {
	start := time.Now()
	r.logger.Infof("Repository: AppendVideoStats start at:%v, video_id=%d", start, input.VideoID)

	ctx, cancel := context.WithTimeout(ctx, time.Duration(r.timeoutSec)*time.Second)
	defer cancel()

	query := `
        INSERT INTO video_stats  (video_id, views, earnings)
        VALUES ($1, $2, $3)
    `

	_, err := r.db.Exec(ctx, query,
		input.VideoID,
		input.Views,
		input.Earnings,
	)
	if err != nil {
		r.logger.Errorf("Repository: AppendVideoStats query error: %v", err)
		return err
	}

	r.logger.Infof("Repository: AppendVideoStats finish at:%v", time.Now())
	return nil
}
func (r *Repository) ListVideosForUpdate(ctx context.Context, minupdateage time.Duration, limit int) ([]*models.Video, error) {
	r.logger.Infof("Repository: ListVideosForUpdate minUpdateAge=%v limit=%d", minupdateage, limit)

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
            updated_at
        FROM videos
		WHERE updated_at <= $1
        ORDER BY updated_at ASC
        LIMIT $2
    `

	rows, err := r.db.Query(ctx, query, cutoff, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*models.Video

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
		); err != nil {
			return nil, err
		}
		result = append(result, &v)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	r.logger.Infof("Repository: ListVideosForUpdate finish at:%v, count=%d", time.Now(), len(result))
	return result, nil
}
func (r *Repository) UpdateVideoAggregates(ctx context.Context, input models.UpdateVideoAggregatesInput) error {
	start := time.Now()
	r.logger.Infof("Repository: UpdateVideoAggregates start at:%v, video_id=%d", start, input.VideoID)

	ctx, cancel := context.WithTimeout(ctx, time.Duration(r.timeoutSec)*time.Second)
	defer cancel()

	query := `
        UPDATE videos
        SET
            current_views    = $1,
            current_earnings = $2
        WHERE id = $3
    `

	_, err := r.db.Exec(ctx, query,
		input.Views,
		input.Earnings,
		input.VideoID,
	)
	if err != nil {
		r.logger.Errorf("Repository: UpdateVideoAggregates query error: %v", err)
		return err
	}

	r.logger.Infof("Repository: UpdateVideoAggregates finish at:%v", time.Now())
	return nil
}
func (r *Repository) GetVideoHistory(ctx context.Context, videoID int64, from, to *time.Time) ([]*models.VideoStatPoint, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(r.timeoutSec)*time.Second)
	defer cancel()

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

	rows, err := r.db.Query(ctx, query, args...)
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

	r.logger.Infof("Repository: GetVideoHistory finish at:%v, count=%d", time.Now(), len(result))
	return result, nil
}
