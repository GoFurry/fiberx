package metadata

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gofurry/fiberx/internal/manifest"
	"github.com/gofurry/fiberx/internal/planner"
	"github.com/gofurry/fiberx/internal/postprocess"
	"github.com/gofurry/fiberx/internal/renderer"
	"github.com/gofurry/fiberx/internal/stack"
	"github.com/gofurry/fiberx/internal/validator"
	"github.com/gofurry/fiberx/internal/version"
	"github.com/gofurry/fiberx/internal/writer"
)

const (
	ManifestDir      = ".fiberx"
	ManifestFilename = "manifest.json"
	SchemaVersionV1  = "v1"

	StatusClean            = "clean"
	StatusLocalModified    = "local_modified"
	StatusGeneratorDrift   = "generator_drift"
	StatusLocalAndGenDrift = "local_and_generator_drift"
)

var renderCache struct {
	mu      sync.Mutex
	entries map[string]renderCacheEntry
}

type renderCacheEntry struct {
	ManagedFiles []ManagedFile
	Fingerprint  string
}

type ProjectManifest struct {
	SchemaVersion string        `json:"schema_version"`
	GeneratedAt   string        `json:"generated_at"`
	Generator     GeneratorInfo `json:"generator"`
	Recipe        Recipe        `json:"recipe"`
	Assets        AssetSet      `json:"assets"`
	Fingerprints  Fingerprints  `json:"fingerprints"`
	ManagedFiles  []ManagedFile `json:"managed_files"`
}

type GeneratorInfo struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
}

type Recipe struct {
	ProjectName  string   `json:"project_name"`
	ModulePath   string   `json:"module_path"`
	Preset       string   `json:"preset"`
	Capabilities []string `json:"capabilities"`
	FiberVersion string   `json:"fiber_version"`
	CLIStyle     string   `json:"cli_style"`
	Logger       string   `json:"logger,omitempty"`
	DB           string   `json:"db,omitempty"`
	DataAccess   string   `json:"data_access,omitempty"`
	JSONLib      string   `json:"json_lib,omitempty"`
}

type AssetSet struct {
	Base            string   `json:"base"`
	PresetPacks     []string `json:"preset_packs"`
	CapabilityPacks []string `json:"capability_packs"`
	RuntimeOverlays []string `json:"runtime_overlays"`
	ReplaceRules    []string `json:"replace_rules"`
	InjectionRules  []string `json:"injection_rules"`
}

type Fingerprints struct {
	TemplateSet    string `json:"template_set"`
	RenderedOutput string `json:"rendered_output"`
}

type ManagedFile struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type DiffReport struct {
	Status              string            `json:"status"`
	Generator           DiffGeneratorInfo `json:"generator"`
	Recipe              Recipe            `json:"recipe"`
	MissingFiles        []string          `json:"missing_files"`
	ChangedFiles        []string          `json:"changed_files"`
	NewManagedFiles     []string          `json:"new_managed_files,omitempty"`
	GeneratorDriftFiles []string          `json:"generator_drift_files"`
}

type DiffGeneratorInfo struct {
	Current   GeneratorInfo `json:"current"`
	Generated GeneratorInfo `json:"generated"`
}

type assetFingerprint struct {
	Kind  string                 `json:"kind"`
	Name  string                 `json:"name"`
	Files []assetFileFingerprint `json:"files"`
}

type assetFileFingerprint struct {
	SourceRel  string `json:"source_rel"`
	OutputPath string `json:"output_path"`
	SHA256     string `json:"sha256"`
}

type injectionFingerprint struct {
	Name       string `json:"name"`
	Target     string `json:"target"`
	Anchor     string `json:"anchor"`
	Order      int    `json:"order"`
	Snippet    string `json:"snippet"`
	SnippetSHA string `json:"snippet_sha256"`
}

type templateFingerprintDescriptor struct {
	Recipe          Recipe                 `json:"recipe"`
	Base            assetFingerprint       `json:"base"`
	PresetPacks     []assetFingerprint     `json:"preset_packs"`
	CapabilityPacks []assetFingerprint     `json:"capability_packs"`
	RuntimeOverlays []assetFingerprint     `json:"runtime_overlays"`
	ReplaceRules    []manifest.ReplaceRule `json:"replace_rules"`
	InjectionRules  []injectionFingerprint `json:"injection_rules"`
}

