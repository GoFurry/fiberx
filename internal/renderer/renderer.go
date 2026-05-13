package renderer

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gofurry/fiberx/internal/manifest"
	"github.com/gofurry/fiberx/internal/planner"
)

type File struct {
	Path    string
	Content []byte
}

type Result struct {
	Files           []File
	Warnings        []string
	ReplaceRuleHits []string
	InjectionHits   []string
}

func Render(plan planner.Plan) (Result, error) {
	rendered := make(map[string][]byte)

	for _, asset := range plan.Assets {
		files, err := manifest.CollectAssetFiles(asset.Dir)
		if err != nil {
			return Result{}, err
		}

		for _, file := range files {
			if _, exists := rendered[file.OutputPath]; exists && !asset.AllowOverride {
				return Result{}, fmt.Errorf("duplicate rendered file %q from asset %q", file.OutputPath, asset.Name)
			}

			data, err := os.ReadFile(file.SourcePath)
			if err != nil {
				return Result{}, fmt.Errorf("read asset file %q: %w", file.SourcePath, err)
			}

			rendered[file.OutputPath] = data
		}
	}

	replaceRuleHits := make([]string, 0, len(plan.ReplaceRules))
	for _, rule := range plan.ReplaceRules {
		replaceRuleHits = append(replaceRuleHits, rule.Name)
		for path, content := range rendered {
			replaced := string(content)
			for _, replacement := range rule.Replacements {
				replaced = strings.ReplaceAll(replaced, replacement.Placeholder, replacementValue(replacement.ValueFrom, plan))
			}
			rendered[path] = []byte(replaced)
		}
	}

	injectionHits := make([]string, 0, len(plan.InjectionRules))
	for _, rule := range plan.InjectionRules {
		content, ok := rendered[rule.Target]
		if !ok {
			return Result{}, fmt.Errorf("injection target %q was not rendered", rule.Target)
		}

		snippet, err := loadSnippet(plan, rule)
		if err != nil {
			return Result{}, err
		}

		source := string(content)
		if !strings.Contains(source, rule.Anchor) {
			return Result{}, fmt.Errorf("anchor %q not found in %q", rule.Anchor, rule.Target)
		}

		source = strings.Replace(source, rule.Anchor, snippet+"\n\t"+rule.Anchor, 1)
		rendered[rule.Target] = []byte(source)
		injectionHits = append(injectionHits, rule.Name)
	}

	paths := make([]string, 0, len(rendered))
	for path := range rendered {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	files := make([]File, 0, len(paths))
	for _, path := range paths {
		files = append(files, File{
			Path:    filepath.ToSlash(path),
			Content: rendered[path],
		})
	}

	return Result{
		Files:           files,
		Warnings:        []string{},
		ReplaceRuleHits: replaceRuleHits,
		InjectionHits:   injectionHits,
	}, nil
}

func replacementValue(name string, plan planner.Plan) string {
	switch name {
	case "project_name":
		return plan.ProjectName
	case "module_path":
		return plan.ModulePath
	case "preset_name":
		return plan.Preset.Name
	case "preset_summary":
		return plan.Preset.Summary
	case "preset_description":
		return plan.Preset.Description
	case "fiber_version":
		return plan.FiberVersion
	case "cli_style":
		return plan.CLIStyle
	case "fiber_module":
		return plan.Options["fiber_module"]
	case "fiber_dependency":
		return plan.Options["fiber_dependency"]
	case "default_stack":
		return plan.Options["default_stack"]
	case "default_logger":
		return plan.Options["default_logger"]
	case "default_database":
		return plan.Options["default_database"]
	case "default_data_access":
		return plan.Options["default_data_access"]
	case "logger_backend":
		return plan.Options["logger_backend"]
	case "db_kind":
		return plan.Options["db_kind"]
	case "db_type_default":
		return plan.Options["db_type_default"]
	case "data_access_kind":
		return plan.Options["data_access_kind"]
	case "json_lib":
		return plan.Options["json_lib"]
	case "json_import":
		return plan.Options["json_import"]
	case "json_encoder":
		return plan.Options["json_encoder"]
	case "json_decoder":
		return plan.Options["json_decoder"]
	case "swagger_enabled":
		return boolString(planHasCapability(plan, "swagger"))
	case "embedded_ui_enabled":
		return boolString(planHasCapability(plan, "embedded-ui"))
	case "redis_enabled":
		return boolString(planHasCapability(plan, "redis"))
	default:
		return ""
	}
}

func planHasCapability(plan planner.Plan, name string) bool {
	for _, capability := range plan.Capabilities {
		if capability.Name == name {
			return true
		}
	}
	return false
}

func boolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func loadSnippet(plan planner.Plan, rule manifest.InjectionRule) (string, error) {
	for _, asset := range snippetAssets(plan, rule.Scope) {
		path := filepath.Join(asset.Dir, filepath.FromSlash(rule.Snippet))
		data, err := os.ReadFile(path)
		if err == nil {
			return strings.TrimRight(string(data), "\r\n"), nil
		}
	}

	return "", fmt.Errorf("snippet %q for rule %q could not be resolved", rule.Snippet, rule.Name)
}

func snippetAssets(plan planner.Plan, scope manifest.Scope) []planner.AssetSelection {
	assets := []planner.AssetSelection{}

	if len(scope.Capabilities) > 0 {
		for _, asset := range plan.CapabilityPacks {
			for _, capability := range scope.Capabilities {
				if asset.Name == capability {
					assets = append(assets, asset)
				}
			}
		}
	}

	if len(scope.Presets) > 0 {
		for _, asset := range plan.PresetPacks {
			for _, preset := range scope.Presets {
				if plan.Preset.Name == preset {
					assets = append(assets, asset)
				}
			}
		}
	}

	if len(assets) == 0 && plan.Base.Name != "" {
		assets = append(assets, plan.Base)
	}

	return assets
}
