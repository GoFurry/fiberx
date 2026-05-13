package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	env "github.com/gofurry/fiberx/v3/light/config"
	"github.com/gofurry/fiberx/v3/light/internal/bootstrap"
	"github.com/gofurry/fiberx/v3/light/internal/transport/http/router"
	"github.com/gofurry/fiberx/v3/light/pkg/common"
	"github.com/gofiber/fiber/v3"
)

func runService() error {
	cfg := env.GetServerConfig()

	debug.SetGCPercent(cfg.Server.GCPercent)
	debug.SetMemoryLimit(int64(cfg.Server.MemoryLimit << 30))

	if err := bootstrap.Start(); err != nil {
		return err
	}

	app := newApp()
	app.fiberApp = router.New().Init()
	return app.run()
}

func appIdentity() (string, string) {
	cfg := env.GetServerConfig()
	appID := cfg.Server.AppID
	if appID == "" {
		appID = common.COMMON_PROJECT_NAME
	}

	appName := cfg.Server.AppName
	if appName == "" {
		appName = appID
	}
	return appID, appName
}

type app struct {
	fiberApp     *fiber.App
	shutdownOnce sync.Once
	stopping     atomic.Bool
}

func newApp() *app {
	return &app{}
}

func (a *app) run() error {
	cfg := env.GetServerConfig()
	addr := cfg.Server.IPAddress + ":" + cfg.Server.Port

	defer func() {
		if err := a.shutdown(); err != nil {
			slog.Error("application shutdown failed", "error", err)
		}
	}()

	if err := a.fiberApp.Listen(addr, fiber.ListenConfig{
		TLSConfig:         nil,
		EnablePrefork:     cfg.Server.EnablePrefork,
		ListenerNetwork:   cfg.Server.Network,
		EnablePrintRoutes: cfg.Server.Mode == "debug",
	}); err != nil {
		if a.stopping.Load() {
			return nil
		}
		return fmt.Errorf("fiber app exited unexpectedly: %w", err)
	}
	return nil
}

func (a *app) shutdown() error {
	var shutdownErr error

	a.shutdownOnce.Do(func() {
		a.stopping.Store(true)

		if a.fiberApp != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := a.fiberApp.ShutdownWithContext(ctx); err != nil {
				shutdownErr = errors.Join(shutdownErr, fmt.Errorf("shutdown fiber app failed: %w", err))
			}
		}

		if err := bootstrap.Shutdown(); err != nil {
			shutdownErr = errors.Join(shutdownErr, err)
		}
	})

	return shutdownErr
}
