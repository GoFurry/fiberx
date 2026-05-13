package validator

import (
	"fmt"
	"os"
	"strings"

	"github.com/gofurry/fiberx/internal/manifest"
	"github.com/gofurry/fiberx/internal/stack"
)

func ValidateCatalog(catalog manifest.Catalog) error {
	seenPresets := make(map[string]struct{}, len(catalog.Presets))
	for _, preset := range catalog.Presets {
		if preset.Name == "" {
			return fmt.Errorf("preset name cannot be empty")
		}
		if strings.TrimSpace(preset.Summary) == "" {
			return fmt.Errorf("preset %q summary cannot be empty", preset.Name)
		}
		if strings.TrimSpace(preset.Description) == "" {
			return fmt.Errorf("preset %q description cannot be empty", preset.Name)
		}
		if preset.Implemented && strings.TrimSpace(preset.Base) == "" {
			return fmt.Errorf("preset %q base cannot be empty when implemented", preset.Name)
		}
		if _, exists := seenPresets[preset.Name]; exists {
			return fmt.Errorf("duplicate preset %q", preset.Name)
		}
		seenPresets[preset.Name] = struct{}{}
	}

	seenCapabilities := make(map[string]struct{}, len(catalog.Capabilities))
	for _, capability := range catalog.Capabilities {
		if capability.Name == "" {
			return fmt.Errorf("capability name cannot be empty")
		}
		if strings.TrimSpace(capability.Summary) == "" {
			return fmt.Errorf("capability %q summary cannot be empty", capability.Name)
		}
		if strings.TrimSpace(capability.Description) == "" {
			return fmt.Errorf("capability %q description cannot be empty", capability.Name)
		}
		if len(capability.AllowedPresets) == 0 {
			return fmt.Errorf("capability %q allowed_presets cannot be empty", capability.Name)
		}
		if _, exists := seenCapabilities[capability.Name]; exists {
			return fmt.Errorf("duplicate capability %q", capability.Name)
		}
		seenCapabilities[capability.Name] = struct{}{}
	}

	for _, preset := range catalog.Presets {
		for _, capabilityName := range preset.DefaultCapabilities {
			if !catalog.HasCapability(capabilityName) {
				return fmt.Errorf("preset %q references unknown default capability %q", preset.Name, capabilityName)
			}
			if !contains(preset.AllowedCapabilities, capabilityName) {
				return fmt.Errorf("preset %q default capability %q must also be in allowed_capabilities", preset.Name, capabilityName)
			}
		}
		for _, capabilityName := range preset.AllowedCapabilities {
			if !catalog.HasCapability(capabilityName) {
				return fmt.Errorf("preset %q references unknown allowed capability %q", preset.Name, capabilityName)
			}
			capability, _ := catalog.FindCapability(capabilityName)
			if !contains(capability.AllowedPresets, preset.Name) {
				return fmt.Errorf("preset %q allowed capability %q must also reference preset %q", preset.Name, capabilityName, preset.Name)
			}
		}
	}

	for _, capability := range catalog.Capabilities {
		for _, presetName := range capability.AllowedPresets {
			if _, ok := catalog.FindPreset(presetName); !ok {
				return fmt.Errorf("capability %q references unknown preset %q", capability.Name, presetName)
			}
			preset, _ := catalog.FindPreset(presetName)
			if !contains(preset.AllowedCapabilities, capability.Name) {
				return fmt.Errorf("capability %q allowed preset %q must also reference capability %q", capability.Name, presetName, capability.Name)
			}
		}
		for _, name := range capability.DependsOn {
			if !catalog.HasCapability(name) {
				return fmt.Errorf("capability %q depends on unknown capability %q", capability.Name, name)
			}
		}
		for _, name := range capability.ConflictsWith {
			if !catalog.HasCapability(name) {
				return fmt.Errorf("capability %q conflicts with unknown capability %q", capability.Name, name)
			}
		}
	}

	seenReplaceRules := make(map[string]struct{}, len(catalog.ReplaceRules))
	for _, rule := range catalog.ReplaceRules {
		if strings.TrimSpace(rule.Name) == "" {
			return fmt.Errorf("replace rule name cannot be empty")
		}
		if _, exists := seenReplaceRules[rule.Name]; exists {
			return fmt.Errorf("duplicate replace rule %q", rule.Name)
		}
		seenReplaceRules[rule.Name] = struct{}{}
		if len(rule.Replacements) == 0 {
			return fmt.Errorf("replace rule %q must define at least one replacement", rule.Name)
		}
		if err := validateScope(rule.Scope, catalog, fmt.Sprintf("replace rule %q", rule.Name)); err != nil {
			return err
		}
		for _, replacement := range rule.Replacements {
			if strings.TrimSpace(replacement.Placeholder) == "" {
				return fmt.Errorf("replace rule %q contains an empty placeholder", rule.Name)
			}
			if strings.TrimSpace(replacement.ValueFrom) == "" {
				return fmt.Errorf("replace rule %q contains an empty value_from", rule.Name)
			}
		}
	}

	seenInjectionRules := make(map[string]struct{}, len(catalog.InjectionRules))
	for _, rule := range catalog.InjectionRules {
		if strings.TrimSpace(rule.Name) == "" {
			return fmt.Errorf("injection rule name cannot be empty")
		}
		if _, exists := seenInjectionRules[rule.Name]; exists {
			return fmt.Errorf("duplicate injection rule %q", rule.Name)
		}
		seenInjectionRules[rule.Name] = struct{}{}
		if err := validateScope(rule.Scope, catalog, fmt.Sprintf("injection rule %q", rule.Name)); err != nil {
			return err
		}
		if strings.TrimSpace(rule.Target) == "" {
			return fmt.Errorf("injection rule %q target cannot be empty", rule.Name)
		}
		if strings.TrimSpace(rule.Anchor) == "" {
			return fmt.Errorf("injection rule %q anchor cannot be empty", rule.Name)
		}
		if strings.TrimSpace(rule.Snippet) == "" {
			return fmt.Errorf("injection rule %q snippet cannot be empty", rule.Name)
		}
	}

	return nil
}

