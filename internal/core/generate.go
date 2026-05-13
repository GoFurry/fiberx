package core

import (
	"path/filepath"
	"time"

	"github.com/gofurry/fiberx/internal/manifest"
	"github.com/gofurry/fiberx/internal/metadata"
	"github.com/gofurry/fiberx/internal/planner"
	"github.com/gofurry/fiberx/internal/postprocess"
	"github.com/gofurry/fiberx/internal/renderer"
	"github.com/gofurry/fiberx/internal/report"
	"github.com/gofurry/fiberx/internal/stack"
	"github.com/gofurry/fiberx/internal/validator"
	"github.com/gofurry/fiberx/internal/version"
	"github.com/gofurry/fiberx/internal/writer"
)

func Generate(req Request) error {
	_, err := Run(req)
	return err
}

func Preview(req Request) (report.Preview, error) {
	options := stack.NormalizeOptions(req.Options)
	if err := stack.ValidateOptions(options); err != nil {
		return report.Preview{}, err
	}

	catalogRoot := manifest.ResolveRoot(options["manifest_root"])

	catalog, err := manifest.LoadCatalog(catalogRoot)
	if err != nil {
		return report.Preview{}, err
	}

	if err := validator.ValidateCatalog(catalog); err != nil {
		return report.Preview{}, err
	}
	if err := validator.ValidateAssets(catalogRoot, catalog); err != nil {
		return report.Preview{}, err
	}

	if err := validator.ValidateRequest(req.ProjectName, req.ModulePath, req.Preset, req.Capabilities, options, catalog); err != nil {
		return report.Preview{}, err
	}

	preset, _ := catalog.FindPreset(req.Preset)
	selectedCapabilityNames := append([]string{}, catalog.AppliedDefaultCapabilities(preset)...)
	selectedCapabilityNames = append(selectedCapabilityNames, req.Capabilities...)
	selectedCapabilities := make([]manifest.CapabilityManifest, 0, len(selectedCapabilityNames))
	for _, name := range selectedCapabilityNames {
		capability, ok := catalog.FindCapability(name)
		if !ok {
			continue
		}
		alreadySelected := false
		for _, selected := range selectedCapabilities {
			if selected.Name == capability.Name {
				alreadySelected = true
				break
			}
		}
		if alreadySelected {
			continue
		}
		selectedCapabilities = append(selectedCapabilities, capability)
	}
	if err := validator.ValidateGenerationSupport(preset, selectedCapabilities); err != nil {
		return report.Preview{}, err
	}

	plan := planner.BuildPlan(req.ProjectName, req.ModulePath, req.Preset, req.Capabilities, options, catalogRoot, catalog)
	rendered, err := renderer.Render(plan)
	if err != nil {
		return report.Preview{}, err
	}

	return report.BuildPreview(
		plan,
		rendered,
		version.Version,
		version.Commit,
		filepath.ToSlash(filepath.Join(metadata.ManifestDir, metadata.ManifestFilename)),
	), nil
}

func Run(req Request) (report.Summary, error) {
	options := stack.NormalizeOptions(req.Options)
	if err := stack.ValidateOptions(options); err != nil {
		return report.Summary{}, err
	}

	catalogRoot := manifest.ResolveRoot(options["manifest_root"])

	catalog, err := manifest.LoadCatalog(catalogRoot)
	if err != nil {
		return report.Summary{}, err
	}

	if err := validator.ValidateCatalog(catalog); err != nil {
		return report.Summary{}, err
	}
	if err := validator.ValidateAssets(catalogRoot, catalog); err != nil {
		return report.Summary{}, err
	}

	if err := validator.ValidateRequest(req.ProjectName, req.ModulePath, req.Preset, req.Capabilities, options, catalog); err != nil {
		return report.Summary{}, err
	}

	preset, _ := catalog.FindPreset(req.Preset)
	selectedCapabilityNames := append([]string{}, catalog.AppliedDefaultCapabilities(preset)...)
	selectedCapabilityNames = append(selectedCapabilityNames, req.Capabilities...)
	selectedCapabilities := make([]manifest.CapabilityManifest, 0, len(selectedCapabilityNames))
	for _, name := range selectedCapabilityNames {
		capability, ok := catalog.FindCapability(name)
		if !ok {
			continue
		}
		alreadySelected := false
		for _, selected := range selectedCapabilities {
			if selected.Name == capability.Name {
				alreadySelected = true
				break
			}
		}
		if alreadySelected {
			continue
		}
		selectedCapabilities = append(selectedCapabilities, capability)
	}
	if err := validator.ValidateGenerationSupport(preset, selectedCapabilities); err != nil {
		return report.Summary{}, err
	}

	plan := planner.BuildPlan(req.ProjectName, req.ModulePath, req.Preset, req.Capabilities, options, catalogRoot, catalog)
	rendered, err := renderer.Render(plan)
	if err != nil {
		return report.Summary{}, err
	}
	writeResult, err := writer.New(options["target_dir"]).Write(rendered)
	if err != nil {
		return report.Summary{}, err
	}
	if err := postprocess.FinalizeGeneratedModule(writeResult.TargetDir); err != nil {
		return report.Summary{}, err
	}

	projectManifest, err := metadata.BuildManifest(plan, rendered, writeResult.TargetDir, time.Now())
	if err != nil {
		return report.Summary{}, err
	}
	if err := metadata.WriteManifest(writeResult.TargetDir, projectManifest); err != nil {
		return report.Summary{}, err
	}

	return report.Build(
		plan,
		rendered,
		writeResult,
		version.Version,
		version.Commit,
		projectManifest.Fingerprints.TemplateSet,
		projectManifest.Fingerprints.RenderedOutput,
		filepath.ToSlash(filepath.Join(metadata.ManifestDir, metadata.ManifestFilename)),
	), nil
}