func ManifestPath(projectDir string) string {
	return filepath.Join(projectDir, ManifestDir, ManifestFilename)
}

func BuildManifest(plan planner.Plan, rendered renderer.Result, targetDir string, generatedAt time.Time) (ProjectManifest, error) {
	managedFiles, renderedOutputFingerprint, err := snapshotManagedFiles(targetDir, rendered.Files)
	if err != nil {
		return ProjectManifest{}, err
	}

	templateSetFingerprint, err := computeTemplateSetFingerprint(plan)
	if err != nil {
		return ProjectManifest{}, err
	}

	return ProjectManifest{
		SchemaVersion: SchemaVersionV1,
		GeneratedAt:   generatedAt.UTC().Format(time.RFC3339),
		Generator: GeneratorInfo{
			Version: version.Version,
			Commit:  version.Commit,
		},
		Recipe:       buildRecipe(plan),
		Assets:       buildAssets(plan, rendered),
		Fingerprints: Fingerprints{TemplateSet: templateSetFingerprint, RenderedOutput: renderedOutputFingerprint},
		ManagedFiles: managedFiles,
	}, nil
}

func WriteManifest(targetDir string, projectManifest ProjectManifest) error {
	manifestPath := ManifestPath(targetDir)
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0o755); err != nil {
		return fmt.Errorf("create metadata directory for %q: %w", manifestPath, err)
	}

	data, err := json.MarshalIndent(projectManifest, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal project manifest: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(manifestPath, data, 0o644); err != nil {
		return fmt.Errorf("write project manifest %q: %w", manifestPath, err)
	}
	return nil
}

func LoadManifest(projectDir string) (ProjectManifest, error) {
	manifestPath := ManifestPath(projectDir)
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ProjectManifest{}, fmt.Errorf("fiberx metadata file %q was not found", filepath.ToSlash(filepath.Join(ManifestDir, ManifestFilename)))
		}
		return ProjectManifest{}, fmt.Errorf("read project manifest %q: %w", manifestPath, err)
	}

	var projectManifest ProjectManifest
	if err := json.Unmarshal(data, &projectManifest); err != nil {
		return ProjectManifest{}, fmt.Errorf("decode project manifest %q: %w", manifestPath, err)
	}
	if projectManifest.SchemaVersion != SchemaVersionV1 {
		return ProjectManifest{}, fmt.Errorf("unsupported fiberx metadata schema %q", projectManifest.SchemaVersion)
	}
	return projectManifest, nil
}

func BuildDiff(projectDir string, catalogRoot string) (DiffReport, error) {
	projectManifest, err := LoadManifest(projectDir)
	if err != nil {
		return DiffReport{}, err
	}

	currentLocalManaged, missingFiles, changedFiles, err := snapshotProjectManagedFiles(projectDir, projectManifest.ManagedFiles)
	if err != nil {
		return DiffReport{}, err
	}

	currentRenderedManaged, _, err := renderRecipeManagedFiles(projectManifest.Recipe, catalogRoot)
	if err != nil {
		return DiffReport{}, err
	}

	newManagedFiles, generatorDriftFiles := compareManagedFiles(projectManifest.ManagedFiles, currentRenderedManaged)
	localChanged := len(missingFiles) > 0 || len(changedFiles) > 0
	generatorChanged := len(newManagedFiles) > 0 || len(generatorDriftFiles) > 0

	status := StatusClean
	switch {
	case localChanged && generatorChanged:
		status = StatusLocalAndGenDrift
	case localChanged:
		status = StatusLocalModified
	case generatorChanged:
		status = StatusGeneratorDrift
	}

	_ = currentLocalManaged

	return DiffReport{
		Status: status,
		Generator: DiffGeneratorInfo{
			Current: GeneratorInfo{
				Version: version.Version,
				Commit:  version.Commit,
			},
			Generated: projectManifest.Generator,
		},
		Recipe:              projectManifest.Recipe,
		MissingFiles:        missingFiles,
		ChangedFiles:        changedFiles,
		NewManagedFiles:     newManagedFiles,
		GeneratorDriftFiles: generatorDriftFiles,
	}, nil
}