func ValidateAssets(root string, catalog manifest.Catalog) error {
	for _, preset := range catalog.Presets {
		if !preset.Implemented {
			continue
		}

		if _, err := os.Stat(manifest.BaseAssetDir(root, preset.Base)); err != nil {
			return fmt.Errorf("preset %q base asset %q is missing", preset.Name, preset.Base)
		}
		if _, err := os.Stat(manifest.BaseAssetDir(root, preset.Base+"-cobra")); err != nil {
			return fmt.Errorf("preset %q cobra base asset %q is missing", preset.Name, preset.Base+"-cobra")
		}

		for _, pack := range preset.Packs {
			if _, err := os.Stat(manifest.PackAssetDir(root, pack)); err != nil {
				return fmt.Errorf("preset %q pack asset %q is missing", preset.Name, pack)
			}
			if _, err := os.Stat(manifest.PackAssetDir(root, pack+"-v3")); err != nil {
				return fmt.Errorf("preset %q fiber v3 pack asset %q is missing", preset.Name, pack+"-v3")
			}
		}
	}

	for _, pack := range []string{
		"runtime-logger-zap",
		"runtime-logger-slog",
		"runtime-data-stdlib",
		"runtime-data-sqlx",
		"runtime-data-sqlc",
	} {
		if _, err := os.Stat(manifest.PackAssetDir(root, pack)); err != nil {
			return fmt.Errorf("runtime overlay asset %q is missing", pack)
		}
	}

	for _, capability := range catalog.Capabilities {
		if !capability.Implemented {
			continue
		}

		for _, pack := range capability.Packs {
			if _, err := os.Stat(manifest.CapabilityAssetDir(root, pack)); err != nil {
				return fmt.Errorf("capability %q pack asset %q is missing", capability.Name, pack)
			}
		}
	}

	for _, rule := range catalog.InjectionRules {
		if err := validateInjectionRuleAssets(root, catalog, rule); err != nil {
			return err
		}
	}

	return nil
}

func ValidateRequest(projectName string, modulePath string, preset string, capabilities []string, options map[string]string, catalog manifest.Catalog) error {
	if strings.TrimSpace(projectName) == "" {
		return fmt.Errorf("project name cannot be empty")
	}

	if strings.TrimSpace(modulePath) == "" {
		return fmt.Errorf("module path cannot be empty")
	}

	if !strings.Contains(modulePath, "/") {
		return fmt.Errorf("module path %q must look like a Go module path", modulePath)
	}

	options = stack.NormalizeOptions(options)
	if err := stack.ValidateOptions(options); err != nil {
		return err
	}

	presetManifest, ok := catalog.FindPreset(preset)
	if !ok {
		return fmt.Errorf("unknown preset %q", preset)
	}
	if presetManifest.Name == "extra-light" {
		if explicitlySet(options, stack.OptionLogger) {
			return fmt.Errorf("preset %q does not support logger option", presetManifest.Name)
		}
		if explicitlySet(options, stack.OptionDB) {
			return fmt.Errorf("preset %q does not support db option", presetManifest.Name)
		}
		if explicitlySet(options, stack.OptionDataAccess) {
			return fmt.Errorf("preset %q does not support data access option", presetManifest.Name)
		}
	}

	seenCapabilities := make(map[string]struct{}, len(capabilities))
	for _, capability := range capabilities {
		if capability == "" {
			return fmt.Errorf("capability name cannot be empty")
		}
		capabilityManifest, ok := catalog.FindCapability(capability)
		if !ok {
			return fmt.Errorf("unknown capability %q", capability)
		}
		if _, exists := seenCapabilities[capability]; exists {
			return fmt.Errorf("duplicate capability %q", capability)
		}
		if !contains(presetManifest.AllowedCapabilities, capabilityManifest.Name) {
			return fmt.Errorf("capability %q is not allowed for preset %q", capabilityManifest.Name, presetManifest.Name)
		}
		if !contains(capabilityManifest.AllowedPresets, presetManifest.Name) {
			return fmt.Errorf("capability %q does not support preset %q", capabilityManifest.Name, presetManifest.Name)
		}
		seenCapabilities[capability] = struct{}{}
	}

	for _, defaultCapability := range presetManifest.DefaultCapabilities {
		if _, exists := seenCapabilities[defaultCapability]; exists {
			continue
		}
		seenCapabilities[defaultCapability] = struct{}{}
	}

	for _, capabilityName := range capabilities {
		capabilityManifest, _ := catalog.FindCapability(capabilityName)
		for _, dependency := range capabilityManifest.DependsOn {
			if _, ok := seenCapabilities[dependency]; !ok {
				return fmt.Errorf("capability %q requires %q", capabilityName, dependency)
			}
		}
		for _, conflict := range capabilityManifest.ConflictsWith {
			if _, ok := seenCapabilities[conflict]; ok {
				return fmt.Errorf("capability %q conflicts with %q", capabilityName, conflict)
			}
		}
	}

	return nil
}

