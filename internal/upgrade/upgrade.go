package upgrade

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/gofurry/fiberx/internal/metadata"
	"github.com/gofurry/fiberx/internal/version"
)

const (
	LevelCompatible   = "compatible"
	LevelManualReview = "manual_review"
	LevelBreaking     = "breaking"

	DiffStatusUnavailable = "unavailable"
)

var releaseVersionPattern = regexp.MustCompile(`^v\d+\.\d+\.\d+$`)

type Assessment struct {
	ProjectDir          string                 `json:"project_dir"`
	GeneratedGenerator  metadata.GeneratorInfo `json:"generated_generator"`
	CurrentGenerator    metadata.GeneratorInfo `json:"current_generator"`
	Recipe              metadata.Recipe        `json:"recipe"`
	DiffStatus          string                 `json:"diff_status"`
	CompatibilityLevel  string                 `json:"compatibility_level"`
	Reasons             []string               `json:"reasons"`
	BlockingIssues      []string               `json:"blocking_issues"`
	LocalModifiedFiles  []string               `json:"local_modified_files"`
	GeneratorDriftFiles []string               `json:"generator_drift_files"`
	MissingFiles        []string               `json:"missing_files"`
	NewManagedFiles     []string               `json:"new_managed_files"`
}

type UpgradePlan struct {
	Assessment           Assessment `json:"assessment"`
	RecommendedSteps     []string   `json:"recommended_steps"`
	ManagedFilesToReview []string   `json:"managed_files_to_review"`
}

type semanticVersion struct {
	Major int
	Minor int
	Patch int
}

func Inspect(projectDir, catalogRoot string) (Assessment, error) {
	absProjectDir, err := filepath.Abs(projectDir)
	if err != nil {
		return Assessment{}, err
	}

	projectManifest, err := metadata.LoadManifest(absProjectDir)
	if err != nil {
		return Assessment{}, err
	}

	assessment := Assessment{
		ProjectDir:         absProjectDir,
		GeneratedGenerator: projectManifest.Generator,
		CurrentGenerator: metadata.GeneratorInfo{
			Version: version.Version,
			Commit:  version.Commit,
		},
		Recipe: projectManifest.Recipe,
	}

	diffReport, err := metadata.BuildDiff(absProjectDir, catalogRoot)
	if err != nil {
		assessment.DiffStatus = DiffStatusUnavailable
		assessment.CompatibilityLevel = LevelBreaking
		assessment.Reasons = []string{"项目配方无法被当前 generator 重新渲染"}
		assessment.BlockingIssues = []string{err.Error()}
		return assessment, nil
	}

	assessment.DiffStatus = diffReport.Status
	assessment.MissingFiles = append([]string(nil), diffReport.MissingFiles...)
	assessment.LocalModifiedFiles = append([]string(nil), diffReport.ChangedFiles...)
	assessment.NewManagedFiles = append([]string(nil), diffReport.NewManagedFiles...)
	assessment.GeneratorDriftFiles = append([]string(nil), diffReport.GeneratorDriftFiles...)

	return finalizeAssessment(assessment), nil
}

func Plan(projectDir, catalogRoot string) (UpgradePlan, error) {
	assessment, err := Inspect(projectDir, catalogRoot)
	if err != nil {
		return UpgradePlan{}, err
	}

	plan := UpgradePlan{
		Assessment:           assessment,
		ManagedFilesToReview: managedFilesToReview(assessment),
	}

	switch assessment.CompatibilityLevel {
	case LevelBreaking:
		return plan, nil
	case LevelCompatible:
		switch assessment.DiffStatus {
		case metadata.StatusClean:
			plan.RecommendedSteps = []string{"项目与当前 generator 一致，无需升级动作。"}
		case metadata.StatusGeneratorDrift:
			plan.RecommendedSteps = []string{
				"在临时目录使用当前 generator 重新生成同一 recipe 的项目副本。",
				"逐个审查 generator drift 涉及的受管文件，并确认哪些变化要合并回项目。",
				"仅在人工确认后再同步这些受管文件变化。",
			}
		default:
			plan.RecommendedSteps = []string{"当前项目可以被当前 generator 识别，但仍建议人工确认受管文件差异。"}
		}
	case LevelManualReview:
		if usesDevelopmentGenerator(assessment.CurrentGenerator) || usesDevelopmentGenerator(assessment.GeneratedGenerator) {
			plan.RecommendedSteps = append(plan.RecommendedSteps, "当前或历史 generator 使用了开发态版本标识；在依赖升级结论前请先人工确认版本来源。")
		}
		switch assessment.DiffStatus {
		case metadata.StatusLocalModified:
			plan.RecommendedSteps = append(plan.RecommendedSteps,
				"先审查本地对受管文件的改动，确认哪些内容是项目自定义保留项。",
				"在临时目录重新生成当前 recipe，并将本地改动与最新模板输出逐个对比。",
			)
		case metadata.StatusLocalAndGenDrift:
			plan.RecommendedSteps = append(plan.RecommendedSteps,
				"先冻结本地对受管文件的修改，避免在升级评估过程中继续叠加变化。",
				"在临时目录重新生成当前 recipe，并按文件合并本地定制与 generator drift。",
			)
		case metadata.StatusGeneratorDrift:
			plan.RecommendedSteps = append(plan.RecommendedSteps,
				"当前 generator 输出已经发生变化；请先人工审查 drift 文件，再决定是否采纳升级结果。",
			)
		case metadata.StatusClean:
			plan.RecommendedSteps = append(plan.RecommendedSteps,
				"虽然受管文件当前没有差异，但版本信息不足以形成可靠兼容判断；请人工确认升级来源后再继续。",
			)
		default:
			plan.RecommendedSteps = append(plan.RecommendedSteps, "请先人工确认当前项目状态，再执行后续升级比对。")
		}
	}

	return plan, nil
}

