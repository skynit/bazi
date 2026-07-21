package bazi

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"strings"
	"sync"
)

const (
	PatternDetectorProfileMigrationLedgerID      = "bazi.pattern-detector-profile-migrations"
	PatternDetectorProfileMigrationLedgerSchema  = "pattern_detector_profile_migration_ledger_v2"
	PatternDetectorProfileMigrationLedgerSHA256  = "a72422e12e07adae349c147b3581f8c4829368f134f00a4f229c9a1c29d21825"
	PatternDetectorProfileMigrationChainScheme   = "pattern_detector_profile_migration_chain_v1"
	PatternDetectorProfileMigrationGenesisSHA256 = "0000000000000000000000000000000000000000000000000000000000000000"
)

//go:embed rules/pattern_detector_profile_migrations.json
var patternDetectorProfileMigrationLedgerJSON string

var (
	patternDetectorProfileMigrationValidationOnce sync.Once
	patternDetectorProfileMigrationLedgerValid    bool
	patternDetectorProfileMigrationReferenceValue PatternDetectorProfileMigrationReference
)

type PatternDetectorProfileMigrationLedger struct {
	Schema        string                                      `json:"schema"`
	ClaimBoundary string                                      `json:"claim_boundary"`
	ChainScheme   string                                      `json:"chain_scheme"`
	ProfileSets   []PatternDetectorProfileMigrationProfileSet `json:"profile_sets"`
	Snapshots     []PatternDetectorProfileMigrationSnapshot   `json:"snapshots"`
	ChangeSets    []PatternDetectorProfileMigrationChangeSet  `json:"change_sets"`
	Migrations    []PatternDetectorProfileMigration           `json:"migrations"`
}

type PatternDetectorProfileMigrationProfileSet struct {
	ProfileSetID     string                         `json:"profile_set_id"`
	DetectorProfiles []PatternDetectorProfileDigest `json:"detector_profiles"`
}

type PatternDetectorProfileMigrationSnapshot struct {
	SnapshotID             string `json:"snapshot_id"`
	EngineVersion          string `json:"engine_version"`
	RuleVersion            string `json:"rule_version"`
	RuleID                 string `json:"rule_id"`
	SchemaVersion          string `json:"schema_version"`
	DetectorProfile        string `json:"detector_profile"`
	DetectorManifestSHA256 string `json:"detector_manifest_sha256"`
	ProfileSetID           string `json:"profile_set_id"`
}

type PatternDetectorProfileMigrationChangeSet struct {
	ChangeSetID string                          `json:"change_set_id"`
	Result      PatternDetectorProfileChangeSet `json:"result"`
}

type PatternDetectorProfileMigration struct {
	MigrationID             string `json:"migration_id"`
	FromSnapshotID          string `json:"from_snapshot_id"`
	ToSnapshotID            string `json:"to_snapshot_id"`
	ExpectedChangeSetID     string `json:"expected_change_set_id"`
	PreviousMigrationSHA256 string `json:"previous_migration_sha256"`
	MigrationSHA256         string `json:"migration_sha256"`
}

type PatternDetectorProfileMigrationReference struct {
	LedgerID             string `json:"ledger_id"`
	Schema               string `json:"schema"`
	SHA256               string `json:"sha256"`
	MigrationCount       int    `json:"migration_count"`
	LatestMigrationID    string `json:"latest_migration_id"`
	LatestFromSnapshotID string `json:"latest_from_snapshot_id"`
	LatestToSnapshotID   string `json:"latest_to_snapshot_id"`
	ChangeScheme         string `json:"change_scheme"`
	ClaimBoundary        string `json:"claim_boundary"`
	ChainScheme          string `json:"chain_scheme"`
	ChainHeadSHA256      string `json:"chain_head_sha256"`
}

type patternDetectorProfileMigrationHashPayload struct {
	Scheme                  string                                  `json:"scheme"`
	PreviousMigrationSHA256 string                                  `json:"previous_migration_sha256"`
	MigrationID             string                                  `json:"migration_id"`
	FromSnapshot            PatternDetectorProfileMigrationSnapshot `json:"from_snapshot"`
	FromDetectorProfiles    []PatternDetectorProfileDigest          `json:"from_detector_profiles"`
	ToSnapshot              PatternDetectorProfileMigrationSnapshot `json:"to_snapshot"`
	ToDetectorProfiles      []PatternDetectorProfileDigest          `json:"to_detector_profiles"`
	ExpectedChangeSetID     string                                  `json:"expected_change_set_id"`
	ExpectedChanges         PatternDetectorProfileChangeSet         `json:"expected_changes"`
}

// PatternDetectorProfileMigrationLedgerSnapshot returns a newly decoded and
// fully validated ledger. Callers cannot mutate later loads through the result.
func PatternDetectorProfileMigrationLedgerSnapshot() (PatternDetectorProfileMigrationLedger, bool) {
	var ledger PatternDetectorProfileMigrationLedger
	if err := json.Unmarshal([]byte(patternDetectorProfileMigrationLedgerJSON), &ledger); err != nil {
		return PatternDetectorProfileMigrationLedger{}, false
	}
	if !validPatternDetectorProfileMigrationLedger(ledger) {
		return PatternDetectorProfileMigrationLedger{}, false
	}
	return ledger, true
}

