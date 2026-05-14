package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/gofurry/fiberx/internal/build"
	"github.com/gofurry/fiberx/internal/buildconfig"
	"github.com/gofurry/fiberx/internal/core"
	"github.com/gofurry/fiberx/internal/manifest"
	"github.com/gofurry/fiberx/internal/metadata"
	"github.com/gofurry/fiberx/internal/report"
	"github.com/gofurry/fiberx/internal/stack"
	"github.com/gofurry/fiberx/internal/upgrade"
	"github.com/gofurry/fiberx/internal/validator"
	"github.com/gofurry/fiberx/internal/version"
)

const (
	currentRelease   = "v0.1.4"
	nextRelease      = "v0.1.5"
	nextReleaseFocus = "release metadata sync, changelog and usage updates, CLI output alignment, and release-surface consistency"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		printUsage(os.Stdout)
		return nil
	}

	switch args[0] {
	case "new":
		return runNew(args[1:])
	case "init":
		return runInit(args[1:])
	case "list":
		return runList(args[1:])
	case "explain":
		return runExplain(args[1:])
	case "inspect":
		return runInspect(args[1:])
	case "diff":
		return runDiff(args[1:])
	case "upgrade":
		return runUpgrade(args[1:])
	case "build":
		return runBuild(args[1:])
	case "validate":
		return runValidate(args[1:])
	case "doctor":
		return runDoctor(args[1:])
	case "help", "-h", "--help":
		printUsage(os.Stdout)
		return nil
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runNew(args []string) error {
	fs := newFlagSet("new")
	modulePath := fs.String("module", "", "go module path")
	preset := fs.String("preset", "light", "preset name")
	with := fs.String("with", "", "comma-separated capability names")
	fiberVersion := fs.String("fiber-version", stack.DefaultFiberVersion(), "fiber version: v3 or v2")
	cliStyle := fs.String("cli-style", stack.DefaultCLIStyle(), "cli style: cobra or native")
	loggerBackend := fs.String("logger", stack.DefaultLogger(), "logger backend: zap or slog")
	dbKind := fs.String("db", stack.DefaultDB(), "database kind: sqlite, pgsql, or mysql")
	dataAccess := fs.String("data-access", stack.DefaultDataAccess(), "data access stack: stdlib, sqlx, or sqlc")
	jsonLib := fs.String("json-lib", stack.DefaultJSONLib(), "json backend: stdlib, sonic, or go-json")
	printPlan := fs.Bool("print-plan", false, "print the generation plan without writing files")
	asJSON := fs.Bool("json", false, "render generation plan output as JSON (requires --print-plan)")

	if err := fs.Parse(reorderArgs(args, map[string]bool{
		"--module":        true,
		"--preset":        true,
		"--with":          true,
		"--fiber-version": true,
		"--cli-style":     true,
		"--logger":        true,
		"--db":            true,
		"--data-access":   true,
		"--json-lib":      true,
		"--print-plan":    false,
		"--json":          false,
	})); err != nil {
		return err
	}
	if *asJSON && !*printPlan {
		return errors.New("--json requires --print-plan")
	}

	positionals := fs.Args()
	if len(positionals) != 1 {
		return errors.New("new requires exactly one project name")
	}

	projectName := positionals[0]
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	targetDir := filepath.Join(cwd, projectName)
	options := map[string]string{
		"command":                "new",
		"output_mode":            "new",
		"target_dir":             targetDir,
		stack.OptionFiberVersion: *fiberVersion,
		stack.OptionCLIStyle:     *cliStyle,
	}
	setOptionalRuntimeFlags(fs, options, *loggerBackend, *dbKind, *dataAccess, *jsonLib)
	req := core.Request{
		ProjectName:  projectName,
		ModulePath:   defaultModulePath(projectName, *modulePath),
		Preset:       *preset,
		Capabilities: parseCapabilities(*with),
		Options:      options,
	}

	if *printPlan {
		preview, err := core.Preview(req)
		if err != nil {
			return err
		}
		if *asJSON {
			return writeJSON(os.Stdout, preview)
		}
		printPreview(os.Stdout, preview)
		return nil
	}

	summary, err := core.Run(req)
	if err != nil {
		return err
	}

	printSummary(os.Stdout, summary)
	return nil
}

func runInit(args []string) error {
	fs := newFlagSet("init")
	name := fs.String("name", "", "project name override")
	modulePath := fs.String("module", "", "go module path")
	preset := fs.String("preset", "light", "preset name")
	with := fs.String("with", "", "comma-separated capability names")
	fiberVersion := fs.String("fiber-version", stack.DefaultFiberVersion(), "fiber version: v3 or v2")
	cliStyle := fs.String("cli-style", stack.DefaultCLIStyle(), "cli style: cobra or native")
	loggerBackend := fs.String("logger", stack.DefaultLogger(), "logger backend: zap or slog")
	dbKind := fs.String("db", stack.DefaultDB(), "database kind: sqlite, pgsql, or mysql")
	dataAccess := fs.String("data-access", stack.DefaultDataAccess(), "data access stack: stdlib, sqlx, or sqlc")
	jsonLib := fs.String("json-lib", stack.DefaultJSONLib(), "json backend: stdlib, sonic, or go-json")
	printPlan := fs.Bool("print-plan", false, "print the generation plan without writing files")
	asJSON := fs.Bool("json", false, "render generation plan output as JSON (requires --print-plan)")

	if err := fs.Parse(reorderArgs(args, map[string]bool{
		"--name":          true,
		"--module":        true,
		"--preset":        true,
		"--with":          true,
		"--fiber-version": true,
		"--cli-style":     true,
		"--logger":        true,
		"--db":            true,
		"--data-access":   true,
		"--json-lib":      true,
		"--print-plan":    false,
		"--json":          false,
	})); err != nil {
		return err
	}
	if *asJSON && !*printPlan {
		return errors.New("--json requires --print-plan")
	}

	if len(fs.Args()) != 0 {
		return errors.New("init does not accept positional arguments")
	}

	projectName := *name
	if projectName == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		projectName = filepath.Base(cwd)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	options := map[string]string{
		"command":                "init",
		"output_mode":            "init",
		"target_dir":             cwd,
		stack.OptionFiberVersion: *fiberVersion,
		stack.OptionCLIStyle:     *cliStyle,
	}
	setOptionalRuntimeFlags(fs, options, *loggerBackend, *dbKind, *dataAccess, *jsonLib)

	req := core.Request{
		ProjectName:  projectName,
		ModulePath:   defaultModulePath(projectName, *modulePath),
		Preset:       *preset,
		Capabilities: parseCapabilities(*with),
		Options:      options,
	}

	if *printPlan {
		preview, err := core.Preview(req)
		if err != nil {
			return err
		}
		if *asJSON {
			return writeJSON(os.Stdout, preview)
		}
		printPreview(os.Stdout, preview)
		return nil
	}

	summary, err := core.Run(req)
	if err != nil {
		return err
	}

	printSummary(os.Stdout, summary)
	return nil
}

func runList(args []string) error {
	if len(args) != 1 {
		return errors.New("list requires one target: presets or capabilities")
	}

	catalog, err := loadCatalog()
	if err != nil {
		return err
	}

	switch args[0] {
	case "presets":
		for _, preset := range catalog.Presets {
			fmt.Printf("%s\timplemented=%t\t%s\n", preset.Name, preset.Implemented, preset.Summary)
		}
		return nil
	case "capabilities":
		for _, capability := range catalog.Capabilities {
			fmt.Printf("%s\timplemented=%t\t%s\n", capability.Name, capability.Implemented, capability.Summary)
		}
		return nil
	default:
		return fmt.Errorf("unknown list target %q", args[0])
	}
}

func runExplain(args []string) error {
	fs := newFlagSet("explain")
	asJSON := fs.Bool("json", false, "render explain matrix output as JSON")
	if err := fs.Parse(reorderArgs(args, map[string]bool{"--json": false})); err != nil {
		return err
	}
	args = fs.Args()
	if len(args) == 1 && args[0] == "matrix" {
		catalog, err := loadCatalog()
		if err != nil {
			return err
		}
		matrix := buildCapabilityMatrix(catalog)
		if *asJSON {
			return writeJSON(os.Stdout, matrix)
		}
		printCapabilityMatrix(os.Stdout, matrix)
		return nil
	}
	if *asJSON {
		return errors.New("--json is only supported with explain matrix")
	}
	if len(args) != 2 {
		return errors.New("explain requires a kind and a name")
	}

	catalog, err := loadCatalog()
	if err != nil {
		return err
	}

	switch args[0] {
	case "preset":
		preset, ok := catalog.FindPreset(args[1])
		if !ok {
			return fmt.Errorf("unknown preset %q", args[1])
		}
		fmt.Printf("preset: %s\nsummary: %s\ndescription: %s\nimplemented: %t\nbase: %s\npacks: %s\ndefault_capabilities: %s\nallowed_capabilities: %s\ndefault_stack: %s\ndefault_logger: %s\ndefault_database: %s\ndefault_data_access: %s\nsupported_fiber_versions: %s\nsupported_cli_styles: %s\n", preset.Name, preset.Summary, preset.Description, preset.Implemented, joinOrNone([]string{preset.Base}), joinOrNone(preset.Packs), joinOrNone(preset.DefaultCapabilities), joinOrNone(preset.AllowedCapabilities), stack.DefaultStackLabel(), defaultLoggerForPreset(preset.Name), defaultDatabaseForPreset(preset.Name), defaultDataAccessForPreset(preset.Name), stack.SupportedFiberVersions(), stack.SupportedCLIStyles())
		if preset.Name == "extra-light" {
			fmt.Println("phase11_runtime_options: unsupported")
		} else {
			fmt.Printf("supported_loggers: %s\nsupported_databases: %s\nsupported_data_access: %s\n", stack.SupportedLoggers(), stack.SupportedDatabases(), stack.SupportedDataAccess())
		}
		return nil
	case "capability":
		capability, ok := catalog.FindCapability(args[1])
		if !ok {
			return fmt.Errorf("unknown capability %q", args[1])
		}
		defaultOn, optionalOn, unsupportedOn := capabilityPresetBoundary(catalog, capability)
		fmt.Printf("capability: %s\nsummary: %s\ndescription: %s\nimplemented: %t\npacks: %s\nallowed_presets: %s\ndefault_on_presets: %s\noptional_on_presets: %s\nunsupported_on_presets: %s\ndepends_on: %s\nconflicts_with: %s\n", capability.Name, capability.Summary, capability.Description, capability.Implemented, joinOrNone(capability.Packs), joinOrNone(orderNames(capability.AllowedPresets, []string{"heavy", "medium", "light", "extra-light"})), joinOrNone(defaultOn), joinOrNone(optionalOn), joinOrNone(unsupportedOn), joinOrNone(capability.DependsOn), joinOrNone(capability.ConflictsWith))
		return nil
	default:
		return fmt.Errorf("unknown explain target %q", args[0])
	}
}

func runValidate(args []string) error {
	fs := newFlagSet("validate")
	verbose := fs.Bool("verbose", false, "show full diagnostics")
	if err := fs.Parse(reorderArgs(args, map[string]bool{})); err != nil {
		return err
	}
	if len(fs.Args()) != 0 {
		return errors.New("validate does not accept positional arguments")
	}

	catalog, err := loadCatalog()
	if err != nil {
		return err
	}
	if err := validator.ValidateCatalog(catalog); err != nil {
		return err
	}
	if err := validator.ValidateAssets(manifest.ResolveRoot(""), catalog); err != nil {
		return err
	}

	if !*verbose {
		fmt.Println("fiberx validate: ok")
		fmt.Printf("release: %s\n", currentRelease)
		fmt.Printf("generator: %s (%s)\n", version.Version, version.Commit)
		fmt.Printf("presets: %s\n", joinOrNone(implementedPresets(catalog)))
		fmt.Printf("capabilities: %s\n", joinOrNone(implementedCapabilities(catalog)))
		fmt.Printf("default stack: %s\n", stack.DefaultStackLabel())
		fmt.Println("note: use --verbose for full diagnostics")
		return nil
	}

	printValidateVerbose(os.Stdout, catalog)
	return nil
}

func runDoctor(args []string) error {
	fs := newFlagSet("doctor")
	verbose := fs.Bool("verbose", false, "show full diagnostics")
	if err := fs.Parse(reorderArgs(args, map[string]bool{})); err != nil {
		return err
	}
	if len(fs.Args()) != 0 {
		return errors.New("doctor does not accept positional arguments")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	root := manifest.ResolveRoot("")
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return err
	}

	projectRoot := detectProjectRoot(cwd)
	doctorMode := detectDoctorMode(cwd, rootAbs)
	if doctorMode == "project" {
		return runProjectDoctor(projectRoot, rootAbs, *verbose)
	}
	if doctorMode == "standalone" {
		if !*verbose {
			fmt.Println("fiberx doctor")
			fmt.Println("mode: standalone")
			fmt.Printf("generator: %s (%s)\n", version.Version, version.Commit)
			fmt.Printf("release: %s\n", currentRelease)
			fmt.Printf("go: %s\n", runtime.Version())
			fmt.Printf("workspace: %s\n", cwd)
			fmt.Println("status: no generator repository or generated project detected")
			fmt.Println("note: use --verbose for full diagnostics")
			return nil
		}
		printSection(os.Stdout, "environment")
		fmt.Fprintf(os.Stdout, "mode: standalone\ncwd: %s\ngo: %s\nrelease: %s\n", cwd, runtime.Version(), currentRelease)
		fmt.Fprintln(os.Stdout)
		printSection(os.Stdout, "generator")
		fmt.Fprintf(os.Stdout, "generator-version: %s\ngenerator-commit: %s\n", version.Version, version.Commit)
		fmt.Fprintln(os.Stdout)
		printSection(os.Stdout, "status")
		fmt.Fprintln(os.Stdout, "status: no generator repository or generated project detected")
		return nil
	}

	catalog, err := manifest.LoadCatalog(root)
	if err != nil {
		return err
	}

	if !*verbose {
		fmt.Println("fiberx doctor")
		fmt.Printf("mode: %s\n", doctorMode)
		fmt.Printf("generator: %s (%s)\n", version.Version, version.Commit)
		fmt.Printf("release: %s\n", currentRelease)
		fmt.Printf("go: %s\n", runtime.Version())
		fmt.Printf("workspace: %s\n", cwd)
		fmt.Printf("manifest root: %s\n", rootAbs)
		fmt.Println("status: ok")
		fmt.Println("note: use --verbose for full diagnostics")
		return nil
	}

	printDoctorVerbose(os.Stdout, cwd, rootAbs, catalog)
	return nil
}

func runInspect(args []string) error {
	fs := newFlagSet("inspect")
	asJSON := fs.Bool("json", false, "render inspect output as JSON")
	if err := fs.Parse(reorderArgs(args, map[string]bool{"--json": true})); err != nil {
		return err
	}

	projectDir, err := resolveProjectDir(fs.Args())
	if err != nil {
		return err
	}
	projectManifest, err := metadata.LoadManifest(projectDir)
	if err != nil {
		return err
	}

	if *asJSON {
		return writeJSON(os.Stdout, projectManifest)
	}

	fmt.Printf("project: %s\n", projectDir)
	fmt.Printf("metadata: %s\n", filepath.ToSlash(filepath.Join(metadata.ManifestDir, metadata.ManifestFilename)))
	fmt.Printf("generated-at: %s\n", projectManifest.GeneratedAt)
	fmt.Printf("generator-version: %s\n", projectManifest.Generator.Version)
	fmt.Printf("generator-commit: %s\n", projectManifest.Generator.Commit)
	fmt.Printf("preset: %s\n", projectManifest.Recipe.Preset)
	fmt.Printf("capabilities: %s\n", joinOrNone(projectManifest.Recipe.Capabilities))
	fmt.Printf("stack: fiber-%s + %s\n", projectManifest.Recipe.FiberVersion, projectManifest.Recipe.CLIStyle)
	if projectManifest.Recipe.Logger != "" || projectManifest.Recipe.DB != "" || projectManifest.Recipe.DataAccess != "" || projectManifest.Recipe.JSONLib != "" {
		fmt.Printf("runtime: logger=%s db=%s data-access=%s json-lib=%s\n", valueOrNone(projectManifest.Recipe.Logger), valueOrNone(projectManifest.Recipe.DB), valueOrNone(projectManifest.Recipe.DataAccess), valueOrNone(projectManifest.Recipe.JSONLib))
	}
	fmt.Printf("base: %s\n", projectManifest.Assets.Base)
	fmt.Printf("preset packs: %s\n", joinOrNone(projectManifest.Assets.PresetPacks))
	fmt.Printf("capability packs: %s\n", joinOrNone(projectManifest.Assets.CapabilityPacks))
	fmt.Printf("runtime overlays: %s\n", joinOrNone(projectManifest.Assets.RuntimeOverlays))
	fmt.Printf("replace rules: %s\n", joinOrNone(projectManifest.Assets.ReplaceRules))
	fmt.Printf("injection rules: %s\n", joinOrNone(projectManifest.Assets.InjectionRules))
	fmt.Printf("template fingerprint: %s\n", projectManifest.Fingerprints.TemplateSet)
	fmt.Printf("rendered fingerprint: %s\n", projectManifest.Fingerprints.RenderedOutput)
	fmt.Printf("managed files: %d\n", len(projectManifest.ManagedFiles))
	return nil
}

func runDiff(args []string) error {
	fs := newFlagSet("diff")
	asJSON := fs.Bool("json", false, "render diff output as JSON")
	if err := fs.Parse(reorderArgs(args, map[string]bool{"--json": true})); err != nil {
		return err
	}

	projectDir, err := resolveProjectDir(fs.Args())
	if err != nil {
		return err
	}

	diffReport, err := metadata.BuildDiff(projectDir, manifest.ResolveRoot(""))
	if err != nil {
		return err
	}

	if *asJSON {
		return writeJSON(os.Stdout, diffReport)
	}

	fmt.Printf("project: %s\n", projectDir)
	fmt.Printf("status: %s\n", diffReport.Status)
	fmt.Printf("generated-by: %s (%s)\n", diffReport.Generator.Generated.Version, diffReport.Generator.Generated.Commit)
	fmt.Printf("current-generator: %s (%s)\n", diffReport.Generator.Current.Version, diffReport.Generator.Current.Commit)
	fmt.Printf("preset: %s\n", diffReport.Recipe.Preset)
	fmt.Printf("capabilities: %s\n", joinOrNone(diffReport.Recipe.Capabilities))
	if diffReport.Recipe.Logger != "" || diffReport.Recipe.DB != "" || diffReport.Recipe.DataAccess != "" || diffReport.Recipe.JSONLib != "" {
		fmt.Printf("runtime: logger=%s db=%s data-access=%s json-lib=%s\n", valueOrNone(diffReport.Recipe.Logger), valueOrNone(diffReport.Recipe.DB), valueOrNone(diffReport.Recipe.DataAccess), valueOrNone(diffReport.Recipe.JSONLib))
	}
	fmt.Printf("missing files: %s\n", joinOrNone(diffReport.MissingFiles))
	fmt.Printf("changed files: %s\n", joinOrNone(diffReport.ChangedFiles))
	fmt.Printf("new managed files: %s\n", joinOrNone(diffReport.NewManagedFiles))
	fmt.Printf("generator drift files: %s\n", joinOrNone(diffReport.GeneratorDriftFiles))
	return nil
}

func runUpgrade(args []string) error {
	if len(args) == 0 {
		return errors.New("upgrade requires a subcommand: inspect or plan")
	}

	switch args[0] {
	case "inspect":
		return runUpgradeInspect(args[1:])
	case "plan":
		return runUpgradePlan(args[1:])
	default:
		return fmt.Errorf("unknown upgrade subcommand %q", args[0])
	}
}

func runUpgradeInspect(args []string) error {
	fs := newFlagSet("upgrade inspect")
	asJSON := fs.Bool("json", false, "render upgrade inspect output as JSON")
	if err := fs.Parse(reorderArgs(args, map[string]bool{"--json": true})); err != nil {
		return err
	}

	projectDir, err := resolveProjectDir(fs.Args())
	if err != nil {
		return err
	}

	assessment, err := upgrade.Inspect(projectDir, manifest.ResolveRoot(""))
	if err != nil {
		return err
	}

	if *asJSON {
		return writeJSON(os.Stdout, assessment)
	}

	fmt.Printf("project: %s\n", assessment.ProjectDir)
	fmt.Printf("generated-by: %s (%s)\n", assessment.GeneratedGenerator.Version, assessment.GeneratedGenerator.Commit)
	fmt.Printf("current-generator: %s (%s)\n", assessment.CurrentGenerator.Version, assessment.CurrentGenerator.Commit)
	fmt.Printf("preset: %s\n", assessment.Recipe.Preset)
	fmt.Printf("capabilities: %s\n", joinOrNone(assessment.Recipe.Capabilities))
	fmt.Printf("stack: fiber-%s + %s\n", assessment.Recipe.FiberVersion, assessment.Recipe.CLIStyle)
	if assessment.Recipe.Logger != "" || assessment.Recipe.DB != "" || assessment.Recipe.DataAccess != "" || assessment.Recipe.JSONLib != "" {
		fmt.Printf("runtime: logger=%s db=%s data-access=%s json-lib=%s\n", valueOrNone(assessment.Recipe.Logger), valueOrNone(assessment.Recipe.DB), valueOrNone(assessment.Recipe.DataAccess), valueOrNone(assessment.Recipe.JSONLib))
	}
	fmt.Printf("diff status: %s\n", assessment.DiffStatus)
	fmt.Printf("compatibility level: %s\n", assessment.CompatibilityLevel)
	fmt.Printf("reasons: %s\n", joinOrNone(assessment.Reasons))
	fmt.Printf("blocking issues: %s\n", joinOrNone(assessment.BlockingIssues))
	fmt.Printf("local modified files: %s\n", joinOrNone(assessment.LocalModifiedFiles))
	fmt.Printf("generator drift files: %s\n", joinOrNone(assessment.GeneratorDriftFiles))
	fmt.Printf("missing files: %s\n", joinOrNone(assessment.MissingFiles))
	fmt.Printf("new managed files: %s\n", joinOrNone(assessment.NewManagedFiles))
	return nil
}

func runUpgradePlan(args []string) error {
	fs := newFlagSet("upgrade plan")
	asJSON := fs.Bool("json", false, "render upgrade plan output as JSON")
	if err := fs.Parse(reorderArgs(args, map[string]bool{"--json": true})); err != nil {
		return err
	}

	projectDir, err := resolveProjectDir(fs.Args())
	if err != nil {
		return err
	}

	upgradePlan, err := upgrade.Plan(projectDir, manifest.ResolveRoot(""))
	if err != nil {
		return err
	}

	if *asJSON {
		return writeJSON(os.Stdout, upgradePlan)
	}

	fmt.Printf("project: %s\n", upgradePlan.Assessment.ProjectDir)
	fmt.Printf("compatibility level: %s\n", upgradePlan.Assessment.CompatibilityLevel)
	fmt.Printf("diff status: %s\n", upgradePlan.Assessment.DiffStatus)
	fmt.Printf("upgrade summary: %s\n", joinOrNone(upgradePlan.Assessment.Reasons))
	fmt.Printf("blocking issues: %s\n", joinOrNone(upgradePlan.Assessment.BlockingIssues))
	fmt.Printf("local modified files: %s\n", joinOrNone(upgradePlan.Assessment.LocalModifiedFiles))
	fmt.Printf("generator drift files: %s\n", joinOrNone(upgradePlan.Assessment.GeneratorDriftFiles))
	fmt.Printf("missing files: %s\n", joinOrNone(upgradePlan.Assessment.MissingFiles))
	fmt.Printf("new managed files: %s\n", joinOrNone(upgradePlan.Assessment.NewManagedFiles))
	fmt.Printf("managed files to review: %s\n", joinOrNone(upgradePlan.ManagedFilesToReview))
	if len(upgradePlan.RecommendedSteps) == 0 {
		fmt.Println("recommended steps: (none)")
		return nil
	}
	fmt.Println("recommended steps:")
	for index, step := range upgradePlan.RecommendedSteps {
		fmt.Printf("  %d. %s\n", index+1, step)
	}
	return nil
}

func runBuild(args []string) error {
	fs := newFlagSet("build")
	clean := fs.Bool("clean", false, "clean the output directory before building")
	platform := fs.String("target", "", "filter builds to a single goos/goarch platform")
	dryRun := fs.Bool("dry-run", false, "print the build plan without writing outputs")
	profile := fs.String("profile", "", "apply a named build profile overlay")
	noHooks := fs.Bool("no-hooks", false, "skip target hooks during the build")
	autoApprove := fs.Bool("yes", false, "approve execution of planned hooks without prompting")
	if err := fs.Parse(reorderArgs(args, map[string]bool{"--target": true, "--profile": true, "--no-hooks": false, "--yes": false})); err != nil {
		return err
	}

	projectDir, err := os.Getwd()
	if err != nil {
		return err
	}

	cfg, err := buildconfig.LoadWithProfile(projectDir, *profile)
	if err != nil {
		return err
	}

	previewResult, err := build.Execute(projectDir, cfg, build.Options{
		TargetNames:    fs.Args(),
		PlatformFilter: *platform,
		Clean:          *clean,
		DryRun:         true,
		Profile:        *profile,
	})
	if err != nil {
		return err
	}

	hooksPresent := resultHasHooks(previewResult)
	confirmationRequired := hooksPresent && !*dryRun && !*noHooks && !*autoApprove
	if confirmationRequired {
		if !isInteractiveTerminal() {
			return errors.New("build hooks are present; rerun with --yes to approve them or --no-hooks to skip them")
		}
		if !promptForHookApproval(previewResult) {
			return errors.New("build aborted by user")
		}
	}

	result, err := build.Execute(projectDir, cfg, build.Options{
		TargetNames:    fs.Args(),
		PlatformFilter: *platform,
		Clean:          *clean,
		DryRun:         *dryRun,
		Profile:        *profile,
		NoHooks:        *noHooks,
		AutoApprove:    *autoApprove,
	})
	if err != nil {
		return err
	}

	if result.DryRun {
		fmt.Printf("build plan project=%s out_dir=%s\n", cfg.Project.Name, filepath.ToSlash(result.OutDir))
	} else {
		fmt.Printf("built project=%s out_dir=%s\n", cfg.Project.Name, filepath.ToSlash(result.OutDir))
	}
	fmt.Printf("version: %s\n", result.Version.Version)
	fmt.Printf("commit: %s\n", result.Version.Commit)
	fmt.Printf("build time: %s\n", result.Version.BuildTime)
	fmt.Printf("profile: %s\n", valueOrNone(result.Profile))
	fmt.Printf("dry-run: %t\n", result.DryRun)
	fmt.Printf("artifacts: %d\n", len(result.Artifacts))
	for _, artifact := range result.Artifacts {
		fmt.Printf("  - target=%s platform=%s output=%s", artifact.TargetName, artifact.Platform, filepath.ToSlash(artifact.OutputPath))
		if artifact.ArchivePath != "" {
			fmt.Printf(" archive=%s", filepath.ToSlash(artifact.ArchivePath))
		}
		if artifact.DistributablePath != "" && artifact.DistributablePath != artifact.OutputPath {
			fmt.Printf(" distributable=%s", filepath.ToSlash(artifact.DistributablePath))
		}
		if len(artifact.PreHooks) > 0 {
			fmt.Printf(" pre_hooks=%s", joinOrNone(artifact.PreHooks))
		}
		if len(artifact.PostHooks) > 0 {
			fmt.Printf(" post_hooks=%s", joinOrNone(artifact.PostHooks))
		}
		if artifact.UPXEnabled {
			fmt.Printf(" upx=enabled(level=%d)", artifact.UPXLevel)
		}
		fmt.Println()
	}
	if hooksPresent {
		switch {
		case *noHooks:
			fmt.Println("hooks: skipped by --no-hooks")
		case *dryRun:
			fmt.Printf("hooks: planned; %s\n", hookActionSummary(previewResult, confirmationRequired))
		case *autoApprove:
			fmt.Println("hooks: approved by --yes")
		default:
			fmt.Println("hooks: executed")
		}
	}
	if result.ChecksumPath != "" {
		fmt.Printf("checksum: %s\n", filepath.ToSlash(result.ChecksumPath))
	}
	fmt.Printf("build metadata: %s\n", filepath.ToSlash(result.BuildMetadataPath))
	fmt.Printf("release manifest: %s\n", filepath.ToSlash(result.ReleaseManifestPath))
	return nil
}

func newFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "fiberx is a CLI-first Fiber project generator.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  fiberx new <name> [--module path] [--preset name] [--with cap1,cap2] [--fiber-version v3|v2] [--cli-style cobra|native] [--logger zap|slog] [--db sqlite|pgsql|mysql] [--data-access stdlib|sqlx|sqlc] [--json-lib stdlib|sonic|go-json] [--print-plan] [--json]")
	fmt.Fprintln(w, "  fiberx init [--name name] [--module path] [--preset name] [--with cap1,cap2] [--fiber-version v3|v2] [--cli-style cobra|native] [--logger zap|slog] [--db sqlite|pgsql|mysql] [--data-access stdlib|sqlx|sqlc] [--json-lib stdlib|sonic|go-json] [--print-plan] [--json]")
	fmt.Fprintln(w, "  fiberx list presets")
	fmt.Fprintln(w, "  fiberx list capabilities")
	fmt.Fprintln(w, "  fiberx explain preset <name>")
	fmt.Fprintln(w, "  fiberx explain capability <name>")
	fmt.Fprintln(w, "  fiberx explain matrix [--json]")
	fmt.Fprintln(w, "  fiberx inspect [path] [--json]")
	fmt.Fprintln(w, "  fiberx diff [path] [--json]")
	fmt.Fprintln(w, "  fiberx upgrade inspect [path] [--json]")
	fmt.Fprintln(w, "  fiberx upgrade plan [path] [--json]")
	fmt.Fprintln(w, "  fiberx build [target...] [--clean] [--target goos/goarch] [--profile name] [--dry-run] [--no-hooks] [--yes]")
	fmt.Fprintln(w, "  fiberx validate [--verbose]")
	fmt.Fprintln(w, "  fiberx doctor [--verbose]")
	fmt.Fprintf(w, "\nDefault stack: %s\n", stack.DefaultStackLabel())
	fmt.Fprintf(w, "Default logger/database/data access: %s / %s / %s\n", stack.DefaultLogger(), stack.DefaultDB(), stack.DefaultDataAccess())
	fmt.Fprintf(w, "Default JSON backend: %s\n", stack.DefaultJSONLib())
	fmt.Fprintln(w, "Capability policy: swagger and embedded-ui default on medium/heavy, optional on light; redis optional on medium/heavy only.")
	fmt.Fprintf(w, "Release: %s completed.\n", currentRelease)
	fmt.Fprintf(w, "Current milestone: %s in progress for %s.\n", nextRelease, nextReleaseFocus)
	fmt.Fprintln(w, "Use `fiberx doctor --verbose` or `fiberx validate --verbose` for full diagnostics.")
}

