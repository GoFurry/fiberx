package bootstrap

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"

	env "github.com/gofurry/fiberx/v3/extra-light/config"
	"github.com/gofurry/fiberx/v3/extra-light/internal/db"
)

var (
	lifecycleMu sync.Mutex
	started     atomic.Bool
)

func Start() error {
	lifecycleMu.Lock()
	defer lifecycleMu.Unlock()

	if started.Load() {
		return nil
	}

	cfg := env.GetServerConfig()
	if err := initLogger(cfg); err != nil {
		return fmt.Errorf("logger init failed: %w", err)
	}

	if err := db.InitDatabaseOnStart(); err != nil {
		_ = closeLogger()
		return fmt.Errorf("database init failed: %w", err)
	}

	started.Store(true)
	slog.Info("application bootstrap completed")
	return nil
}

func Shutdown() error {
	lifecycleMu.Lock()
	defer lifecycleMu.Unlock()

	if !started.Load() {
		return nil
	}

	var shutdownErr error
	db.Close()
	if err := closeLogger(); err != nil {
		shutdownErr = errors.Join(shutdownErr, fmt.Errorf("logger close failed: %w", err))
	}

	started.Store(false)
	return shutdownErr
}
