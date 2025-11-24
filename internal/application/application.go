package application

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
	"ttanalytic/internal/api"
	"ttanalytic/internal/api/handlers"
	"ttanalytic/internal/config"
	pgprovider "ttanalytic/internal/infrastructure"
	"ttanalytic/internal/repo"
	"ttanalytic/internal/service"

	"go.uber.org/zap"
)

type Application struct {
	cfg     *config.Config
	logger  *zap.SugaredLogger
	db      *pgprovider.Provider
	service handlers.Service
	repo    *repo.Repository
	router  *api.Router
	wg      sync.WaitGroup
}

func NewApplication() *Application {
	return &Application{}
}
func (a *Application) Start(ctx context.Context) error {
	if err := a.initConfig(); err != nil {
		return fmt.Errorf("init config: %w", err)
	}

	if err := a.initLogger(); err != nil {
		return fmt.Errorf("init logger: %w", err)
	}

	if err := a.initDatabase(ctx); err != nil {
		return fmt.Errorf("init database: %w", err)
	}

	if err := a.initRepository(); err != nil {
		return fmt.Errorf("init repository: %w", err)
	}
	if err := a.initService(); err != nil {
		return fmt.Errorf("init service: %w", err)
	}
	if err := a.initRouter(); err != nil {
		return fmt.Errorf("init router: %w", err)
	}

	a.startHTTPServer()

	a.logger.Info("Application started successfully")

	return nil
}

func (a *Application) Wait(ctx context.Context, cancel context.CancelFunc) error {
	defer cancel()

	<-ctx.Done()
	a.logger.Info("Shutdown signal received, starting graceful shutdown...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := a.router.Shutdown(shutdownCtx); err != nil {
		a.logger.Errorf("HTTP server shutdown error: %v", err)
	}

	a.db.Close()
	a.logger.Info("Database connections closed")

	a.wg.Wait()
	a.logger.Info("Graceful shutdown completed")

	return nil
}
func (a *Application) initConfig() error {
	cfg, err := config.ParseConfig()
	if err != nil {
		return err
	}

	a.cfg = cfg
	return nil
}
func (a *Application) initLogger() error {
	log, err := zap.NewProduction()
	if err != nil {
		return err
	}

	a.logger = log.Sugar()

	return nil
}
func (a *Application) initDatabase(ctx context.Context) error {
	a.db = pgprovider.NewProvider(
		a.logger,
		a.cfg.SQLDataBase.Server,
		a.cfg.SQLDataBase.Database,
		a.cfg.SQLDataBase.Username,
		a.cfg.SQLDataBase.Password,
		a.cfg.SQLDataBase.Port,
		a.cfg.SQLDataBase.MaxIdleConns,
		a.cfg.SQLDataBase.MaxOpenConns,
		a.cfg.SQLDataBase.ConnMaxLifetimeMin,
	)

	if err := a.db.Open(ctx); err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	a.logger.Info("Database connection established")

	return nil
}
func (a *Application) initRepository() error {
	a.repo = repo.NewRepository(
		a.db.DB(),
		a.logger,
		a.cfg.SQLDataBase.QueryTimeoutSec,
	)

	return nil
}
func (a *Application) initService() error {
	a.service = service.NewService(
		a.repo,
		a.logger,
	)

	return nil
}
func (a *Application) initRouter() error {
	h := handlers.NewHandler(
		a.service,
		a.logger,
	)

	a.router = api.NewRouter(a.cfg, h)
	return nil
}
func (a *Application) startHTTPServer() {
	a.wg.Add(1)

	go func() {
		defer a.wg.Done()

		a.logger.Infof("Starting HTTP server on %s", a.cfg.ListenAddr)
		if err := a.router.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.logger.Infof("HTTP server error: %w", err)
		}
	}()
}
