package ziwei

import (
	"encoding/json"
	"strings"
	"testing"
)

type recordingStarPlugin struct {
	descriptor PluginDescriptor
	marker     string
	star       string
}

func (p recordingStarPlugin) Descriptor() PluginDescriptor { return p.descriptor }
func (p recordingStarPlugin) AdditionalStars(_ PalaceInfo, palaceIdx int) []StarOutput {
	if p.star == "" || palaceIdx != 0 {
		return nil
	}
	return []StarOutput{{Name: p.star, Type: "soft", Scope: "origin", Brightness: "旺"}}
}
func (p recordingStarPlugin) MutatePalace(palace *PalaceInfo) bool {
	palace.AdjectiveStars = append(palace.AdjectiveStars, p.marker)
	return true
}

func TestPluginExecutionPlanUsesProfileOrderAndStableHash(t *testing.T) {
	pluginA := recordingStarPlugin{descriptor: PluginDescriptor{ID: "plugin-a", Version: "1.0.0"}, marker: "A"}
	pluginB := recordingStarPlugin{descriptor: PluginDescriptor{ID: "plugin-b", Version: "2.0.0", DependsOn: []string{"plugin-a"}}, marker: "B"}
	catalog, err := newPluginCatalog([]StarPlugin{pluginB, pluginA})
	if err != nil {
		t.Fatal(err)
	}
	profile := CalculationProfile{
		ID: "test-profile", PluginManifest: []PluginRequirement{
			{ID: "plugin-a", Version: "1.0.0"},
			{ID: "plugin-b", Version: "2.0.0"},
		},
	}
	plan, err := buildPluginExecutionPlan(profile, catalog)
	if err != nil {
		t.Fatal(err)
	}
	chart := &ZiWeiChart{}
	plan.Apply(chart)
	if got := strings.Join(chart.Palaces[0].AdjectiveStars, ","); got != "A,B" {
		t.Fatalf("plugin execution order = %q, want A,B", got)
	}
	if len(plan.hash) != 64 || plan.hash != pluginManifestHash(profile.PluginManifest) {
		t.Fatalf("manifest hash = %q", plan.hash)
	}
	reversed := []PluginRequirement{
		{ID: "plugin-b", Version: "2.0.0"},
		{ID: "plugin-a", Version: "1.0.0"},
	}
	if plan.hash == pluginManifestHash(reversed) {
		t.Fatal("plugin manifest hash ignored execution order")
	}
}

func TestPluginExecutionPlanRejectsDependencyConflictAndVersionMismatch(t *testing.T) {
	pluginA := recordingStarPlugin{descriptor: PluginDescriptor{ID: "plugin-a", Version: "1"}}
	pluginB := recordingStarPlugin{descriptor: PluginDescriptor{ID: "plugin-b", Version: "1", DependsOn: []string{"plugin-a"}}}
	pluginC := recordingStarPlugin{descriptor: PluginDescriptor{ID: "plugin-c", Version: "1", ConflictsWith: []string{"plugin-a"}}}
	catalog, err := newPluginCatalog([]StarPlugin{pluginA, pluginB, pluginC})
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name     string
		manifest []PluginRequirement
		contains string
	}{
		{name: "dependency order", manifest: []PluginRequirement{{ID: "plugin-b", Version: "1"}, {ID: "plugin-a", Version: "1"}}, contains: "earlier"},
		{name: "conflict", manifest: []PluginRequirement{{ID: "plugin-a", Version: "1"}, {ID: "plugin-c", Version: "1"}}, contains: "conflicts"},
		{name: "version", manifest: []PluginRequirement{{ID: "plugin-a", Version: "2"}}, contains: "catalog provides"},
		{name: "missing", manifest: []PluginRequirement{{ID: "missing", Version: "1"}}, contains: "missing ziwei plugin"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, planErr := buildPluginExecutionPlan(CalculationProfile{ID: "test", PluginManifest: tt.manifest}, catalog)
			if planErr == nil || !strings.Contains(planErr.Error(), tt.contains) {
				t.Fatalf("error = %v, want %q", planErr, tt.contains)
			}
		})
	}
}

func TestZiWeiChartStampsExecutedPluginManifest(t *testing.T) {
	pluginA := recordingStarPlugin{descriptor: PluginDescriptor{ID: "plugin-a", Version: "1"}, marker: "A", star: "测试星"}
	catalog, err := newPluginCatalog([]StarPlugin{pluginA})
	if err != nil {
		t.Fatal(err)
	}
	profile := CalculationProfile{
		ID: "test-profile", EngineVersion: "test-engine", RuleVersion: "test-rule", School: "test-school",
		RuntimeRuleTablesSchema: ZiWeiRuntimeRuleTablesSchema,
		RuntimeRuleTablesHash:   ZiWeiRuntimeRuleTablesSHA256,
		PluginManifest:          []PluginRequirement{{ID: "plugin-a", Version: "1"}},
	}
	profile.PluginManifestHash = pluginManifestHash(profile.PluginManifest)
	service := &ZiWeiService{profile: profile, pluginCatalog: catalog}
	chart, err := service.calculateChart(profile, 2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatal(err)
	}
	if chart.PluginManifestHash != profile.PluginManifestHash || !equalPluginManifest(chart.PluginManifest, profile.PluginManifest) {
		t.Fatalf("chart manifest = %+v hash=%q", chart.PluginManifest, chart.PluginManifestHash)
	}
	if len(chart.Palaces[0].AdjectiveStars) == 0 || chart.Palaces[0].AdjectiveStars[len(chart.Palaces[0].AdjectiveStars)-1] != "A" {
		t.Fatalf("profile plugin was not applied: %+v", chart.Palaces[0].AdjectiveStars)
	}
	foundStar := false
	for _, star := range chart.Palaces[0].Stars {
		if star.Name == "测试星" && star.Brightness == "旺" {
			foundStar = true
		}
	}
	if !foundStar {
		t.Fatalf("profile plugin star missing from JSON-facing Stars: %+v", chart.Palaces[0].Stars)
	}
	encoded, err := json.Marshal(chart)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{`"plugin_manifest"`, `"plugin_manifest_hash"`, `"plugin-a"`} {
		if !strings.Contains(string(encoded), field) {
			t.Fatalf("serialized chart missing %s", field)
		}
	}
}
