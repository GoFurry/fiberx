package core

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gofurry/fiberx/internal/metadata"
	"github.com/gofurry/fiberx/internal/stack"
)

func TestRunSupportsV1PresetMatrix(t *testing.T) {
	testCases := []struct {
		name             string
		preset           string
		capabilities     []string
		fiberVersion     string
		cliStyle         string
		routerPath       string
		routerSnippet    string
		expectRedis      bool
		expectMedium     bool
		expectHeavy      bool
		expectLight      bool
		expectExtraLight bool
		expectDocsAsset  bool
		expectUIAsset    bool
	}{
		{name: "heavy default", preset: "heavy", routerPath: filepath.Join("internal", "transport", "http", "router", "router.go"), routerSnippet: `registerMetricsRoutes(app, deps)`, expectHeavy: true, expectDocsAsset: true, expectUIAsset: true},
		{name: "heavy fiber-v3 native", preset: "heavy", fiberVersion: stack.FiberV3, cliStyle: stack.CLINative, routerPath: filepath.Join("internal", "transport", "http", "router", "router.go"), routerSnippet: `registerMetricsRoutes(app, deps)`, expectHeavy: true, expectDocsAsset: true, expectUIAsset: true},
		{name: "heavy fiber-v2 cobra", preset: "heavy", fiberVersion: stack.FiberV2, cliStyle: stack.CLICobra, routerPath: filepath.Join("internal", "transport", "http", "router", "router.go"), routerSnippet: `registerMetricsRoutes(app, deps)`, expectHeavy: true, expectDocsAsset: true, expectUIAsset: true},
		{name: "heavy fiber-v2 native", preset: "heavy", fiberVersion: stack.FiberV2, cliStyle: stack.CLINative, routerPath: filepath.Join("internal", "transport", "http", "router", "router.go"), routerSnippet: `registerMetricsRoutes(app, deps)`, expectHeavy: true, expectDocsAsset: true, expectUIAsset: true},
		{name: "heavy with redis", preset: "heavy", capabilities: []string{"redis"}, routerPath: filepath.Join("internal", "transport", "http", "router", "router.go"), routerSnippet: `registerMetricsRoutes(app, deps)`, expectRedis: true, expectHeavy: true, expectDocsAsset: true, expectUIAsset: true},
		{name: "medium default", preset: "medium", routerPath: filepath.Join("internal", "transport", "http", "router", "router.go"), routerSnippet: `registerSwaggerRoutes(app, deps.Config)`, expectMedium: true, expectDocsAsset: true, expectUIAsset: true},
		{name: "medium fiber-v3 native", preset: "medium", fiberVersion: stack.FiberV3, cliStyle: stack.CLINative, routerPath: filepath.Join("internal", "transport", "http", "router", "router.go"), routerSnippet: `registerSwaggerRoutes(app, deps.Config)`, expectMedium: true, expectDocsAsset: true, expectUIAsset: true},
		{name: "medium fiber-v2 cobra", preset: "medium", fiberVersion: stack.FiberV2, cliStyle: stack.CLICobra, routerPath: filepath.Join("internal", "transport", "http", "router", "router.go"), routerSnippet: `registerSwaggerRoutes(app, deps.Config)`, expectMedium: true, expectDocsAsset: true, expectUIAsset: true},
		{name: "medium fiber-v2 native", preset: "medium", fiberVersion: stack.FiberV2, cliStyle: stack.CLINative, routerPath: filepath.Join("internal", "transport", "http", "router", "router.go"), routerSnippet: `registerSwaggerRoutes(app, deps.Config)`, expectMedium: true, expectDocsAsset: true, expectUIAsset: true},
		{name: "medium with redis", preset: "medium", capabilities: []string{"redis"}, routerPath: filepath.Join("internal", "transport", "http", "router", "router.go"), routerSnippet: `registerSwaggerRoutes(app, deps.Config)`, expectRedis: true, expectMedium: true, expectDocsAsset: true, expectUIAsset: true},
		{name: "light default", preset: "light", routerPath: filepath.Join("internal", "transport", "http", "router", "router.go"), routerSnippet: `wrapTimeoutRouter(app.Group("/api"), deps.Config.Middleware.Timeout)`, expectLight: true},
		{name: "light fiber-v2 native", preset: "light", fiberVersion: stack.FiberV2, cliStyle: stack.CLINative, routerPath: filepath.Join("internal", "transport", "http", "router", "router.go"), routerSnippet: `wrapTimeoutRouter(app.Group("/api"), deps.Config.Middleware.Timeout)`, expectLight: true},
		{name: "light fiber-v3 native", preset: "light", fiberVersion: stack.FiberV3, cliStyle: stack.CLINative, routerPath: filepath.Join("internal", "transport", "http", "router", "router.go"), routerSnippet: `wrapTimeoutRouter(app.Group("/api"), deps.Config.Middleware.Timeout)`, expectLight: true},
		{name: "light fiber-v2 cobra", preset: "light", fiberVersion: stack.FiberV2, cliStyle: stack.CLICobra, routerPath: filepath.Join("internal", "transport", "http", "router", "router.go"), routerSnippet: `wrapTimeoutRouter(app.Group("/api"), deps.Config.Middleware.Timeout)`, expectLight: true},
		{name: "light with swagger", preset: "light", capabilities: []string{"swagger"}, routerPath: filepath.Join("internal", "transport", "http", "router", "router.go"), routerSnippet: `registerSwaggerRoutes(app, deps.Config)`, expectLight: true, expectDocsAsset: true},
		{name: "light with embedded-ui", preset: "light", capabilities: []string{"embedded-ui"}, routerPath: filepath.Join("internal", "transport", "http", "router", "router.go"), routerSnippet: `registerEmbeddedUIRoutes(app, deps.Config)`, expectLight: true, expectUIAsset: true},
		{name: "extra-light default", preset: "extra-light", routerPath: filepath.Join("internal", "transport", "http", "router", "router.go"), routerSnippet: `registerHealthRoutes(app, deps)`, expectExtraLight: true},
		{name: "extra-light fiber-v2 native", preset: "extra-light", fiberVersion: stack.FiberV2, cliStyle: stack.CLINative, routerPath: filepath.Join("internal", "transport", "http", "router", "router.go"), routerSnippet: `registerHealthRoutes(app, deps)`, expectExtraLight: true},
		{name: "extra-light fiber-v3 native", preset: "extra-light", fiberVersion: stack.FiberV3, cliStyle: stack.CLINative, routerPath: filepath.Join("internal", "transport", "http", "router", "router.go"), routerSnippet: `registerHealthRoutes(app, deps)`, expectExtraLight: true},
		{name: "extra-light fiber-v2 cobra", preset: "extra-light", fiberVersion: stack.FiberV2, cliStyle: stack.CLICobra, routerPath: filepath.Join("internal", "transport", "http", "router", "router.go"), routerSnippet: `registerHealthRoutes(app, deps)`, expectExtraLight: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			targetDir := t.TempDir()
			options := requestOptionsForTest(targetDir, tc.fiberVersion, tc.cliStyle)
			req := Request{
				ProjectName:  "demo",
				ModulePath:   "github.com/example/demo",
				Preset:       tc.preset,
				Capabilities: tc.capabilities,
				Options:      options,
			}

			summary, err := Run(req)
			if err != nil {
				t.Fatalf("Run() returned error: %v", err)
			}

			if summary.Preset != tc.preset {
				t.Fatalf("expected preset %q, got %q", tc.preset, summary.Preset)
			}
			if summary.TargetDir != targetDir {
				t.Fatalf("expected target dir %q, got %q", targetDir, summary.TargetDir)
			}
			if summary.FiberVersion != expectedFiberVersion(tc.fiberVersion) {
				t.Fatalf("expected fiber version %q, got %q", expectedFiberVersion(tc.fiberVersion), summary.FiberVersion)
			}
			if summary.CLIStyle != expectedCLIStyle(tc.cliStyle) {
				t.Fatalf("expected cli style %q, got %q", expectedCLIStyle(tc.cliStyle), summary.CLIStyle)
			}
			if summary.Base != expectedBaseName(tc.cliStyle) {
				t.Fatalf("expected base %q, got %q", expectedBaseName(tc.cliStyle), summary.Base)
			}

			assertGeneratedFileContains(t, targetDir, "README.md", tc.preset)
			assertGeneratedFileContains(t, targetDir, "README.md", "github.com/example/demo")
			assertGeneratedFileContains(t, targetDir, "README.md", "fiber version: `"+expectedFiberVersion(tc.fiberVersion)+"`")
			assertGeneratedFileContains(t, targetDir, "README.md", "cli style: `"+expectedCLIStyle(tc.cliStyle)+"`")
			assertGeneratedFileContains(t, targetDir, "README.md", "json backend: `stdlib`")
			assertGeneratedFileContains(t, targetDir, "README.md", "docs/runbook.md")
			assertGeneratedFileContains(t, targetDir, filepath.Join("docs", "runbook.md"), "config/server.prod.yaml")
			assertGeneratedFileContains(t, targetDir, filepath.Join("docs", "configuration.md"), "config/server.dev.yaml")
			assertGeneratedFileContains(t, targetDir, filepath.Join("docs", "api-contract.md"), `"code": 1`)
			assertGeneratedFileContains(t, targetDir, filepath.Join("docs", "verification.md"), "/healthz")
			assertGeneratedFileContains(t, targetDir, filepath.Join("config", "server.dev.yaml"), `mode: "debug"`)
			assertGeneratedFileContains(t, targetDir, filepath.Join("config", "server.prod.yaml"), `mode: "release"`)
			assertGeneratedFileContains(t, targetDir, "fiberx.yaml", "version:")
			assertGeneratedFileContains(t, targetDir, "fiberx.yaml", "source: git")
			assertGeneratedFileContains(t, targetDir, "fiberx.yaml", "parallel: false")
			assertGeneratedFileContains(t, targetDir, "fiberx.yaml", "checksum:")
			assertGeneratedFileContains(t, targetDir, "fiberx.yaml", "archive:")
			assertGeneratedFileContains(t, targetDir, filepath.Join("internal", "version", "version.go"), `Version   = "dev"`)
			assertGeneratedFileContains(t, targetDir, tc.routerPath, tc.routerSnippet)
			assertGeneratedFileContains(t, targetDir, tc.routerPath, `github.com/gofiber/fiber/`+expectedFiberVersion(tc.fiberVersion))
			assertGeneratedFileContains(t, targetDir, tc.routerPath, `JSONEncoder:`)
			assertGeneratedFileContains(t, targetDir, tc.routerPath, `json.Marshal`)
			assertGeneratedFileContains(t, targetDir, tc.routerPath, `JSONDecoder:`)
			assertGeneratedFileContains(t, targetDir, tc.routerPath, `json.Unmarshal`)
			assertGeneratedFileNotContains(t, targetDir, tc.routerPath, `HealthzPath`)
			assertGeneratedFileContains(t, targetDir, tc.routerPath, `message := "internal server error"`)
			assertGeneratedFileNotContains(t, targetDir, tc.routerPath, `message := err.Error()`)
			assertGeneratedFileContains(t, targetDir, filepath.Join("internal", "bootstrap", "serve.go"), `signal.NotifyContext`)
			assertGeneratedFileContains(t, targetDir, filepath.Join("internal", "bootstrap", "serve.go"), `app.Shutdown()`)
			assertGeneratedFileContains(t, targetDir, "go.mod", expectedFiberDependency(tc.fiberVersion))
			if expectedCLIStyle(tc.cliStyle) == stack.CLICobra {
				assertGeneratedFileContains(t, targetDir, filepath.Join("cmd", "root.go"), `github.com/spf13/cobra`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("cmd", "root.go"), `github.com/spf13/viper`)
			} else {
				assertGeneratedFileContains(t, targetDir, filepath.Join("cmd", "root.go"), `bootstrap.Main(args)`)
			}
			if tc.expectDocsAsset {
				assertGeneratedFileContains(t, targetDir, filepath.Join("config", "server.yaml"), `route_prefix: "/docs"`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("docs", "openapi.yaml"), "openapi: 3.0.3")
			} else {
				assertGeneratedFileMissing(t, targetDir, filepath.Join("docs", "openapi.yaml"))
			}
			if tc.expectUIAsset {
				assertGeneratedFileContains(t, targetDir, filepath.Join("config", "server.yaml"), `route_prefix: "/ui"`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("internal", "transport", "http", "webui", "dist", "index.html"), "embedded UI ships")
			} else {
				assertGeneratedFileMissing(t, targetDir, filepath.Join("internal", "transport", "http", "webui", "dist", "index.html"))
			}
			if tc.expectHeavy {
				assertGeneratedFileContains(t, targetDir, filepath.Join("config", "server.yaml"), `/metrics`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("config", "server.yaml"), `route_prefix: "/metrics"`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("config", "server.yaml"), `interval: "1s"`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("internal", "bootstrap", "serve.go"), `jobs.Start(cfg.Scheduler, logger, metricsCollector)`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("internal", "bootstrap", "serve.go"), `"metrics:http"`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("internal", "bootstrap", "serve.go"), `"jobs:scheduler"`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("internal", "infra", "metrics", "metrics.go"), "fiberx_requests_total")
				assertGeneratedFileContains(t, targetDir, filepath.Join("internal", "jobs", "scheduler.go"), "fiberx heavy demo job ran")
			}
			if tc.expectMedium || tc.expectHeavy || tc.expectLight {
				assertGeneratedFileContains(t, targetDir, tc.routerPath, `wrapTimeoutRouter(app.Group("/api"), deps.Config.Middleware.Timeout)`)
				assertGeneratedFileNotContains(t, targetDir, tc.routerPath, `RouteRegistrar`)
				assertGeneratedFileNotContains(t, targetDir, tc.routerPath, `registerApplicationRoutes(`)
				assertGeneratedFileNotContains(t, targetDir, tc.routerPath, `UserController`)
				assertGeneratedFileNotContains(t, targetDir, tc.routerPath, `registerUserRoutes(`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("internal", "transport", "http", "router", "timeout_router.go"), `type timeoutRouter struct`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("internal", "transport", "http", "router", "timeout_router.go"), `func wrapTimeoutRouter(`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("internal", "transport", "http", "router", "timeout_router.go"), `"request timeout"`)
				if expectedFiberVersion(tc.fiberVersion) == stack.FiberV3 {
					assertGeneratedFileContains(t, targetDir, filepath.Join("internal", "transport", "http", "router", "timeout_router.go"), `wrapped := make([]any, 0, len(handlers))`)
				}
				assertGeneratedFileContains(t, targetDir, filepath.Join("internal", "transport", "http", "router", "router.go"), `type AppModules struct`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("internal", "transport", "http", "router", "router.go"), `func NewApp(deps Dependencies, modules AppModules)`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("internal", "transport", "http", "router", "router.go"), `api(wrapTimeoutRouter(app.Group("/api"), deps.Config.Middleware.Timeout), modules)`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("internal", "transport", "http", "router", "url.go"), `func api(root fiber.Router, modules AppModules)`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("internal", "transport", "http", "router", "url.go"), `func v1(root fiber.Router, modules AppModules)`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("internal", "transport", "http", "router", "url.go"), `func userRoutes(root fiber.Router, userAPI *usercontroller.API)`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("internal", "transport", "http", "router", "url.go"), `root.Post("/", userAPI.Create)`)
				assertGeneratedFileNotContains(t, targetDir, filepath.Join("internal", "transport", "http", "router", "url.go"), `UserBasePath = "/api/v1/user"`)
				assertGeneratedFileMissing(t, targetDir, filepath.Join("internal", "bootstrap", "route_registrars.go"))
				assertGeneratedFileContains(t, targetDir, filepath.Join("pkg", "common", "constant.go"), `TimeFormatDigitDay`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("pkg", "common", "constant.go"), `ReturnSuccess`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("pkg", "common", "constant.go"), `ContentTypeJSON`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("pkg", "common", "error.go"), `type APIError interface`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("pkg", "common", "error.go"), `type AppError struct`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("pkg", "common", "error.go"), `NewValidationError`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("pkg", "common", "error.go"), `NewInternalError`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("pkg", "common", "error.go"), `func (e AppError) Unwrap() error`)
				assertGeneratedFileNotContains(t, targetDir, filepath.Join("pkg", "common", "error.go"), `type AppErrorContract interface`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("pkg", "common", "response.go"), `func Success(`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("pkg", "common", "response.go"), `func Error(`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("pkg", "common", "response.go"), `func NewResponse(`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("pkg", "common", "response.go"), `normalizeError(`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("pkg", "common", "response.go"), `return NewResponse(c).SuccessWithData(data)`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("pkg", "common", "response.go"), `return NewInternalError(value)`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("config", "config.go"), `type TimeoutConfig struct`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("config", "config.go"), `cfg := defaultConfig()`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("config", "config.go"), `func defaultConfig() Config`)
				assertGeneratedFileNotContains(t, targetDir, filepath.Join("config", "config.go"), `if !c.Log.LogCompress`)
				assertGeneratedFileNotContains(t, targetDir, filepath.Join("config", "config.go"), `if !c.Log.LogShowLine`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("config", "config.go"), `middleware.timeout.duration_seconds must be greater than 0 when timeout is enabled`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("config", "server.yaml"), `timeout:`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("config", "server.yaml"), `duration_seconds: 15`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("config", "server.yaml"), `exclude_paths:`)
				assertGeneratedFileContains(t, targetDir, tc.routerPath, `app.Use(httpmiddleware.AccessLog(logger))`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("internal", "infra", "db", "sqlite.go"), `ensureSQLiteParentDir(dsn)`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("internal", "infra", "db", "sqlite.go"), `return os.MkdirAll(dir, 0o755)`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("internal", "bootstrap", "serve.go"), `modules := httprouter.AppModules{`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("internal", "bootstrap", "serve.go"), `usercontroller.New(userSvc)`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("internal", "bootstrap", "serve.go"), `httprouter.NewApp(`)
				assertGeneratedFileNotContains(t, targetDir, filepath.Join("internal", "bootstrap", "serve.go"), `registrars := buildRouteRegistrars(`)
				assertGeneratedFileNotContains(t, targetDir, filepath.Join("internal", "bootstrap", "serve.go"), `userStore :=`)
				assertGeneratedFileNotContains(t, targetDir, filepath.Join("internal", "bootstrap", "serve.go"), `userservice.Init(`)
				assertGeneratedFileNotContains(t, targetDir, filepath.Join("internal", "bootstrap", "serve.go"), `userService :=`)
				assertGeneratedFileNotContains(t, targetDir, filepath.Join("internal", "bootstrap", "serve.go"), `userController :=`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("internal", "app", "user", "controller", "user_controller.go"), `type API struct`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("internal", "app", "user", "controller", "user_controller.go"), `func New(svc *service.Service) *API`)
				assertGeneratedFileNotContains(t, targetDir, filepath.Join("internal", "app", "user", "controller", "user_controller.go"), `err.Error()`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("internal", "app", "user", "controller", "user_controller.go"), `common.NewResponse(c).Error(`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("internal", "app", "user", "controller", "user_controller.go"), `mapServiceError(err)`)
				assertGeneratedFileNotContains(t, targetDir, filepath.Join("internal", "app", "user", "controller", "user_controller.go"), `return common.Error(c, fiber.StatusInternalServerError, "request failed")`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("internal", "app", "user", "service", "user_service.go"), `func New(`)
				assertGeneratedFileNotContains(t, targetDir, filepath.Join("internal", "app", "user", "service", "user_service.go"), `func GetUserService() *Service`)
				assertGeneratedFileNotContains(t, targetDir, filepath.Join("internal", "app", "user", "service", "user_service.go"), `func Init(`)
				assertGeneratedFileNotContains(t, targetDir, filepath.Join("internal", "app", "user", "service", "user_service.go"), `cacheKey := fmt.Sprintf("users:`)
				if expectedFiberVersion(tc.fiberVersion) == stack.FiberV3 {
					assertGeneratedFileContains(t, targetDir, filepath.Join("internal", "bootstrap", "serve.go"), `registerAppHooks(app, cfg, logger)`)
					assertGeneratedFileContains(t, targetDir, filepath.Join("internal", "bootstrap", "app_hooks.go"), `OnPreStartupMessage`)
					assertGeneratedFileContains(t, targetDir, filepath.Join("internal", "bootstrap", "app_hooks.go"), `OnPostStartupMessage`)
					assertGeneratedFileContains(t, targetDir, filepath.Join("internal", "bootstrap", "app_hooks.go"), `OnPreShutdown`)
					assertGeneratedFileContains(t, targetDir, filepath.Join("internal", "bootstrap", "app_hooks.go"), `OnPostShutdown`)
				} else {
					assertGeneratedFileMissing(t, targetDir, filepath.Join("internal", "bootstrap", "app_hooks.go"))
					assertGeneratedFileNotContains(t, targetDir, filepath.Join("internal", "bootstrap", "serve.go"), `registerAppHooks(`)
				}
			}
			if tc.expectExtraLight {
				assertGeneratedFileMissing(t, targetDir, filepath.Join("pkg", "common", "constant.go"))
				assertGeneratedFileMissing(t, targetDir, filepath.Join("pkg", "common", "error.go"))
				assertGeneratedFileMissing(t, targetDir, filepath.Join("internal", "transport", "http", "router", "timeout_router.go"))
				assertGeneratedFileNotContains(t, targetDir, tc.routerPath, `AccessLog(`)
				assertGeneratedFileNotContains(t, targetDir, filepath.Join("internal", "transport", "http", "router", "url.go"), `HealthzPath`)
				if expectedFiberVersion(tc.fiberVersion) == stack.FiberV3 {
					assertGeneratedFileContains(t, targetDir, filepath.Join("internal", "bootstrap", "app_hooks.go"), `OnPreStartupMessage`)
				} else {
					assertGeneratedFileMissing(t, targetDir, filepath.Join("internal", "bootstrap", "app_hooks.go"))
				}
			}

			bootstrap := readGeneratedFile(t, targetDir, filepath.Join("internal", "bootstrap", "bootstrap.go"))
			if tc.expectRedis {
				if !strings.Contains(bootstrap, `"cache:redis"`) {
					t.Fatalf("expected redis injection in bootstrap, got:\n%s", bootstrap)
				}
			} else if strings.Contains(bootstrap, `"cache:redis"`) {
				t.Fatalf("did not expect redis injection in bootstrap, got:\n%s", bootstrap)
			}
			if tc.expectMedium {
				if !strings.Contains(bootstrap, `"docs:swagger"`) || !strings.Contains(bootstrap, `"ui:embedded"`) {
					t.Fatalf("expected default medium capabilities in bootstrap, got:\n%s", bootstrap)
				}
			}
			if tc.expectHeavy {
				if !strings.Contains(bootstrap, `"docs:swagger"`) || !strings.Contains(bootstrap, `"ui:embedded"`) {
					t.Fatalf("expected default heavy capabilities in bootstrap, got:\n%s", bootstrap)
				}
			}
			if tc.expectLight {
				if hasCapability(tc.capabilities, "swagger") {
					assertGeneratedFileContains(t, targetDir, filepath.Join("config", "server.yaml"), "swagger:\n  enabled: true")
				} else {
					assertGeneratedFileContains(t, targetDir, filepath.Join("config", "server.yaml"), "swagger:\n  enabled: false")
				}
				if hasCapability(tc.capabilities, "embedded-ui") {
					assertGeneratedFileContains(t, targetDir, filepath.Join("config", "server.yaml"), "embedded_ui:\n  enabled: true")
				} else {
					assertGeneratedFileContains(t, targetDir, filepath.Join("config", "server.yaml"), "embedded_ui:\n  enabled: false")
				}
				if hasCapability(tc.capabilities, "swagger") {
					if !strings.Contains(bootstrap, `"docs:swagger"`) {
						t.Fatalf("expected light swagger capability in bootstrap, got:\n%s", bootstrap)
					}
				} else if strings.Contains(bootstrap, `"docs:swagger"`) {
					t.Fatalf("did not expect light swagger capability by default, got:\n%s", bootstrap)
				}

				if hasCapability(tc.capabilities, "embedded-ui") {
					if !strings.Contains(bootstrap, `"ui:embedded"`) {
						t.Fatalf("expected light embedded-ui capability in bootstrap, got:\n%s", bootstrap)
					}
				} else if strings.Contains(bootstrap, `"ui:embedded"`) {
					t.Fatalf("did not expect light embedded-ui capability by default, got:\n%s", bootstrap)
				}
			}
			if tc.expectExtraLight {
				if strings.Contains(bootstrap, `"docs:swagger"`) || strings.Contains(bootstrap, `"ui:embedded"`) || strings.Contains(bootstrap, `"cache:redis"`) {
					t.Fatalf("did not expect extra-light optional services in bootstrap, got:\n%s", bootstrap)
				}
			}

			runGeneratedProjectTests(t, targetDir)
			if tc.expectMedium {
				runMediumBlackBoxScenario(t, targetDir, tc.expectRedis)
			}
			if tc.expectHeavy {
				runHeavyBlackBoxScenario(t, targetDir, tc.expectRedis)
			}
			if tc.expectLight {
				if (expectedFiberVersion(tc.fiberVersion) == stack.FiberV3 && expectedCLIStyle(tc.cliStyle) == stack.CLICobra) || (expectedFiberVersion(tc.fiberVersion) == stack.FiberV2 && expectedCLIStyle(tc.cliStyle) == stack.CLINative) || len(tc.capabilities) > 0 {
					runLightBlackBoxScenario(t, targetDir, hasCapability(tc.capabilities, "swagger"), hasCapability(tc.capabilities, "embedded-ui"))
				} else {
					runLightStartupSmokeScenario(t, targetDir)
				}
			}
			if tc.expectExtraLight {
				if (expectedFiberVersion(tc.fiberVersion) == stack.FiberV3 && expectedCLIStyle(tc.cliStyle) == stack.CLICobra) || (expectedFiberVersion(tc.fiberVersion) == stack.FiberV2 && expectedCLIStyle(tc.cliStyle) == stack.CLINative) {
					runExtraLightBlackBoxScenario(t, targetDir)
				} else {
					runExtraLightStartupSmokeScenario(t, targetDir)
				}
			}
		})
	}
}