func finalizeAssessment(assessment Assessment) Assessment {
	reasons := make([]string, 0, 4)

	if compare, ok := compareReleaseVersions(assessment.CurrentGenerator.Version, assessment.GeneratedGenerator.Version); ok && compare < 0 {
		assessment.CompatibilityLevel = LevelBreaking
		assessment.Reasons = []string{"当前 generator 版本低于项目记录的生成器版本"}
		assessment.BlockingIssues = []string{
			fmt.Sprintf("current generator %s is lower than generated version %s", assessment.CurrentGenerator.Version, assessment.GeneratedGenerator.Version),
		}
		return assessment
	}

	switch assessment.DiffStatus {
	case metadata.StatusClean:
		reasons = append(reasons, "当前项目与历史 metadata 以及当前 generator 输出一致")
	case metadata.StatusLocalModified:
		reasons = append(reasons, "项目存在受管文件本地改动")
	case metadata.StatusGeneratorDrift:
		reasons = append(reasons, "当前 generator 对同一 recipe 的输出已经发生变化")
	case metadata.StatusLocalAndGenDrift:
		reasons = append(reasons, "项目同时存在本地受管文件改动和 generator drift")
	case DiffStatusUnavailable:
		reasons = append(reasons, "当前 generator 无法完成差异重放")
	}

	if len(assessment.MissingFiles) > 0 {
		reasons = append(reasons, "项目中缺失部分受管文件")
	}
	if len(assessment.NewManagedFiles) > 0 {
		reasons = append(reasons, "当前 generator 会引入新的受管文件")
	}

	developmentMetadata := usesDevelopmentGenerator(assessment.CurrentGenerator) || usesDevelopmentGenerator(assessment.GeneratedGenerator)
	if developmentMetadata {
		reasons = append(reasons, "当前或历史 generator 使用开发态版本标识，兼容判断需要人工复核")
	}

	switch {
	case assessment.CompatibilityLevel == LevelBreaking:
	case assessment.DiffStatus == metadata.StatusLocalModified || assessment.DiffStatus == metadata.StatusLocalAndGenDrift || developmentMetadata:
		assessment.CompatibilityLevel = LevelManualReview
	case assessment.DiffStatus == metadata.StatusClean || assessment.DiffStatus == metadata.StatusGeneratorDrift:
		assessment.CompatibilityLevel = LevelCompatible
	default:
		assessment.CompatibilityLevel = LevelManualReview
	}

	assessment.Reasons = reasons
	return assessment
}

func managedFilesToReview(assessment Assessment) []string {
	seen := map[string]bool{}
	files := []string{}
	for _, group := range [][]string{
		assessment.MissingFiles,
		assessment.LocalModifiedFiles,
		assessment.NewManagedFiles,
		assessment.GeneratorDriftFiles,
	} {
		for _, path := range group {
			if seen[path] {
				continue
			}
			seen[path] = true
			files = append(files, path)
		}
	}
	sort.Strings(files)
	return files
}

func usesDevelopmentGenerator(info metadata.GeneratorInfo) bool {
	if strings.TrimSpace(info.Version) == "" || strings.TrimSpace(info.Commit) == "" {
		return true
	}
	if info.Commit == "unknown" {
		return true
	}
	return strings.Contains(info.Version, "-dev")
}

func compareReleaseVersions(current, generated string) (int, bool) {
	currentVersion, ok := parseReleaseVersion(current)
	if !ok {
		return 0, false
	}
	generatedVersion, ok := parseReleaseVersion(generated)
	if !ok {
		return 0, false
	}

	switch {
	case currentVersion.Major != generatedVersion.Major:
		if currentVersion.Major < generatedVersion.Major {
			return -1, true
		}
		return 1, true
	case currentVersion.Minor != generatedVersion.Minor:
		if currentVersion.Minor < generatedVersion.Minor {
			return -1, true
		}
		return 1, true
	case currentVersion.Patch != generatedVersion.Patch:
		if currentVersion.Patch < generatedVersion.Patch {
			return -1, true
		}
		return 1, true
	default:
		return 0, true
	}
}

func parseReleaseVersion(raw string) (semanticVersion, bool) {
	if !releaseVersionPattern.MatchString(raw) {
		return semanticVersion{}, false
	}

	var version semanticVersion
	if _, err := fmt.Sscanf(raw, "v%d.%d.%d", &version.Major, &version.Minor, &version.Patch); err != nil {
		return semanticVersion{}, false
	}
	return version, true
}
