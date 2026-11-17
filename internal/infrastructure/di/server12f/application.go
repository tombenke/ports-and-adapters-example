package server12f

import (
	"context"
	"log/slog"
	"sync"

	"github.com/tombenke/go-12f-common/v2/apprun"
	"github.com/tombenke/go-12f-common/v2/buildinfo"
	"github.com/tombenke/go-12f-common/v2/log"
)

var _ apprun.Application = (*Application)(nil)

// Application represents the overall application object.
type Application struct {
	config *Config
	err    error

	// The internal components of the application.
	components []apprun.ComponentLifecycleManager
}

// NewApplication creates a new application instance.
func NewApplication(config *Config) (apprun.Application, error) {
	if config == nil {
		config = &Config{}
	}

	components := []apprun.ComponentLifecycleManager{
		//	webComponent,
	}

	// Create and return the application object
	return &Application{
		config: config,
		//web:               webComponent,
		err: nil,

		components: components,
	}, nil
}

func (a *Application) Components(_ context.Context) []apprun.ComponentLifecycleManager {
	return a.components
}

// Startup the application.
func (a *Application) AfterStartup(ctx context.Context, _ *sync.WaitGroup) error {
	a.getLogger(ctx).Info("AfterStartup")
	a.getLogger(ctx).Debug("BuildInfo", "AppName", buildinfo.AppName(), "Version", buildinfo.Version())

	return nil
}

// BeforeShutdown shuts down the application.
func (a *Application) BeforeShutdown(ctx context.Context) error {
	a.getLogger(ctx).Info("BeforeShutdown")
	return nil
}

func (*Application) getLogger(ctx context.Context) *slog.Logger {
	return log.GetFromContextOrDefault(ctx).With("app", "Application")
}
