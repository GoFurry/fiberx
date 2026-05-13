package cache

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	env "github.com/gofurry/fiberx/v3/heavy/config"
	log "github.com/gofurry/fiberx/v3/heavy/internal/infra/logging"
	"github.com/gofurry/fiberx/v3/heavy/pkg/common"
	goredis "github.com/redis/go-redis/v9"
)

var (
	service           *Service
	backgroundContext = context.Background()
)

type Config struct {
	Addr     string
	Username string
	Password string
	DB       int
	PoolSize int
}

type Service struct {
	cfg Config
	raw *goredis.Client
}

type Pipeliner = goredis.Pipeliner

func New(ctx context.Context, cfg Config) (*Service, error) {
	normalized, err := normalizeConfig(cfg)
	if err != nil {
		return nil, err
	}

	client := goredis.NewClient(&goredis.Options{
		Addr:      normalized.Addr,
		Username:  normalized.Username,
		Password:  normalized.Password,
		DB:        normalized.DB,
		PoolSize:  normalized.PoolSize,
		OnConnect: onConnect,
	})

	svc := &Service{
		cfg: normalized,
		raw: client,
	}

	connCtx, cancel := context.WithTimeout(ctxOrBackground(ctx), 5*time.Second)
	defer cancel()
	if err := svc.Ping(connCtx); err != nil {
		_ = svc.Close()
		return nil, fmt.Errorf("ping redis failed: %w", err)
	}

	return svc, nil
}

func GetRedisService() *goredis.Client {
	if service == nil {
		return nil
	}
	return service.Raw()
}

func Raw() *goredis.Client {
	return GetRedisService()
}

func RedisReady() bool {
	if service == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return service.Ping(ctx) == nil
}

func InitRedisOnStart() error {
	cfg := env.GetServerConfig().Redis
	svc, err := New(context.Background(), Config{
		Addr:     cfg.RedisAddr,
		Username: cfg.RedisUsername,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
		PoolSize: cfg.RedisPoolSize,
	})
	if err != nil {
		return fmt.Errorf("connect to redis failed: %w", err)
	}

	service = svc
	log.Debug("redis connected")
	return nil
}

func Close() error {
	if service == nil {
		return nil
	}

	err := service.Close()
	service = nil
	return err
}

func (s *Service) Raw() *goredis.Client {
	if s == nil {
		return nil
	}
	return s.raw
}

func (s *Service) Ping(ctx context.Context) error {
	if s == nil || s.raw == nil {
		return errors.New("redis service is not initialized")
	}
	return s.raw.Ping(ctxOrBackground(ctx)).Err()
}

func (s *Service) Close() error {
	if s == nil || s.raw == nil {
		return nil
	}
	err := s.raw.Close()
	s.raw = nil
	return err
}

func (s *Service) Del(ctx context.Context, keys ...string) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	return s.raw.Del(ctxOrBackground(ctx), keys...).Err()
}

func (s *Service) SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error) {
	if err := s.ensureReady(); err != nil {
		return false, err
	}

	result, err := s.raw.SetArgs(ctxOrBackground(ctx), key, value, goredis.SetArgs{
		TTL:  expiration,
		Mode: "NX",
	}).Result()
	switch {
	case errors.Is(err, goredis.Nil):
		return false, nil
	case err != nil:
		return false, err
	default:
		return strings.EqualFold(result, "OK"), nil
	}
}

func (s *Service) Set(ctx context.Context, key string, value any) error {
	return s.SetExpire(ctx, key, value, 0)
}

func (s *Service) SetExpire(ctx context.Context, key string, value any, expiration time.Duration) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	return s.raw.Set(ctxOrBackground(ctx), key, value, expiration).Err()
}

func (s *Service) GetString(ctx context.Context, key string) (string, error) {
	if err := s.ensureReady(); err != nil {
		return "", err
	}

	value, err := s.raw.Get(ctxOrBackground(ctx), key).Result()
	switch {
	case errors.Is(err, goredis.Nil):
		return "", nil
	case err != nil:
		return "", err
	default:
		return strings.TrimSpace(value), nil
	}
}

func (s *Service) HSetMap(ctx context.Context, key string, values map[string]string) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	return s.raw.HSet(ctxOrBackground(ctx), key, values).Err()
}

