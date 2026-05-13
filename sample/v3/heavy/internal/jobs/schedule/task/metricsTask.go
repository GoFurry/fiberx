package task

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	env "github.com/gofurry/fiberx/v3/heavy/config"
	cs "github.com/gofurry/fiberx/v3/heavy/internal/infra/cache"
	log "github.com/gofurry/fiberx/v3/heavy/internal/infra/logging"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

var (
	promClient v1.API
	once       sync.Once
)

func initPromClient() v1.API {
	once.Do(func() {
		cfg := env.GetServerConfig()
		if !cfg.Prometheus.Enabled || cfg.Prometheus.Url == "" {
			return
		}

		client, err := api.NewClient(api.Config{Address: cfg.Prometheus.Url})
		if err != nil {
			log.Error("[initPromClient] create prom client err:", err)
			return
		}
		promClient = v1.NewAPI(client)
		log.Infof("[initPromClient] prometheus client init success, url: %s", cfg.Prometheus.Url)
	})
	return promClient
}

func getPromClient() v1.API {
	if promClient == nil {
		return initPromClient()
	}
	return promClient
}

func UpdateMetricsCache() {
	cfg := env.GetServerConfig()
	if !cfg.Prometheus.Enabled || cfg.Prometheus.Url == "" {
		return
	}
	if !cs.RedisReady() {
		log.Warn("[UpdateMetricsCache] redis is not ready, skip metrics update")
		return
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	log.Debug("[UpdateMetricsCache] start metrics cache update")

	prom := getPromClient()
	if prom == nil {
		log.Error("[UpdateMetricsCache] prometheus client is nil, skip metrics update")
		return
	}

	updateNodeMetrics(ctx, prom)
	updateServiceMetrics(ctx, prom, cfg.Prometheus.ServiceMetrics)
	updateServicePerPathMetrics(ctx, prom, cfg.Prometheus.ServiceMetrics)

	log.Debugf("[UpdateMetricsCache] metrics cache update finished, cost: %v", time.Since(start))
}

func updateNodeMetrics(ctx context.Context, prom v1.API) {
	metrics := map[string]struct {
		Query     string
		IsHistory bool
	}{
		"cpu_usage":       {`100 - avg(irate(node_cpu_seconds_total{job="servers",mode="idle"}[5m])) * 100`, true},
		"mem_usage":       {`sum(node_memory_MemTotal_bytes{job="servers"}) - sum(node_memory_MemAvailable_bytes{job="servers"})`, true},
		"disk_usage":      {`100 * (sum(node_filesystem_size_bytes{job="servers",fstype!~"tmpfs|overlay"}) - sum(node_filesystem_avail_bytes{job="servers",fstype!~"tmpfs|overlay"})) / sum(node_filesystem_size_bytes{job="servers",fstype!~"tmpfs|overlay"})`, false},
		"net_rx_1d":       {`sum(increase(node_network_receive_bytes_total{job="servers", device!~"lo|docker0"}[1d]))`, false},
		"net_tx_1d":       {`sum(increase(node_network_transmit_bytes_total{job="servers", device!~"lo|docker0"}[1d]))`, false},
		"tcp_connections": {`sum(node_netstat_Tcp_CurrEstab{job="servers"}) or vector(0)`, true},
		"uptime":          {`avg(time() - node_boot_time_seconds{job="servers"})`, false},
	}

	nodeCurrentFields := make(map[string]string)
	for key, metricCfg := range metrics {
		val, ok := queryPromAgg(ctx, prom, metricCfg.Query)
		if !ok {
			continue
		}
		nodeCurrentFields[key] = fmt.Sprintf("%.4f", val)

		if metricCfg.IsHistory {
			cacheHistory(ctx, "prom:node:history:"+key, val, 7)
		}
	}

	if len(nodeCurrentFields) > 0 {
		err := cs.GetRedisService().HSet(ctx, "prom:node:current", nodeCurrentFields).Err()
		if err != nil {
			log.Errorf("[cacheNodeCurrentBatch] err: %v", err)
		}
	}
}

func updateServiceMetrics(ctx context.Context, prom v1.API, services []string) {
	for _, svc := range services {
		queries := map[string]string{
			"http_requests_1d": fmt.Sprintf(`sum(increase(%s_http_requests_total[1d]))`, svc),
			"http_requests_7d": fmt.Sprintf(`sum(increase(%s_http_requests_total[7d]))`, svc),
			"avg_response_1h":  fmt.Sprintf(`sum(rate(%s_http_request_duration_seconds_sum[1h])) / sum(rate(%s_http_request_duration_seconds_count[1h])) or vector(0)`, svc, svc),
			"p99_response_1h":  fmt.Sprintf(`histogram_quantile(0.99, sum by(le) (rate(%s_http_request_duration_seconds_bucket[1h])))`, svc),
			"p95_response_1h":  fmt.Sprintf(`histogram_quantile(0.95, sum by(le) (rate(%s_http_request_duration_seconds_bucket[1h])))`, svc),
			"fail_rate_1h":     fmt.Sprintf(`sum(rate(%s_http_requests_total{status!~"2.."}[1h])) / sum(rate(%s_http_requests_total[1h])) or vector(0)`, svc, svc),
		}

		svcCurrentFields := make(map[string]string)
		for key, query := range queries {
			val, ok := queryPromAgg(ctx, prom, query)
			if !ok {
				continue
			}
			svcCurrentFields[key] = fmt.Sprintf("%.4f", val)
		}

		if len(svcCurrentFields) > 0 {
			redisKey := "prom:service:" + svc + ":current"
			err := cs.GetRedisService().HSet(ctx, redisKey, svcCurrentFields).Err()
			if err != nil {
				log.Errorf("[cacheServiceCurrentBatch] svc=%s err=%v", svc, err)
			}
		}
	}
}

func updateServicePerPathMetrics(ctx context.Context, prom v1.API, services []string) {
	queries := map[string]string{
		"http_requests_1d": `sum by(path) (increase(%s_http_requests_total[1d]))`,
		"http_requests_7d": `sum by(path) (increase(%s_http_requests_total[7d]))`,
		"avg_response_1h":  `sum by(path) (rate(%s_http_request_duration_seconds_sum[1h])) / sum by(path) (rate(%s_http_request_duration_seconds_count[1h]))`,
	}

	for _, svc := range services {
		for metric, q := range queries {
			query := fmt.Sprintf(q, svc, svc)
			result, ok := queryPromVector(ctx, prom, query)
			if !ok {
				continue
			}
			for path, val := range result {
				cacheServicePathMetric(ctx, svc, metric, path, val)
			}
		}
	}
}

func queryPromAgg(ctx context.Context, api v1.API, query string) (float64, bool) {
	result, _, err := api.Query(ctx, query, time.Now())
	if err != nil {
		return 0, false
	}

	switch v := result.(type) {
	case model.Vector:
		var sum float64
		for _, s := range v {
			sum += float64(s.Value)
		}
		return sum, !math.IsNaN(sum)
	case *model.Scalar:
		return float64(v.Value), true
	}
	return 0, false
}

func queryPromVector(ctx context.Context, api v1.API, query string) (map[string]float64, bool) {
	result, _, err := api.Query(ctx, query, time.Now())
	if err != nil {
		return nil, false
	}

	vec, ok := result.(model.Vector)
	if !ok {
		return nil, false
	}

	out := make(map[string]float64)
	for _, s := range vec {
		out[string(s.Metric["path"])] = float64(s.Value)
	}
	return out, true
}

func cacheServicePathMetric(ctx context.Context, svc, metric, path string, val float64) {
	redisClient := cs.GetRedisService()
	if redisClient == nil {
		return
	}

	key := fmt.Sprintf("prom:service:%s:path:%s", svc, metric)
	redisClient.HSet(ctx, key, path, fmt.Sprintf("%.4f", val))
	redisClient.Expire(ctx, key, 10*time.Minute)
}

func cacheHistory(ctx context.Context, key string, val float64, days int) {
	redisClient := cs.GetRedisService()
	if redisClient == nil {
		return
	}

	ts := time.Now().Unix()
	value := fmt.Sprintf("%.4f", val)

	if _, err := redisClient.Do(ctx, "ZADD", key, ts, value).Result(); err != nil {
		log.Errorf("[cacheHistory RedisZADD] key=%s err=%v", key, err)
		return
	}

	expireTS := time.Now().Add(-time.Duration(days*24) * time.Hour).Unix()
	redisClient.Do(ctx, "ZREMRANGEBYSCORE", key, 0, expireTS)
	redisClient.Do(ctx, "EXPIRE", key, days*24*3600)
}
