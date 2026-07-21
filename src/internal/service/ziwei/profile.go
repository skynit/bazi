package ziwei

import (
	"fmt"
	"strings"
)

const (
	DefaultProfileID   = "ziwei-local-composite-v2"
	ZiWeiEngineVersion = "ziwei-local-go-2026-07-20.76"
	ZiWeiRuleVersion   = "ziwei-rules-2026-07-20.76"
	ZiWeiRuleSchool    = "紫微斗数-本地综合规则-v2"

	SiHuaRuleID       = "ziwei.sihua.ten-stem.iztro-v1"
	SiHuaSourceRepo   = "https://github.com/SylarLong/iztro"
	SiHuaSourceCommit = "2dfe3ecb41d725b2bea1084bbdfe4dd655e37b13"
	SiHuaSourcePath   = "src/data/heavenlyStems.ts"
	SiHuaSourceSHA256 = "f50a96b4fda42f834b2ff7aea36533a38bd5bcee94b74d5ecc1deb22ee07279c"

	StarBrightnessRuleID     = "ziwei.star-brightness.iztro-v1"
	StarBrightnessSourcePath = "src/data/stars.ts"
	StarBrightnessSHA256     = "87d2fdbb4501db7b6aa054237cba4de38210e7b3dbd97e849ce1936c0f735b95"

	LeapMonthRuleID     = "ziwei.leap-month.normalization.iztro-v1"
	LeapMonthSourcePath = "src/utils/index.ts"
	LeapMonthSHA256     = "9f3699115c9c149dc261f9a293e6c7ad40ac4f744884dba21b367b96d9bbfe16"

	MonthlyStarsRuleID     = "ziwei.monthly-stars.iztro-v1"
	MonthlyStarsSourcePath = "src/star/location.ts"
	MonthlyStarsSHA256     = "082f2b5883e57e7f6858cc64573efd5916baecedd578a09fdefdbc291849caef"

	AdjectiveStarsRuleID     = "ziwei.adjective-stars.iztro-v1"
	AdjectiveStarsSourcePath = "src/star/adjectiveStar.ts"
	AdjectiveStarsSHA256     = "be0e7afce2ffb75155e92116add91e4e59fd91b796ebff870b40e0dfb19eb02e"

	PeriodChronologyRuleID     = "ziwei.period-chronology.iztro-normal-v1"
	PeriodChronologySourcePath = "src/astro/FunctionalAstrolabe.ts"
	PeriodChronologySHA256     = "1748e1c8210a19aac1da7479e50237b30b160ac694d3ecbef5bb8ba4a0822f27"

	TransitStarsRuleID     = "ziwei.transit-stars.iztro-v1"
	TransitStarsSourcePath = "src/star/horoscopeStar.ts"
	TransitStarsSHA256     = "0cce679f149711bee610f62cdeabbbc12697a242847f297275cc846669bf29f8"

	CorePatternRuleID       = "ziwei.pattern.core.renhuai-v1"
	CorePatternSourceRepo   = "https://github.com/Renhuai123/ziwei-doushu"
	CorePatternSourceCommit = "88194a404242bfe5c6d5cc512e4117e3e245cdd5"
	CorePatternSourcePath   = "lib/ziwei/patterns.ts"
	CorePatternSourceSHA256 = "c0866fa76cc67e2d636aeface0016cfb0bc2e08da9b48cca7fcbdd38f814dcce"

	QiShaChaoDouRuleID       = "ziwei.pattern.qisha-chaodou.iztro-v1"
	QiShaChaoDouSourcePath   = "docs/learn/pattern.html"
	QiShaChaoDouSourceSHA256 = "ca80d6c9edab47c3364fdf27de2c711442df52d3e2ba88533e76be05e5172294"
)

// RuleSourceRef pins an external rule source without promoting it to Gold.
type RuleSourceRef struct {
	RuleID           string `json:"rule_id"`
	Repository       string `json:"repository"`
	Commit           string `json:"commit"`
	Path             string `json:"path"`
	SHA256           string `json:"sha256"`
	License          string `json:"license"`
	SourceTier       string `json:"source_tier"`
	ValidationStatus string `json:"validation_status"`
}

// CalculationProfile identifies a complete, reproducible ZiWei calculation
// convention. A new profile ID or rule version is required whenever a
// school-specific placement rule changes.
type CalculationProfile struct {
	ID                      string              `json:"id"`
	EngineVersion           string              `json:"engine_version"`
	RuleVersion             string              `json:"rule_version"`
	School                  string              `json:"school"`
	RuleSources             []RuleSourceRef     `json:"rule_sources"`
	RuntimeRuleTablesSchema string              `json:"runtime_rule_tables_schema"`
	RuntimeRuleTablesHash   string              `json:"runtime_rule_tables_hash"`
	PluginManifest          []PluginRequirement `json:"plugin_manifest"`
	PluginManifestHash      string              `json:"plugin_manifest_hash"`
}

