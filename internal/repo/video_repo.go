package repo

import (
	"context"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

// PgDriver is the interface for database operations
type PgDriver interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row
	Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error)
	Close()
}

type Metrics interface {
	RecordDBQuery(operation string, success bool, duration time.Duration)
	SetInternalCallTime(method string, startTime time.Time)
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
	metrics    Metrics
	logger     Logger
	timeoutSec int
}

// NewRepository creates a new repository instance
func NewRepository(db PgDriver, metrics Metrics, logger Logger, timeoutSec int) *Repository {
	return &Repository{
		db:         db,
		metrics:    metrics,
		logger:     logger,
		timeoutSec: timeoutSec,
	}
}
