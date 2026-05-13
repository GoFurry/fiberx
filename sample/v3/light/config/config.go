package env

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofurry/fiberx/v3/light/pkg/common"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

var (
	configuration *serverConfig
	configErr     error
	configOnce    sync.Once
	configOptions = configLoaderOptions{
		projectName: common.COMMON_PROJECT_NAME,
		fileName:    "server.yaml",
	}
	configOptionsMu sync.Mutex
)

type configLoaderOptions struct {
	projectName string
	fileName    string
	configFile  string
}

type serverConfig struct {
	ClusterId  int              `yaml:"cluster_id"`
	Server     ServerConfig     `yaml:"server"`
	DataBase   DataBaseConfig   `yaml:"database"`
	Log        LogConfig        `yaml:"log"`
	Middleware MiddlewareConfig `yaml:"middleware"`
}

type ServerConfigHolder = serverConfig

type MiddlewareConfig struct {
	Cors        CorsConfig        `yaml:"cors"`
	RequestID   RequestIDConfig   `yaml:"request_id"`
	AccessLog   AccessLogConfig   `yaml:"access_log"`
	Timeout     TimeoutConfig     `yaml:"timeout"`
	Health      HealthConfig      `yaml:"health"`
	Compression CompressionConfig `yaml:"compression"`
	Limiter     LimiterConfig     `yaml:"limiter"`
	ETag        ETagConfig        `yaml:"etag"`
}

type RequestIDConfig struct {
	Enabled bool   `yaml:"enabled"`
	Header  string `yaml:"header"`
}

type AccessLogConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Format     string `yaml:"format"`
	TimeFormat string `yaml:"time_format"`
	TimeZone   string `yaml:"time_zone"`
}

type TimeoutConfig struct {
	Enabled         bool     `yaml:"enabled"`
	DurationSeconds int      `yaml:"duration_seconds"`
	ExcludePaths    []string `yaml:"exclude_paths"`
}

type HealthConfig struct {
	Enabled       bool `yaml:"enabled"`
	IncludeLegacy bool `yaml:"include_legacy"`
}

type CompressionConfig struct {
	Enabled bool   `yaml:"enabled"`
	Level   string `yaml:"level"`
}

type LimiterConfig struct {
	Enabled                bool          `yaml:"enabled"`
	MaxRequests            int           `yaml:"max_requests"`
	Expiration             time.Duration `yaml:"expiration"`
	Strategy               string        `yaml:"strategy"`
	KeySource              string        `yaml:"key_source"`
	KeyHeader              string        `yaml:"key_header"`
	SkipFailedRequests     bool          `yaml:"skip_failed_requests"`
	SkipSuccessfulRequests bool          `yaml:"skip_successful_requests"`
	DisableHeaders         bool          `yaml:"disable_headers"`
	ExcludePaths           []string      `yaml:"exclude_paths"`
}

type ETagConfig struct {
	Enabled bool `yaml:"enabled"`
	Weak    bool `yaml:"weak"`
}

type CorsConfig struct {
	AllowOrigins []string `yaml:"allow_origins"`
}

type LogConfig struct {
	LogLevel      string `yaml:"log_level"`
	LogMode       string `yaml:"log_mode"`
	LogPath       string `yaml:"log_path"`
	LogMaxSize    int    `yaml:"log_max_size"`
	LogMaxBackups int    `yaml:"log_max_backups"`
	LogMaxAge     int    `yaml:"log_max_age"`
}

type DataBaseConfig struct {
	Enabled     bool                 `yaml:"enabled"`
	AutoMigrate bool                 `yaml:"auto_migrate"`
	DBType      string               `yaml:"db_type"`
	SQLite      SQLiteDataBaseConfig `yaml:"sqlite"`
	DSN         string               `yaml:"dsn"`
	DBName      string               `yaml:"db_name"`
	SQLPath     string               `yaml:"sqlite_path"`
}

type SQLiteDataBaseConfig struct {
	DSN  string `yaml:"dsn"`
	Path string `yaml:"path"`
}

type ServerConfig struct {
	AppID         string `yaml:"app_id"`
	AppName       string `yaml:"app_name"`
	AppVersion    string `yaml:"app_version"`
	Mode          string `yaml:"mode"`
	IPAddress     string `yaml:"ip_address"`
	Port          string `yaml:"port"`
	MemoryLimit   int    `yaml:"memory_limit"`
	GCPercent     int    `yaml:"gc_percent"`
	Network       string `yaml:"network"`
	EnablePrefork bool   `yaml:"enable_prefork"`
	IsFullStack   bool   `yaml:"is_full_stack"`
}

