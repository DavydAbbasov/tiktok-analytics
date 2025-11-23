package repo

// PgDriver is the interface for database operations
type PgDriver interface {
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