func patternDetectorProfileMigrationReference() PatternDetectorProfileMigrationReference {
	patternDetectorProfileMigrationValidationOnce.Do(func() {
		ledger, valid := PatternDetectorProfileMigrationLedgerSnapshot()
		patternDetectorProfileMigrationLedgerValid = valid && patternDetectorProfileMigrationLedgerSHA256() == PatternDetectorProfileMigrationLedgerSHA256
		if patternDetectorProfileMigrationLedgerValid {
			latest := ledger.Migrations[len(ledger.Migrations)-1]
			patternDetectorProfileMigrationReferenceValue = PatternDetectorProfileMigrationReference{
				LedgerID:             PatternDetectorProfileMigrationLedgerID,
				Schema:               ledger.Schema,
				SHA256:               PatternDetectorProfileMigrationLedgerSHA256,
				MigrationCount:       len(ledger.Migrations),
				LatestMigrationID:    latest.MigrationID,
				LatestFromSnapshotID: latest.FromSnapshotID,
				LatestToSnapshotID:   latest.ToSnapshotID,
				ChangeScheme:         PatternDetectorProfileChangeScheme,
				ClaimBoundary:        ledger.ClaimBoundary,
				ChainScheme:          ledger.ChainScheme,
				ChainHeadSHA256:      latest.MigrationSHA256,
			}
		}
	})
	if !patternDetectorProfileMigrationLedgerValid {
		panic("invalid embedded pattern detector profile migration ledger")
	}
	return patternDetectorProfileMigrationReferenceValue
}

func validPatternDetectorProfileMigrationLedger(ledger PatternDetectorProfileMigrationLedger) bool {
	if ledger.Schema != PatternDetectorProfileMigrationLedgerSchema || ledger.ClaimBoundary != PatternDetectorProfileChangeEvidenceLimit ||
		ledger.ChainScheme != PatternDetectorProfileMigrationChainScheme ||
		len(ledger.ProfileSets) == 0 || len(ledger.Snapshots) == 0 || len(ledger.ChangeSets) == 0 || len(ledger.Migrations) == 0 {
		return false
	}

	profileSets := make(map[string][]PatternDetectorProfileDigest, len(ledger.ProfileSets))
	for _, profileSet := range ledger.ProfileSets {
		if profileSet.ProfileSetID == "" || !canonicalPatternDetectorProfileDigests(profileSet.DetectorProfiles) {
			return false
		}
		if _, duplicate := profileSets[profileSet.ProfileSetID]; duplicate {
			return false
		}
		profileSets[profileSet.ProfileSetID] = profileSet.DetectorProfiles
	}

	snapshots := make(map[string]PatternDetectorProfileMigrationSnapshot, len(ledger.Snapshots))
	for _, snapshot := range ledger.Snapshots {
		if snapshot.SnapshotID == "" || snapshot.EngineVersion == "" || snapshot.RuleVersion == "" || snapshot.RuleID == "" ||
			snapshot.SchemaVersion == "" || snapshot.DetectorProfile == "" || !validPatternDetectorDigest(snapshot.DetectorManifestSHA256) ||
			snapshot.ProfileSetID == "" || snapshot.SnapshotID != snapshot.RuleID {
			return false
		}
		if _, ok := profileSets[snapshot.ProfileSetID]; !ok {
			return false
		}
		if _, duplicate := snapshots[snapshot.SnapshotID]; duplicate {
			return false
		}
		snapshots[snapshot.SnapshotID] = snapshot
	}

	changeSets := make(map[string]PatternDetectorProfileChangeSet, len(ledger.ChangeSets))
	for _, changeSet := range ledger.ChangeSets {
		if changeSet.ChangeSetID == "" || !validPatternDetectorProfileChangeSet(changeSet.Result) {
			return false
		}
		if _, duplicate := changeSets[changeSet.ChangeSetID]; duplicate {
			return false
		}
		changeSets[changeSet.ChangeSetID] = changeSet.Result
	}

	migrationIDs := make(map[string]struct{}, len(ledger.Migrations))
	previousMigrationSHA256 := PatternDetectorProfileMigrationGenesisSHA256
	for index, migration := range ledger.Migrations {
		from, fromOK := snapshots[migration.FromSnapshotID]
		to, toOK := snapshots[migration.ToSnapshotID]
		expected, expectedOK := changeSets[migration.ExpectedChangeSetID]
		wantMigrationID, validMigrationID := patternDetectorProfileMigrationID(migration.FromSnapshotID, migration.ToSnapshotID)
		if !validMigrationID || migration.MigrationID != wantMigrationID || !fromOK || !toOK || !expectedOK || from.SnapshotID == to.SnapshotID ||
			migration.PreviousMigrationSHA256 != previousMigrationSHA256 || !validPatternDetectorDigest(migration.MigrationSHA256) {
			return false
		}
		if _, duplicate := migrationIDs[migration.MigrationID]; duplicate {
			return false
		}
		migrationIDs[migration.MigrationID] = struct{}{}
		got := ComparePatternDetectorProfiles(profileSets[from.ProfileSetID], profileSets[to.ProfileSetID])
		if !reflect.DeepEqual(got, expected) {
			return false
		}
		migrationSHA256 := patternDetectorProfileMigrationSHA256(ledger.ChainScheme, migration, from, profileSets[from.ProfileSetID], to, profileSets[to.ProfileSetID], expected)
		if migrationSHA256 != migration.MigrationSHA256 {
			return false
		}
		previousMigrationSHA256 = migration.MigrationSHA256
		if index > 0 && ledger.Migrations[index-1].ToSnapshotID != migration.FromSnapshotID {
			return false
		}
	}

	latestMigration := ledger.Migrations[len(ledger.Migrations)-1]
	latest, ok := snapshots[latestMigration.ToSnapshotID]
	if !ok || latest.EngineVersion != EngineVersion || latest.RuleVersion != RuleVersion || latest.RuleID != PatternRuleID ||
		latest.SchemaVersion != PatternSchemaVersion || latest.DetectorProfile != PatternDetectorProfile {
		return false
	}
	detectors := patternDetectorRegistry()
	return latest.DetectorManifestSHA256 == patternDetectorManifestSHA256(detectors) &&
		reflect.DeepEqual(profileSets[latest.ProfileSetID], patternDetectorProfileDigests(detectors))
}

