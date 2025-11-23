package application

import (
	"context"
	"fmt"
	"sync"
	"ttanalytic/internal/config"
)

type Application struct {
	cfg *config.Config

	errChan chan error
	wg      sync.WaitGroup

	//flag for ready
	ready    bool
	readyMux sync.RWMutex
}

func NewApplication() *Application {
	return &Application{
		errChan: make(chan error),
		ready:   false,
	}
}
func (a *Application) Start(ctx context.Context) error {
	if err := a.initConfig(); err != nil {
		return fmt.Errorf("init config: %w", err)
	}

	a.setReady(true)
	//a.logger.Info("Application started successfully")

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
func (a *Application) setReady(ready bool) {
	a.readyMux.Lock()
	defer a.readyMux.Unlock()

	a.ready = ready
}
func (a *Application) IsReady() bool {
	a.readyMux.RLock()
	defer a.readyMux.RUnlock()

	return a.ready
}