func TestGeneratedConfigPreservesExplicitFalse(t *testing.T) {
	targetDir := t.TempDir()
	req := Request{
		ProjectName: "demo",
		ModulePath:  "github.com/example/demo",
		Preset:      "medium",
		Options:     requestOptionsForTest(targetDir, stack.FiberV3, stack.CLICobra),
	}

	if _, err := Run(req); err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	testFile := `package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExplicitFalsePreserved(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "server.yaml")
	body := []byte("server:\n  app_name: \"demo\"\nlog:\n  log_compress: false\n  log_show_line: false\nmiddleware:\n  timeout:\n    enabled: false\nswagger:\n  enabled: false\nembedded_ui:\n  enabled: false\n")
	if err := os.WriteFile(configPath, body, 0o644); err != nil {
		t.Fatalf("write config failed: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.Log.LogCompress {
		t.Fatalf("expected log_compress to remain false")
	}
	if cfg.Log.LogShowLine {
		t.Fatalf("expected log_show_line to remain false")
	}
	if cfg.Middleware.Timeout.Enabled {
		t.Fatalf("expected timeout.enabled to remain false")
	}
	if cfg.Swagger.Enabled {
		t.Fatalf("expected swagger.enabled to remain false")
	}
	if cfg.EmbeddedUI.Enabled {
		t.Fatalf("expected embedded_ui.enabled to remain false")
	}
}
`
	testPath := filepath.Join(targetDir, "config", "config_explicit_false_test.go")
	if err := os.WriteFile(testPath, []byte(testFile), 0o644); err != nil {
		t.Fatalf("write generated config test failed: %v", err)
	}

	cmd := exec.Command("go", "test", "./config", "-run", "TestExplicitFalsePreserved")
	cmd.Dir = targetDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("generated config test failed: %v\n%s", err, string(output))
	}
}