func printValidateVerbose(w io.Writer, catalog manifest.Catalog) {
	printSection(w, "summary")
	fmt.Fprintf(w, "state 4 generator validated successfully: presets=%d capabilities=%d replace_rules=%d injection_rules=%d\n", len(catalog.Presets), len(catalog.Capabilities), len(catalog.ReplaceRules), len(catalog.InjectionRules))
	fmt.Fprintf(w, "implemented presets: %s\n", joinOrNone(implementedPresets(catalog)))
	fmt.Fprintf(w, "implemented capabilities: %s\n", joinOrNone(implementedCapabilities(catalog)))
	fmt.Fprintf(w, "deferred capabilities: %s\n", joinOrNone(deferredCapabilities(catalog)))
	fmt.Fprintln(w, "stable production baseline: medium")
	fmt.Fprintln(w, "completed production track: heavy")
	fmt.Fprintln(w)
	printSection(w, "release")
	fmt.Fprintf(w, "release: %s\n", currentRelease)
	fmt.Fprintf(w, "current milestone: %s\n", nextRelease)
	fmt.Fprintln(w, "generator mainline: pure generator repository")
	fmt.Fprintln(w)
	printSection(w, "capabilities")
	fmt.Fprintln(w, "default medium experience: swagger,embedded-ui")
	fmt.Fprintln(w, "default heavy experience: swagger,embedded-ui")
	fmt.Fprintln(w, "light optional experience: swagger,embedded-ui")
	fmt.Fprintln(w, "extra-light optional experience: none")
	printCapabilityPolicy(w, catalog)
	fmt.Fprintln(w)
	printSection(w, "runtime")
	fmt.Fprintf(w, "default stack: %s\n", stack.DefaultStackLabel())
	fmt.Fprintf(w, "supported fiber versions: %s\n", stack.SupportedFiberVersions())
	fmt.Fprintf(w, "supported cli styles: %s\n", stack.SupportedCLIStyles())
	fmt.Fprintf(w, "default logger: %s\n", stack.DefaultLogger())
	fmt.Fprintf(w, "default database: %s\n", stack.DefaultDB())
	fmt.Fprintf(w, "default data access: %s\n", stack.DefaultDataAccess())
	fmt.Fprintf(w, "default json lib: %s\n", stack.DefaultJSONLib())
	fmt.Fprintf(w, "supported loggers: %s\n", stack.SupportedLoggers())
	fmt.Fprintf(w, "supported databases: %s\n", stack.SupportedDatabases())
	fmt.Fprintf(w, "supported data access: %s\n", stack.SupportedDataAccess())
	fmt.Fprintf(w, "supported json libs: %s\n", stack.SupportedJSONLibs())
	fmt.Fprintln(w, "runtime-option presets: medium,heavy,light")
	fmt.Fprintln(w, "runtime-option unavailable presets: extra-light")
}