var calculationProfiles = map[string]CalculationProfile{
	DefaultProfileID: {
		ID:                      DefaultProfileID,
		EngineVersion:           ZiWeiEngineVersion,
		RuleVersion:             ZiWeiRuleVersion,
		School:                  ZiWeiRuleSchool,
		RuntimeRuleTablesSchema: ZiWeiRuntimeRuleTablesSchema,
		RuntimeRuleTablesHash:   ZiWeiRuntimeRuleTablesSHA256,
		RuleSources: []RuleSourceRef{{
			RuleID: SiHuaRuleID, Repository: SiHuaSourceRepo, Commit: SiHuaSourceCommit,
			Path: SiHuaSourcePath, SHA256: SiHuaSourceSHA256, License: "MIT", SourceTier: "silver_external",
			ValidationStatus: "cross_checked_not_gold",
		}, {
			RuleID: StarBrightnessRuleID, Repository: SiHuaSourceRepo, Commit: SiHuaSourceCommit,
			Path: StarBrightnessSourcePath, SHA256: StarBrightnessSHA256, License: "MIT", SourceTier: "silver_external",
			ValidationStatus: "cross_checked_not_gold",
		}, {
			RuleID: LeapMonthRuleID, Repository: SiHuaSourceRepo, Commit: SiHuaSourceCommit,
			Path: LeapMonthSourcePath, SHA256: LeapMonthSHA256, License: "MIT", SourceTier: "silver_external",
			ValidationStatus: "cross_checked_not_gold",
		}, {
			RuleID: MonthlyStarsRuleID, Repository: SiHuaSourceRepo, Commit: SiHuaSourceCommit,
			Path: MonthlyStarsSourcePath, SHA256: MonthlyStarsSHA256, License: "MIT", SourceTier: "silver_external",
			ValidationStatus: "cross_checked_not_gold",
		}, {
			RuleID: AdjectiveStarsRuleID, Repository: SiHuaSourceRepo, Commit: SiHuaSourceCommit,
			Path: AdjectiveStarsSourcePath, SHA256: AdjectiveStarsSHA256, License: "MIT", SourceTier: "silver_external",
			ValidationStatus: "cross_checked_not_gold",
		}, {
			RuleID: PeriodChronologyRuleID, Repository: SiHuaSourceRepo, Commit: SiHuaSourceCommit,
			Path: PeriodChronologySourcePath, SHA256: PeriodChronologySHA256, License: "MIT", SourceTier: "silver_external",
			ValidationStatus: "cross_checked_not_gold",
		}, {
			RuleID: TransitStarsRuleID, Repository: SiHuaSourceRepo, Commit: SiHuaSourceCommit,
			Path: TransitStarsSourcePath, SHA256: TransitStarsSHA256, License: "MIT", SourceTier: "silver_external",
			ValidationStatus: "cross_checked_not_gold",
		}, {
			RuleID: CorePatternRuleID, Repository: CorePatternSourceRepo, Commit: CorePatternSourceCommit,
			Path: CorePatternSourcePath, SHA256: CorePatternSourceSHA256, License: "MIT", SourceTier: "silver_external",
			ValidationStatus: "cross_checked_not_gold",
		}, {
			RuleID: QiShaChaoDouRuleID, Repository: SiHuaSourceRepo, Commit: SiHuaSourceCommit,
			Path: QiShaChaoDouSourcePath, SHA256: QiShaChaoDouSourceSHA256, License: "MIT", SourceTier: "silver_external",
			ValidationStatus: "cross_checked_not_gold",
		}},
	},
}

// ResolveProfile returns the default profile for an empty ID and rejects
// unknown profiles instead of silently falling back to a different school.
func ResolveProfile(id string) (CalculationProfile, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		id = DefaultProfileID
	}
	profile, ok := calculationProfiles[id]
	if !ok {
		return CalculationProfile{}, fmt.Errorf("unknown ziwei profile %q", id)
	}
	manifest, err := normalizePluginManifest(profile.PluginManifest)
	if err != nil {
		return CalculationProfile{}, fmt.Errorf("invalid ziwei profile %q: %w", id, err)
	}
	profile.PluginManifest = manifest
	profile.PluginManifestHash = pluginManifestHash(manifest)
	profile.RuleSources = cloneRuleSources(profile.RuleSources)
	return profile, nil
}

func chartMatchesProfile(chart *ZiWeiChart, profile CalculationProfile) bool {
	return chart != nil && chart.ProfileID == profile.ID &&
		chart.EngineVersion == profile.EngineVersion &&
		chart.RuleVersion == profile.RuleVersion &&
		chart.RuleSchool == profile.School &&
		equalRuleSources(chart.RuleSources, profile.RuleSources) &&
		chart.RuntimeRuleTablesSchema == profile.RuntimeRuleTablesSchema &&
		chart.RuntimeRuleTablesHash == profile.RuntimeRuleTablesHash &&
		chart.PluginManifestHash == profile.PluginManifestHash &&
		equalPluginManifest(chart.PluginManifest, profile.PluginManifest) &&
		validateRuntimeRuleTables(profile) == nil &&
		validChartContentHash(chart) &&
		validPublishedZiWeiChartStructure(chart, profile)
}

func chartMatchesDeclaredProfile(chart *ZiWeiChart) bool {
	if chart == nil {
		return false
	}
	profile, err := ResolveProfile(chart.ProfileID)
	return err == nil && chartMatchesProfile(chart, profile)
}

func stampChartProfile(chart *ZiWeiChart, profile CalculationProfile, plan *pluginExecutionPlan) {
	chart.ProfileID = profile.ID
	chart.EngineVersion = profile.EngineVersion
	chart.RuleVersion = profile.RuleVersion
	chart.RuleSchool = profile.School
	chart.RuleSources = cloneRuleSources(profile.RuleSources)
	chart.RuntimeRuleTablesSchema = profile.RuntimeRuleTablesSchema
	chart.RuntimeRuleTablesHash = profile.RuntimeRuleTablesHash
	chart.PluginManifest = clonePluginManifest(plan.manifest)
	chart.PluginManifestHash = plan.hash
}

func cloneRuleSources(sources []RuleSourceRef) []RuleSourceRef {
	if len(sources) == 0 {
		return []RuleSourceRef{}
	}
	return append([]RuleSourceRef(nil), sources...)
}

func equalRuleSources(left, right []RuleSourceRef) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func equalPluginManifest(left, right []PluginRequirement) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