func TestGenerateRejectsUnsupportedCombinations(t *testing.T) {
	testCases := []struct {
		name         string
		preset       string
		capabilities []string
		want         string
	}{
		{name: "light with redis", preset: "light", capabilities: []string{"redis"}, want: `not allowed for preset "light"`},
		{name: "light with swagger redis", preset: "light", capabilities: []string{"swagger", "redis"}, want: `not allowed for preset "light"`},
		{name: "light with embedded-ui redis", preset: "light", capabilities: []string{"embedded-ui", "redis"}, want: `not allowed for preset "light"`},
		{name: "light with full capability set", preset: "light", capabilities: []string{"swagger", "embedded-ui", "redis"}, want: `not allowed for preset "light"`},
		{name: "extra-light with redis", preset: "extra-light", capabilities: []string{"redis"}, want: `not allowed for preset "extra-light"`},
		{name: "extra-light with embedded-ui", preset: "extra-light", capabilities: []string{"embedded-ui"}, want: `not allowed for preset "extra-light"`},
		{name: "extra-light with swagger", preset: "extra-light", capabilities: []string{"swagger"}, want: `not allowed for preset "extra-light"`},
		{name: "extra-light with swagger embedded-ui", preset: "extra-light", capabilities: []string{"swagger", "embedded-ui"}, want: `not allowed for preset "extra-light"`},
		{name: "extra-light with swagger redis", preset: "extra-light", capabilities: []string{"swagger", "redis"}, want: `not allowed for preset "extra-light"`},
		{name: "extra-light with embedded-ui redis", preset: "extra-light", capabilities: []string{"embedded-ui", "redis"}, want: `not allowed for preset "extra-light"`},
		{name: "extra-light with full capability set", preset: "extra-light", capabilities: []string{"swagger", "embedded-ui", "redis"}, want: `not allowed for preset "extra-light"`},
		{name: "invalid fiber version", preset: "medium", capabilities: []string{}, want: `fiber version "v9" is not supported`},
		{name: "invalid cli style", preset: "medium", capabilities: []string{}, want: `cli style "bash" is not supported`},
		{name: "invalid logger", preset: "medium", capabilities: []string{}, want: `logger "printf" is not supported`},
		{name: "invalid db", preset: "medium", capabilities: []string{}, want: `database "oracle" is not supported`},
		{name: "invalid data access", preset: "medium", capabilities: []string{}, want: `data access "gorm" is not supported`},
		{name: "extra-light logger unsupported", preset: "extra-light", capabilities: []string{}, want: `does not support logger option`},
		{name: "extra-light db unsupported", preset: "extra-light", capabilities: []string{}, want: `does not support db option`},
		{name: "extra-light data access unsupported", preset: "extra-light", capabilities: []string{}, want: `does not support data access option`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			options := map[string]string{
				"manifest_root": "../../generator",
				"target_dir":    t.TempDir(),
			}
			if tc.name == "invalid fiber version" {
				options[stack.OptionFiberVersion] = "v9"
			}
			if tc.name == "invalid cli style" {
				options[stack.OptionCLIStyle] = "bash"
			}
			if tc.name == "invalid logger" {
				options[stack.OptionLogger] = "printf"
			}
			if tc.name == "invalid db" {
				options[stack.OptionDB] = "oracle"
			}
			if tc.name == "invalid data access" {
				options[stack.OptionDataAccess] = "gorm"
			}
			if tc.name == "extra-light logger unsupported" {
				options[stack.OptionLogger] = "zap"
			}
			if tc.name == "extra-light db unsupported" {
				options[stack.OptionDB] = stack.DBPgSQL
			}
			if tc.name == "extra-light data access unsupported" {
				options[stack.OptionDataAccess] = stack.DataAccessSQLX
			}
			req := Request{
				ProjectName:  "demo",
				ModulePath:   "github.com/example/demo",
				Preset:       tc.preset,
				Capabilities: tc.capabilities,
				Options:      options,
			}

			err := Generate(req)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected error containing %q, got %v", tc.want, err)
			}
		})
	}
}