func ValidateGenerationSupport(preset manifest.PresetManifest, capabilities []manifest.CapabilityManifest) error {
	if !preset.Implemented {
		return fmt.Errorf("preset %q assets are not implemented yet", preset.Name)
	}

	for _, capability := range capabilities {
		if !capability.Implemented {
			return fmt.Errorf("capability %q assets are not implemented yet", capability.Name)
		}
	}

	return nil
}

func validateScope(scope manifest.Scope, catalog manifest.Catalog, context string) error {
	for _, presetName := range scope.Presets {
		if _, ok := catalog.FindPreset(presetName); !ok {
			return fmt.Errorf("%s references unknown preset %q", context, presetName)
		}
	}
	for _, capabilityName := range scope.Capabilities {
		if !catalog.HasCapability(capabilityName) {
			return fmt.Errorf("%s references unknown capability %q", context, capabilityName)
		}
	}
	return nil
}

func validateInjectionRuleAssets(root string, catalog manifest.Catalog, rule manifest.InjectionRule) error {
	roots := matchingSnippetRoots(root, catalog, rule.Scope)
	if len(roots) == 0 {
		return nil
	}

	snippetFound := false
	for _, assetRoot := range roots {
		if manifest.SnippetExists(assetRoot, rule.Snippet) {
			snippetFound = true
			break
		}
	}
	if !snippetFound {
		return fmt.Errorf("injection rule %q snippet %q does not exist in any matching asset root", rule.Name, rule.Snippet)
	}

	targetFound := false
	for _, assetRoot := range implementedAssetRoots(root, catalog) {
		files, err := manifest.CollectAssetFiles(assetRoot)
		if err != nil {
			return err
		}
		for _, file := range files {
			if file.OutputPath == rule.Target {
				targetFound = true
				break
			}
		}
		if targetFound {
			break
		}
	}
	if !targetFound {
		return fmt.Errorf("injection rule %q target %q does not exist in any matching asset root", rule.Name, rule.Target)
	}

	return nil
}

func matchingSnippetRoots(root string, catalog manifest.Catalog, scope manifest.Scope) []string {
	roots := []string{}

	for _, capabilityName := range scope.Capabilities {
		capability, ok := catalog.FindCapability(capabilityName)
		if !ok || !capability.Implemented {
			continue
		}
		for _, pack := range capability.Packs {
			roots = append(roots, manifest.CapabilityAssetDir(root, pack))
		}
	}

	for _, presetName := range scope.Presets {
		preset, ok := catalog.FindPreset(presetName)
		if !ok || !preset.Implemented {
			continue
		}
		for _, pack := range preset.Packs {
			roots = append(roots, manifest.PackAssetDir(root, pack))
		}
	}

	return roots
}

func implementedAssetRoots(root string, catalog manifest.Catalog) []string {
	roots := []string{}

	for _, preset := range catalog.Presets {
		if !preset.Implemented {
			continue
		}
		if preset.Base != "" {
			roots = append(roots, manifest.BaseAssetDir(root, preset.Base))
		}
		for _, pack := range preset.Packs {
			roots = append(roots, manifest.PackAssetDir(root, pack))
		}
	}

	for _, capability := range catalog.Capabilities {
		if !capability.Implemented {
			continue
		}
		for _, pack := range capability.Packs {
			roots = append(roots, manifest.CapabilityAssetDir(root, pack))
		}
	}

	return roots
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func explicitlySet(options map[string]string, key string) bool {
	return strings.EqualFold(strings.TrimSpace(options["_explicit_"+key]), "true")
}
