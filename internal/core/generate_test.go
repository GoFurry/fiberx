package core

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

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

		})
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

		})
	}
}

func TestGeneratedProjectCompileSmoke(t *testing.T) {
	testCases := []struct {
		name         string
		preset       string
		capabilities []string
	}{
		{name: "extra-light default", preset: "extra-light"},
		{name: "heavy default", preset: "heavy"},
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
