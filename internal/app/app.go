package app

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/jcmturner/gokrb5/v8/config"
	"go.uber.org/zap"
)

type Application struct {
	cfg      *config.Config
	logger   *zap.SugaredLogger
	metrics  *metrics.Metrics
	db       *pgprovider.Provider
	repo     *repository.Repository
	service  *service.Service
	router   *api.Router
	wg       sync.WaitGroup
	ready    bool
	readyMux sync.RWMutex
}

func NewApplication() *Application {
	return &Application{
		ready: false,
	}
}
func (a *Application) Start(ctx context.Context) error {

	return nil
}

func (a *Application) Wait(ctx context.Context, cancel context.CancelFunc) error {
	defer cancel()

	select {
	case <-ctx.Done():
		a.logger.Info("Shutdown signal received, starting graceful shutdown...")
	case err := <-a.errChan:
		a.logger.Errorf("Error received, initiating shutdown: %v", err)
		return err
	}

	a.setReady(false)

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
func (a *Application) Wait(ctx context.Context, cancel context.CancelFunc) error {}
func (a *Application) IsReady() bool                                             {}
func (a *Application) initConfig() error                                         {}
func (a *Application) initLogger() error                                         {}
func (a *Application) initMetrics() error                                        {}
func (a *Application) initDatabase(ctx context.Context) error                    {}
func (a *Application) initRepository() error                                     {}
func (a *Application) initService() error                                        {}
func (a *Application) initRouter() error                                         {}
func (a *Application) startHTTPServer()                                          {}
func (a *Application) startUtilityServer()                                       {}
func (a *Application) healthHandler(writer http.ResponseWriter, _ *http.Request) {}
func (a *Application) readyHandler(writer http.ResponseWriter, _ *http.Request)  {}
