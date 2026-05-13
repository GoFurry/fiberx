package bootstrap

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"

	env "github.com/gofurry/fiberx/v3/light/config"
	usermodels "github.com/gofurry/fiberx/v3/light/internal/app/user/models"
	"github.com/gofurry/fiberx/v3/light/internal/infra/db"
	log "github.com/gofurry/fiberx/v3/light/internal/infra/logging"
	"github.com/gofurry/fiberx/v3/light/pkg/common"
)

var (
	lifecycleMu sync.Mutex
	started     atomic.Bool
)

func databaseModels() []any {
	return []any{
		&usermodels.User{},
	}
}

func Start() error {
	lifecycleMu.Lock()
	defer lifecycleMu.Unlock()

	if started.Load() {
		return nil
	}

	cfg := env.GetServerConfig()
	if err := initLogger(cfg); err != nil {
		return err
	}

	cleanupOnError := func(cause error) error {
		started.Store(false)
		return errors.Join(cause, shutdownComponents(cfg))
	}

	if cfg.DataBase.Enabled {
		if err := db.InitDatabaseOnStart(databaseModels()...); err != nil {
			return cleanupOnError(fmt.Errorf("database init failed: %w", err))
		}
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

	cfg := env.GetServerConfig()
	err := shutdownComponents(cfg)
	started.Store(false)
	return err
}

func initLogger(cfg *env.ServerConfigHolder) error {
	logCfg := &log.Config{
		ShowLine:   true,
		TimeFormat: common.TIME_FORMAT_DATE,
	}

	if cfg.Server.Mode == "debug" {
		logCfg.Level = "debug"
		logCfg.Mode = "dev"
		logCfg.EncodeJson = false
	} else {
		logCfg.Level = cfg.Log.LogLevel
		logCfg.Mode = cfg.Log.LogMode
		logCfg.FilePath = cfg.Log.LogPath
		logCfg.MaxSize = cfg.Log.LogMaxSize
		logCfg.MaxBackups = cfg.Log.LogMaxBackups
		logCfg.MaxAge = cfg.Log.LogMaxAge
		logCfg.Compress = true
		logCfg.EncodeJson = true
		logCfg.TimeFormat = common.TIME_FORMAT_LOG
	}

	if err := log.InitLogger(logCfg); err != nil {
		return fmt.Errorf("logger init failed: %w", err)
	}
	return nil
}

func shutdownComponents(cfg *env.ServerConfigHolder) error {
	var shutdownErr error

	if cfg.DataBase.Enabled {
		db.Orm.Close()
	}
	if err := log.Sync(); err != nil {
		shutdownErr = errors.Join(shutdownErr, fmt.Errorf("logger sync failed: %w", err))
	}

	return shutdownErr
}
