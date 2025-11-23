package application

import (
	"context"
	"fmt"
	"sync"
	"ttanalytic/internal/config"
)

type Application struct {
	cfg *config.Config

	wg sync.WaitGroup
}

func NewApplication() *Application {
	return &Application{}
}
func (a *Application) Start(ctx context.Context) error {
	if err := a.initConfig(); err != nil {
		return fmt.Errorf("init config: %w", err)
	}

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
