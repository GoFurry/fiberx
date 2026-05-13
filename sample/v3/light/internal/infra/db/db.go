package db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	env "github.com/gofurry/fiberx/v3/light/config"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var Orm = &orm{}
var once sync.Once

func initOrm() {
	Orm.initErr = Orm.loadDBConfig()
}

type orm struct {
	engine  *gorm.DB
	driver  string
	initErr error
}

func InitDatabaseOnStart(models ...any) error {
	cfg := env.GetServerConfig().DataBase
	if !cfg.Enabled {
		slog.Info("database service disabled by config, skip initialization")
		return nil
	}

	engine := Orm.DB()
	if err := Orm.Err(); err != nil {
		return fmt.Errorf("initialize database failed: %w", err)
	}
	if engine == nil {
		return fmt.Errorf("database engine is nil")
	}

	if !cfg.AutoMigrate {
		slog.Info("database auto migrate disabled by config")
		return nil
	}

	if err := Orm.AutoMigrate(models...); err != nil {
		return fmt.Errorf("auto migrate database failed: %w", err)
	}

	slog.Info("database service initialized", "driver", Orm.Driver())
	return nil
}

func (db *orm) loadDBConfig() error {
	if db.engine != nil {
		return nil
	}

	cfg := env.GetServerConfig().DataBase
	dsn, err := buildSQLiteDSN(cfg.SQLite)
	if err != nil {
		return fmt.Errorf("build sqlite dsn failed: %w", err)
	}

	engine, err := gorm.Open(sqlite.Open(dsn))
	if err != nil {
		return fmt.Errorf("open sqlite database failed: %w", err)
	}

	sqlDB, err := engine.DB()
	if err != nil {
		return fmt.Errorf("get sql db instance failed: %w", err)
	}

	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetConnMaxLifetime(0)
	sqlDB.SetConnMaxIdleTime(0)

	if err = sqlDB.Ping(); err != nil {
		_ = sqlDB.Close()
		return fmt.Errorf("ping sqlite database failed: %w", err)
	}

	db.engine = engine
	db.driver = "sqlite"
	slog.Info("database connected", "driver", db.driver)
	return nil
}

func (db *orm) DB() *gorm.DB {
	once.Do(initOrm)
	return db.engine
}

func (db *orm) Err() error {
	once.Do(initOrm)
	return db.initErr
}

func (db *orm) Ready() bool {
	if db.engine == nil {
		return false
	}

	sqlDB, err := db.engine.DB()
	if err != nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return sqlDB.PingContext(ctx) == nil
}

func (db *orm) Driver() string {
	return db.driver
}

func (db *orm) AutoMigrate(models ...any) error {
	engine := db.DB()
	if err := db.Err(); err != nil {
		return fmt.Errorf("initialize database failed: %w", err)
	}
	if engine == nil {
		return errors.New("database is not initialized")
	}
	if len(models) == 0 {
		slog.Info("no database models registered, skip auto migrate")
		return nil
	}
	return engine.AutoMigrate(models...)
}

func (db *orm) Close() {
	if db.engine == nil {
		return
	}

	sqlDB, err := db.engine.DB()
	if err != nil {
		slog.Error("get sql db instance failed", "error", err)
		return
	}
	if err = sqlDB.Close(); err != nil {
		slog.Error("close database pool failed", "error", err)
		return
	}

	db.engine = nil
	db.driver = ""
	db.initErr = nil
	once = sync.Once{}
	slog.Info("database pool closed")
}

func buildSQLiteDSN(cfg env.SQLiteDataBaseConfig) (string, error) {
	dsn := strings.TrimSpace(cfg.DSN)
	if dsn == "" {
		dsn = strings.TrimSpace(cfg.Path)
	}
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