func buildRecipe(plan planner.Plan) Recipe {
	recipe := Recipe{
		ProjectName:  plan.ProjectName,
		ModulePath:   plan.ModulePath,
		Preset:       plan.Preset.Name,
		Capabilities: capabilityNames(plan.Capabilities),
		FiberVersion: plan.FiberVersion,
		CLIStyle:     plan.CLIStyle,
	}
	if plan.Preset.Name != "extra-light" {
		recipe.Logger = plan.Logger
		recipe.DB = plan.Database
		recipe.DataAccess = plan.DataAccess
	}
	recipe.JSONLib = plan.JSONLib
	return recipe
}

func buildAssets(plan planner.Plan, rendered renderer.Result) AssetSet {
	return AssetSet{
		Base:            plan.Base.Name,
		PresetPacks:     assetNames(plan.PresetPacks),
		CapabilityPacks: assetNames(plan.CapabilityPacks),
		RuntimeOverlays: assetNames(plan.RuntimeOverlays),
		ReplaceRules:    append([]string(nil), rendered.ReplaceRuleHits...),
		InjectionRules:  append([]string(nil), rendered.InjectionHits...),
	}
}

func capabilityNames(capabilities []manifest.CapabilityManifest) []string {
	names := make([]string, 0, len(capabilities))
	for _, capability := range capabilities {
		names = append(names, capability.Name)
	}
	return names
}

func assetNames(assets []planner.AssetSelection) []string {
	names := make([]string, 0, len(assets))
	for _, asset := range assets {
		names = append(names, asset.Name)
	}
	return names
}

func snapshotManagedFiles(targetDir string, renderedFiles []renderer.File) ([]ManagedFile, string, error) {
	managedFiles := make([]ManagedFile, 0, len(renderedFiles))
	for _, renderedFile := range renderedFiles {
		target := filepath.Join(targetDir, filepath.FromSlash(renderedFile.Path))
		data, err := os.ReadFile(target)
		if err != nil {
			return nil, "", fmt.Errorf("read managed file %q: %w", target, err)
		}
		managedFiles = append(managedFiles, ManagedFile{
			Path:   filepath.ToSlash(renderedFile.Path),
			SHA256: hashBytes(data),
		})
	}

	sort.SliceStable(managedFiles, func(i, j int) bool {
		return managedFiles[i].Path < managedFiles[j].Path
	})

	fingerprint, err := hashJSON(managedFiles)
	if err != nil {
		return nil, "", err
	}
	return managedFiles, fingerprint, nil
}

func snapshotProjectManagedFiles(projectDir string, manifestFiles []ManagedFile) ([]ManagedFile, []string, []string, error) {
	currentFiles := make([]ManagedFile, 0, len(manifestFiles))
	missingFiles := []string{}
	changedFiles := []string{}

	for _, managedFile := range manifestFiles {
		target := filepath.Join(projectDir, filepath.FromSlash(managedFile.Path))
		data, err := os.ReadFile(target)
		if err != nil {
			if os.IsNotExist(err) {
				missingFiles = append(missingFiles, managedFile.Path)
				continue
			}
			return nil, nil, nil, fmt.Errorf("read managed file %q: %w", target, err)
		}
		current := ManagedFile{
			Path:   managedFile.Path,
			SHA256: hashBytes(data),
		}
		currentFiles = append(currentFiles, current)
		if current.SHA256 != managedFile.SHA256 {
			changedFiles = append(changedFiles, managedFile.Path)
		}
	}

	sort.Strings(missingFiles)
	sort.Strings(changedFiles)
	return currentFiles, missingFiles, changedFiles, nil
}

