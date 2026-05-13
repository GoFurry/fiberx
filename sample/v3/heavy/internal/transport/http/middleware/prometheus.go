package middleware

import (
	"errors"
	"strconv"
	"strings"
	"time"

	log "github.com/gofurry/fiberx/v3/heavy/internal/infra/logging"
	"github.com/gofurry/fiberx/v3/heavy/internal/infra/observability/metrics"
	"github.com/gofurry/fiberx/v3/heavy/pkg/common"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/gofiber/utils/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type FiberPromConf struct {
	Namespace         string
	SkipPaths         []string
	IgnoreStatusCodes []int
}

type Prometheus struct {
	metrics           *metrics.HTTPMetrics
	metricsHandler    fiber.Handler
	skipPaths         map[string]bool
	ignoreStatusCodes map[int]bool
}

func NewPrometheus(conf FiberPromConf) *Prometheus {
	collectors := metrics.New(conf.Namespace)
	registry := prometheus.NewRegistry()
	registry.MustRegister(collectors.HttpRequestsTotal)
	registry.MustRegister(collectors.HttpRequestDuration)
	registry.MustRegister(collectors.HttpActiveRequests)

	prometheusMiddleware := &Prometheus{
		metrics:           collectors,
		metricsHandler:    adaptor.HTTPHandler(promhttp.HandlerFor(registry, promhttp.HandlerOpts{})),
		skipPaths:         buildSkipPaths(conf.SkipPaths),
		ignoreStatusCodes: buildIgnoreStatusCodes(conf.IgnoreStatusCodes),
	}

	log.Debug("[prometheus] middleware initialized")
	return prometheusMiddleware
}

func (pm *Prometheus) Handler(c fiber.Ctx) error {
	method := utils.CopyString(c.Method())

	pm.metrics.HttpActiveRequests.Inc()
	defer pm.metrics.HttpActiveRequests.Dec()

	start := time.Now()
	err := c.Next()

	routePath := normalizePath(resolveRoutePath(c))
	if pm.skipPaths[routePath] {
		return err
	}

	status := fiber.StatusInternalServerError
	if err != nil {
		if fiberErr, ok := err.(*fiber.Error); ok {
			status = fiberErr.Code
		} else if appErr, ok := errors.AsType[common.Error](err); ok {
			status = appErr.GetHTTPStatus()
		}
	} else {
		status = c.Response().StatusCode()
	}
	if pm.ignoreStatusCodes[status] {
		return err
	}

	pm.metrics.HttpRequestsTotal.WithLabelValues(method, routePath, strconv.Itoa(status)).Inc()
	pm.metrics.HttpRequestDuration.WithLabelValues(method, routePath).Observe(time.Since(start).Seconds())
	return err
}

func (pm *Prometheus) MetricsHandler() fiber.Handler {
	return pm.metricsHandler
}

func resolveRoutePath(c fiber.Ctx) string {
	route := c.Route()
	if route == nil {
		return c.Path()
	}

	path := utils.CopyString(route.Path)
	if path == "" || path == "/" {
		path = utils.CopyString(c.Path())
	}

	return path
}

func normalizePath(routePath string) string {
	normalized := strings.TrimRight(strings.TrimSpace(routePath), "/")
	if normalized == "" {
		return "/"
	}
	return normalized
}

func buildSkipPaths(paths []string) map[string]bool {
	result := map[string]bool{
		"/metrics": true,
	}
	for _, path := range paths {
		path = normalizePath(path)
		if path == "" {
			continue
		}
		result[path] = true
	}
	return result
}

func buildIgnoreStatusCodes(codes []int) map[int]bool {
	result := make(map[int]bool, len(codes))
	for _, code := range codes {
		result[code] = true
	}
	return result
}
