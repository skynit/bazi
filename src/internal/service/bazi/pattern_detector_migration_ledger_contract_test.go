package bazi

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"bazi/internal/model"
)

func TestPatternDetectorProfileMigrationLedgerRecomputesCanonicalChain(t *testing.T) {
	ledger, ok := PatternDetectorProfileMigrationLedgerSnapshot()
	if !ok {
		t.Fatal("embedded detector profile migration ledger is invalid")
	}
	if ledger.Schema != "pattern_detector_profile_migration_ledger_v2" || ledger.ClaimBoundary != "digest_evidence_only" ||
		ledger.ChainScheme != "pattern_detector_profile_migration_chain_v1" || len(ledger.ProfileSets) != 1 ||
		len(ledger.Snapshots) != 5 || len(ledger.ChangeSets) != 1 || len(ledger.Migrations) != 4 {
		t.Fatalf("unexpected ledger shape: %+v", ledger)
	}
	if got := patternDetectorProfileMigrationLedgerSHA256(); got != PatternDetectorProfileMigrationLedgerSHA256 {
		t.Fatalf("migration ledger SHA-256 = %s, want %s", got, PatternDetectorProfileMigrationLedgerSHA256)
	}

	wantSnapshots := []PatternDetectorProfileMigrationSnapshot{
		{SnapshotID: "bazi.pattern-candidate-set-v30", EngineVersion: "bazi-engine-2026-07-17.23", RuleVersion: "bazi-rules-2026-07-17.23", RuleID: "bazi.pattern-candidate-set-v30", SchemaVersion: "pattern-candidates-2026-07-17.23", DetectorProfile: "classical_structural_detectors_v41", DetectorManifestSHA256: expectedPatternDetectorManifestSHA256, ProfileSetID: "pattern-detector-profiles-2026-07-17.23"},
		{SnapshotID: "bazi.pattern-candidate-set-v31", EngineVersion: "bazi-engine-2026-07-17.24", RuleVersion: "bazi-rules-2026-07-17.24", RuleID: "bazi.pattern-candidate-set-v31", SchemaVersion: "pattern-candidates-2026-07-17.24", DetectorProfile: "classical_structural_detectors_v42", DetectorManifestSHA256: expectedPatternDetectorManifestSHA256, ProfileSetID: "pattern-detector-profiles-2026-07-17.23"},
		{SnapshotID: "bazi.pattern-candidate-set-v32", EngineVersion: "bazi-engine-2026-07-17.25", RuleVersion: "bazi-rules-2026-07-17.25", RuleID: "bazi.pattern-candidate-set-v32", SchemaVersion: "pattern-candidates-2026-07-17.25", DetectorProfile: "classical_structural_detectors_v43", DetectorManifestSHA256: expectedPatternDetectorManifestSHA256, ProfileSetID: "pattern-detector-profiles-2026-07-17.23"},
		{SnapshotID: "bazi.pattern-candidate-set-v33", EngineVersion: "bazi-engine-2026-07-17.26", RuleVersion: "bazi-rules-2026-07-17.26", RuleID: "bazi.pattern-candidate-set-v33", SchemaVersion: "pattern-candidates-2026-07-17.26", DetectorProfile: "classical_structural_detectors_v44", DetectorManifestSHA256: expectedPatternDetectorManifestSHA256, ProfileSetID: "pattern-detector-profiles-2026-07-17.23"},
		{SnapshotID: "bazi.pattern-candidate-set-v34", EngineVersion: "bazi-engine-2026-07-17.27", RuleVersion: "bazi-rules-2026-07-17.27", RuleID: "bazi.pattern-candidate-set-v34", SchemaVersion: "pattern-candidates-2026-07-17.27", DetectorProfile: "classical_structural_detectors_v45", DetectorManifestSHA256: expectedPatternDetectorManifestSHA256, ProfileSetID: "pattern-detector-profiles-2026-07-17.23"},
	}
	if !reflect.DeepEqual(ledger.Snapshots, wantSnapshots) {
		t.Fatalf("migration snapshots = %+v, want %+v", ledger.Snapshots, wantSnapshots)
	}

	profiles := ledger.ProfileSets[0].DetectorProfiles
	if !reflect.DeepEqual(profiles, patternDetectorProfileDigests(patternDetectorRegistry())) {
		t.Fatal("migration profile set does not match the current registry")
	}
	wantChanges := ComparePatternDetectorProfiles(profiles, profiles)
	if !reflect.DeepEqual(ledger.ChangeSets[0].Result, wantChanges) {
		t.Fatalf("stored migration result = %+v, recomputed %+v", ledger.ChangeSets[0].Result, wantChanges)
	}
	wantMigrationSHA256 := []string{
		"0f71d210f4475e0c3a81480a612b5b0e39b86601085e4aa79d8d364660f6ccc9",
		"aea51f9ca734acc61861aaacfff161c522671ff68b87c669389fde442bbc1a53",
		"ea922b348fa81df44a70ece07f84b30fc5d8b50d2958e0012219d353ea5de2aa",
		"07dc296ad9e5dd0f834e40256c1e0f6033eb0ded435d0c76be6a0602ae0113bd",
	}
	previousMigrationSHA256 := PatternDetectorProfileMigrationGenesisSHA256
	for index, migration := range ledger.Migrations {
		if migration.ExpectedChangeSetID != ledger.ChangeSets[0].ChangeSetID {
			t.Errorf("migration %d change-set reference = %q", index, migration.ExpectedChangeSetID)
		}
		if migration.PreviousMigrationSHA256 != previousMigrationSHA256 || migration.MigrationSHA256 != wantMigrationSHA256[index] {
			t.Errorf("migration %d hash link = %s/%s", index, migration.PreviousMigrationSHA256, migration.MigrationSHA256)
		}
		gotSHA256 := patternDetectorProfileMigrationSHA256(
			ledger.ChainScheme, migration, ledger.Snapshots[index], profiles,
			ledger.Snapshots[index+1], profiles, ledger.ChangeSets[0].Result,
		)
		if gotSHA256 != migration.MigrationSHA256 {
			t.Errorf("migration %d recomputed SHA-256 = %s, want %s", index, gotSHA256, migration.MigrationSHA256)
		}
		if index > 0 && ledger.Migrations[index-1].ToSnapshotID != migration.FromSnapshotID {
			t.Errorf("migration chain is discontinuous at %d: %+v", index, ledger.Migrations)
		}
		previousMigrationSHA256 = migration.MigrationSHA256
	}
	reference := patternDetectorProfileMigrationReference()
	latest := ledger.Migrations[len(ledger.Migrations)-1]
	if reference.MigrationCount != len(ledger.Migrations) || reference.LatestMigrationID != latest.MigrationID ||
		reference.LatestFromSnapshotID != latest.FromSnapshotID || reference.LatestToSnapshotID != latest.ToSnapshotID ||
		reference.ChainScheme != ledger.ChainScheme || reference.ChainHeadSHA256 != latest.MigrationSHA256 {
		t.Fatalf("migration reference was not derived from ledger tail: %+v / %+v", reference, latest)
	}

	ledger.ProfileSets[0].DetectorProfiles[0].RuleID = "mutated"
	fresh, ok := PatternDetectorProfileMigrationLedgerSnapshot()
	if !ok || fresh.ProfileSets[0].DetectorProfiles[0].RuleID == "mutated" {
		t.Fatal("fresh migration ledger inherited caller mutation")
	}
}