func TestRunSupportsPhase12CapabilityGenerationMatrix(t *testing.T) {
	testCases := []struct {
		name         string
		preset       string
		capabilities []string
	}{
		{name: "heavy default", preset: "heavy"},
		{name: "heavy redis", preset: "heavy", capabilities: []string{"redis"}},
		{name: "heavy swagger", preset: "heavy", capabilities: []string{"swagger"}},
		{name: "heavy embedded-ui", preset: "heavy", capabilities: []string{"embedded-ui"}},
		{name: "heavy swagger embedded-ui", preset: "heavy", capabilities: []string{"swagger", "embedded-ui"}},
		{name: "heavy swagger redis", preset: "heavy", capabilities: []string{"swagger", "redis"}},
		{name: "heavy embedded-ui redis", preset: "heavy", capabilities: []string{"embedded-ui", "redis"}},
		{name: "heavy full", preset: "heavy", capabilities: []string{"swagger", "embedded-ui", "redis"}},
		{name: "medium default", preset: "medium"},
		{name: "medium redis", preset: "medium", capabilities: []string{"redis"}},
		{name: "medium swagger", preset: "medium", capabilities: []string{"swagger"}},
		{name: "medium embedded-ui", preset: "medium", capabilities: []string{"embedded-ui"}},
		{name: "medium swagger embedded-ui", preset: "medium", capabilities: []string{"swagger", "embedded-ui"}},
		{name: "medium swagger redis", preset: "medium", capabilities: []string{"swagger", "redis"}},
		{name: "medium embedded-ui redis", preset: "medium", capabilities: []string{"embedded-ui", "redis"}},
		{name: "medium full", preset: "medium", capabilities: []string{"swagger", "embedded-ui", "redis"}},
		{name: "light default", preset: "light"},
		{name: "light swagger", preset: "light", capabilities: []string{"swagger"}},
		{name: "light embedded-ui", preset: "light", capabilities: []string{"embedded-ui"}},
		{name: "light swagger embedded-ui", preset: "light", capabilities: []string{"swagger", "embedded-ui"}},
		{name: "extra-light default", preset: "extra-light"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			targetDir := t.TempDir()
			summary, err := Run(Request{
				ProjectName:  "demo",
				ModulePath:   "github.com/example/demo",
				Preset:       tc.preset,
				Capabilities: tc.capabilities,
				Options:      requestOptionsForTest(targetDir, stack.FiberV3, stack.CLICobra),
			})
			if err != nil {
				t.Fatalf("Run() returned error: %v", err)
			}

			if summary.Preset != tc.preset {
				t.Fatalf("expected preset %q, got %q", tc.preset, summary.Preset)
			}

			assertPhase12CapabilityArtifacts(t, targetDir, tc.preset, tc.capabilities)
			runGeneratedProjectTests(t, targetDir)
		})
	}
}

func TestPhase12CapabilityMatrixBlackBox(t *testing.T) {
	testCases := []struct {
		name         string
		preset       string
		capabilities []string
		runScenario  func(t *testing.T, targetDir string)
	}{
		{
			name:        "light default",
			preset:      "light",
			runScenario: func(t *testing.T, targetDir string) { runLightBlackBoxScenario(t, targetDir, false, false) },
		},
		{
			name:         "light swagger",
			preset:       "light",
			capabilities: []string{"swagger"},
			runScenario:  func(t *testing.T, targetDir string) { runLightBlackBoxScenario(t, targetDir, true, false) },
		},
		{
			name:         "light embedded-ui",
			preset:       "light",
			capabilities: []string{"embedded-ui"},
			runScenario:  func(t *testing.T, targetDir string) { runLightBlackBoxScenario(t, targetDir, false, true) },
		},
		{
			name:         "light full",
			preset:       "light",
			capabilities: []string{"swagger", "embedded-ui"},
			runScenario:  func(t *testing.T, targetDir string) { runLightBlackBoxScenario(t, targetDir, true, true) },
		},
		{
			name:        "medium default",
			preset:      "medium",
			runScenario: func(t *testing.T, targetDir string) { runMediumBlackBoxScenario(t, targetDir, false) },
		},
		{
			name:         "medium redis",
			preset:       "medium",
			capabilities: []string{"redis"},
			runScenario:  func(t *testing.T, targetDir string) { runMediumBlackBoxScenario(t, targetDir, true) },
		},
		{
			name:         "medium full",
			preset:       "medium",
			capabilities: []string{"swagger", "embedded-ui", "redis"},
			runScenario:  func(t *testing.T, targetDir string) { runMediumBlackBoxScenario(t, targetDir, true) },
		},
		{
			name:        "heavy default",
			preset:      "heavy",
			runScenario: func(t *testing.T, targetDir string) { runHeavyBlackBoxScenario(t, targetDir, false) },
		},
		{
			name:         "heavy redis",
			preset:       "heavy",
			capabilities: []string{"redis"},
			runScenario:  func(t *testing.T, targetDir string) { runHeavyBlackBoxScenario(t, targetDir, true) },
		},
		{
			name:         "heavy full",
			preset:       "heavy",
			capabilities: []string{"swagger", "embedded-ui", "redis"},
			runScenario:  func(t *testing.T, targetDir string) { runHeavyBlackBoxScenario(t, targetDir, true) },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			targetDir := t.TempDir()
			_, err := Run(Request{
				ProjectName:  "demo",
				ModulePath:   "github.com/example/demo",
				Preset:       tc.preset,
				Capabilities: tc.capabilities,
				Options:      requestOptionsForTest(targetDir, stack.FiberV3, stack.CLICobra),
			})
			if err != nil {
				t.Fatalf("Run() returned error: %v", err)
			}

			tc.runScenario(t, targetDir)
		})
	}
}

func TestRunWritesPhase13MetadataManifest(t *testing.T) {
	targetDir := t.TempDir()
	summary, err := Run(Request{
		ProjectName: "demo",
		ModulePath:  "github.com/example/demo",
		Preset:      "medium",
		Options:     requestOptionsForTest(targetDir, stack.FiberV3, stack.CLICobra),
	})
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	projectManifest, err := metadata.LoadManifest(targetDir)
	if err != nil {
		t.Fatalf("LoadManifest() returned error: %v", err)
	}

	if projectManifest.SchemaVersion != metadata.SchemaVersionV1 {
		t.Fatalf("expected schema version %q, got %q", metadata.SchemaVersionV1, projectManifest.SchemaVersion)
	}
	if projectManifest.Recipe.Preset != "medium" || projectManifest.Recipe.ModulePath != "github.com/example/demo" {
		t.Fatalf("unexpected manifest recipe: %+v", projectManifest.Recipe)
	}
	if strings.Join(projectManifest.Recipe.Capabilities, ",") != "swagger,embedded-ui" {
		t.Fatalf("unexpected manifest capabilities: %+v", projectManifest.Recipe.Capabilities)
	}
	if projectManifest.Recipe.Logger != stack.DefaultLogger() || projectManifest.Recipe.DB != stack.DefaultDB() || projectManifest.Recipe.DataAccess != stack.DefaultDataAccess() || projectManifest.Recipe.JSONLib != stack.DefaultJSONLib() {
		t.Fatalf("unexpected manifest runtime recipe: %+v", projectManifest.Recipe)
	}
	if projectManifest.Assets.Base != "service-base-cobra" {
		t.Fatalf("unexpected manifest base asset: %+v", projectManifest.Assets)
	}
	if len(projectManifest.ManagedFiles) == 0 {
		t.Fatalf("expected managed files to be recorded")
	}
	if projectManifest.Fingerprints.TemplateSet == "" || projectManifest.Fingerprints.RenderedOutput == "" {
		t.Fatalf("expected manifest fingerprints to be populated: %+v", projectManifest.Fingerprints)
	}
	if summary.MetadataPath != ".fiberx/manifest.json" {
		t.Fatalf("expected metadata path summary, got %+v", summary)
	}
	if summary.TemplateSetFingerprint != projectManifest.Fingerprints.TemplateSet || summary.RenderedOutputFingerprint != projectManifest.Fingerprints.RenderedOutput {
		t.Fatalf("expected summary fingerprints to match manifest: summary=%+v manifest=%+v", summary, projectManifest.Fingerprints)
	}

	firstManaged := projectManifest.ManagedFiles[0]
	if strings.TrimSpace(firstManaged.Path) == "" || strings.TrimSpace(firstManaged.SHA256) == "" {
		t.Fatalf("expected managed file entry to include path and hash: %+v", firstManaged)
	}
}

func TestRunSupportsPhase11RuntimeSelections(t *testing.T) {
	testCases := []struct {
		name       string
		preset     string
		logger     string
		dbKind     string
		dataAccess string
	}{
		{name: "medium slog pgsql sqlx", preset: "medium", logger: stack.LoggerSlog, dbKind: stack.DBPgSQL, dataAccess: stack.DataAccessSQLX},
		{name: "light zap mysql sqlc", preset: "light", logger: stack.LoggerZap, dbKind: stack.DBMySQL, dataAccess: stack.DataAccessSQLC},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			targetDir := t.TempDir()
			options := requestOptionsForTest(targetDir, stack.FiberV3, stack.CLICobra)
			options[stack.OptionLogger] = tc.logger
			options[stack.OptionDB] = tc.dbKind
			options[stack.OptionDataAccess] = tc.dataAccess

			summary, err := Run(Request{
				ProjectName: "demo",
				ModulePath:  "github.com/example/demo",
				Preset:      tc.preset,
				Options:     options,
			})
			if err != nil {
				t.Fatalf("Run() returned error: %v", err)
			}

			if summary.Logger != tc.logger || summary.Database != tc.dbKind || summary.DataAccess != tc.dataAccess || summary.JSONLib != stack.DefaultJSONLib() {
				t.Fatalf("unexpected runtime summary: %+v", summary)
			}

			assertGeneratedFileContains(t, targetDir, "README.md", "logger: `"+tc.logger+"`")
			assertGeneratedFileContains(t, targetDir, "README.md", "json backend: `stdlib`")
			switch tc.dbKind {
			case stack.DBPgSQL:
				assertGeneratedFileContains(t, targetDir, filepath.Join("config", "server.yaml"), `db_type: "postgres"`)
			case stack.DBMySQL:
				assertGeneratedFileContains(t, targetDir, filepath.Join("config", "server.yaml"), `db_type: "mysql"`)
			}
			if tc.dataAccess == stack.DataAccessSQLC {
				assertGeneratedFileContains(t, targetDir, "sqlc.yaml", `engine: "mysql"`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("internal", "app", "user", "dao", "query.sql"), `-- name: CreateUser :one`)
			}

			runGeneratedProjectTests(t, targetDir)
		})
	}
}

