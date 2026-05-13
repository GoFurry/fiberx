package bootstrap

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	env "github.com/gofurry/fiberx/v3/extra-light/config"
)

var logFile *os.File

func initLogger(cfg *env.ServerConfigHolder) error {
	writers := make([]io.Writer, 0, 2)
	writers = append(writers, os.Stdout)

	path := strings.TrimSpace(cfg.Log.LogPath)
	if path != "" {
		dir := filepath.Dir(path)
		if dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return err
			}
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		logFile = file
		writers = append(writers, file)
	}

	level := parseLogLevel(cfg.Log.LogLevel)
	logger := slog.New(slog.NewTextHandler(io.MultiWriter(writers...), &slog.HandlerOptions{
		Level: level,
	}))
	slog.SetDefault(logger)
	return nil
}

func closeLogger() error {
	if logFile == nil {
		return nil
	}
	err := logFile.Close()
	logFile = nil
	return err
}

func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