func TestPatternDetectorProfileMigrationLedgerRejectsBrokenEvidence(t *testing.T) {
	ledger, ok := PatternDetectorProfileMigrationLedgerSnapshot()
	if !ok {
		t.Fatal("embedded detector profile migration ledger is invalid")
	}
	mutations := []func(*PatternDetectorProfileMigrationLedger){
		func(value *PatternDetectorProfileMigrationLedger) { value.ClaimBoundary = "mutated" },
		func(value *PatternDetectorProfileMigrationLedger) { value.ChainScheme = "mutated" },
		func(value *PatternDetectorProfileMigrationLedger) {
			value.ProfileSets[0].DetectorProfiles[0].AlgorithmSHA256 = strings.Repeat("0", 64)
		},
		func(value *PatternDetectorProfileMigrationLedger) {
			value.ProfileSets[0].DetectorProfiles[1].RuleID = value.ProfileSets[0].DetectorProfiles[0].RuleID
		},
		func(value *PatternDetectorProfileMigrationLedger) { value.Snapshots[2].EngineVersion = "mutated" },
		func(value *PatternDetectorProfileMigrationLedger) {
			value.Snapshots[2].DetectorManifestSHA256 = strings.Repeat("0", 64)
		},
		func(value *PatternDetectorProfileMigrationLedger) {
			value.ChangeSets[0].Result.Changes[0].Classes[0] = PatternDetectorAlgorithmDigestChanged
		},
		func(value *PatternDetectorProfileMigrationLedger) {
			value.Migrations[1].FromSnapshotID = value.Migrations[0].FromSnapshotID
		},
		func(value *PatternDetectorProfileMigrationLedger) { value.Migrations[1].MigrationID = "mutated" },
		func(value *PatternDetectorProfileMigrationLedger) {
			value.Migrations[1].PreviousMigrationSHA256 = strings.Repeat("0", 64)
		},
		func(value *PatternDetectorProfileMigrationLedger) {
			value.Migrations[0].MigrationSHA256 = strings.Repeat("f", 64)
		},
		func(value *PatternDetectorProfileMigrationLedger) {
			value.Migrations[0], value.Migrations[1] = value.Migrations[1], value.Migrations[0]
		},
	}
	for index, mutate := range mutations {
		broken := clonePatternDetectorProfileMigrationLedger(t, ledger)
		mutate(&broken)
		if validPatternDetectorProfileMigrationLedger(broken) {
			t.Errorf("broken migration ledger %d passed validation", index)
		}
	}
}