func TestPhase11RuntimeMatrixDefaultStack(t *testing.T) {
	presets := []string{"medium", "heavy", "light"}
	databases := []string{stack.DBSQLite, stack.DBPgSQL, stack.DBMySQL}
	dataAccessStacks := []string{stack.DataAccessStdlib, stack.DataAccessSQLX, stack.DataAccessSQLC}

	for _, preset := range presets {
		for _, dbKind := range databases {
			for _, dataAccess := range dataAccessStacks {
				name := preset + " " + dbKind + " " + dataAccess
				t.Run(name, func(t *testing.T) {
					database, ok := lookupRuntimeDatabaseForMatrix(t, dbKind)
					if !ok {
						t.Skipf("runtime database for %s is not configured", dbKind)
					}

					targetDir := t.TempDir()
					options := requestOptionsForTest(targetDir, stack.FiberV3, stack.CLICobra)
					options[stack.OptionDB] = dbKind
					options[stack.OptionDataAccess] = dataAccess

					summary, err := Run(Request{
						ProjectName: "demo",
						ModulePath:  "github.com/example/demo",
						Preset:      preset,
						Options:     options,
					})
					if err != nil {
						t.Fatalf("Run() returned error: %v", err)
					}
					if summary.Logger != stack.DefaultLogger() {
						t.Fatalf("expected default logger %q, got %q", stack.DefaultLogger(), summary.Logger)
					}
					if summary.Database != dbKind {
						t.Fatalf("expected database %q, got %q", dbKind, summary.Database)
					}
					if summary.DataAccess != dataAccess {
						t.Fatalf("expected data access %q, got %q", dataAccess, summary.DataAccess)
					}
					if summary.JSONLib != stack.DefaultJSONLib() {
						t.Fatalf("expected json lib %q, got %q", stack.DefaultJSONLib(), summary.JSONLib)
					}
					assertPhase11RuntimeArtifacts(t, targetDir, dbKind, dataAccess)

					runGeneratedProjectTests(t, targetDir)

					switch preset {
					case "medium":
						runMediumBlackBoxScenarioWithDatabase(t, targetDir, false, database)
					case "heavy":
						runHeavyBlackBoxScenarioWithDatabase(t, targetDir, false, database)
					case "light":
						runLightBlackBoxScenarioWithDatabase(t, targetDir, false, false, database)
					default:
						t.Fatalf("unsupported preset %q", preset)
					}
				})
			}
		}
	}
}

func TestRunSupportsJSONLibSelections(t *testing.T) {
	testCases := []struct {
		name          string
		preset        string
		fiberVersion  string
		jsonLib       string
		wantImport    string
		wantEncoder   string
		wantDecoder   string
		expectHooksV3 bool
	}{
		{
			name:          "medium v3 sonic",
			preset:        "medium",
			fiberVersion:  stack.FiberV3,
			jsonLib:       stack.JSONLibSonic,
			wantImport:    `"github.com/bytedance/sonic"`,
			wantEncoder:   `sonic.Marshal`,
			wantDecoder:   `sonic.Unmarshal`,
			expectHooksV3: true,
		},
		{
			name:          "light v2 go-json",
			preset:        "light",
			fiberVersion:  stack.FiberV2,
			jsonLib:       stack.JSONLibGoJSON,
			wantImport:    `json "github.com/goccy/go-json"`,
			wantEncoder:   `json.Marshal`,
			wantDecoder:   `json.Unmarshal`,
			expectHooksV3: false,
		},
		{
			name:          "extra-light v3 stdlib",
			preset:        "extra-light",
			fiberVersion:  stack.FiberV3,
			jsonLib:       stack.JSONLibStdlib,
			wantImport:    `"encoding/json"`,
			wantEncoder:   `json.Marshal`,
			wantDecoder:   `json.Unmarshal`,
			expectHooksV3: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			targetDir := t.TempDir()
			options := requestOptionsForTest(targetDir, tc.fiberVersion, stack.CLICobra)
			options[stack.OptionJSONLib] = tc.jsonLib

			summary, err := Run(Request{
				ProjectName: "demo",
				ModulePath:  "github.com/example/demo",
				Preset:      tc.preset,
				Options:     options,
			})
			if err != nil {
				t.Fatalf("Run() returned error: %v", err)
			}
			if summary.JSONLib != tc.jsonLib {
				t.Fatalf("expected json lib %q, got %q", tc.jsonLib, summary.JSONLib)
			}

			routerPath := filepath.Join("internal", "transport", "http", "router", "router.go")
			assertGeneratedFileContains(t, targetDir, routerPath, tc.wantImport)
			assertGeneratedFileContains(t, targetDir, routerPath, tc.wantEncoder)
			assertGeneratedFileContains(t, targetDir, routerPath, tc.wantDecoder)
			assertGeneratedFileContains(t, targetDir, "README.md", "json backend: `"+tc.jsonLib+"`")

			projectManifest, err := metadata.LoadManifest(targetDir)
			if err != nil {
				t.Fatalf("LoadManifest() returned error: %v", err)
			}
			if projectManifest.Recipe.JSONLib != tc.jsonLib {
				t.Fatalf("expected manifest json lib %q, got %q", tc.jsonLib, projectManifest.Recipe.JSONLib)
			}

			hooksPath := filepath.Join("internal", "bootstrap", "app_hooks.go")
			if tc.expectHooksV3 {
				assertGeneratedFileContains(t, targetDir, hooksPath, `OnPreStartupMessage`)
			} else {
				assertGeneratedFileMissing(t, targetDir, hooksPath)
			}

			runGeneratedProjectTests(t, targetDir)
		})
	}
}

func assertPhase11RuntimeArtifacts(t *testing.T, targetDir, dbKind, dataAccess string) {
	t.Helper()

	switch dbKind {
	case stack.DBSQLite:
		assertGeneratedFileContains(t, targetDir, filepath.Join("config", "server.yaml"), `db_type: "sqlite"`)
	case stack.DBPgSQL:
		assertGeneratedFileContains(t, targetDir, filepath.Join("config", "server.yaml"), `db_type: "postgres"`)
	case stack.DBMySQL:
		assertGeneratedFileContains(t, targetDir, filepath.Join("config", "server.yaml"), `db_type: "mysql"`)
	}

	if dataAccess == stack.DataAccessSQLC {
		switch dbKind {
		case stack.DBSQLite:
			assertGeneratedFileContains(t, targetDir, "sqlc.yaml", `engine: "sqlite"`)
		case stack.DBPgSQL:
			assertGeneratedFileContains(t, targetDir, "sqlc.yaml", `engine: "postgres"`)
		case stack.DBMySQL:
			assertGeneratedFileContains(t, targetDir, "sqlc.yaml", `engine: "mysql"`)
		}
		assertGeneratedFileContains(t, targetDir, filepath.Join("internal", "app", "user", "dao", "query.sql"), `-- name: CreateUser :one`)
		return
	}

	assertGeneratedFileMissing(t, targetDir, "sqlc.yaml")
}

func assertPhase12CapabilityArtifacts(t *testing.T, targetDir, preset string, capabilities []string) {
	t.Helper()

	expectSwagger, expectEmbeddedUI, expectRedis := expectedCapabilityMaterialization(preset, capabilities)
	bootstrap := readGeneratedFile(t, targetDir, filepath.Join("internal", "bootstrap", "bootstrap.go"))
	config := normalizeGeneratedText(readGeneratedFile(t, targetDir, filepath.Join("config", "server.yaml")))

	if expectSwagger {
		assertGeneratedFileContains(t, targetDir, filepath.Join("docs", "openapi.yaml"), "openapi: 3.0.3")
		if !strings.Contains(bootstrap, `"docs:swagger"`) {
			t.Fatalf("expected swagger service registration in bootstrap, got:\n%s", bootstrap)
		}
	} else {
		assertGeneratedFileMissing(t, targetDir, filepath.Join("docs", "openapi.yaml"))
		if strings.Contains(bootstrap, `"docs:swagger"`) {
			t.Fatalf("did not expect swagger service registration in bootstrap, got:\n%s", bootstrap)
		}
	}

	if expectEmbeddedUI {
		assertGeneratedFileContains(t, targetDir, filepath.Join("internal", "transport", "http", "webui", "dist", "index.html"), "embedded UI ships")
		if !strings.Contains(bootstrap, `"ui:embedded"`) {
			t.Fatalf("expected embedded-ui service registration in bootstrap, got:\n%s", bootstrap)
		}
	} else {
		assertGeneratedFileMissing(t, targetDir, filepath.Join("internal", "transport", "http", "webui", "dist", "index.html"))
		if strings.Contains(bootstrap, `"ui:embedded"`) {
			t.Fatalf("did not expect embedded-ui service registration in bootstrap, got:\n%s", bootstrap)
		}
	}

	if expectRedis {
		assertGeneratedFileContains(t, targetDir, filepath.Join("internal", "infra", "cache", "redis.go"), "github.com/redis/go-redis/v9")
		if !strings.Contains(bootstrap, `"cache:redis"`) {
			t.Fatalf("expected redis service registration in bootstrap, got:\n%s", bootstrap)
		}
	} else {
		assertGeneratedFileMissing(t, targetDir, filepath.Join("internal", "infra", "cache", "redis.go"))
		if strings.Contains(bootstrap, `"cache:redis"`) {
			t.Fatalf("did not expect redis service registration in bootstrap, got:\n%s", bootstrap)
		}
	}

	switch preset {
	case "heavy", "medium", "light":
		if !strings.Contains(config, "swagger:\n  enabled: "+strconv.FormatBool(expectSwagger)) {
			t.Fatalf("expected swagger enabled=%t in config, got:\n%s", expectSwagger, config)
		}
		if !strings.Contains(config, "embedded_ui:\n  enabled: "+strconv.FormatBool(expectEmbeddedUI)) {
			t.Fatalf("expected embedded_ui enabled=%t in config, got:\n%s", expectEmbeddedUI, config)
		}
	case "extra-light":
		assertGeneratedFileMissing(t, targetDir, filepath.Join("docs", "openapi.yaml"))
		assertGeneratedFileMissing(t, targetDir, filepath.Join("internal", "transport", "http", "webui", "dist", "index.html"))
	}

	switch preset {
	case "heavy", "medium":
		if !strings.Contains(config, "redis:\n  enabled: "+strconv.FormatBool(expectRedis)) {
			t.Fatalf("expected redis enabled=%t in config, got:\n%s", expectRedis, config)
		}
	case "light", "extra-light":
		if strings.Contains(config, "redis:\n  enabled: true") {
			t.Fatalf("did not expect redis to be enabled in %s config", preset)
		}
	}
}