func printDoctorVerbose(w io.Writer, cwd string, rootAbs string, catalog manifest.Catalog) {
	printSection(w, "environment")
	fmt.Fprintf(w, "cwd: %s\n", cwd)
	fmt.Fprintf(w, "go: %s\n", runtime.Version())
	fmt.Fprintf(w, "release: %s\n", currentRelease)
	fmt.Fprintf(w, "current milestone: %s\n", nextRelease)
	fmt.Fprintf(w, "manifest-root: %s\n", rootAbs)
	fmt.Fprintln(w)
	printSection(w, "catalog")
	fmt.Fprintf(w, "presets: %d\n", len(catalog.Presets))
	fmt.Fprintf(w, "capabilities: %d\n", len(catalog.Capabilities))
	fmt.Fprintf(w, "replace-rules: %d\n", len(catalog.ReplaceRules))
	fmt.Fprintf(w, "injection-rules: %d\n", len(catalog.InjectionRules))
	fmt.Fprintf(w, "implemented-presets: %s\n", joinOrNone(implementedPresets(catalog)))
	fmt.Fprintf(w, "implemented-capabilities: %s\n", joinOrNone(implementedCapabilities(catalog)))
	fmt.Fprintf(w, "deferred-capabilities: %s\n", joinOrNone(deferredCapabilities(catalog)))
	fmt.Fprintln(w)
	printSection(w, "capability policy")
	fmt.Fprintf(w, "default-medium-capabilities: %s\n", "swagger,embedded-ui")
	fmt.Fprintf(w, "default-heavy-capabilities: %s\n", "swagger,embedded-ui")
	fmt.Fprintf(w, "light-optional-capabilities: %s\n", "swagger,embedded-ui")
	fmt.Fprintf(w, "extra-light-optional-capabilities: %s\n", "none")
	printCapabilityPolicy(w, catalog)
	fmt.Fprintln(w)
	printSection(w, "runtime")
	fmt.Fprintf(w, "default-stack: %s\n", stack.DefaultStackLabel())
	fmt.Fprintf(w, "supported-fiber-versions: %s\n", stack.SupportedFiberVersions())
	fmt.Fprintf(w, "supported-cli-styles: %s\n", stack.SupportedCLIStyles())
	fmt.Fprintf(w, "default-logger: %s\n", stack.DefaultLogger())
	fmt.Fprintf(w, "default-database: %s\n", stack.DefaultDB())
	fmt.Fprintf(w, "default-data-access: %s\n", stack.DefaultDataAccess())
	fmt.Fprintf(w, "default-json-lib: %s\n", stack.DefaultJSONLib())
	fmt.Fprintf(w, "supported-loggers: %s\n", stack.SupportedLoggers())
	fmt.Fprintf(w, "supported-databases: %s\n", stack.SupportedDatabases())
	fmt.Fprintf(w, "supported-data-access: %s\n", stack.SupportedDataAccess())
	fmt.Fprintf(w, "supported-json-libs: %s\n", stack.SupportedJSONLibs())
	fmt.Fprintln(w)
	printSection(w, "generator")
	fmt.Fprintf(w, "generator-version: %s\n", version.Version)
	fmt.Fprintf(w, "generator-commit: %s\n", version.Commit)
	fmt.Fprintln(w, "writer-mode: real-write")
}