type configKey struct {
	name string
	kind string
}

var knownConfigKeys = []configKey{
	{name: "cluster_id", kind: "int"},
	{name: "server.app_id", kind: "string"},
	{name: "server.app_name", kind: "string"},
	{name: "server.app_version", kind: "string"},
	{name: "server.mode", kind: "string"},
	{name: "server.ip_address", kind: "string"},
	{name: "server.port", kind: "string"},
	{name: "server.memory_limit", kind: "int"},
	{name: "server.gc_percent", kind: "int"},
	{name: "server.network", kind: "string"},
	{name: "server.enable_prefork", kind: "bool"},
	{name: "server.is_full_stack", kind: "bool"},
	{name: "database.enabled", kind: "bool"},
	{name: "database.auto_migrate", kind: "bool"},
	{name: "database.db_type", kind: "string"},
	{name: "database.sqlite.dsn", kind: "string"},
	{name: "database.sqlite.path", kind: "string"},
	{name: "database.dsn", kind: "string"},
	{name: "database.db_name", kind: "string"},
	{name: "database.sqlite_path", kind: "string"},
	{name: "log.log_level", kind: "string"},
	{name: "log.log_mode", kind: "string"},
	{name: "log.log_path", kind: "string"},
	{name: "log.log_max_size", kind: "int"},
	{name: "log.log_max_backups", kind: "int"},
	{name: "log.log_max_age", kind: "int"},
	{name: "middleware.cors.allow_origins", kind: "string_slice"},
	{name: "middleware.request_id.enabled", kind: "bool"},
	{name: "middleware.request_id.header", kind: "string"},
	{name: "middleware.access_log.enabled", kind: "bool"},
	{name: "middleware.access_log.format", kind: "string"},
	{name: "middleware.access_log.time_format", kind: "string"},
	{name: "middleware.access_log.time_zone", kind: "string"},
	{name: "middleware.timeout.enabled", kind: "bool"},
	{name: "middleware.timeout.duration_seconds", kind: "int"},
	{name: "middleware.timeout.exclude_paths", kind: "string_slice"},
	{name: "middleware.health.enabled", kind: "bool"},
	{name: "middleware.health.include_legacy", kind: "bool"},
	{name: "middleware.compression.enabled", kind: "bool"},
	{name: "middleware.compression.level", kind: "string"},
	{name: "middleware.limiter.enabled", kind: "bool"},
	{name: "middleware.limiter.max_requests", kind: "int"},
	{name: "middleware.limiter.expiration", kind: "int"},
	{name: "middleware.limiter.strategy", kind: "string"},
	{name: "middleware.limiter.key_source", kind: "string"},
	{name: "middleware.limiter.key_header", kind: "string"},
	{name: "middleware.limiter.skip_failed_requests", kind: "bool"},
	{name: "middleware.limiter.skip_successful_requests", kind: "bool"},
	{name: "middleware.limiter.disable_headers", kind: "bool"},
	{name: "middleware.limiter.exclude_paths", kind: "string_slice"},
	{name: "middleware.etag.enabled", kind: "bool"},
	{name: "middleware.etag.weak", kind: "bool"},
}

func ConfigureServerConfig(projectName, fileName, configFile string) {
	configOptionsMu.Lock()
	defer configOptionsMu.Unlock()

	if configuration != nil {
		return
	}

	if projectName = strings.TrimSpace(projectName); projectName != "" {
		configOptions.projectName = projectName
	}
	if fileName = strings.TrimSpace(fileName); fileName != "" {
		configOptions.fileName = fileName
	}
	configOptions.configFile = strings.TrimSpace(configFile)
}

func InitServerConfig(projectName string) error {
	opts := currentConfigOptions()
	ConfigureServerConfig(projectName, opts.fileName, opts.configFile)
	ensureServerConfig()
	return configErr
}

func MustInitServerConfig(projectName, configFile string) error {
	ConfigureServerConfig(projectName, "server.yaml", configFile)
	ensureServerConfig()
	return configErr
}