func (s *Service) HSet(ctx context.Context, key string, fieldName string, fieldValue string) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	return s.raw.HSet(ctxOrBackground(ctx), key, fieldName, fieldValue).Err()
}

func (s *Service) HGet(ctx context.Context, key string, fieldName string) (string, error) {
	if err := s.ensureReady(); err != nil {
		return "", err
	}

	value, err := s.raw.HGet(ctxOrBackground(ctx), key, fieldName).Result()
	switch {
	case errors.Is(err, goredis.Nil):
		return "", fmt.Errorf("redis hash field not found: %s.%s", key, fieldName)
	case err != nil:
		return "", err
	default:
		return value, nil
	}
}

func (s *Service) HMGet(ctx context.Context, key string, fields ...string) ([]interface{}, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}

	value, err := s.raw.HMGet(ctxOrBackground(ctx), key, fields...).Result()
	switch {
	case errors.Is(err, goredis.Nil):
		return nil, errors.New("redis hash not found")
	case err != nil:
		return nil, err
	default:
		return value, nil
	}
}

func (s *Service) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}

	value, err := s.raw.HGetAll(ctxOrBackground(ctx), key).Result()
	switch {
	case errors.Is(err, goredis.Nil):
		return nil, errors.New("redis hash not found")
	case err != nil:
		return nil, err
	default:
		return value, nil
	}
}

func (s *Service) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	if err := s.ensureReady(); err != nil {
		return 0, err
	}

	value, err := s.raw.HDel(ctxOrBackground(ctx), key, fields...).Result()
	switch {
	case errors.Is(err, goredis.Nil):
		return 0, errors.New("redis hash not found")
	case err != nil:
		return 0, err
	default:
		return value, nil
	}
}

func (s *Service) Incr(ctx context.Context, key string) (int64, error) {
	if err := s.ensureReady(); err != nil {
		return 0, err
	}
	return s.raw.Incr(ctxOrBackground(ctx), key).Result()
}

func (s *Service) CountByPrefix(ctx context.Context, prefix string) (int64, error) {
	if err := s.ensureReady(); err != nil {
		return 0, err
	}

	var (
		cursor uint64
		count  int
	)
	pattern := prefix + "*"

	for {
		keys, nextCursor, err := s.raw.Scan(ctxOrBackground(ctx), cursor, pattern, 100).Result()
		if err != nil {
			return 0, err
		}

		count += len(keys)
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return int64(count), nil
}

func (s *Service) FindByPrefix(ctx context.Context, prefix string) ([]string, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}

	var (
		cursor  uint64
		results []string
	)
	pattern := prefix + "*"

	for {
		keys, nextCursor, err := s.raw.Scan(ctxOrBackground(ctx), cursor, pattern, 100).Result()
		if err != nil {
			return nil, err
		}

		results = append(results, keys...)
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return results, nil
}

func (s *Service) DelByPrefix(ctx context.Context, prefix string) error {
	if err := s.ensureReady(); err != nil {
		return err
	}

	keys, err := s.FindByPrefix(ctx, prefix)
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	return s.Del(ctx, keys...)
}

func (s *Service) PipelineExec(ctx context.Context, fn func(pipe Pipeliner)) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	if fn == nil {
		return errors.New("redis pipeline function is required")
	}

	pipe := s.raw.Pipeline()
	fn(pipe)
	_, err := pipe.Exec(ctxOrBackground(ctx))
	return err
}

func Get(key string) *goredis.Cmd {
	if service == nil || service.Raw() == nil {
		cmd := goredis.NewCmd(backgroundContext)
		cmd.SetErr(errors.New("redis service is not ready"))
		return cmd
	}
	return service.Raw().Do(backgroundContext, "GET", key)
}

func Del(keys ...string) common.Error {
	if err := current().Del(backgroundContext, keys...); err != nil {
		log.Errorf("redis DEL failed: %v", err)
		return common.NewServiceError("delete redis keys failed")
	}
	return nil
}

func SetNX(key string, value any, expiration time.Duration) bool {
	ok, err := current().SetNX(backgroundContext, key, value, expiration)
	if err != nil {
		log.Errorf("redis SETNX failed: %v", err)
		return false
	}
	return ok
}

func Set(key string, value any) common.Error {
	return SetExpire(key, value, 0)
}

func SetExpire(key string, value any, expiration time.Duration) common.Error {
	if err := current().SetExpire(backgroundContext, key, value, expiration); err != nil {
		log.Errorf("redis SET failed: %v", err)
		return common.NewServiceError("set redis key failed")
	}
	return nil
}