func loadCatalog() (manifest.Catalog, error) {
	return manifest.LoadCatalog(manifest.ResolveRoot(""))
}

func implementedPresets(catalog manifest.Catalog) []string {
	names := make([]string, 0, len(catalog.Presets))
	for _, preset := range catalog.Presets {
		if preset.Implemented {
			names = append(names, preset.Name)
		}
	}
	return orderNames(names, []string{"heavy", "medium", "light", "extra-light"})
}

func implementedCapabilities(catalog manifest.Catalog) []string {
	names := make([]string, 0, len(catalog.Capabilities))
	for _, capability := range catalog.Capabilities {
		if capability.Implemented {
			names = append(names, capability.Name)
		}
	}
	return orderNames(names, []string{"redis", "swagger", "embedded-ui"})
}

func deferredCapabilities(catalog manifest.Catalog) []string {
	names := make([]string, 0, len(catalog.Capabilities))
	for _, capability := range catalog.Capabilities {
		if capability.Implemented {
			continue
		}
		names = append(names, capability.Name)
	}
	return orderNames(names, []string{"swagger", "embedded-ui", "redis"})
}

func joinOrNone(items []string) string {
	filtered := make([]string, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item) == "" {
			continue
		}
		filtered = append(filtered, item)
	}
	if len(filtered) == 0 {
		return "(none)"
	}
	return strings.Join(filtered, ",")
}

