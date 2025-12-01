package pgprovider

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Logger interface {
	Errorf(format string, args ...any)
	Infof(format string, args ...any)
}
type Provider struct {
	db        *pgxpool.Pool
	logger    Logger
	cs        string
	idlConns  int32
	openConns int32
	lifetime  int
}

func NewProvider(
	logger Logger,
	server string,
	database string,
	username string,
	password string,
	port int,
	idlConns int32,
	openConns int32,
	lifetime int,
) *Provider {
	cs := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		username, password, server, port, database,
	)

	return &Provider{
		logger:    logger,
		cs:        cs,
		idlConns:  idlConns,
		openConns: openConns,
		lifetime:  lifetime,
	}
}
func (p *Provider) Open(ctx context.Context) error {
	cfg, err := pgxpool.ParseConfig(p.cs)
	if err != nil {
		return fmt.Errorf("parse config: %w", err)
	}

	cfg.MaxConns = p.openConns
	cfg.MinConns = p.idlConns
	cfg.MaxConnLifetime = time.Duration(p.lifetime) * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("create pool: %w", err)
	}

	if pingErr := pool.Ping(ctx); pingErr != nil {
		pool.Close()

		return fmt.Errorf("ping database: %w", pingErr)
	}

	p.db = pool
	p.logger.Infof("Connected to PostgreSQL: %s/%s", cfg.ConnConfig.Host, cfg.ConnConfig.Database)

	return nil
}
func (p *Provider) DB() *pgxpool.Pool {
	return p.db
}

func (p *Provider) Close() {
	if p.db != nil {
		p.db.Close()
		p.logger.Infof("PostgreSQL connection closed")
	}
}

func (p *Provider) Stats() *pgxpool.Stat {
	if p.db != nil {
		return p.db.Stat()
	}

	return nil
}
