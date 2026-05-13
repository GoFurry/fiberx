package env

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gofurry/fiberx/v3/extra-light/pkg/common"
	"gopkg.in/yaml.v2"
)

var (
	configuration *serverConfig
	configErr     error
	configOnce    sync.Once
	configPath    string
	configMu      sync.Mutex
)

type serverConfig struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Log      LogConfig      `yaml:"log"`
}

type ServerConfigHolder = serverConfig

type ServerConfig struct {
	AppName   string `yaml:"app_name"`
	Mode      string `yaml:"mode"`
	IPAddress string `yaml:"ip_address"`
	Port      string `yaml:"port"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type LogConfig struct {
	LogLevel string `yaml:"log_level"`
	LogPath  string `yaml:"log_path"`
}

func InitServerConfig(filePath string) error {
	configMu.Lock()
	if configPath == "" {
		configPath = strings.TrimSpace(filePath)
	}
	configMu.Unlock()

	ensureServerConfig()
	return configErr
}

func MustInitServerConfig(filePath string) error {
	return InitServerConfig(filePath)
}

func ensureServerConfig() {
	configOnce.Do(func() {
		cfg := defaultConfig()
		path := resolveConfigPath(currentConfigPath())

		content, err := os.ReadFile(path)
		if err != nil {
			configErr = fmt.Errorf("read config file failed: %w", err)
			return
		}
		if err := yaml.Unmarshal(content, cfg); err != nil {
			configErr = fmt.Errorf("parse config file failed: %w", err)
			return
		}
		cfg.normalize()
		if err := cfg.validate(); err != nil {
			configErr = err
			return
		}
		configuration = cfg
	})
}

func defaultConfig() *serverConfig {
	return &serverConfig{
		Server: ServerConfig{
			AppName:   common.COMMON_PROJECT_NAME,
			Mode:      "debug",
			IPAddress: "127.0.0.1",
			Port:      "9999",
		},
		Database: DatabaseConfig{
			Path: "./data/app.db",
		},
		Log: LogConfig{
			LogLevel: "debug",
			LogPath:  "./logs/app.log",
		},
	}
}

func (cfg *serverConfig) normalize() {
	if strings.TrimSpace(cfg.Server.AppName) == "" {
		cfg.Server.AppName = common.COMMON_PROJECT_NAME
	}
	if strings.TrimSpace(cfg.Server.Mode) == "" {
		cfg.Server.Mode = "debug"
	}
	if strings.TrimSpace(cfg.Server.IPAddress) == "" {
		cfg.Server.IPAddress = "127.0.0.1"
	}
	if strings.TrimSpace(cfg.Server.Port) == "" {
		cfg.Server.Port = "9999"
	}
	if strings.TrimSpace(cfg.Database.Path) == "" {
		cfg.Database.Path = "./data/app.db"
	}
	if strings.TrimSpace(cfg.Log.LogLevel) == "" {
		cfg.Log.LogLevel = "debug"
	}
}

func (cfg *serverConfig) validate() error {
	mode := strings.ToLower(strings.TrimSpace(cfg.Server.Mode))
	switch mode {
	case "debug", "release", "prod":
	default:
		return fmt.Errorf("server.mode must be one of debug, release, prod")
	}
	if strings.TrimSpace(cfg.Server.Port) == "" {
		return fmt.Errorf("server.port is required")
	}
	if strings.TrimSpace(cfg.Database.Path) == "" {
		return fmt.Errorf("database.path is required")
	}
	return nil
}

func resolveConfigPath(filePath string) string {
	if filePath != "" {
		return filePath
	}
	pwd, err := os.Getwd()
	if err != nil {
		return filepath.Join("config", "server.yaml")
	}
	return filepath.Join(pwd, "config", "server.yaml")
}

func currentConfigPath() string {
	configMu.Lock()
	defer configMu.Unlock()
	return configPath
}

func GetServerConfig() *serverConfig {
	ensureServerConfig()
	if configuration != nil {
		return configuration
	}
	return defaultConfig()
}