func TestPatternAnalysisPublishesAndValidatesMigrationLedgerReference(t *testing.T) {
	want := PatternDetectorProfileMigrationReference{
		LedgerID:             "bazi.pattern-detector-profile-migrations",
		Schema:               "pattern_detector_profile_migration_ledger_v2",
		SHA256:               "a72422e12e07adae349c147b3581f8c4829368f134f00a4f229c9a1c29d21825",
		MigrationCount:       4,
		LatestMigrationID:    "bazi.pattern-candidate-set-v33_to_v34",
		LatestFromSnapshotID: "bazi.pattern-candidate-set-v33",
		LatestToSnapshotID:   "bazi.pattern-candidate-set-v34",
		ChangeScheme:         "layered_detector_digest_delta_v1",
		ClaimBoundary:        "digest_evidence_only",
		ChainScheme:          "pattern_detector_profile_migration_chain_v1",
		ChainHeadSHA256:      "07dc296ad9e5dd0f834e40256c1e0f6033eb0ded435d0c76be6a0602ae0113bd",
	}
	pillars := []model.Pillar{
		{Gan: "丙", Zhi: "子"}, {Gan: "丙", Zhi: "寅"},
		{Gan: "甲", Zhi: "寅"}, {Gan: "戊", Zhi: "辰"},
	}
	for _, analysis := range []PatternAnalysis{
		AnalyzePatternExtended(pillars, "寅"),
		AnalyzePatternExtended(pillars[:3], "寅"),
	} {
		if !reflect.DeepEqual(analysis.DetectorMigration, want) {
			t.Errorf("detector migration reference = %+v, want %+v", analysis.DetectorMigration, want)
		}
		payload, err := json.Marshal(analysis)
		if err != nil {
			t.Fatal(err)
		}
		for _, fragment := range []string{`"detector_profile_migration"`, `"migration_count":4`, `"bazi.pattern-candidate-set-v33_to_v34"`, want.SHA256, want.ChainHeadSHA256} {
			if !strings.Contains(string(payload), fragment) {
				t.Errorf("pattern JSON missing migration evidence %q: %s", fragment, payload)
			}
		}
	}

	analysis := AnalyzePatternExtended(pillars, "寅")
	for _, mutate := range []func(*PatternDetectorProfileMigrationReference){
		func(reference *PatternDetectorProfileMigrationReference) { reference.LedgerID = "mutated" },
		func(reference *PatternDetectorProfileMigrationReference) { reference.Schema = "mutated" },
		func(reference *PatternDetectorProfileMigrationReference) { reference.SHA256 = strings.Repeat("0", 64) },
		func(reference *PatternDetectorProfileMigrationReference) { reference.MigrationCount++ },
		func(reference *PatternDetectorProfileMigrationReference) { reference.LatestMigrationID = "mutated" },
		func(reference *PatternDetectorProfileMigrationReference) { reference.LatestFromSnapshotID = "mutated" },
		func(reference *PatternDetectorProfileMigrationReference) { reference.LatestToSnapshotID = "mutated" },
		func(reference *PatternDetectorProfileMigrationReference) { reference.ChangeScheme = "mutated" },
		func(reference *PatternDetectorProfileMigrationReference) { reference.ClaimBoundary = "mutated" },
		func(reference *PatternDetectorProfileMigrationReference) { reference.ChainScheme = "mutated" },
		func(reference *PatternDetectorProfileMigrationReference) {
			reference.ChainHeadSHA256 = strings.Repeat("0", 64)
		},
	} {
		tampered := analysis
		mutate(&tampered.DetectorMigration)
		if reflect.DeepEqual(tampered, analysis) || ValidPatternAnalysis(tampered, pillars, "寅") {
			t.Fatal("tampered migration ledger reference passed strict validation")
		}
	}
}