func (cfg *serverConfig) normalize() {
	if cfg.ClusterId == 0 {
		cfg.ClusterId = 1
	}
	if cfg.Server.AppID == "" {
		cfg.Server.AppID = common.COMMON_PROJECT_NAME
	}
	if cfg.Server.AppName == "" {
		cfg.Server.AppName = cfg.Server.AppID
	}
	if cfg.Server.AppVersion == "" {
		cfg.Server.AppVersion = "v1.0.0"
	}
	if cfg.Server.Mode == "" {
		cfg.Server.Mode = "debug"
	}
	if cfg.Server.IPAddress == "" {
		cfg.Server.IPAddress = "127.0.0.1"
	}
	if cfg.Server.Port == "" {
		cfg.Server.Port = "9999"
	}
	if cfg.Server.Network == "" {
		cfg.Server.Network = "tcp"
	}

	cfg.DataBase.normalize()

	if cfg.Middleware.RequestID.Header == "" {
		cfg.Middleware.RequestID.Header = "X-Request-ID"
	}
	if cfg.Middleware.AccessLog.Format == "" {
		cfg.Middleware.AccessLog.Format = "${time} | ${status} | ${latency} | ${method} | ${path} | rid=${respHeader:X-Request-ID}"
	}
	if cfg.Middleware.AccessLog.TimeFormat == "" {
		cfg.Middleware.AccessLog.TimeFormat = common.TIME_FORMAT_LOG
	}
	if cfg.Middleware.AccessLog.TimeZone == "" {
		cfg.Middleware.AccessLog.TimeZone = "Local"
	}
	if cfg.Middleware.Timeout.DurationSeconds <= 0 {
		cfg.Middleware.Timeout.DurationSeconds = 15
	}
	cfg.Middleware.Timeout.ExcludePaths = normalizeStringList(cfg.Middleware.Timeout.ExcludePaths)
	if cfg.Middleware.Compression.Level == "" {
		cfg.Middleware.Compression.Level = "default"
	}
	cfg.Middleware.Compression.Level = strings.ToLower(strings.TrimSpace(cfg.Middleware.Compression.Level))
	if cfg.Middleware.Limiter.Strategy == "" {
		cfg.Middleware.Limiter.Strategy = "fixed"
	}
	cfg.Middleware.Limiter.Strategy = strings.ToLower(strings.TrimSpace(cfg.Middleware.Limiter.Strategy))
	if cfg.Middleware.Limiter.KeySource == "" {
		cfg.Middleware.Limiter.KeySource = "ip"
	}
	cfg.Middleware.Limiter.KeySource = strings.ToLower(strings.TrimSpace(cfg.Middleware.Limiter.KeySource))
	cfg.Middleware.Limiter.ExcludePaths = normalizeStringList(cfg.Middleware.Limiter.ExcludePaths)
}

func (cfg *serverConfig) validate() error {
	var errs []error

	switch cfg.Server.Mode {
	case "debug", "release", "prod":
	default:
		errs = append(errs, fmt.Errorf("server.mode must be one of debug, release, prod"))
	}

	if port, err := strconv.Atoi(cfg.Server.Port); err != nil || port <= 0 || port > 65535 {
		errs = append(errs, fmt.Errorf("server.port must be a valid port"))
	}
	if cfg.Server.MemoryLimit < 0 {
		errs = append(errs, fmt.Errorf("server.memory_limit must be >= 0"))
	}
	if cfg.Server.GCPercent < 0 {
		errs = append(errs, fmt.Errorf("server.gc_percent must be >= 0"))
	}

	if cfg.Middleware.Limiter.Enabled {
		if cfg.Middleware.Limiter.MaxRequests <= 0 {
			errs = append(errs, fmt.Errorf("middleware.limiter.max_requests must be > 0 when limiter is enabled"))
		}
		if cfg.Middleware.Limiter.Expiration <= 0 {
			errs = append(errs, fmt.Errorf("middleware.limiter.expiration must be > 0 when limiter is enabled"))
		}
		switch cfg.Middleware.Limiter.Strategy {
		case "fixed", "sliding":
		default:
			errs = append(errs, fmt.Errorf("middleware.limiter.strategy must be one of fixed, sliding"))
		}
		switch cfg.Middleware.Limiter.KeySource {
		case "ip", "path", "ip_path", "header":
		default:
			errs = append(errs, fmt.Errorf("middleware.limiter.key_source must be one of ip, path, ip_path, header"))
		}
		if cfg.Middleware.Limiter.KeySource == "header" && strings.TrimSpace(cfg.Middleware.Limiter.KeyHeader) == "" {
			errs = append(errs, fmt.Errorf("middleware.limiter.key_header is required when key_source is header"))
		}
	}

	if cfg.Middleware.Timeout.Enabled && cfg.Middleware.Timeout.DurationSeconds <= 0 {
		errs = append(errs, fmt.Errorf("middleware.timeout.duration_seconds must be > 0 when timeout is enabled"))
	}

	switch cfg.Middleware.Compression.Level {
	case "", "default", "best_speed", "best_compression":
	default:
		errs = append(errs, fmt.Errorf("middleware.compression.level must be one of default, best_speed, best_compression"))
	}

	switch cfg.DataBase.DBType {
	case "", "sqlite":
	default:
		errs = append(errs, fmt.Errorf("database.db_type %q is not supported in light", cfg.DataBase.DBType))
	}

	if cfg.DataBase.Enabled && strings.TrimSpace(cfg.DataBase.SQLite.DSN) == "" && strings.TrimSpace(cfg.DataBase.SQLite.Path) == "" {
		errs = append(errs, fmt.Errorf("database.sqlite.path or database.sqlite.dsn is required when sqlite is enabled"))
	}

	return errors.Join(errs...)
}