func expectedCapabilityMaterialization(preset string, capabilities []string) (bool, bool, bool) {
	switch preset {
	case "heavy", "medium":
		return true, true, hasCapability(capabilities, "redis")
	case "light":
		return hasCapability(capabilities, "swagger"), hasCapability(capabilities, "embedded-ui"), false
	case "extra-light":
		return false, false, false
	default:
		return false, false, false
	}
}

func assertGeneratedFileContains(t *testing.T, targetDir string, relativePath string, want string) {
	t.Helper()

	content := normalizeGeneratedText(readGeneratedFile(t, targetDir, relativePath))
	if !strings.Contains(content, normalizeGeneratedText(want)) {
		t.Fatalf("expected %s to contain %q, got:\n%s", relativePath, want, content)
	}
}

func assertGeneratedFileNotContains(t *testing.T, targetDir string, relativePath string, want string) {
	t.Helper()

	content := normalizeGeneratedText(readGeneratedFile(t, targetDir, relativePath))
	if strings.Contains(content, normalizeGeneratedText(want)) {
		t.Fatalf("expected %s to not contain %q, got:\n%s", relativePath, want, content)
	}
}

func readGeneratedFile(t *testing.T, targetDir string, relativePath string) string {
	t.Helper()

	path := filepath.Join(targetDir, relativePath)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read generated file %q: %v", path, err)
	}
	return string(data)
}

func normalizeGeneratedText(content string) string {
	return strings.ReplaceAll(content, "\r\n", "\n")
}

func assertGeneratedFileMissing(t *testing.T, targetDir string, relativePath string) {
	t.Helper()

	path := filepath.Join(targetDir, relativePath)
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("expected generated file %q to be absent", path)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat generated file %q: %v", path, err)
	}
}

func requestOptionsForTest(targetDir, fiberVersion, cliStyle string) map[string]string {
	options := map[string]string{
		"command":       "new",
		"manifest_root": "../../generator",
		"target_dir":    targetDir,
	}
	if fiberVersion != "" {
		options[stack.OptionFiberVersion] = fiberVersion
	}
	if cliStyle != "" {
		options[stack.OptionCLIStyle] = cliStyle
	}
	return options
}

func expectedFiberVersion(raw string) string {
	if raw == "" {
		return stack.DefaultFiberVersion()
	}
	return raw
}

func expectedCLIStyle(raw string) string {
	if raw == "" {
		return stack.DefaultCLIStyle()
	}
	return raw
}

func expectedBaseName(cliStyle string) string {
	if expectedCLIStyle(cliStyle) == stack.CLICobra {
		return "service-base-cobra"
	}
	return "service-base"
}

func expectedFiberDependency(fiberVersion string) string {
	if expectedFiberVersion(fiberVersion) == stack.FiberV2 {
		return "github.com/gofiber/fiber/v2 v2."
	}
	return "github.com/gofiber/fiber/v3 v3."
}

func hasCapability(capabilities []string, want string) bool {
	for _, capability := range capabilities {
		if capability == want {
			return true
		}
	}
	return false
}

func runGeneratedProjectTests(t *testing.T, targetDir string) {
	t.Helper()

	cmd := exec.Command("go", "test", "./...")
	cmd.Dir = targetDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("generated project go test failed: %v\n%s", err, string(output))
	}
}

func runLightStartupSmokeScenario(t *testing.T, targetDir string) {
	t.Helper()

	port := randomPort(t)
	tempDir := t.TempDir()
	databasePath := filepath.Join(tempDir, "data", "app.db")
	logPath := filepath.Join(tempDir, "logs", "app.log")
	if err := os.MkdirAll(filepath.Dir(databasePath), 0o755); err != nil {
		t.Fatalf("create database dir failed: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		t.Fatalf("create log dir failed: %v", err)
	}

	configBody := `server:
  app_id: "fiberx"
  app_name: "demo"
  app_version: "v1.0.0"
  mode: "debug"
  ip_address: "127.0.0.1"
  port: "` + port + `"
database:
  enabled: true
  auto_migrate: true
  db_type: "sqlite"
  sqlite:
    path: "` + filepath.ToSlash(databasePath) + `"
log:
  log_level: "debug"
  log_mode: "text"
  log_path: "` + filepath.ToSlash(logPath) + `"
middleware:
  cors:
    allow_origins:
      - "*"
  gzip:
    enabled: true
swagger:
  enabled: false
  route_prefix: "/docs"
embedded_ui:
  enabled: false
  route_prefix: "/ui"
`
	configPath := filepath.Join(tempDir, "server.yaml")
	if err := os.WriteFile(configPath, []byte(configBody), 0o644); err != nil {
		t.Fatalf("write config failed: %v", err)
	}

	binaryPath := buildBinary(t, targetDir)
	cmd := exec.Command(binaryPath, "serve", "--config", configPath)
	cmd.Dir = targetDir
	var output strings.Builder
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := cmd.Start(); err != nil {
		t.Fatalf("start light smoke service failed: %v", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	}()

	baseURL := "http://127.0.0.1:" + port
	waitForReady(t, baseURL+"/healthz", &output, cmd)
	health := doJSONRequest(t, "GET", baseURL+"/healthz", nil, nil)
	if health.StatusCode != 200 || health.Code != 1 {
		t.Fatalf("unexpected light smoke health response: %+v", health)
	}
}

func runExtraLightStartupSmokeScenario(t *testing.T, targetDir string) {
	t.Helper()

	port := randomPort(t)
	tempDir := t.TempDir()
	databasePath := filepath.Join(tempDir, "data", "app.db")
	logPath := filepath.Join(tempDir, "logs", "app.log")
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		t.Fatalf("create log dir failed: %v", err)
	}

	configBody := `server:
  app_name: "demo"
  mode: "debug"
  ip_address: "127.0.0.1"
  port: "` + port + `"
database:
  path: "` + filepath.ToSlash(databasePath) + `"
log:
  log_level: "debug"
  log_path: "` + filepath.ToSlash(logPath) + `"
`
	configPath := filepath.Join(tempDir, "server.yaml")
	if err := os.WriteFile(configPath, []byte(configBody), 0o644); err != nil {
		t.Fatalf("write config failed: %v", err)
	}

	binaryPath := buildBinary(t, targetDir)
	cmd := exec.Command(binaryPath, "serve", "--config", configPath)
	cmd.Dir = targetDir
	var output strings.Builder
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := cmd.Start(); err != nil {
		t.Fatalf("start extra-light smoke service failed: %v", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	}()

	baseURL := "http://127.0.0.1:" + port
	waitForReady(t, baseURL+"/healthz", &output, cmd)
	health := doJSONRequest(t, "GET", baseURL+"/healthz", nil, nil)
	if health.StatusCode != 200 || health.Code != 1 {
		t.Fatalf("unexpected extra-light smoke health response: %+v", health)
	}
}

type runtimeDatabaseConfig struct {
	Kind       string
	SQLitePath string
	DSN        string
}

func newSQLiteRuntimeDatabaseConfig(t *testing.T) runtimeDatabaseConfig {
	t.Helper()

	tempDir := t.TempDir()
	databasePath := filepath.Join(tempDir, "data", "app.db")
	if err := os.MkdirAll(filepath.Dir(databasePath), 0o755); err != nil {
		t.Fatalf("create database dir failed: %v", err)
	}

	return runtimeDatabaseConfig{
		Kind:       "sqlite",
		SQLitePath: databasePath,
	}
}

func lookupExternalRuntimeDatabase(dbKind string) (runtimeDatabaseConfig, bool) {
	switch dbKind {
	case stack.DBPgSQL:
		dsn := strings.TrimSpace(os.Getenv("FIBERX_TEST_PGSQL_DSN"))
		if dsn == "" {
			return runtimeDatabaseConfig{}, false
		}
		return runtimeDatabaseConfig{Kind: "postgres", DSN: dsn}, true
	case stack.DBMySQL:
		dsn := strings.TrimSpace(os.Getenv("FIBERX_TEST_MYSQL_DSN"))
		if dsn == "" {
			return runtimeDatabaseConfig{}, false
		}
		return runtimeDatabaseConfig{Kind: "mysql", DSN: dsn}, true
	default:
		return runtimeDatabaseConfig{}, false
	}
}

func lookupRuntimeDatabaseForMatrix(t *testing.T, dbKind string) (runtimeDatabaseConfig, bool) {
	t.Helper()

	if dbKind == stack.DBSQLite {
		return newSQLiteRuntimeDatabaseConfig(t), true
	}
	return lookupExternalRuntimeDatabase(dbKind)
}

func renderDatabaseConfigBlock(cfg runtimeDatabaseConfig) string {
	switch cfg.Kind {
	case "sqlite":
		return `database:
  enabled: true
  auto_migrate: true
  db_type: "sqlite"
  sqlite:
    path: ` + yamlSingleQuoted(filepath.ToSlash(cfg.SQLitePath)) + `
`
	case "postgres":
		return `database:
  enabled: true
  auto_migrate: true
  db_type: "postgres"
  postgres:
    dsn: ` + yamlSingleQuoted(cfg.DSN) + `
`
	case "mysql":
		return `database:
  enabled: true
  auto_migrate: true
  db_type: "mysql"
  mysql:
    dsn: ` + yamlSingleQuoted(cfg.DSN) + `
`
	default:
		panic("unsupported runtime database kind: " + cfg.Kind)
	}
}

func yamlSingleQuoted(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func runLightBlackBoxScenario(t *testing.T, targetDir string, enableSwagger bool, enableEmbeddedUI bool) {
	t.Helper()

	runLightBlackBoxScenarioWithDatabase(t, targetDir, enableSwagger, enableEmbeddedUI, newSQLiteRuntimeDatabaseConfig(t))
}

func runLightBlackBoxScenarioWithDatabase(t *testing.T, targetDir string, enableSwagger bool, enableEmbeddedUI bool, database runtimeDatabaseConfig) {
	t.Helper()

	port := randomPort(t)
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "logs", "app.log")
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		t.Fatalf("create log dir failed: %v", err)
	}

	configBody := `server:
  app_id: "fiberx"
  app_name: "demo"
  app_version: "v1.0.0"
  mode: "debug"
  ip_address: "127.0.0.1"
  port: "` + port + `"
` + renderDatabaseConfigBlock(database) + `log:
  log_level: "debug"
  log_mode: "text"
  log_path: ` + yamlSingleQuoted(filepath.ToSlash(logPath)) + `
middleware:
  cors:
    allow_origins:
      - "*"
  gzip:
    enabled: true
swagger:
  enabled: ` + strconv.FormatBool(enableSwagger) + `
  route_prefix: "/docs"
embedded_ui:
  enabled: ` + strconv.FormatBool(enableEmbeddedUI) + `
  route_prefix: "/ui"
`
	baseURL, _, _ := startGeneratedService(t, targetDir, configBody)

	health := doJSONRequest(t, "GET", baseURL+"/healthz", nil, nil)
	if health.StatusCode != 200 || health.Code != 1 {
		t.Fatalf("unexpected health response: %+v", health)
	}
	if health.Headers.Get("X-Request-ID") == "" || health.Headers.Get("X-Content-Type-Options") == "" {
		t.Fatalf("expected request-id and security headers on light health response")
	}
	if !healthBodyContainsService(t, health.Data, "http:light") {
		t.Fatalf("expected light service to appear in health payload: %s", string(health.Data))
	}
	if enableSwagger && !healthBodyContainsService(t, health.Data, "docs:swagger") {
		t.Fatalf("expected swagger service to appear in light health payload: %s", string(health.Data))
	}
	if !enableSwagger && healthBodyContainsService(t, health.Data, "docs:swagger") {
		t.Fatalf("did not expect swagger service in default light health payload: %s", string(health.Data))
	}
	if enableEmbeddedUI && !healthBodyContainsService(t, health.Data, "ui:embedded") {
		t.Fatalf("expected embedded-ui service to appear in light health payload: %s", string(health.Data))
	}
	if !enableEmbeddedUI && healthBodyContainsService(t, health.Data, "ui:embedded") {
		t.Fatalf("did not expect embedded-ui service in default light health payload: %s", string(health.Data))
	}

	for _, path := range []string{"/livez", "/readyz", "/startupz"} {
		resp := doJSONRequest(t, "GET", baseURL+path, nil, nil)
		if resp.StatusCode != 200 {
			t.Fatalf("expected %s to return 200, got %+v", path, resp)
		}
	}

	if enableSwagger {
		assertDocsRoute(t, baseURL)
	} else {
		assertRouteMissing(t, baseURL+"/docs/openapi.yaml")
	}
	if enableEmbeddedUI {
		assertUIRoute(t, baseURL)
	} else {
		assertRouteMissing(t, baseURL+"/ui")
	}
	assertRouteMissing(t, baseURL+"/metrics")

	runUserCRUDScenario(t, baseURL, database)
}

func runExtraLightBlackBoxScenario(t *testing.T, targetDir string) {
	t.Helper()

	port := randomPort(t)
	tempDir := t.TempDir()
	databasePath := filepath.Join(tempDir, "data", "app.db")
	logPath := filepath.Join(tempDir, "logs", "app.log")
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		t.Fatalf("create log dir failed: %v", err)
	}

	configBody := `server:
  app_name: "demo"
  mode: "debug"
  ip_address: "127.0.0.1"
  port: "` + port + `"
database:
  path: "` + filepath.ToSlash(databasePath) + `"
log:
  log_level: "debug"
  log_path: "` + filepath.ToSlash(logPath) + `"
`
	configPath := filepath.Join(tempDir, "server.yaml")
	if err := os.WriteFile(configPath, []byte(configBody), 0o644); err != nil {
		t.Fatalf("write config failed: %v", err)
	}

	binaryPath := buildBinary(t, targetDir)
	cmd := exec.Command(binaryPath, "serve", "--config", configPath)
	cmd.Dir = targetDir
	var output strings.Builder
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := cmd.Start(); err != nil {
		t.Fatalf("start extra-light service failed: %v", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	}()

	baseURL := "http://127.0.0.1:" + port
	waitForReady(t, baseURL+"/healthz", &output, cmd)

	health := doJSONRequest(t, "GET", baseURL+"/healthz", nil, nil)
	if health.StatusCode != 200 || health.Code != 1 {
		t.Fatalf("unexpected extra-light health response: %+v", health)
	}
	if health.Headers.Get("X-Request-ID") != "" || health.Headers.Get("X-Content-Type-Options") != "" {
		t.Fatalf("did not expect request-id or security headers on extra-light health response")
	}
	payload := decodeExtraLightHealthData(t, health.Data)
	if !payload.Database.Ready {
		t.Fatalf("expected extra-light database readiness payload, got %+v", payload)
	}
	if !containsService(payload.Services, "http:extra-light") {
		t.Fatalf("expected extra-light service in health payload, got %+v", payload)
	}

	for _, path := range []string{"/livez", "/readyz", "/startupz"} {
		resp := doJSONRequest(t, "GET", baseURL+path, nil, nil)
		if resp.StatusCode != 200 {
			t.Fatalf("expected %s to return 200, got %+v", path, resp)
		}
	}

	assertRouteMissing(t, baseURL+"/docs/openapi.yaml")
	assertRouteMissing(t, baseURL+"/ui")
	assertRouteMissing(t, baseURL+"/api/v1/user/")
	assertRouteMissing(t, baseURL+"/metrics")

	if _, err := os.Stat(databasePath); err != nil {
		t.Fatalf("expected sqlite database at %s: %v", databasePath, err)
	}
}

func runMediumBlackBoxScenario(t *testing.T, targetDir string, enableRedis bool) {
	t.Helper()

	runMediumBlackBoxScenarioWithDatabase(t, targetDir, enableRedis, newSQLiteRuntimeDatabaseConfig(t))
}

func runMediumBlackBoxScenarioWithDatabase(t *testing.T, targetDir string, enableRedis bool, database runtimeDatabaseConfig) {
	t.Helper()

	port := randomPort(t)
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "logs", "app.log")
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		t.Fatalf("create log dir failed: %v", err)
	}

	configBody := `server:
  app_id: "fiberx"
  app_name: "demo"
  app_version: "v1.0.0"
  mode: "debug"
  ip_address: "127.0.0.1"
  port: "` + port + `"
` + renderDatabaseConfigBlock(database) + `log:
  log_level: "debug"
  log_mode: "text"
  log_path: ` + yamlSingleQuoted(filepath.ToSlash(logPath)) + `
middleware:
  cors:
    allow_origins:
      - "*"
  gzip:
    enabled: true
swagger:
  enabled: true
  route_prefix: "/docs"
embedded_ui:
  enabled: true
  route_prefix: "/ui"
redis:
  enabled: ` + strconv.FormatBool(enableRedis) + `
  addr: "127.0.0.1:0"
  password: ""
  db: 0
`
	baseURL, _, _ := startGeneratedService(t, targetDir, configBody)

	health := doJSONRequest(t, "GET", baseURL+"/healthz", nil, nil)
	if health.StatusCode != 200 || health.Code != 1 {
		t.Fatalf("unexpected health response: %+v", health)
	}
	if health.Headers.Get("X-Request-ID") == "" || health.Headers.Get("X-Content-Type-Options") == "" {
		t.Fatalf("expected request-id and security headers on health response")
	}
	if enableRedis && !healthBodyContainsService(t, health.Data, "cache:redis") {
		t.Fatalf("expected redis service to appear in health payload: %s", string(health.Data))
	}

	for _, path := range []string{"/livez", "/readyz", "/startupz"} {
		resp := doJSONRequest(t, "GET", baseURL+path, nil, nil)
		if resp.StatusCode != 200 {
			t.Fatalf("expected %s to return 200, got %+v", path, resp)
		}
	}

	assertDocsRoute(t, baseURL)
	assertUIRoute(t, baseURL)
	runUserCRUDScenario(t, baseURL, database)
}

