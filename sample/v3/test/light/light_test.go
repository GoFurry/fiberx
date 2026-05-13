package light_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gofurry/fiberx/v3/test/internaltest"
)

func TestLightTemplateBlackBox(t *testing.T) {
	port := "18103"
	workdir := internaltest.TemplateRoot(t, "light")
	configPath, databasePath := writeLightConfig(t, port)
	baseURL := internaltest.FormatBaseURL(port)

	stop := internaltest.StartService(t, workdir, configPath, baseURL, "/healthz")
	t.Cleanup(stop)

	assertLightScenario(t, baseURL, databasePath)
}

func assertLightScenario(t *testing.T, baseURL, databasePath string) {
	t.Helper()

	if _, err := os.Stat(databasePath); err != nil {
		t.Fatalf("sqlite database file not created: %v", err)
	}

	health := internaltest.DoRequest(t, "GET", baseURL+"/healthz", nil)
	if health.Status != 200 {
		t.Fatalf("healthz returned unexpected status: %+v", health)
	}
	if health.Code != 1 || strings.ToLower(health.Message) != "success" {
		t.Fatalf("healthz returned unexpected envelope: %+v", health)
	}
	if health.Headers.Get("X-Request-ID") == "" {
		t.Fatalf("expected request id header on healthz response")
	}

	for _, path := range []string{"/livez", "/readyz", "/startupz"} {
		internaltest.AssertStatus(t, baseURL+path, 200)
	}

	created := createUsers(t, baseURL)

	list := internaltest.DoRequest(t, "GET", baseURL+"/api/v1/user/?page_num=1&page_size=50", nil)
	if list.Headers.Get("ETag") == "" {
		t.Fatalf("expected ETag header on list response")
	}
	var users internaltest.UserListResponse
	internaltest.MustDecode(t, list.Data, &users)
	if users.Total < int64(len(created)) {
		t.Fatalf("unexpected list payload: %+v", users)
	}

	compressed := internaltest.RawRequest(t, "GET", baseURL+"/api/v1/user/?page_num=1&page_size=50", nil, map[string]string{
		"Accept-Encoding": "gzip",
	})
	defer compressed.Body.Close()
	if compressed.Header.Get("Content-Encoding") != "gzip" {
		t.Fatalf("expected compressed response when Accept-Encoding is set")
	}

	first := created[0]
	get := internaltest.DoRequest(t, "GET", baseURL+"/api/v1/user/"+internaltest.ToString(first.ID), nil)
	var fetched internaltest.UserResponse
	internaltest.MustDecode(t, get.Data, &fetched)
	if fetched.Email != first.Email {
		t.Fatalf("unexpected fetched user: %+v", fetched)
	}

	updateBody := map[string]any{
		"name":   "Alice Updated",
		"email":  "alice.updated@example.com",
		"age":    29,
		"status": "inactive",
	}
	update := internaltest.DoRequest(t, "PUT", baseURL+"/api/v1/user/"+internaltest.ToString(first.ID), updateBody)
	var updated internaltest.UserResponse
	internaltest.MustDecode(t, update.Data, &updated)
	if updated.Name != "Alice Updated" || updated.Status != "inactive" {
		t.Fatalf("unexpected updated user: %+v", updated)
	}

	deleteResponse := internaltest.DoRequest(t, "DELETE", baseURL+"/api/v1/user/"+internaltest.ToString(first.ID), nil)
	if deleteResponse.Code != 1 {
		t.Fatalf("delete returned unexpected envelope: %+v", deleteResponse)
	}

	notFound := internaltest.DoRequest(t, "GET", baseURL+"/api/v1/user/"+internaltest.ToString(first.ID), nil)
	if notFound.Code == 1 || !strings.Contains(strings.ToLower(notFound.Message), "not found") {
		t.Fatalf("expected deleted user to be missing, got %+v", notFound)
	}
}

func createUsers(t *testing.T, baseURL string) []internaltest.UserResponse {
	t.Helper()

	users := make([]internaltest.UserResponse, 0, 12)
	for index := 0; index < 12; index++ {
		body := map[string]any{
			"name":   fmt.Sprintf("User-%02d", index+1),
			"email":  fmt.Sprintf("user-%02d@example.com", index+1),
			"age":    20 + index,
			"status": "active",
		}
		resp := internaltest.DoRequest(t, "POST", baseURL+"/api/v1/user/", body)
		if resp.Code != 1 {
			t.Fatalf("create user returned unexpected envelope: %+v", resp)
		}
		var created internaltest.UserResponse
		internaltest.MustDecode(t, resp.Data, &created)
		if created.ID == 0 {
			t.Fatalf("expected created user id, got %+v", created)
		}
		users = append(users, created)
	}
	return users
}

func writeLightConfig(t *testing.T, port string) (string, string) {
	t.Helper()

	tempDir := t.TempDir()
	databasePath := filepath.Join(tempDir, "data", "app.db")
	logPath := filepath.Join(tempDir, "logs", "app.log")
	if err := os.MkdirAll(filepath.Dir(databasePath), 0o755); err != nil {
		t.Fatalf("create database dir failed: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		t.Fatalf("create log dir failed: %v", err)
	}

	configPath := filepath.Join(tempDir, "server.yaml")
	configBody := fmt.Sprintf(`server:
  app_id: "fiberx"
  app_name: "fiberx"
  app_version: "v1.0.0"
  mode: "debug"
  ip_address: "127.0.0.1"
  port: %q
  memory_limit: 1
  gc_percent: 1000
  network: "tcp"
  enable_prefork: false
  is_full_stack: false
database:
  enabled: true
  auto_migrate: true
  db_type: "sqlite"
  sqlite:
    path: %q
log:
  log_level: "debug"
  log_mode: "dev"
  log_path: %q
middleware:
  cors:
    allow_origins: ["http://127.0.0.1:8888"]
  limiter:
    enabled: false
`, port, filepath.ToSlash(databasePath), filepath.ToSlash(logPath))

	if err := os.WriteFile(configPath, []byte(configBody), 0o644); err != nil {
		t.Fatalf("write test config failed: %v", err)
	}

	return configPath, filepath.FromSlash(databasePath)
}
