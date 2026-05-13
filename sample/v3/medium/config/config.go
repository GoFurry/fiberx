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

	"github.com/gofurry/fiberx/v3/medium/pkg/common"
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
	Redis      RedisConfig      `yaml:"redis"`
	Middleware MiddlewareConfig `yaml:"middleware"`
	Waf        WafConfig        `yaml:"waf"`
}

type ServerConfigHolder = serverConfig

type WafConfig struct {
	Enabled  bool     `yaml:"enabled"`
	ConfPath []string `yaml:"conf_path"`
}

type MiddlewareConfig struct {
	Swagger         SwaggerConfig         `yaml:"swagger"`
	Cors            CorsConfig            `yaml:"cors"`
	RequestID       RequestIDConfig       `yaml:"request_id"`
	AccessLog       AccessLogConfig       `yaml:"access_log"`
	Timeout         TimeoutConfig         `yaml:"timeout"`
	Health          HealthConfig          `yaml:"health"`
	SecurityHeaders SecurityHeadersConfig `yaml:"security_headers"`
	Compression     CompressionConfig     `yaml:"compression"`
	Limiter         LimiterConfig         `yaml:"limiter"`
	CSRF            CSRFConfig            `yaml:"csrf"`
	ETag            ETagConfig            `yaml:"etag"`
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

type SecurityHeadersConfig struct {
	Enabled               bool   `yaml:"enabled"`
	ContentSecurityPolicy string `yaml:"content_security_policy"`
	PermissionPolicy      string `yaml:"permission_policy"`
	HSTSMaxAge            int    `yaml:"hsts_max_age"`
	HSTSExcludeSubdomains bool   `yaml:"hsts_exclude_subdomains"`
	HSTSPreloadEnabled    bool   `yaml:"hsts_preload_enabled"`
	CSPReportOnly         bool   `yaml:"csp_report_only"`
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

type CSRFConfig struct {
	Enabled            bool     `yaml:"enabled"`
	TokenPath          string   `yaml:"token_path"`
	CookieName         string   `yaml:"cookie_name"`
	CookieSameSite     string   `yaml:"cookie_same_site"`
	CookieSecure       bool     `yaml:"cookie_secure"`
	CookieHTTPOnly     bool     `yaml:"cookie_http_only"`
	CookieSessionOnly  bool     `yaml:"cookie_session_only"`
	IdleTimeoutSeconds int      `yaml:"idle_timeout_seconds"`
	SingleUseToken     bool     `yaml:"single_use_token"`
	TrustedOrigins     []string `yaml:"trusted_origins"`
	ExcludePaths       []string `yaml:"exclude_paths"`
}

type ETagConfig struct {
	Enabled bool `yaml:"enabled"`
	Weak    bool `yaml:"weak"`
}

type CorsConfig struct {
	AllowOrigins []string `yaml:"allow_origins"`
}

type SwaggerConfig struct {
	Enabled  bool   `yaml:"enabled"`
	FilePath string `yaml:"file_path"`
	BasePath string `yaml:"base_path"`
	Path     string `yaml:"path"`
	Title    string `yaml:"title"`
}

type RedisConfig struct {
	Enabled       bool   `yaml:"enabled"`
	RedisUsername string `yaml:"redis_username"`
	RedisAddr     string `yaml:"redis_addr"`
	RedisPassword string `yaml:"redis_password"`
	RedisDB       int    `yaml:"redis_db"`
	RedisPoolSize int    `yaml:"redis_pool_size"`
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
	Postgres    SQLDataBaseConfig    `yaml:"postgres"`
	MySQL       SQLDataBaseConfig    `yaml:"mysql"`
	DSN         string               `yaml:"dsn"`
	DBName      string               `yaml:"db_name"`
	DBHost      string               `yaml:"db_host"`
	DBPort      string               `yaml:"db_port"`
	DBUser      string               `yaml:"db_username"`
	DBPass      string               `yaml:"db_password"`
	SQLPath     string               `yaml:"sqlite_path"`
}

type SQLDataBaseConfig struct {
	DSN    string `yaml:"dsn"`
	DBName string `yaml:"db_name"`
	DBHost string `yaml:"db_host"`
	DBPort string `yaml:"db_port"`
	DBUser string `yaml:"db_username"`
	DBPass string `yaml:"db_password"`
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
	{name: "database.postgres.dsn", kind: "string"},
	{name: "database.postgres.db_name", kind: "string"},
	{name: "database.postgres.db_host", kind: "string"},
	{name: "database.postgres.db_port", kind: "string"},
	{name: "database.postgres.db_username", kind: "string"},
	{name: "database.postgres.db_password", kind: "string"},
	{name: "database.mysql.dsn", kind: "string"},
	{name: "database.mysql.db_name", kind: "string"},
	{name: "database.mysql.db_host", kind: "string"},
	{name: "database.mysql.db_port", kind: "string"},
	{name: "database.mysql.db_username", kind: "string"},
	{name: "database.mysql.db_password", kind: "string"},
	{name: "database.dsn", kind: "string"},
	{name: "database.db_name", kind: "string"},
	{name: "database.db_host", kind: "string"},
	{name: "database.db_port", kind: "string"},
	{name: "database.db_username", kind: "string"},
	{name: "database.db_password", kind: "string"},
	{name: "database.sqlite_path", kind: "string"},
	{name: "log.log_level", kind: "string"},
	{name: "log.log_mode", kind: "string"},
	{name: "log.log_path", kind: "string"},
	{name: "log.log_max_size", kind: "int"},
	{name: "log.log_max_backups", kind: "int"},
	{name: "log.log_max_age", kind: "int"},
	{name: "redis.enabled", kind: "bool"},
	{name: "redis.redis_username", kind: "string"},
	{name: "redis.redis_addr", kind: "string"},
	{name: "redis.redis_password", kind: "string"},
	{name: "redis.redis_db", kind: "int"},
	{name: "redis.redis_pool_size", kind: "int"},
	{name: "middleware.swagger.enabled", kind: "bool"},
	{name: "middleware.swagger.file_path", kind: "string"},
	{name: "middleware.swagger.base_path", kind: "string"},
	{name: "middleware.swagger.path", kind: "string"},
	{name: "middleware.swagger.title", kind: "string"},
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
	{name: "middleware.security_headers.enabled", kind: "bool"},
	{name: "middleware.security_headers.content_security_policy", kind: "string"},
	{name: "middleware.security_headers.permission_policy", kind: "string"},
	{name: "middleware.security_headers.hsts_max_age", kind: "int"},
	{name: "middleware.security_headers.hsts_exclude_subdomains", kind: "bool"},
	{name: "middleware.security_headers.hsts_preload_enabled", kind: "bool"},
	{name: "middleware.security_headers.csp_report_only", kind: "bool"},
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
	{name: "middleware.csrf.enabled", kind: "bool"},
	{name: "middleware.csrf.token_path", kind: "string"},
	{name: "middleware.csrf.cookie_name", kind: "string"},
	{name: "middleware.csrf.cookie_same_site", kind: "string"},
	{name: "middleware.csrf.cookie_secure", kind: "bool"},
	{name: "middleware.csrf.cookie_http_only", kind: "bool"},
	{name: "middleware.csrf.cookie_session_only", kind: "bool"},
	{name: "middleware.csrf.idle_timeout_seconds", kind: "int"},
	{name: "middleware.csrf.single_use_token", kind: "bool"},
	{name: "middleware.csrf.trusted_origins", kind: "string_slice"},
	{name: "middleware.csrf.exclude_paths", kind: "string_slice"},
	{name: "middleware.etag.enabled", kind: "bool"},
	{name: "middleware.etag.weak", kind: "bool"},
	{name: "waf.enabled", kind: "bool"},
	{name: "waf.conf_path", kind: "string_slice"},
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

	if cfg.Middleware.Swagger.Title == "" {
		cfg.Middleware.Swagger.Title = cfg.Server.AppName
	}
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
	if cfg.Middleware.CSRF.TokenPath == "" {
		cfg.Middleware.CSRF.TokenPath = "/csrf/token"
	}
	if cfg.Middleware.CSRF.CookieName == "" {
		cfg.Middleware.CSRF.CookieName = "csrf_"
	}
	if cfg.Middleware.CSRF.CookieSameSite == "" {
		cfg.Middleware.CSRF.CookieSameSite = "Lax"
	}
	if cfg.Middleware.CSRF.IdleTimeoutSeconds <= 0 {
		cfg.Middleware.CSRF.IdleTimeoutSeconds = 1800
	}
	cfg.Middleware.CSRF.TrustedOrigins = normalizeStringList(cfg.Middleware.CSRF.TrustedOrigins)
	cfg.Middleware.CSRF.ExcludePaths = normalizeStringList(cfg.Middleware.CSRF.ExcludePaths)
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

	if cfg.Redis.Enabled && strings.TrimSpace(cfg.Redis.RedisAddr) == "" {
		errs = append(errs, fmt.Errorf("redis.redis_addr is required when redis.enabled is true"))
	}
	if cfg.Redis.RedisDB < 0 {
		errs = append(errs, fmt.Errorf("redis.redis_db must be >= 0"))
	}
	if cfg.Redis.RedisPoolSize < 0 {
		errs = append(errs, fmt.Errorf("redis.redis_pool_size must be >= 0"))
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

	if cfg.Middleware.CSRF.Enabled {
		if strings.TrimSpace(cfg.Middleware.CSRF.TokenPath) == "" {
			errs = append(errs, fmt.Errorf("middleware.csrf.token_path is required when csrf is enabled"))
		}
		if cfg.Middleware.CSRF.IdleTimeoutSeconds <= 0 {
			errs = append(errs, fmt.Errorf("middleware.csrf.idle_timeout_seconds must be > 0 when csrf is enabled"))
		}
	}

	if cfg.Waf.Enabled && len(cfg.Waf.ConfPath) == 0 {
		errs = append(errs, fmt.Errorf("waf.conf_path is required when waf.enabled is true"))
	}

	switch cfg.DataBase.DBType {
	case "postgres", "postgresql", "mysql", "sqlite":
	case "":
		errs = append(errs, fmt.Errorf("database.db_type is required when database.enabled is true"))
	default:
		errs = append(errs, fmt.Errorf("database.db_type %q is not supported", cfg.DataBase.DBType))
	}

	if cfg.DataBase.Enabled && cfg.DataBase.DBType == "sqlite" {
		if strings.TrimSpace(cfg.DataBase.SQLite.DSN) == "" && strings.TrimSpace(cfg.DataBase.SQLite.Path) == "" {
			errs = append(errs, fmt.Errorf("database.sqlite.path or database.sqlite.dsn is required when sqlite is enabled"))
		}
	}

	return errors.Join(errs...)
}

func (cfg *DataBaseConfig) normalize() {
	cfg.DBType = strings.ToLower(strings.TrimSpace(cfg.DBType))
	if cfg.DBType == "" {
		cfg.DBType = "postgres"
	}

	cfg.applyLegacyConfig()

	if cfg.SQLite.Path == "" {
		cfg.SQLite.Path = "./data/app.db"
	}

	normalizeSQLDefaults(&cfg.Postgres, SQLDataBaseConfig{
		DBHost: "127.0.0.1",
		DBPort: "5432",
		DBName: "gf",
		DBUser: "postgres",
		DBPass: "123456",
	})

	normalizeSQLDefaults(&cfg.MySQL, SQLDataBaseConfig{
		DBHost: "127.0.0.1",
		DBPort: "3306",
		DBName: "gf",
		DBUser: "root",
		DBPass: "123456",
	})
}

func (cfg *DataBaseConfig) applyLegacyConfig() {
	switch cfg.DBType {
	case "sqlite":
		if cfg.SQLite.DSN == "" {
			cfg.SQLite.DSN = strings.TrimSpace(cfg.DSN)
		}
		if cfg.SQLite.Path == "" {
			cfg.SQLite.Path = strings.TrimSpace(cfg.SQLPath)
		}
		if cfg.SQLite.Path == "" {
			cfg.SQLite.Path = strings.TrimSpace(cfg.DBName)
		}
	case "mysql":
		applyLegacySQLConfig(&cfg.MySQL, cfg)
	default:
		applyLegacySQLConfig(&cfg.Postgres, cfg)
	}
}

func applyLegacySQLConfig(target *SQLDataBaseConfig, legacy *DataBaseConfig) {
	if target.DSN == "" {
		target.DSN = strings.TrimSpace(legacy.DSN)
	}
	if target.DBName == "" {
		target.DBName = strings.TrimSpace(legacy.DBName)
	}
	if target.DBHost == "" {
		target.DBHost = strings.TrimSpace(legacy.DBHost)
	}
	if target.DBPort == "" {
		target.DBPort = strings.TrimSpace(legacy.DBPort)
	}
	if target.DBUser == "" {
		target.DBUser = strings.TrimSpace(legacy.DBUser)
	}
	if target.DBPass == "" {
		target.DBPass = strings.TrimSpace(legacy.DBPass)
	}
}

func normalizeSQLDefaults(target *SQLDataBaseConfig, defaults SQLDataBaseConfig) {
	if target.DBHost == "" {
		target.DBHost = defaults.DBHost
	}
	if target.DBPort == "" {
		target.DBPort = defaults.DBPort
	}
	if target.DBName == "" {
		target.DBName = defaults.DBName
	}
	if target.DBUser == "" {
		target.DBUser = defaults.DBUser
	}
	if target.DBPass == "" {
		target.DBPass = defaults.DBPass
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

	applyDefaults(v, projectName)
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

func applyDefaults(v *viper.Viper, projectName string) {
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
	v.SetDefault("database.db_type", "sqlite")
	v.SetDefault("database.auto_migrate", true)
	v.SetDefault("database.sqlite.path", "./data/app.db")
	v.SetDefault("database.postgres.db_host", "127.0.0.1")
	v.SetDefault("database.postgres.db_port", "5432")
	v.SetDefault("database.postgres.db_name", "postgres")
	v.SetDefault("database.postgres.db_username", "postgres")
	v.SetDefault("database.postgres.db_password", "123456")
	v.SetDefault("database.mysql.db_host", "127.0.0.1")
	v.SetDefault("database.mysql.db_port", "3306")
	v.SetDefault("database.mysql.db_name", "mysql")
	v.SetDefault("database.mysql.db_username", "root")
	v.SetDefault("database.mysql.db_password", "123456")
	v.SetDefault("redis.redis_username", "")
	v.SetDefault("redis.redis_addr", "127.0.0.1:6379")
	v.SetDefault("redis.redis_password", "")
	v.SetDefault("redis.redis_db", 0)
	v.SetDefault("redis.redis_pool_size", 10)
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
	v.SetDefault("middleware.security_headers.enabled", true)
	v.SetDefault("middleware.security_headers.hsts_max_age", 0)
	v.SetDefault("middleware.security_headers.hsts_exclude_subdomains", false)
	v.SetDefault("middleware.security_headers.hsts_preload_enabled", false)
	v.SetDefault("middleware.security_headers.csp_report_only", false)
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
	v.SetDefault("middleware.csrf.enabled", false)
	v.SetDefault("middleware.csrf.token_path", "/csrf/token")
	v.SetDefault("middleware.csrf.cookie_name", "csrf_")
	v.SetDefault("middleware.csrf.cookie_same_site", "Lax")
	v.SetDefault("middleware.csrf.cookie_secure", false)
	v.SetDefault("middleware.csrf.cookie_http_only", false)
	v.SetDefault("middleware.csrf.cookie_session_only", false)
	v.SetDefault("middleware.csrf.idle_timeout_seconds", 1800)
	v.SetDefault("middleware.csrf.single_use_token", false)
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