func runHeavyBlackBoxScenario(t *testing.T, targetDir string, enableRedis bool) {
	t.Helper()

	runHeavyBlackBoxScenarioWithDatabase(t, targetDir, enableRedis, newSQLiteRuntimeDatabaseConfig(t))
}

func runHeavyBlackBoxScenarioWithDatabase(t *testing.T, targetDir string, enableRedis bool, database runtimeDatabaseConfig) {
	t.Helper()

	port := randomPort(t)
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "logs", "app.log")
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		t.Fatalf("create log dir failed: %v", err)
	}

	configBody := `server:
  app_id: "fiberx"
  app_name: "demo"
  app_version: "v1.0.0"
  mode: "debug"
  ip_address: "127.0.0.1"
  port: "` + port + `"
` + renderDatabaseConfigBlock(database) + `log:
  log_level: "debug"
  log_mode: "text"
  log_path: ` + yamlSingleQuoted(filepath.ToSlash(logPath)) + `
middleware:
  cors:
    allow_origins:
      - "*"
  gzip:
    enabled: true
swagger:
  enabled: true
  route_prefix: "/docs"
embedded_ui:
  enabled: true
  route_prefix: "/ui"
metrics:
  enabled: true
  route_prefix: "/metrics"
scheduler:
  enabled: true
  interval: "100ms"
redis:
  enabled: ` + strconv.FormatBool(enableRedis) + `
  addr: "127.0.0.1:0"
  password: ""
  db: 0
`
	baseURL, _, output := startGeneratedService(t, targetDir, configBody)

	health := doJSONRequest(t, "GET", baseURL+"/healthz", nil, nil)
	if health.StatusCode != 200 || health.Code != 1 {
		t.Fatalf("unexpected health response: %+v", health)
	}
	if health.Headers.Get("X-Request-ID") == "" || health.Headers.Get("X-Content-Type-Options") == "" {
		t.Fatalf("expected request-id and security headers on health response")
	}

	healthData := decodeHeavyHealthData(t, health.Data)
	if !containsService(healthData.Services, "metrics:http") || !containsService(healthData.Services, "jobs:scheduler") {
		t.Fatalf("expected metrics and jobs services in health payload: %+v", healthData)
	}
	if !containsService(healthData.Services, "docs:swagger") || !containsService(healthData.Services, "ui:embedded") {
		t.Fatalf("expected default heavy capability services in health payload: %+v", healthData)
	}
	if enableRedis && !containsService(healthData.Services, "cache:redis") {
		t.Fatalf("expected redis service to appear in health payload: %+v", healthData)
	}

	for _, path := range []string{"/livez", "/readyz", "/startupz"} {
		resp := doJSONRequest(t, "GET", baseURL+path, nil, nil)
		if resp.StatusCode != 200 {
			t.Fatalf("expected %s to return 200, got %+v", path, resp)
		}
	}

	waitForHeavyJobRuns(t, baseURL)
	assertDocsRoute(t, baseURL)
	assertUIRoute(t, baseURL)
	assertMetricsRoute(t, baseURL)
	runUserCRUDScenario(t, baseURL, database)
	if database.Kind != "sqlite" && strings.TrimSpace(output.String()) == "" {
		t.Log("heavy external-db scenario completed without startup stderr/stdout noise")
	}
}

type heavyHealthData struct {
	Services       []string `json:"services"`
	RequestsTotal  uint64   `json:"requests_total"`
	JobRunsTotal   int64    `json:"job_runs_total"`
	LastJobRunUnix int64    `json:"last_job_run_unix"`
}