func renderRecipeManagedFiles(recipe Recipe, catalogRoot string) ([]ManagedFile, string, error) {
	cacheKey, err := renderCacheKey(recipe, catalogRoot)
	if err != nil {
		return nil, "", err
	}
	if managedFiles, fingerprint, ok := lookupRenderCache(cacheKey); ok {
		return managedFiles, fingerprint, nil
	}

	options := map[string]string{
		"command":                "diff",
		"manifest_root":          catalogRoot,
		stack.OptionFiberVersion: recipe.FiberVersion,
		stack.OptionCLIStyle:     recipe.CLIStyle,
	}
	if recipe.Logger != "" {
		options[stack.OptionLogger] = recipe.Logger
	}
	if recipe.DB != "" {
		options[stack.OptionDB] = recipe.DB
	}
	if recipe.DataAccess != "" {
		options[stack.OptionDataAccess] = recipe.DataAccess
	}
	if recipe.JSONLib != "" {
		options[stack.OptionJSONLib] = recipe.JSONLib
	}

	tempDir, err := os.MkdirTemp("", "fiberx-phase13-diff-*")
	if err != nil {
		return nil, "", fmt.Errorf("create temp dir for diff render: %w", err)
	}
	defer os.RemoveAll(tempDir)
	options["target_dir"] = tempDir

	catalog, err := manifest.LoadCatalog(catalogRoot)
	if err != nil {
		return nil, "", err
	}
	if err := validator.ValidateCatalog(catalog); err != nil {
		return nil, "", err
	}
	if err := validator.ValidateAssets(catalogRoot, catalog); err != nil {
		return nil, "", err
	}
	if err := validator.ValidateRequest(recipe.ProjectName, recipe.ModulePath, recipe.Preset, recipe.Capabilities, options, catalog); err != nil {
		return nil, "", err
	}

	plan := planner.BuildPlan(recipe.ProjectName, recipe.ModulePath, recipe.Preset, recipe.Capabilities, options, catalogRoot, catalog)
	rendered, err := renderer.Render(plan)
	if err != nil {
		return nil, "", err
	}
	if _, err := writer.New(tempDir).Write(rendered); err != nil {
		return nil, "", err
	}
	if err := postprocess.FinalizeGeneratedModule(tempDir); err != nil {
		return nil, "", err
	}

	managedFiles, fingerprint, err := snapshotManagedFiles(tempDir, rendered.Files)
	if err != nil {
		return nil, "", err
	}
	storeRenderCache(cacheKey, managedFiles, fingerprint)
	return managedFiles, fingerprint, nil
}

func compareManagedFiles(manifestFiles []ManagedFile, renderedFiles []ManagedFile) ([]string, []string) {
	manifestByPath := make(map[string]string, len(manifestFiles))
	renderedByPath := make(map[string]string, len(renderedFiles))
	for _, file := range manifestFiles {
		manifestByPath[file.Path] = file.SHA256
	}
	for _, file := range renderedFiles {
		renderedByPath[file.Path] = file.SHA256
	}

	newManagedFiles := []string{}
	generatorDriftFiles := []string{}
	for path, renderedSHA := range renderedByPath {
		manifestSHA, ok := manifestByPath[path]
		if !ok {
			newManagedFiles = append(newManagedFiles, path)
			generatorDriftFiles = append(generatorDriftFiles, path)
			continue
		}
		if manifestSHA != renderedSHA {
			generatorDriftFiles = append(generatorDriftFiles, path)
		}
	}
	for path := range manifestByPath {
		if _, ok := renderedByPath[path]; !ok {
			generatorDriftFiles = append(generatorDriftFiles, path)
		}
	}

	sort.Strings(newManagedFiles)
	sort.Strings(generatorDriftFiles)
	return newManagedFiles, uniqueStrings(generatorDriftFiles)
}

func computeTemplateSetFingerprint(plan planner.Plan) (string, error) {
	base, err := fingerprintAsset(plan.Base)
	if err != nil {
		return "", err
	}

	presetPacks, err := fingerprintAssets(plan.PresetPacks)
	if err != nil {
		return "", err
	}
	capabilityPacks, err := fingerprintAssets(plan.CapabilityPacks)
	if err != nil {
		return "", err
	}
	runtimeOverlays, err := fingerprintAssets(plan.RuntimeOverlays)
	if err != nil {
		return "", err
	}
	injectionRules, err := fingerprintInjectionRules(plan)
	if err != nil {
		return "", err
	}

	descriptor := templateFingerprintDescriptor{
		Recipe:          normalizedRecipeForFingerprint(buildRecipe(plan)),
		Base:            base,
		PresetPacks:     presetPacks,
		CapabilityPacks: capabilityPacks,
		RuntimeOverlays: runtimeOverlays,
		ReplaceRules:    append([]manifest.ReplaceRule(nil), plan.ReplaceRules...),
		InjectionRules:  injectionRules,
	}
	return hashJSON(descriptor)
}

