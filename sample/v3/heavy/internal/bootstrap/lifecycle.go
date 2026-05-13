package bootstrap

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"

	env "github.com/gofurry/fiberx/v3/heavy/config"
	usermodels "github.com/gofurry/fiberx/v3/heavy/internal/app/user/models"
	cache "github.com/gofurry/fiberx/v3/heavy/internal/infra/cache"
	"github.com/gofurry/fiberx/v3/heavy/internal/infra/db"
	log "github.com/gofurry/fiberx/v3/heavy/internal/infra/logging"
	scheduler "github.com/gofurry/fiberx/v3/heavy/internal/infra/scheduler"
	"github.com/gofurry/fiberx/v3/heavy/internal/jobs/schedule"
	"github.com/gofurry/fiberx/v3/heavy/pkg/common"
	fibercoraza "github.com/gofiber/contrib/v3/coraza"
)

var (
	lifecycleMu sync.Mutex
	started     atomic.Bool
	wafEngine   *fibercoraza.Engine
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

	if cfg.Waf.Enabled {
		if err := initWAF(cfg); err != nil {
			return cleanupOnError(err)
		}
	}

	if cfg.DataBase.Enabled {
		if err := db.InitDatabaseOnStart(databaseModels()...); err != nil {
			return cleanupOnError(fmt.Errorf("database init failed: %w", err))
		}
	}

	if cfg.Redis.Enabled {
		if err := cache.InitRedisOnStart(); err != nil {
			return cleanupOnError(fmt.Errorf("redis init failed: %w", err))
		}
	}

	if cfg.Schedule.Enabled {
		if err := scheduler.InitTimeWheelOnStart(); err != nil {
			return cleanupOnError(fmt.Errorf("scheduler init failed: %w", err))
		}
		if err := registerScheduledJobs(schedule.Jobs()); err != nil {
			return cleanupOnError(fmt.Errorf("schedule registration failed: %w", err))
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

func initWAF(cfg *env.ServerConfigHolder) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("waf init panic: %v", recovered)
		}
	}()

	corazaCfg := fibercoraza.ConfigDefault
	corazaCfg.DirectivesFile = append([]string(nil), cfg.Waf.ConfPath...)
	corazaCfg.BlockMessage = "Request blocked by CorazaWAF"

	engine, initErr := fibercoraza.NewEngine(corazaCfg)
	if initErr != nil {
		wafEngine = nil
		return fmt.Errorf("waf init failed: %w", initErr)
	}

	wafEngine = engine
	return nil
}

func WAFEngine() *fibercoraza.Engine {
	return wafEngine
}

func shutdownComponents(cfg *env.ServerConfigHolder) error {
	var shutdownErr error

	if cfg.Waf.Enabled {
		wafEngine = nil
	}

	if cfg.Schedule.Enabled {
		scheduler.Stop()
	}

	if cfg.Redis.Enabled {
		if err := cache.Close(); err != nil {
			shutdownErr = errors.Join(shutdownErr, fmt.Errorf("redis shutdown failed: %w", err))
		}
	}

	if cfg.DataBase.Enabled {
		db.Orm.Close()
	}

	if err := log.Sync(); err != nil {
		shutdownErr = errors.Join(shutdownErr, fmt.Errorf("logger sync failed: %w", err))
	}

	return shutdownErr
}

func registerScheduledJobs(jobs []schedule.Job) error {
	for _, job := range jobs {
		if job.Run == nil {
			continue
		}
		if job.Interval <= 0 {
			return fmt.Errorf("scheduled job %q interval must be greater than 0", job.Name)
		}

		if job.RunOnStart {
			go job.Run()
		}
		scheduler.AddCronJob(job.Interval, job.Run)
		slog.Info("scheduled job registered", "name", job.Name, "interval", job.Interval.String())
	}
	return nil
}
