package report

import (
	"github.com/gofurry/fiberx/internal/planner"
	"github.com/gofurry/fiberx/internal/renderer"
	"github.com/gofurry/fiberx/internal/writer"
)

type Summary struct {
	Base                      string
	FiberVersion              string
	CLIStyle                  string
	Logger                    string
	Database                  string
	DataAccess                string
	JSONLib                   string
	GeneratorVersion          string
	GeneratorCommit           string
	TemplateSetFingerprint    string
	RenderedOutputFingerprint string
	MetadataPath              string
	PresetPacks               []string
	CapabilityPacks           []string
	RuntimeOverlays           []string
	Preset                    string
	Capabilities              []string
	ReplaceRules              []string
	InjectionRules            []string
	WrittenFiles              int
	WrittenPaths              []string
	Warnings                  []string
	DryRun                    bool
	TargetDir                 string
}

type Preview struct {
	Base             string   `json:"base"`
	FiberVersion     string   `json:"fiber_version"`
	CLIStyle         string   `json:"cli_style"`
	Logger           string   `json:"logger"`
	Database         string   `json:"database"`
	DataAccess       string   `json:"data_access"`
	JSONLib          string   `json:"json_lib"`
	ProjectName      string   `json:"project_name"`
	ModulePath       string   `json:"module_path"`
	TargetDir        string   `json:"target_dir"`
	MetadataPath     string   `json:"metadata_path"`
	Preset           string   `json:"preset"`
	Capabilities     []string `json:"capabilities"`
	PresetPacks      []string `json:"preset_packs"`
	CapabilityPacks  []string `json:"capability_packs"`
	RuntimeOverlays  []string `json:"runtime_overlays"`
	ReplaceRules     []string `json:"replace_rules"`
	InjectionRules   []string `json:"injection_rules"`
	EstimatedFiles   int      `json:"estimated_files"`
	EstimatedPaths   []string `json:"estimated_paths"`
	Warnings         []string `json:"warnings"`
	GeneratorVersion string   `json:"generator_version"`
	GeneratorCommit  string   `json:"generator_commit"`
}

func Build(plan planner.Plan, rendered renderer.Result, writeResult writer.Result, generatorVersion, generatorCommit, templateSetFingerprint, renderedOutputFingerprint, metadataPath string) Summary {
	capabilities := make([]string, 0, len(plan.Capabilities))
	for _, capability := range plan.Capabilities {
		capabilities = append(capabilities, capability.Name)
	}

	presetPacks := make([]string, 0, len(plan.PresetPacks))
	for _, pack := range plan.PresetPacks {
		presetPacks = append(presetPacks, pack.Name)
	}

	capabilityPacks := make([]string, 0, len(plan.CapabilityPacks))
	for _, pack := range plan.CapabilityPacks {
		capabilityPacks = append(capabilityPacks, pack.Name)
	}

	runtimeOverlays := make([]string, 0, len(plan.RuntimeOverlays))
	for _, pack := range plan.RuntimeOverlays {
		runtimeOverlays = append(runtimeOverlays, pack.Name)
	}

	return Summary{
		Base:                      plan.Base.Name,
		FiberVersion:              plan.FiberVersion,
		CLIStyle:                  plan.CLIStyle,
		Logger:                    plan.Logger,
		Database:                  plan.Database,
		DataAccess:                plan.DataAccess,
		JSONLib:                   plan.JSONLib,
		GeneratorVersion:          generatorVersion,
		GeneratorCommit:           generatorCommit,
		TemplateSetFingerprint:    templateSetFingerprint,
		RenderedOutputFingerprint: renderedOutputFingerprint,
		MetadataPath:              metadataPath,
		PresetPacks:               presetPacks,
		CapabilityPacks:           capabilityPacks,
		RuntimeOverlays:           runtimeOverlays,
		Preset:                    plan.Preset.Name,
		Capabilities:              capabilities,
		ReplaceRules:              append([]string(nil), rendered.ReplaceRuleHits...),
		InjectionRules:            append([]string(nil), rendered.InjectionHits...),
		WrittenFiles:              writeResult.WrittenFiles,
		WrittenPaths:              append([]string(nil), writeResult.WrittenPaths...),
		Warnings:                  append([]string(nil), rendered.Warnings...),
		DryRun:                    writeResult.DryRun,
		TargetDir:                 writeResult.TargetDir,
	}
}

func BuildPreview(plan planner.Plan, rendered renderer.Result, generatorVersion, generatorCommit, metadataPath string) Preview {
	capabilities := make([]string, 0, len(plan.Capabilities))
	for _, capability := range plan.Capabilities {
		capabilities = append(capabilities, capability.Name)
	}

	presetPacks := make([]string, 0, len(plan.PresetPacks))
	for _, pack := range plan.PresetPacks {
		presetPacks = append(presetPacks, pack.Name)
	}

	capabilityPacks := make([]string, 0, len(plan.CapabilityPacks))
	for _, pack := range plan.CapabilityPacks {
		capabilityPacks = append(capabilityPacks, pack.Name)
	}

	runtimeOverlays := make([]string, 0, len(plan.RuntimeOverlays))
	for _, pack := range plan.RuntimeOverlays {
		runtimeOverlays = append(runtimeOverlays, pack.Name)
	}

	estimatedPaths := make([]string, 0, len(rendered.Files))
	for _, file := range rendered.Files {
		estimatedPaths = append(estimatedPaths, file.Path)
	}

	return Preview{
		Base:             plan.Base.Name,
		FiberVersion:     plan.FiberVersion,
		CLIStyle:         plan.CLIStyle,
		Logger:           plan.Logger,
		Database:         plan.Database,
		DataAccess:       plan.DataAccess,
		JSONLib:          plan.JSONLib,
		ProjectName:      plan.ProjectName,
		ModulePath:       plan.ModulePath,
		TargetDir:        plan.TargetDir,
		MetadataPath:     metadataPath,
		Preset:           plan.Preset.Name,
		Capabilities:     capabilities,
		PresetPacks:      presetPacks,
		CapabilityPacks:  capabilityPacks,
		RuntimeOverlays:  runtimeOverlays,
		ReplaceRules:     append([]string(nil), rendered.ReplaceRuleHits...),
		InjectionRules:   append([]string(nil), rendered.InjectionHits...),
		EstimatedFiles:   len(rendered.Files),
		EstimatedPaths:   estimatedPaths,
		Warnings:         append([]string(nil), rendered.Warnings...),
		GeneratorVersion: generatorVersion,
		GeneratorCommit:  generatorCommit,
	}
}
