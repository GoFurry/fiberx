package db

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	env "github.com/gofurry/fiberx/v3/extra-light/config"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var (
	engine  *gorm.DB
	initErr error
	once    sync.Once
)

func InitDatabaseOnStart(models ...any) error {
	db := DB()
	if initErr != nil {
		return initErr
	}
	if db == nil {
		return fmt.Errorf("database engine is nil")
	}
	if len(models) == 0 {
		slog.Info("no database models registered, skip auto migrate")
		return nil
	}
	if err := db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("auto migrate database failed: %w", err)
	}
	slog.Info("database service initialized", "driver", "sqlite")
	return nil
}

func DB() *gorm.DB {
	once.Do(func() {
		cfg := env.GetServerConfig()
		dsn, err := buildSQLiteDSN(cfg.Database.Path)
		if err != nil {
			initErr = err
			return
		}

		gdb, err := gorm.Open(sqlite.Open(dsn))
		if err != nil {
			initErr = fmt.Errorf("open sqlite database failed: %w", err)
			return
		}

		sqlDB, err := gdb.DB()
		if err != nil {
			initErr = fmt.Errorf("get sql db instance failed: %w", err)
			return
		}

		sqlDB.SetMaxIdleConns(1)
		sqlDB.SetMaxOpenConns(1)
		sqlDB.SetConnMaxLifetime(0)
		sqlDB.SetConnMaxIdleTime(0)

		if err = sqlDB.Ping(); err != nil {
			_ = sqlDB.Close()
			initErr = fmt.Errorf("ping sqlite database failed: %w", err)
			return
		}

		engine = gdb
		slog.Info("database connected", "driver", "sqlite")
	})

	return engine
}

func Ready() bool {
	if engine == nil {
		return false
	}

	sqlDB, err := engine.DB()
	if err != nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return sqlDB.PingContext(ctx) == nil
}

func Close() {
	if engine == nil {
		return
	}

	sqlDB, err := engine.DB()
	if err != nil {
		slog.Error("get sql db instance failed", "error", err)
		return
	}
	if err := sqlDB.Close(); err != nil {
		slog.Error("close database pool failed", "error", err)
		return
	}

	engine = nil
	initErr = nil
	once = sync.Once{}
	slog.Info("database pool closed")
}

func buildSQLiteDSN(path string) (string, error) {
	dsn := strings.TrimSpace(path)
	if dsn == "" {
		dsn = "./data/app.db"
	}

	if dsn == ":memory:" || strings.HasPrefix(dsn, "file:") {
		return dsn, nil
	}

	dir := filepath.Dir(dsn)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", fmt.Errorf("create sqlite directory failed: %w", err)
		}
	}

	return dsn, nil
}