func orderNames(items []string, preferred []string) []string {
	if len(items) <= 1 {
		return append([]string(nil), items...)
	}

	order := make(map[string]int, len(preferred))
	for index, name := range preferred {
		order[name] = index
	}

	ordered := append([]string(nil), items...)
	sort.SliceStable(ordered, func(i int, j int) bool {
		left, leftOK := order[ordered[i]]
		right, rightOK := order[ordered[j]]
		switch {
		case leftOK && rightOK:
			return left < right
		case leftOK:
			return true
		case rightOK:
			return false
		default:
			return ordered[i] < ordered[j]
		}
	})

	return ordered
}

func capabilityPresetBoundary(catalog manifest.Catalog, capability manifest.CapabilityManifest) ([]string, []string, []string) {
	defaultOn := []string{}
	optionalOn := []string{}
	unsupportedOn := []string{}
	for _, presetName := range implementedPresets(catalog) {
		preset, ok := catalog.FindPreset(presetName)
		if !ok {
			continue
		}
		if contains(preset.DefaultCapabilities, capability.Name) {
			defaultOn = append(defaultOn, presetName)
			continue
		}
		if contains(capability.AllowedPresets, presetName) {
			optionalOn = append(optionalOn, presetName)
			continue
		}
		unsupportedOn = append(unsupportedOn, presetName)
	}
	return defaultOn, optionalOn, unsupportedOn
}

