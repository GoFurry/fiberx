package planner

import (
	"testing"

	"github.com/gofurry/fiberx/internal/manifest"
	"github.com/gofurry/fiberx/internal/validator"
)

func TestBuildPlanSelectsMediumRedisAssetsAndRules(t *testing.T) {
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

	plan := BuildPlan("demo", "github.com/example/demo", "medium", []string{"redis"}, map[string]string{"target_dir": t.TempDir()}, root, catalog)

	if plan.Base.Name != "service-base-cobra" {
		t.Fatalf("expected default base service-base-cobra, got %q", plan.Base.Name)
	}
	if plan.FiberVersion != "v3" || plan.CLIStyle != "cobra" {
		t.Fatalf("expected default stack v3/cobra, got fiber=%q cli=%q", plan.FiberVersion, plan.CLIStyle)
	}
	if len(plan.PresetPacks) != 1 || plan.PresetPacks[0].Name != "preset-medium-v3" {
		t.Fatalf("expected one preset pack preset-medium-v3, got %#v", plan.PresetPacks)
	}
	if len(plan.CapabilityPacks) != 3 {
		t.Fatalf("expected three capability packs for medium defaults + redis, got %#v", plan.CapabilityPacks)
	}
	if plan.Logger != "zap" || plan.Database != "sqlite" || plan.DataAccess != "stdlib" || plan.JSONLib != "stdlib" {
		t.Fatalf("expected default runtime options zap/sqlite/stdlib/stdlib, got logger=%q db=%q data=%q json=%q", plan.Logger, plan.Database, plan.DataAccess, plan.JSONLib)
	}
	if len(plan.RuntimeOverlays) != 2 {
		t.Fatalf("expected two runtime overlays, got %#v", plan.RuntimeOverlays)
	}
	assertPlanHasPack(t, plan.RuntimeOverlays, "runtime-logger-zap")
	assertPlanHasPack(t, plan.RuntimeOverlays, "runtime-data-stdlib")
	assertPlanHasPack(t, plan.CapabilityPacks, "swagger")
	assertPlanHasPack(t, plan.CapabilityPacks, "embedded-ui")
	assertPlanHasPack(t, plan.CapabilityPacks, "redis")
	if len(plan.ReplaceRules) != 1 {
		t.Fatalf("expected 1 replace rule, got %d", len(plan.ReplaceRules))
	}
	if len(plan.InjectionRules) != 3 {
		t.Fatalf("expected 3 injection rules, got %d", len(plan.InjectionRules))
	}
}

func assertPlanHasPack(t *testing.T, assets []AssetSelection, name string) {
	t.Helper()

	for _, asset := range assets {
		if asset.Name == name {
			return
		}
	}

	t.Fatalf("expected asset selection %q in %#v", name, assets)
}