func (cfg *DataBaseConfig) normalize() {
	cfg.DBType = strings.ToLower(strings.TrimSpace(cfg.DBType))
	if cfg.DBType == "" {
		cfg.DBType = "sqlite"
	}
	if cfg.SQLite.DSN == "" {
		cfg.SQLite.DSN = strings.TrimSpace(cfg.DSN)
	}
	if cfg.SQLite.Path == "" {
		cfg.SQLite.Path = strings.TrimSpace(cfg.SQLPath)
	}
	if cfg.SQLite.Path == "" {
		cfg.SQLite.Path = strings.TrimSpace(cfg.DBName)
	}
	if cfg.SQLite.Path == "" {
		cfg.SQLite.Path = "./data/app.db"
	}
}

func normalizeStringList(items []string) []string {
	if len(items) == 0 {
		return nil
	}

	result := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}

func InitConfig(projectName, fileName, configFile string, conf interface{}) error {
	v := viper.New()
	configFile = strings.TrimSpace(configFile)

	if configFile != "" {
		v.SetConfigFile(configFile)
		if ext := strings.TrimPrefix(filepath.Ext(configFile), "."); ext != "" {
			v.SetConfigType(ext)
		}
	} else {
		configName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
		configType := strings.TrimPrefix(filepath.Ext(fileName), ".")
		if configName == "" {
			configName = fileName
		}
		if configType == "" {
			configType = "yaml"
		}

		v.SetConfigName(configName)
		v.SetConfigType(configType)
		v.AddConfigPath(filepath.Join("/etc", projectName))

		pwd, err := os.Getwd()
		if err != nil {
			fmt.Println("Error loading pwd dir:", err.Error())
		} else {
			v.AddConfigPath(filepath.Join(pwd, "config"))
		}
	}

	applyDefaults(v)
	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	bindKnownEnvKeys(v)

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("can not find any %s file: %w", fileName, err)
	}

	fmt.Println("load config:" + v.ConfigFileUsed())

	settings := collectSettings(v)
	raw, err := yaml.Marshal(settings)
	if err != nil {
		return fmt.Errorf("marshal merged config failed: %w", err)
	}
	if err := yaml.Unmarshal(raw, conf); err != nil {
		return fmt.Errorf("unmarshal merged config failed: %w", err)
	}

	return nil
}