func GetString(key string) (string, common.Error) {
	value, err := current().GetString(backgroundContext, key)
	if err != nil {
		log.Errorf("redis GET failed: %v", err)
		return "", common.NewServiceError("get redis key failed")
	}
	return value, nil
}

func HSetMap(key string, kvMap map[string]string) common.Error {
	if err := current().HSetMap(backgroundContext, key, kvMap); err != nil {
		log.Errorf("redis HSET map failed: %v", err)
		return common.NewServiceError("set redis hash failed")
	}
	return nil
}

func HSet(key string, fieldName string, fieldVal string) common.Error {
	if err := current().HSet(backgroundContext, key, fieldName, fieldVal); err != nil {
		log.Errorf("redis HSET failed: %v", err)
		return common.NewServiceError("set redis hash field failed")
	}
	return nil
}

func HGet(key string, fieldName string) (string, common.Error) {
	value, err := current().HGet(backgroundContext, key, fieldName)
	if err != nil {
		log.Errorf("redis HGET failed: %v", err)
		return "", common.NewServiceError("get redis hash field failed")
	}
	return value, nil
}

func HMGet(key string, fields ...string) ([]interface{}, common.Error) {
	value, err := current().HMGet(backgroundContext, key, fields...)
	if err != nil {
		log.Errorf("redis HMGET failed: %v", err)
		return nil, common.NewServiceError("get redis hash fields failed")
	}
	return value, nil
}

func HGetAll(key string) (map[string]string, common.Error) {
	value, err := current().HGetAll(backgroundContext, key)
	if err != nil {
		log.Errorf("redis HGETALL failed: %v", err)
		return nil, common.NewServiceError("get redis hash failed")
	}
	return value, nil
}

func HDel(key string, fields ...string) (int64, common.Error) {
	value, err := current().HDel(backgroundContext, key, fields...)
	if err != nil {
		log.Errorf("redis HDEL failed: %v", err)
		return 0, common.NewServiceError("delete redis hash fields failed")
	}
	return value, nil
}

func Incr(key string) {
	if _, err := current().Incr(backgroundContext, key); err != nil {
		log.Errorf("redis INCR failed: %v", err)
	}
}

func CountByPrefix(prefix string) (int64, common.Error) {
	value, err := current().CountByPrefix(backgroundContext, prefix)
	if err != nil {
		log.Errorf("redis SCAN failed: %v", err)
		return 0, common.NewServiceError("scan redis keys failed")
	}
	return value, nil
}

func DelByPrefix(prefix string) common.Error {
	if err := current().DelByPrefix(backgroundContext, prefix); err != nil {
		log.Errorf("redis delete by prefix failed: %v", err)
		return common.NewServiceError("scan redis keys failed")
	}
	return nil
}

func FindByPrefix(prefix string) ([]string, common.Error) {
	value, err := current().FindByPrefix(backgroundContext, prefix)
	if err != nil {
		log.Errorf("redis scan failed: %v", err)
		return nil, common.NewServiceError("scan redis keys failed")
	}
	return value, nil
}

func PipelineExec(fn func(pipe Pipeliner)) common.Error {
	if err := current().PipelineExec(backgroundContext, fn); err != nil {
		log.Errorf("redis pipeline execution failed: %v", err)
		return common.NewServiceError("execute redis pipeline failed")
	}
	return nil
}

func current() *Service {
	if service == nil {
		return &Service{}
	}
	return service
}

func (s *Service) ensureReady() error {
	if s == nil || s.raw == nil {
		return errors.New("redis service is not initialized")
	}
	return nil
}

func normalizeConfig(cfg Config) (Config, error) {
	normalized := cfg
	normalized.Addr = strings.TrimSpace(normalized.Addr)
	normalized.Username = strings.TrimSpace(normalized.Username)
	normalized.Password = strings.TrimSpace(normalized.Password)

	if normalized.Addr == "" {
		return Config{}, errors.New("redis addr is required")
	}
	if normalized.PoolSize < 0 {
		return Config{}, errors.New("redis pool size cannot be negative")
	}
	return normalized, nil
}

func onConnect(ctx context.Context, cn *goredis.Conn) error {
	log.Debug("redis connection opened")
	return nil
}

func ctxOrBackground(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