func patternDetectorProfileMigrationSHA256(
	scheme string,
	migration PatternDetectorProfileMigration,
	from PatternDetectorProfileMigrationSnapshot,
	fromProfiles []PatternDetectorProfileDigest,
	to PatternDetectorProfileMigrationSnapshot,
	toProfiles []PatternDetectorProfileDigest,
	expected PatternDetectorProfileChangeSet,
) string {
	payload, err := json.Marshal(patternDetectorProfileMigrationHashPayload{
		Scheme:                  scheme,
		PreviousMigrationSHA256: migration.PreviousMigrationSHA256,
		MigrationID:             migration.MigrationID,
		FromSnapshot:            from,
		FromDetectorProfiles:    fromProfiles,
		ToSnapshot:              to,
		ToDetectorProfiles:      toProfiles,
		ExpectedChangeSetID:     migration.ExpectedChangeSetID,
		ExpectedChanges:         expected,
	})
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

func patternDetectorProfileMigrationID(fromSnapshotID, toSnapshotID string) (string, bool) {
	const prefix = "bazi.pattern-candidate-set-"
	if !strings.HasPrefix(fromSnapshotID, prefix) || !strings.HasPrefix(toSnapshotID, prefix) {
		return "", false
	}
	fromVersion := strings.TrimPrefix(fromSnapshotID, prefix)
	toVersion := strings.TrimPrefix(toSnapshotID, prefix)
	if fromVersion == "" || toVersion == "" {
		return "", false
	}
	return prefix + fromVersion + "_to_" + toVersion, true
}

func canonicalPatternDetectorProfileDigests(digests []PatternDetectorProfileDigest) bool {
	if len(digests) == 0 {
		return false
	}
	if _, ok := indexPatternDetectorProfileDigests(digests); !ok {
		return false
	}
	for index := 1; index < len(digests); index++ {
		if digests[index-1].RuleID >= digests[index].RuleID {
			return false
		}
	}
	return true
}

func validPatternDetectorProfileChangeSet(result PatternDetectorProfileChangeSet) bool {
	if result.Scheme != PatternDetectorProfileChangeScheme || result.Status != PatternDetectorProfilesCompared || result.Changes == nil {
		return false
	}
	for index, change := range result.Changes {
		if change.RuleID == "" || len(change.Classes) == 0 || (index > 0 && result.Changes[index-1].RuleID >= change.RuleID) ||
			!canonicalPatternDetectorProfileChangeClasses(change.Classes) {
			return false
		}
	}
	return true
}

func canonicalPatternDetectorProfileChangeClasses(classes []PatternDetectorProfileChangeClass) bool {
	if len(classes) == 1 {
		switch classes[0] {
		case PatternDetectorAdded, PatternDetectorRemoved, PatternDetectorAlgorithmDigestChanged,
			PatternDetectorBehaviorDigestChanged, PatternDetectorSemanticDigestChanged, PatternDetectorLayeredDigestsUnchanged:
			return true
		default:
			return false
		}
	}
	previous := 0
	for _, class := range classes {
		order := 0
		switch class {
		case PatternDetectorAlgorithmDigestChanged:
			order = 1
		case PatternDetectorBehaviorDigestChanged:
			order = 2
		case PatternDetectorSemanticDigestChanged:
			order = 3
		default:
			return false
		}
		if order <= previous {
			return false
		}
		previous = order
	}
	return true
}

func patternDetectorProfileMigrationLedgerSHA256() string {
	var compact bytes.Buffer
	if err := json.Compact(&compact, []byte(patternDetectorProfileMigrationLedgerJSON)); err != nil {
		return ""
	}
	sum := sha256.Sum256(compact.Bytes())
	return hex.EncodeToString(sum[:])
}
