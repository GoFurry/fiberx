package renderer

import (
	"strings"
	"testing"

	"github.com/gofurry/fiberx/internal/manifest"
	"github.com/gofurry/fiberx/internal/planner"
	"github.com/gofurry/fiberx/internal/validator"
)

func TestRenderAppliesReplacementsAndInjection(t *testing.T) {
	root := "../../generator"
	catalog, err := manifest.LoadCatalog(root)
	if err != nil {
		t.Fatalf("LoadCatalog() returned error: %v", err)
	}
	if err := validator.ValidateCatalog(catalog); err != nil {
		t.Fatalf("ValidateCatalog() returned error: %v", err)
	}
	if err := validator.ValidateAssets(root, catalog); err != nil {
		t.Fatalf("ValidateAssets() returned error: %v", err)
	}

	plan := planner.BuildPlan("demo", "github.com/example/demo", "medium", []string{"redis"}, map[string]string{"target_dir": t.TempDir()}, root, catalog)
	result, err := Render(plan)
	if err != nil {
		t.Fatalf("Render() returned error: %v", err)
	}

	bootstrap := findRenderedFile(t, result, "internal/bootstrap/bootstrap.go")
	if !strings.Contains(bootstrap, `services = append(services, "cache:redis")`) {
		t.Fatalf("expected redis injection in bootstrap.go, got:\n%s", bootstrap)
	}
	if !strings.Contains(bootstrap, `services = append(services, "docs:swagger")`) {
		t.Fatalf("expected swagger injection in bootstrap.go, got:\n%s", bootstrap)
	}
	if !strings.Contains(bootstrap, `services = append(services, "ui:embedded")`) {
		t.Fatalf("expected embedded-ui injection in bootstrap.go, got:\n%s", bootstrap)
	}

	goMod := findRenderedFile(t, result, "go.mod")
	if !strings.Contains(goMod, "module github.com/example/demo") {
		t.Fatalf("expected rendered go.mod module path, got:\n%s", goMod)
	}

	openAPI := findRenderedFile(t, result, "docs/openapi.yaml")
	if !strings.Contains(openAPI, "demo API") {
		t.Fatalf("expected swagger docs to be rendered, got:\n%s", openAPI)
	}
}

func findRenderedFile(t *testing.T, result Result, path string) string {
	t.Helper()

	for _, file := range result.Files {
		if file.Path == path {
			return string(file.Content)
		}
	}

	t.Fatalf("rendered file %q not found", path)
	return ""
}