func TestPatternDetectorProfileMigrationConsumersAndMetadataAreSynchronized(t *testing.T) {
	checks := map[string][]string{
		"../../../../API.md": {
			`"detector_profile_migration"`, `"pattern_detector_profile_migration_ledger_v2"`, PatternDetectorProfileMigrationLedgerSHA256, PatternDetectorProfileMigrationChainScheme,
		},
		"../../../../vue/src/api/chart.ts": {
			"export interface PatternDetectorProfileMigrationReference", "detector_profile_migration: PatternDetectorProfileMigrationReference", PatternDetectorProfileMigrationLedgerSHA256, PatternDetectorProfileMigrationChainScheme,
		},
	}
	for path, requiredValues := range checks {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, required := range requiredValues {
			if !strings.Contains(string(source), required) {
				t.Errorf("consumer %s missing migration ledger contract %q", path, required)
			}
		}
	}

	if EngineVersion != "bazi-engine-2026-07-17.27" || RuleVersion != "bazi-rules-2026-07-17.27" || PatternRuleID != "bazi.pattern-candidate-set-v34" ||
		PatternSchemaVersion != "pattern-candidates-2026-07-17.27" || PatternDetectorProfile != "classical_structural_detectors_v45" {
		t.Fatalf("migration version contract = %s/%s/%s/%s/%s", EngineVersion, RuleVersion, PatternRuleID, PatternSchemaVersion, PatternDetectorProfile)
	}
	for _, table := range DefaultRuleMeta().Tables {
		if table.Key != "pattern_candidates" {
			continue
		}
		if table.Version != "2026-07-17.27" || strings.Count(table.Description, "三层差异仍只存在于响应快照和人工版本说明中") != 1 {
			t.Errorf("migration metadata contract is incomplete: %+v", table)
		}
		for _, fragment := range []string{
			"pattern-candidate-set-v32新增pattern_detector_profile_migration_ledger_v1",
			"v30、v31与v32三份版本快照",
			"两条连续迁移记录",
			"由layered_detector_digest_delta_v1重算",
			"47d8ce51013f556366d069c0d9c83d5d239099c68c3888e3846184b4f78feae1",
			"只证明版本快照与摘要差异说明一致",
		} {
			if !strings.Contains(table.Description, fragment) {
				t.Errorf("pattern description missing %q", fragment)
			}
		}
		if count := strings.Count(table.Description, "迁移账本虽有整体摘要，但没有逐项前项链接"); count != 1 {
			t.Errorf("migration hash-chain metadata statement count = %d, want 1", count)
		}
		for _, fragment := range []string{
			"pattern-candidate-set-v33升级为pattern_detector_profile_migration_ledger_v2",
			"pattern_detector_profile_migration_chain_v1",
			"首项previous_migration_sha256固定为64个零",
			"绑定前后完整快照、解析后的逐规则摘要和预期分类",
			"v30至v33四份快照和三条连续迁移",
			"ea922b348fa81df44a70ece07f84b30fc5d8b50d2958e0012219d353ea5de2aa",
			"0c9258b2d186ee641df7455469fb3797a6f2f32cc931c553909fd15e0ab1be2f",
			"只证明追加链和工程证据未被静默重写",
		} {
			if !strings.Contains(table.Description, fragment) {
				t.Errorf("pattern description missing %q", fragment)
			}
		}
		return
	}
	t.Fatal("pattern-candidate rule table not found")
}

func clonePatternDetectorProfileMigrationLedger(t *testing.T, ledger PatternDetectorProfileMigrationLedger) PatternDetectorProfileMigrationLedger {
	t.Helper()
	payload, err := json.Marshal(ledger)
	if err != nil {
		t.Fatal(err)
	}
	var clone PatternDetectorProfileMigrationLedger
	if err := json.Unmarshal(payload, &clone); err != nil {
		t.Fatal(err)
	}
	return clone
}