func printCapabilityPolicy(w io.Writer, catalog manifest.Catalog) {
	for _, capabilityName := range implementedCapabilities(catalog) {
		capability, ok := catalog.FindCapability(capabilityName)
		if !ok {
			continue
		}
		defaultOn, optionalOn, unsupportedOn := capabilityPresetBoundary(catalog, capability)
		fmt.Fprintf(w, "capability-policy-%s: default=%s optional=%s unsupported=%s\n", capability.Name, joinOrNone(defaultOn), joinOrNone(optionalOn), joinOrNone(unsupportedOn))
	}
}

func printSection(w io.Writer, title string) {
	fmt.Fprintf(w, "==== %s ====\n", title)
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func printSummary(w io.Writer, summary report.Summary) {
	fmt.Fprintf(w, "generated preset=%s target=%s\n", summary.Preset, summary.TargetDir)
	fmt.Fprintf(w, "stack: fiber-%s + %s", summary.FiberVersion, summary.CLIStyle)
	if summary.CLIStyle == stack.CLICobra {
		fmt.Fprintf(w, " + viper")
	}
	fmt.Fprintln(w)
	fmt.Fprintf(w, "runtime: logger=%s db=%s data-access=%s json-lib=%s\n", summary.Logger, summary.Database, summary.DataAccess, summary.JSONLib)
	fmt.Fprintf(w, "base: %s\n", summary.Base)
	fmt.Fprintf(w, "preset packs: %s\n", joinOrNone(summary.PresetPacks))
	fmt.Fprintf(w, "capabilities: %s\n", joinOrNone(summary.Capabilities))
	fmt.Fprintf(w, "capability packs: %s\n", joinOrNone(summary.CapabilityPacks))
	fmt.Fprintf(w, "runtime overlays: %s\n", joinOrNone(summary.RuntimeOverlays))
	fmt.Fprintf(w, "replace rules: %s\n", joinOrNone(summary.ReplaceRules))
	fmt.Fprintf(w, "injection rules: %s\n", joinOrNone(summary.InjectionRules))
	fmt.Fprintf(w, "generator: %s (%s)\n", summary.GeneratorVersion, summary.GeneratorCommit)
	fmt.Fprintf(w, "template fingerprint: %s\n", summary.TemplateSetFingerprint)
	fmt.Fprintf(w, "rendered fingerprint: %s\n", summary.RenderedOutputFingerprint)
	fmt.Fprintf(w, "metadata: %s\n", summary.MetadataPath)
	fmt.Fprintf(w, "written files: %d\n", summary.WrittenFiles)
	for _, path := range summary.WrittenPaths {
		fmt.Fprintf(w, "  - %s\n", path)
	}
	if len(summary.Warnings) > 0 {
		fmt.Fprintf(w, "warnings: %s\n", joinOrNone(summary.Warnings))
	}
}

type capabilityMatrix struct {
	Presets      []string                     `json:"presets"`
	Capabilities []string                     `json:"capabilities"`
	Matrix       map[string]map[string]string `json:"matrix"`
}

func printPreview(w io.Writer, preview report.Preview) {
	fmt.Fprintf(w, "generation plan preset=%s target=%s\n", preview.Preset, preview.TargetDir)
	fmt.Fprintf(w, "project: %s\n", preview.ProjectName)
	fmt.Fprintf(w, "module: %s\n", preview.ModulePath)
	fmt.Fprintf(w, "stack: fiber-%s + %s", preview.FiberVersion, preview.CLIStyle)
	if preview.CLIStyle == stack.CLICobra {
		fmt.Fprintf(w, " + viper")
	}
	fmt.Fprintln(w)
	fmt.Fprintf(w, "runtime: logger=%s db=%s data-access=%s json-lib=%s\n", preview.Logger, preview.Database, preview.DataAccess, preview.JSONLib)
	fmt.Fprintf(w, "base: %s\n", preview.Base)
	fmt.Fprintf(w, "preset packs: %s\n", joinOrNone(preview.PresetPacks))
	fmt.Fprintf(w, "capabilities: %s\n", joinOrNone(preview.Capabilities))
	fmt.Fprintf(w, "capability packs: %s\n", joinOrNone(preview.CapabilityPacks))
	fmt.Fprintf(w, "runtime overlays: %s\n", joinOrNone(preview.RuntimeOverlays))
	fmt.Fprintf(w, "replace rules: %s\n", joinOrNone(preview.ReplaceRules))
	fmt.Fprintf(w, "injection rules: %s\n", joinOrNone(preview.InjectionRules))
	fmt.Fprintf(w, "generator: %s (%s)\n", preview.GeneratorVersion, preview.GeneratorCommit)
	fmt.Fprintf(w, "metadata: %s\n", preview.MetadataPath)
	fmt.Fprintf(w, "estimated files: %d\n", preview.EstimatedFiles)
	for _, path := range preview.EstimatedPaths {
		fmt.Fprintf(w, "  - %s\n", path)
	}
	if len(preview.Warnings) > 0 {
		fmt.Fprintf(w, "warnings: %s\n", joinOrNone(preview.Warnings))
	}
}

func buildCapabilityMatrix(catalog manifest.Catalog) capabilityMatrix {
	presets := []string{"heavy", "medium", "light", "extra-light"}
	capabilities := []string{"redis", "swagger", "embedded-ui"}
	matrix := capabilityMatrix{
		Presets:      presets,
		Capabilities: capabilities,
		Matrix:       map[string]map[string]string{},
	}
	for _, presetName := range presets {
		row := map[string]string{}
		for _, capabilityName := range capabilities {
			capability, ok := catalog.FindCapability(capabilityName)
			if !ok {
				row[capabilityName] = "unsupported"
				continue
			}
			defaultOn, optionalOn, _ := capabilityPresetBoundary(catalog, capability)
			switch {
			case contains(defaultOn, presetName):
				row[capabilityName] = "default"
			case contains(optionalOn, presetName):
				row[capabilityName] = "optional"
			default:
				row[capabilityName] = "unsupported"
			}
		}
		matrix.Matrix[presetName] = row
	}
	return matrix
}

func printCapabilityMatrix(w io.Writer, matrix capabilityMatrix) {
	printSection(w, "capability matrix")
	fmt.Fprintln(w, "preset        redis        swagger      embedded-ui")
	for _, preset := range matrix.Presets {
		row := matrix.Matrix[preset]
		fmt.Fprintf(w, "%-13s %-12s %-12s %s\n", preset, row["redis"], row["swagger"], row["embedded-ui"])
	}
}

func detectDoctorMode(cwd, rootAbs string) string {
	if detectProjectRoot(cwd) != "" {
		return "project"
	}
	if detectGeneratorRepoRoot(cwd, rootAbs) != "" {
		return "generator"
	}
	return "standalone"
}

func detectProjectRoot(start string) string {
	for _, candidate := range ascendPaths(start) {
		if _, err := metadata.LoadManifest(candidate); err == nil {
			return candidate
		}
	}
	return ""
}

func detectGeneratorRepoRoot(start, rootAbs string) string {
	for _, candidate := range ascendPaths(start) {
		if pathExists(filepath.Join(candidate, "cmd", "fiberx", "main.go")) && filepath.Clean(filepath.Join(candidate, manifest.DefaultRoot())) == filepath.Clean(rootAbs) {
			return candidate
		}
	}
	return ""
}

func ascendPaths(start string) []string {
	paths := []string{}
	current := filepath.Clean(start)
	for {
		paths = append(paths, current)
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return paths
}

func runProjectDoctor(projectDir, rootAbs string, verbose bool) error {
	projectManifest, err := metadata.LoadManifest(projectDir)
	if err != nil {
		return err
	}
	assessment, err := upgrade.Inspect(projectDir, rootAbs)
	if err != nil {
		return err
	}
	if !verbose {
		fmt.Println("fiberx doctor")
		fmt.Println("mode: project")
		fmt.Printf("project: %s\n", projectDir)
		fmt.Printf("generated-by: %s (%s)\n", assessment.GeneratedGenerator.Version, assessment.GeneratedGenerator.Commit)
		fmt.Printf("current-generator: %s (%s)\n", assessment.CurrentGenerator.Version, assessment.CurrentGenerator.Commit)
		fmt.Printf("preset: %s\n", assessment.Recipe.Preset)
		fmt.Printf("capabilities: %s\n", joinOrNone(assessment.Recipe.Capabilities))
		if assessment.Recipe.Logger != "" || assessment.Recipe.DB != "" || assessment.Recipe.DataAccess != "" || assessment.Recipe.JSONLib != "" {
			fmt.Printf("runtime: logger=%s db=%s data-access=%s json-lib=%s\n", valueOrNone(assessment.Recipe.Logger), valueOrNone(assessment.Recipe.DB), valueOrNone(assessment.Recipe.DataAccess), valueOrNone(assessment.Recipe.JSONLib))
		}
		fmt.Printf("diff status: %s\n", assessment.DiffStatus)
		fmt.Printf("compatibility level: %s\n", assessment.CompatibilityLevel)
		fmt.Printf("status: %s\n", joinOrNone(assessment.Reasons))
		fmt.Println("note: use --verbose for full diagnostics")
		return nil
	}

	fmt.Fprintln(os.Stdout, "fiberx doctor")
	printSection(os.Stdout, "project")
	fmt.Fprintf(os.Stdout, "mode: project\nproject: %s\n", projectDir)
	fmt.Printf("manifest: %s\n", filepath.ToSlash(filepath.Join(projectDir, metadata.ManifestDir, metadata.ManifestFilename)))
	fmt.Printf("generated-at: %s\n", projectManifest.GeneratedAt)
	fmt.Printf("generated-by: %s (%s)\n", assessment.GeneratedGenerator.Version, assessment.GeneratedGenerator.Commit)
	fmt.Printf("current-generator: %s (%s)\n", assessment.CurrentGenerator.Version, assessment.CurrentGenerator.Commit)
	fmt.Fprintln(os.Stdout)
	printSection(os.Stdout, "recipe")
	fmt.Printf("preset: %s\n", assessment.Recipe.Preset)
	fmt.Printf("capabilities: %s\n", joinOrNone(assessment.Recipe.Capabilities))
	if assessment.Recipe.Logger != "" || assessment.Recipe.DB != "" || assessment.Recipe.DataAccess != "" || assessment.Recipe.JSONLib != "" {
		fmt.Printf("runtime: logger=%s db=%s data-access=%s json-lib=%s\n", valueOrNone(assessment.Recipe.Logger), valueOrNone(assessment.Recipe.DB), valueOrNone(assessment.Recipe.DataAccess), valueOrNone(assessment.Recipe.JSONLib))
	}
	fmt.Fprintln(os.Stdout)
	printSection(os.Stdout, "drift")
	fmt.Printf("diff status: %s\n", assessment.DiffStatus)
	fmt.Printf("compatibility level: %s\n", assessment.CompatibilityLevel)
	fmt.Printf("managed files: %d\n", len(projectManifest.ManagedFiles))
	fmt.Printf("missing files: %s\n", joinOrNone(assessment.MissingFiles))
	fmt.Printf("local modified files: %s\n", joinOrNone(assessment.LocalModifiedFiles))
	fmt.Printf("generator drift files: %s\n", joinOrNone(assessment.GeneratorDriftFiles))
	fmt.Printf("new managed files: %s\n", joinOrNone(assessment.NewManagedFiles))
	fmt.Printf("reasons: %s\n", joinOrNone(assessment.Reasons))
	fmt.Printf("blocking issues: %s\n", joinOrNone(assessment.BlockingIssues))
	if cfg, cfgErr := buildconfig.Load(projectDir); cfgErr == nil {
		fmt.Fprintln(os.Stdout)
		printSection(os.Stdout, "build")
		fmt.Printf("build config: present (%s)\n", buildconfig.Filename)
		profileNames := make([]string, 0, len(cfg.Build.Profiles))
		hookSummary := []string{}
		for name := range cfg.Build.Profiles {
			profileNames = append(profileNames, name)
		}
		sort.Strings(profileNames)
		for _, target := range cfg.Build.Targets {
			if len(target.PreHooks) > 0 || len(target.PostHooks) > 0 {
				hookSummary = append(hookSummary, fmt.Sprintf("%s(pre=%d,post=%d)", target.Name, len(target.PreHooks), len(target.PostHooks)))
			}
		}
		fmt.Printf("build profiles: %s\n", joinOrNone(profileNames))
		fmt.Printf("hooks: %s\n", joinOrNone(hookSummary))
	} else {
		fmt.Fprintln(os.Stdout)
		printSection(os.Stdout, "build")
		fmt.Printf("build config: absent (%v)\n", cfgErr)
	}
	return nil
}

func resultHasHooks(result build.Result) bool {
	for _, artifact := range result.Artifacts {
		if len(artifact.PreHooks) > 0 || len(artifact.PostHooks) > 0 {
			return true
		}
	}
	return false
}

func isInteractiveTerminal() bool {
	if info, err := os.Stdin.Stat(); err != nil || (info.Mode()&os.ModeCharDevice) == 0 {
		return false
	}
	if info, err := os.Stdout.Stat(); err != nil || (info.Mode()&os.ModeCharDevice) == 0 {
		return false
	}
	return true
}

func promptForHookApproval(result build.Result) bool {
	fmt.Println("build hooks are present for this build:")
	for _, artifact := range result.Artifacts {
		if len(artifact.PreHooks) == 0 && len(artifact.PostHooks) == 0 {
			continue
		}
		fmt.Printf("  - target=%s platform=%s pre_hooks=%s post_hooks=%s\n", artifact.TargetName, artifact.Platform, joinOrNone(artifact.PreHooks), joinOrNone(artifact.PostHooks))
	}
	fmt.Print("continue and execute hooks? [y/N]: ")
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true
	default:
		return false
	}
}

func hookActionSummary(result build.Result, confirmationRequired bool) string {
	taskCount := 0
	for _, artifact := range result.Artifacts {
		if len(artifact.PreHooks) > 0 || len(artifact.PostHooks) > 0 {
			taskCount++
		}
	}
	if confirmationRequired {
		return fmt.Sprintf("%d task(s) include hooks; interactive confirmation would be required unless --yes or --no-hooks is used", taskCount)
	}
	return fmt.Sprintf("%d task(s) include hooks", taskCount)
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func defaultLoggerForPreset(presetName string) string {
	if presetName == "extra-light" {
		return "slog"
	}
	return stack.DefaultLogger()
}

func defaultDatabaseForPreset(presetName string) string {
	return stack.DefaultDB()
}

func defaultDataAccessForPreset(presetName string) string {
	if presetName == "extra-light" {
		return "builtin"
	}
	return stack.DefaultDataAccess()
}

func resolveProjectDir(args []string) (string, error) {
	if len(args) > 1 {
		return "", errors.New("command accepts at most one project path")
	}
	if len(args) == 0 {
		return os.Getwd()
	}
	return filepath.Abs(args[0])
}

func writeJSON(w io.Writer, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = w.Write(data)
	return err
}

func valueOrNone(value string) string {
	if strings.TrimSpace(value) == "" {
		return "(none)"
	}
	return value
}

func parseCapabilities(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{}
	}

	parts := strings.Split(raw, ",")
	capabilities := make([]string, 0, len(parts))
	for _, part := range parts {
		name := strings.TrimSpace(part)
		if name == "" {
			continue
		}
		capabilities = append(capabilities, name)
	}

	return capabilities
}

func defaultModulePath(projectName string, explicit string) string {
	if explicit != "" {
		return explicit
	}

	slug := strings.ToLower(strings.TrimSpace(projectName))
	slug = strings.ReplaceAll(slug, " ", "-")
	if slug == "" {
		slug = "fiberx-app"
	}

	return "github.com/example/" + slug
}

func reorderArgs(args []string, valueFlags map[string]bool) []string {
	reordered := make([]string, 0, len(args))
	positionals := make([]string, 0, len(args))

	for index := 0; index < len(args); index++ {
		current := args[index]
		if strings.HasPrefix(current, "-") {
			reordered = append(reordered, current)
			if valueFlags[current] && index+1 < len(args) {
				index++
				reordered = append(reordered, args[index])
			}
			continue
		}

		positionals = append(positionals, current)
	}

	return append(reordered, positionals...)
}

func setOptionalRuntimeFlags(fs *flag.FlagSet, options map[string]string, loggerBackend, dbKind, dataAccess, jsonLib string) {
	visited := map[string]bool{}
	fs.Visit(func(f *flag.Flag) {
		visited[f.Name] = true
	})

	if visited["logger"] {
		options[stack.OptionLogger] = loggerBackend
	}
	if visited["db"] {
		options[stack.OptionDB] = dbKind
	}
	if visited["data-access"] {
		options[stack.OptionDataAccess] = dataAccess
	}
	if visited["json-lib"] {
		options[stack.OptionJSONLib] = jsonLib
	}
}