func fingerprintAssets(assets []planner.AssetSelection) ([]assetFingerprint, error) {
	result := make([]assetFingerprint, 0, len(assets))
	for _, asset := range assets {
		fingerprint, err := fingerprintAsset(asset)
		if err != nil {
			return nil, err
		}
		result = append(result, fingerprint)
	}
	return result, nil
}

func fingerprintAsset(asset planner.AssetSelection) (assetFingerprint, error) {
	files, err := manifest.CollectAssetFiles(asset.Dir)
	if err != nil {
		return assetFingerprint{}, err
	}

	fingerprints := make([]assetFileFingerprint, 0, len(files))
	for _, file := range files {
		data, err := os.ReadFile(file.SourcePath)
		if err != nil {
			return assetFingerprint{}, fmt.Errorf("read asset file %q: %w", file.SourcePath, err)
		}
		rel, err := filepath.Rel(asset.Dir, file.SourcePath)
		if err != nil {
			return assetFingerprint{}, fmt.Errorf("resolve relative asset path %q: %w", file.SourcePath, err)
		}
		fingerprints = append(fingerprints, assetFileFingerprint{
			SourceRel:  filepath.ToSlash(rel),
			OutputPath: file.OutputPath,
			SHA256:     hashBytes(data),
		})
	}

	return assetFingerprint{
		Kind:  asset.Kind,
		Name:  asset.Name,
		Files: fingerprints,
	}, nil
}

func fingerprintInjectionRules(plan planner.Plan) ([]injectionFingerprint, error) {
	result := make([]injectionFingerprint, 0, len(plan.InjectionRules))
	for _, rule := range plan.InjectionRules {
		snippet, err := resolveSnippet(plan, rule)
		if err != nil {
			return nil, err
		}
		result = append(result, injectionFingerprint{
			Name:       rule.Name,
			Target:     rule.Target,
			Anchor:     rule.Anchor,
			Order:      rule.Order,
			Snippet:    rule.Snippet,
			SnippetSHA: hashBytes(snippet),
		})
	}
	return result, nil
}

func resolveSnippet(plan planner.Plan, rule manifest.InjectionRule) ([]byte, error) {
	for _, asset := range snippetAssets(plan, rule.Scope) {
		path := filepath.Join(asset.Dir, filepath.FromSlash(rule.Snippet))
		data, err := os.ReadFile(path)
		if err == nil {
			return data, nil
		}
	}
	return nil, fmt.Errorf("snippet %q for rule %q could not be resolved", rule.Snippet, rule.Name)
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

func normalizedRecipeForFingerprint(recipe Recipe) Recipe {
	normalized := recipe
	normalized.Capabilities = append([]string(nil), recipe.Capabilities...)
	sort.Strings(normalized.Capabilities)
	return normalized
}

func hashJSON(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("marshal fingerprint payload: %w", err)
	}
	return hashBytes(data), nil
}

func hashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func uniqueStrings(items []string) []string {
	if len(items) == 0 {
		return []string{}
	}
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	sort.Strings(result)
	return result
}

func renderCacheKey(recipe Recipe, catalogRoot string) (string, error) {
	payload := struct {
		Root   string `json:"root"`
		Recipe Recipe `json:"recipe"`
	}{
		Root:   filepath.Clean(catalogRoot),
		Recipe: normalizedRecipeForFingerprint(recipe),
	}
	return hashJSON(payload)
}

func lookupRenderCache(key string) ([]ManagedFile, string, bool) {
	renderCache.mu.Lock()
	defer renderCache.mu.Unlock()

	entry, ok := renderCache.entries[key]
	if !ok {
		return nil, "", false
	}
	return append([]ManagedFile(nil), entry.ManagedFiles...), entry.Fingerprint, true
}

func storeRenderCache(key string, managedFiles []ManagedFile, fingerprint string) {
	renderCache.mu.Lock()
	defer renderCache.mu.Unlock()

	if renderCache.entries == nil {
		renderCache.entries = make(map[string]renderCacheEntry)
	}
	renderCache.entries[key] = renderCacheEntry{
		ManagedFiles: append([]ManagedFile(nil), managedFiles...),
		Fingerprint:  fingerprint,
	}
}

func JoinCapabilities(items []string) string {
	if len(items) == 0 {
		return "(none)"
	}
	filtered := append([]string(nil), items...)
	sort.Strings(filtered)
	return strings.Join(filtered, ",")
}
