package bazi

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

const (
	PatternDetectorProfileReleaseAnchorSchema              = "pattern_detector_profile_release_anchor_v1"
	PatternDetectorProfileReleaseAnchorID                  = "bazi.pattern-detector-profile-release-anchor-v34"
	PatternDetectorProfileReleaseAnchorArtifactPath        = "release/pattern-detector-profile-anchor.json"
	PatternDetectorProfileReleaseAnchorSHA256              = "ebd6323f28715695aa3c4ee9038e74d261c9fa34b422037266c4b097e3086a2e"
	PatternDetectorProfileReleaseAnchorVerificationProfile = "repository_ci_cross_check_v1"
	PatternDetectorProfileReleaseAnchorTrustBoundary       = "unsigned_repository_ci_artifact"
)

type PatternDetectorProfileReleaseAnchor struct {
	Schema                 string `json:"schema"`
	AnchorID               string `json:"anchor_id"`
	EngineVersion          string `json:"engine_version"`
	RuleVersion            string `json:"rule_version"`
	RuleID                 string `json:"rule_id"`
	SchemaVersion          string `json:"schema_version"`
	DetectorProfile        string `json:"detector_profile"`
	DetectorManifestSHA256 string `json:"detector_manifest_sha256"`
	DetectorProfilesSHA256 string `json:"detector_profiles_sha256"`
	LedgerID               string `json:"ledger_id"`
	LedgerSchema           string `json:"ledger_schema"`
	LedgerSHA256           string `json:"ledger_sha256"`
	MigrationCount         int    `json:"migration_count"`
	ChainScheme            string `json:"chain_scheme"`
	ChainHeadSHA256        string `json:"chain_head_sha256"`
	VerificationProfile    string `json:"verification_profile"`
	TrustBoundary          string `json:"trust_boundary"`
	ClaimBoundary          string `json:"claim_boundary"`
}

type PatternDetectorProfileReleaseAnchorReference struct {
	Schema              string `json:"schema"`
	AnchorID            string `json:"anchor_id"`
	ArtifactPath        string `json:"artifact_path"`
	SHA256              string `json:"sha256"`
	VerificationProfile string `json:"verification_profile"`
	TrustBoundary       string `json:"trust_boundary"`
	ClaimBoundary       string `json:"claim_boundary"`
}

func patternDetectorProfileReleaseAnchorReference() PatternDetectorProfileReleaseAnchorReference {
	return PatternDetectorProfileReleaseAnchorReference{
		Schema:              PatternDetectorProfileReleaseAnchorSchema,
		AnchorID:            PatternDetectorProfileReleaseAnchorID,
		ArtifactPath:        PatternDetectorProfileReleaseAnchorArtifactPath,
		SHA256:              PatternDetectorProfileReleaseAnchorSHA256,
		VerificationProfile: PatternDetectorProfileReleaseAnchorVerificationProfile,
		TrustBoundary:       PatternDetectorProfileReleaseAnchorTrustBoundary,
		ClaimBoundary:       PatternDetectorProfileChangeEvidenceLimit,
	}
}

func ValidPatternDetectorProfileReleaseAnchor(anchor PatternDetectorProfileReleaseAnchor) bool {
	detectors := patternDetectorRegistry()
	migration := patternDetectorProfileMigrationReference()
	return anchor.Schema == PatternDetectorProfileReleaseAnchorSchema &&
		anchor.AnchorID == PatternDetectorProfileReleaseAnchorID &&
		anchor.EngineVersion == EngineVersion && anchor.RuleVersion == RuleVersion && anchor.RuleID == PatternRuleID &&
		anchor.SchemaVersion == PatternSchemaVersion && anchor.DetectorProfile == PatternDetectorProfile &&
		anchor.DetectorManifestSHA256 == patternDetectorManifestSHA256(detectors) &&
		anchor.DetectorProfilesSHA256 == patternDetectorProfilesSHA256(patternDetectorProfileDigests(detectors)) &&
		anchor.LedgerID == migration.LedgerID && anchor.LedgerSchema == migration.Schema && anchor.LedgerSHA256 == migration.SHA256 &&
		anchor.MigrationCount == migration.MigrationCount && anchor.ChainScheme == migration.ChainScheme && anchor.ChainHeadSHA256 == migration.ChainHeadSHA256 &&
		anchor.VerificationProfile == PatternDetectorProfileReleaseAnchorVerificationProfile &&
		anchor.TrustBoundary == PatternDetectorProfileReleaseAnchorTrustBoundary && anchor.ClaimBoundary == PatternDetectorProfileChangeEvidenceLimit
}

func patternDetectorProfilesSHA256(profiles []PatternDetectorProfileDigest) string {
	payload, err := json.Marshal(profiles)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}
