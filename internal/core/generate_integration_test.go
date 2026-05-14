//go:build integration

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

	"github.com/gofurry/fiberx/internal/stack"
)

func TestPhase12CapabilityMatrixBlackBox(t *testing.T) {
	testCases := []struct {
		name         string
		preset       string
		capabilities []string
		runScenario  func(t *testing.T, targetDir string)
	}{
		{
			name:        "extra-light default",
			preset:      "extra-light",
			runScenario: runExtraLightBlackBoxScenario,
		},
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