type extraLightHealthData struct {
	Services []string `json:"services"`
	Database struct {
		Ready bool `json:"ready"`
	} `json:"database"`
}

func decodeHeavyHealthData(t *testing.T, raw json.RawMessage) heavyHealthData {
	t.Helper()

	var payload heavyHealthData
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("decode heavy health payload failed: %v", err)
	}
	return payload
}

func decodeExtraLightHealthData(t *testing.T, raw json.RawMessage) extraLightHealthData {
	t.Helper()

	var payload extraLightHealthData
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("decode extra-light health payload failed: %v", err)
	}
	return payload
}

func waitForHeavyJobRuns(t *testing.T, baseURL string) {
	t.Helper()

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		health := doJSONRequest(t, "GET", baseURL+"/healthz", nil, nil)
		payload := decodeHeavyHealthData(t, health.Data)
		if payload.JobRunsTotal > 0 && payload.LastJobRunUnix > 0 {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}

	t.Fatalf("expected heavy scheduler job to run within deadline")
}

func assertDocsRoute(t *testing.T, baseURL string) {
	t.Helper()

	docsResp, err := http.Get(baseURL + "/docs/openapi.yaml")
	if err != nil {
		t.Fatalf("fetch docs failed: %v", err)
	}
	body, _ := io.ReadAll(docsResp.Body)
	_ = docsResp.Body.Close()
	if docsResp.StatusCode != 200 || !strings.Contains(string(body), "openapi: 3.0.3") {
		t.Fatalf("unexpected docs response: status=%d body=%s", docsResp.StatusCode, string(body))
	}
}

func assertUIRoute(t *testing.T, baseURL string) {
	t.Helper()

	uiResp, err := http.Get(baseURL + "/ui")
	if err != nil {
		t.Fatalf("fetch ui failed: %v", err)
	}
	uiBody, _ := io.ReadAll(uiResp.Body)
	_ = uiResp.Body.Close()
	if uiResp.StatusCode != 200 || !strings.Contains(string(uiBody), "embedded UI ships") {
		t.Fatalf("unexpected ui response: status=%d body=%s", uiResp.StatusCode, string(uiBody))
	}
}

func assertMetricsRoute(t *testing.T, baseURL string) {
	t.Helper()

	metricsResp, err := http.Get(baseURL + "/metrics")
	if err != nil {
		t.Fatalf("fetch metrics failed: %v", err)
	}
	body, _ := io.ReadAll(metricsResp.Body)
	_ = metricsResp.Body.Close()
	metricsBody := string(body)
	if metricsResp.StatusCode != 200 || !strings.Contains(metricsBody, "fiberx_requests_total") || !strings.Contains(metricsBody, "fiberx_job_runs_total") {
		t.Fatalf("unexpected metrics response: status=%d body=%s", metricsResp.StatusCode, metricsBody)
	}
}

func assertRouteMissing(t *testing.T, url string) {
	t.Helper()

	resp, err := rawRequest("GET", url, nil, nil)
	if err != nil {
		t.Fatalf("fetch missing route failed: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected missing route %s to return 404, got status=%d body=%s", url, resp.StatusCode, string(body))
	}
	if !strings.Contains(strings.ToLower(resp.Header.Get("Content-Type")), "application/json") {
		t.Fatalf("expected missing route %s to return json envelope, got content-type=%q body=%s", url, resp.Header.Get("Content-Type"), string(body))
	}
	var decoded apiResponse
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("decode missing route response failed: %v body=%s", err, string(body))
	}
	message := strings.ToLower(decoded.Message)
	if decoded.Code != 0 || (!strings.Contains(message, "not found") && !strings.Contains(message, "cannot get")) {
		t.Fatalf("expected missing route %s to use error envelope, got %+v", url, decoded)
	}
}

func runUserCRUDScenario(t *testing.T, baseURL string, database runtimeDatabaseConfig) {
	t.Helper()

	suffix := strconv.FormatInt(time.Now().UnixNano(), 36)
	createPayload := map[string]any{
		"name":   "Alice",
		"email":  "alice-" + suffix + "@example.com",
		"age":    28,
		"status": "active",
	}
	create := doJSONRequest(t, "POST", baseURL+"/api/v1/user/", createPayload, nil)
	if create.StatusCode != 200 || create.Code != 1 {
		t.Fatalf("create user failed: %+v", create)
	}
	var created struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(create.Data, &created); err != nil {
		t.Fatalf("decode created user failed: %v", err)
	}

	for index := 0; index < 8; index++ {
		payload := map[string]any{
			"name":   "User " + strconv.Itoa(index),
			"email":  "user" + strconv.Itoa(index) + "-" + suffix + "@example.com",
			"age":    20 + index,
			"status": "active",
		}
		resp := doJSONRequest(t, "POST", baseURL+"/api/v1/user/", payload, nil)
		if resp.StatusCode != 200 || resp.Code != 1 {
			t.Fatalf("seed user %d failed: %+v", index, resp)
		}
	}

	list := doJSONRequest(t, "GET", baseURL+"/api/v1/user/?page_num=1&page_size=20", nil, nil)
	if list.StatusCode != 200 || list.Code != 1 || list.Headers.Get("ETag") == "" {
		t.Fatalf("list users failed: %+v", list)
	}

	compressed, err := rawRequest("GET", baseURL+"/api/v1/user/?page_num=1&page_size=20", nil, map[string]string{"Accept-Encoding": "gzip"})
	if err != nil {
		t.Fatalf("compressed request failed: %v", err)
	}
	_, _ = io.ReadAll(compressed.Body)
	_ = compressed.Body.Close()
	if compressed.Header.Get("Content-Encoding") != "gzip" {
		t.Fatalf("expected gzip response, got headers %#v", compressed.Header)
	}

	get := doJSONRequest(t, "GET", baseURL+"/api/v1/user/"+strconv.FormatInt(created.ID, 10), nil, nil)
	if get.StatusCode != 200 || get.Code != 1 {
		t.Fatalf("get user failed: %+v", get)
	}

	update := doJSONRequest(t, "PUT", baseURL+"/api/v1/user/"+strconv.FormatInt(created.ID, 10), map[string]any{
		"name":   "Alice Updated",
		"email":  "alice.updated-" + suffix + "@example.com",
		"age":    29,
		"status": "inactive",
	}, nil)
	if update.StatusCode != 200 || update.Code != 1 {
		t.Fatalf("update user failed: %+v", update)
	}

	deleteResponse := doJSONRequest(t, "DELETE", baseURL+"/api/v1/user/"+strconv.FormatInt(created.ID, 10), nil, nil)
	if deleteResponse.StatusCode != 200 || deleteResponse.Code != 1 {
		t.Fatalf("delete user failed: %+v", deleteResponse)
	}

	notFound := doJSONRequest(t, "GET", baseURL+"/api/v1/user/"+strconv.FormatInt(created.ID, 10), nil, nil)
	if notFound.StatusCode != 404 || notFound.Code == 1 || !strings.Contains(strings.ToLower(notFound.Message), "not found") {
		t.Fatalf("expected deleted user to be missing, got %+v", notFound)
	}

	if database.Kind == "sqlite" {
		if _, err := os.Stat(database.SQLitePath); err != nil {
			t.Fatalf("expected sqlite database at %s: %v", database.SQLitePath, err)
		}
	}
}

type apiResponse struct {
	StatusCode int
	Headers    http.Header
	Code       int             `json:"code"`
	Message    string          `json:"message"`
	Data       json.RawMessage `json:"data"`
}

func doJSONRequest(t *testing.T, method string, url string, body any, headers map[string]string) apiResponse {
	t.Helper()

	resp, err := rawRequest(method, url, body, headers)
	if err != nil {
		t.Fatalf("request %s %s failed: %v", method, url, err)
	}
	defer resp.Body.Close()

	var decoded apiResponse
	decoded.StatusCode = resp.StatusCode
	decoded.Headers = resp.Header.Clone()
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	return decoded
}

func rawRequest(method string, url string, body any, headers map[string]string) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = strings.NewReader(string(raw))
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	return (&http.Client{Timeout: 10 * time.Second}).Do(req)
}

func startGeneratedService(t *testing.T, targetDir string, configBody string) (string, *exec.Cmd, *strings.Builder) {
	t.Helper()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "server.yaml")
	if err := os.WriteFile(configPath, []byte(configBody), 0o644); err != nil {
		t.Fatalf("write config failed: %v", err)
	}

	binaryPath := buildBinary(t, targetDir)
	cmd := exec.Command(binaryPath, "serve", "--config", configPath)
	cmd.Dir = targetDir
	var output strings.Builder
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := cmd.Start(); err != nil {
		t.Fatalf("start generated service failed: %v", err)
	}
	t.Cleanup(func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		_, _ = cmd.Process.Wait()
	})
	host := "127.0.0.1"
	port := extractConfigPort(configBody)
	baseURL := "http://" + net.JoinHostPort(host, port)
	waitForReady(t, baseURL+"/healthz", &output, cmd)
	return baseURL, cmd, &output
}

func extractConfigPort(configBody string) string {
	for _, line := range strings.Split(configBody, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, `port:`) {
			continue
		}
		value := strings.TrimSpace(strings.TrimPrefix(trimmed, "port:"))
		value = strings.Trim(value, `"`)
		if value != "" {
			return value
		}
	}
	return "3000"
}

func buildBinary(t *testing.T, targetDir string) string {
	t.Helper()

	binaryName := "service"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	binaryPath := filepath.Join(t.TempDir(), binaryName)
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = targetDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("build binary failed: %v", err)
	}
	if _, err := os.Stat(binaryPath); err != nil {
		t.Fatalf("expected built binary to exist: %v", err)
	}
	return binaryPath
}

func waitForReady(t *testing.T, url string, output *strings.Builder, cmd *exec.Cmd) {
	t.Helper()

	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
			t.Fatalf("service exited before readiness:\n%s", output.String())
		}

		resp, err := http.Get(url)
		if err == nil {
			_, _ = io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			if resp.StatusCode == 200 {
				return
			}
		}
		time.Sleep(250 * time.Millisecond)
	}

	t.Fatalf("service did not become ready:\n%s", output.String())
}

func healthBodyContainsService(t *testing.T, raw json.RawMessage, want string) bool {
	t.Helper()

	var payload struct {
		Services []string `json:"services"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("decode health payload failed: %v", err)
	}
	for _, service := range payload.Services {
		if service == want {
			return true
		}
	}
	return false
}

func containsService(services []string, want string) bool {
	for _, service := range services {
		if service == want {
			return true
		}
	}
	return false
}

func randomPort(t *testing.T) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve port failed: %v", err)
	}
	defer listener.Close()

	_, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatalf("parse port failed: %v", err)
	}
	return port
}