func ensureServerConfig() {
	configOnce.Do(func() {
		opts := currentConfigOptions()
		cfg := new(serverConfig)
		if err := InitConfig(opts.projectName, opts.fileName, opts.configFile, cfg); err != nil {
			configErr = err
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

func applyDefaults(v *viper.Viper) {
	v.SetDefault("cluster_id", 1)
	v.SetDefault("server.app_id", common.COMMON_PROJECT_NAME)
	v.SetDefault("server.app_name", common.COMMON_PROJECT_NAME)
	v.SetDefault("server.app_version", "v1.0.0")
	v.SetDefault("server.mode", "debug")
	v.SetDefault("server.ip_address", "127.0.0.1")
	v.SetDefault("server.port", "9999")
	v.SetDefault("server.memory_limit", 1)
	v.SetDefault("server.gc_percent", 1000)
	v.SetDefault("server.network", "tcp")
	v.SetDefault("server.enable_prefork", false)
	v.SetDefault("server.is_full_stack", false)
	v.SetDefault("database.enabled", true)
	v.SetDefault("database.db_type", "sqlite")
	v.SetDefault("database.auto_migrate", true)
	v.SetDefault("database.sqlite.path", "./data/app.db")
	v.SetDefault("middleware.request_id.enabled", true)
	v.SetDefault("middleware.request_id.header", "X-Request-ID")
	v.SetDefault("middleware.access_log.enabled", true)
	v.SetDefault("middleware.access_log.format", "${time} | ${status} | ${latency} | ${method} | ${path} | rid=${respHeader:X-Request-ID}")
	v.SetDefault("middleware.access_log.time_format", common.TIME_FORMAT_LOG)
	v.SetDefault("middleware.access_log.time_zone", "Local")
	v.SetDefault("middleware.timeout.enabled", true)
	v.SetDefault("middleware.timeout.duration_seconds", 15)
	v.SetDefault("middleware.timeout.exclude_paths", []string{"/livez", "/readyz", "/startupz", "/healthz"})
	v.SetDefault("middleware.health.enabled", true)
	v.SetDefault("middleware.health.include_legacy", true)
	v.SetDefault("middleware.compression.enabled", true)
	v.SetDefault("middleware.compression.level", "default")
	v.SetDefault("middleware.limiter.enabled", true)
	v.SetDefault("middleware.limiter.max_requests", 3000)
	v.SetDefault("middleware.limiter.expiration", 60)
	v.SetDefault("middleware.limiter.strategy", "fixed")
	v.SetDefault("middleware.limiter.key_source", "ip")
	v.SetDefault("middleware.limiter.skip_failed_requests", false)
	v.SetDefault("middleware.limiter.skip_successful_requests", false)
	v.SetDefault("middleware.limiter.disable_headers", false)
	v.SetDefault("middleware.limiter.exclude_paths", []string{"/livez", "/readyz", "/startupz", "/healthz"})
	v.SetDefault("middleware.etag.enabled", true)
	v.SetDefault("middleware.etag.weak", false)
}

func bindKnownEnvKeys(v *viper.Viper) {
	for _, key := range knownConfigKeys {
		_ = v.BindEnv(key.name)
	}
}

func collectSettings(v *viper.Viper) map[string]interface{} {
	settings := make(map[string]interface{})
	for _, key := range knownConfigKeys {
		setNestedValue(settings, key.name, collectValue(v, key))
	}
	return settings
}

func collectValue(v *viper.Viper, key configKey) interface{} {
	switch key.kind {
	case "bool":
		return v.GetBool(key.name)
	case "int":
		return v.GetInt(key.name)
	case "string_slice":
		if raw, ok := os.LookupEnv(envVariableName(key.name)); ok {
			raw = strings.TrimSpace(raw)
			items := strings.Split(raw, ",")
			result := make([]string, 0, len(items))
			for _, item := range items {
				item = strings.TrimSpace(item)
				if item == "" {
					continue
				}
				result = append(result, item)
			}
			return result
		}
		return v.GetStringSlice(key.name)
	default:
		return v.GetString(key.name)
	}
}

func envVariableName(key string) string {
	replacer := strings.NewReplacer(".", "_", "-", "_")
	return "APP_" + strings.ToUpper(replacer.Replace(key))
}

func setNestedValue(target map[string]interface{}, key string, value interface{}) {
	parts := strings.Split(key, ".")
	current := target
	for index, part := range parts {
		if index == len(parts)-1 {
			current[part] = value
			return
		}

		next, ok := current[part].(map[string]interface{})
		if !ok {
			next = make(map[string]interface{})
			current[part] = next
		}
		current = next
	}
}

func GetServerConfig() *serverConfig {
	ensureServerConfig()
	if configuration != nil {
		return configuration
	}
	cfg := new(serverConfig)
	cfg.normalize()
	return cfg
}

func currentConfigOptions() configLoaderOptions {
	configOptionsMu.Lock()
	defer configOptionsMu.Unlock()
	return configOptions
}
