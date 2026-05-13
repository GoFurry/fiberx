package extra_light_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gofurry/fiberx/v3/test/internaltest"
)

func TestExtraLightTemplateBlackBox(t *testing.T) {
	port := "18104"
	workdir := internaltest.TemplateRoot(t, "extra-light")
	configPath, databasePath := writeExtraLightConfig(t, port)
	baseURL := internaltest.FormatBaseURL(port)

	stop := internaltest.StartService(t, workdir, configPath, baseURL, "/healthz")
	t.Cleanup(stop)

	assertExtraLightScenario(t, baseURL, databasePath)
}

func assertExtraLightScenario(t *testing.T, baseURL, databasePath string) {
	t.Helper()

	if _, err := os.Stat(databasePath); err != nil {
		t.Fatalf("sqlite database file not created: %v", err)
	}

	for _, path := range []string{"/livez", "/readyz", "/startupz", "/healthz"} {
		internaltest.AssertStatus(t, baseURL+path, 200)
	}
}

func writeExtraLightConfig(t *testing.T, port string) (string, string) {
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
  app_name: "fiberx"
  mode: "debug"
  ip_address: "127.0.0.1"
  port: %q
database:
  path: %q
log:
  log_level: "debug"
  log_path: %q
`, port, filepath.ToSlash(databasePath), filepath.ToSlash(logPath))

	if err := os.WriteFile(configPath, []byte(configBody), 0o644); err != nil {
		t.Fatalf("write test config failed: %v", err)
	}

	return configPath, filepath.FromSlash(databasePath)
}
