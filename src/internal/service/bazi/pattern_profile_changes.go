package bazi

import "sort"

const (
	PatternDetectorProfileChangeScheme        = "layered_detector_digest_delta_v1"
	PatternDetectorProfileChangeAlignmentKey  = "rule_id"
	PatternDetectorBehaviorEvidenceScope      = "simple_full_truth_table_complex_partial_contract"
	PatternDetectorProfileChangeEvidenceLimit = "digest_evidence_only"
)

type PatternDetectorProfileChangeClass string

const (
	PatternDetectorAdded                   PatternDetectorProfileChangeClass = "detector_added"
	PatternDetectorRemoved                 PatternDetectorProfileChangeClass = "detector_removed"
	PatternDetectorAlgorithmDigestChanged  PatternDetectorProfileChangeClass = "algorithm_digest_changed"
	PatternDetectorBehaviorDigestChanged   PatternDetectorProfileChangeClass = "behavior_evidence_digest_changed"
	PatternDetectorSemanticDigestChanged   PatternDetectorProfileChangeClass = "semantic_profile_digest_changed"
	PatternDetectorLayeredDigestsUnchanged PatternDetectorProfileChangeClass = "layered_digests_unchanged"
)

type PatternDetectorProfileChangeContract struct {
	Scheme                string                              `json:"scheme"`
	AlignmentKey          string                              `json:"alignment_key"`
	Classes               []PatternDetectorProfileChangeClass `json:"classes"`
	BehaviorEvidenceScope string                              `json:"behavior_evidence_scope"`
	InferenceBoundary     string                              `json:"inference_boundary"`
}

type PatternDetectorProfileComparisonStatus string

const (
	PatternDetectorProfilesCompared     PatternDetectorProfileComparisonStatus = "compared"
	PatternDetectorProfilesInvalidInput PatternDetectorProfileComparisonStatus = "invalid_input"
)

type PatternDetectorProfileChange struct {
	RuleID  string                              `json:"rule_id"`
	Classes []PatternDetectorProfileChangeClass `json:"classes"`
}

type PatternDetectorProfileChangeSet struct {
	Scheme  string                                 `json:"scheme"`
	Status  PatternDetectorProfileComparisonStatus `json:"status"`
	Changes []PatternDetectorProfileChange         `json:"changes"`
}

func patternDetectorProfileChangeContract() PatternDetectorProfileChangeContract {
	return PatternDetectorProfileChangeContract{
		Scheme:       PatternDetectorProfileChangeScheme,
		AlignmentKey: PatternDetectorProfileChangeAlignmentKey,
		Classes: []PatternDetectorProfileChangeClass{
			PatternDetectorAdded,
			PatternDetectorRemoved,
			PatternDetectorAlgorithmDigestChanged,
			PatternDetectorBehaviorDigestChanged,
			PatternDetectorSemanticDigestChanged,
			PatternDetectorLayeredDigestsUnchanged,
		},
		BehaviorEvidenceScope: PatternDetectorBehaviorEvidenceScope,
		InferenceBoundary:     PatternDetectorProfileChangeEvidenceLimit,
	}
}

// ComparePatternDetectorProfiles compares two detector-profile snapshots.
// It reports digest evidence only and does not infer traditional or predictive equivalence.
func ComparePatternDetectorProfiles(before, after []PatternDetectorProfileDigest) PatternDetectorProfileChangeSet {
	result := PatternDetectorProfileChangeSet{
		Scheme:  PatternDetectorProfileChangeScheme,
		Status:  PatternDetectorProfilesInvalidInput,
		Changes: make([]PatternDetectorProfileChange, 0),
	}
	beforeByRule, ok := indexPatternDetectorProfileDigests(before)
	if !ok {
		return result
	}
	afterByRule, ok := indexPatternDetectorProfileDigests(after)
	if !ok {
		return result
	}

	ruleIDs := make([]string, 0, len(beforeByRule)+len(afterByRule))
	seen := make(map[string]struct{}, len(beforeByRule)+len(afterByRule))
	for ruleID := range beforeByRule {
		ruleIDs = append(ruleIDs, ruleID)
		seen[ruleID] = struct{}{}
	}
	for ruleID := range afterByRule {
		if _, exists := seen[ruleID]; !exists {
			ruleIDs = append(ruleIDs, ruleID)
		}
	}
	sort.Strings(ruleIDs)

	result.Status = PatternDetectorProfilesCompared
	for _, ruleID := range ruleIDs {
		beforeDigest, beforeExists := beforeByRule[ruleID]
		afterDigest, afterExists := afterByRule[ruleID]
		change := PatternDetectorProfileChange{RuleID: ruleID, Classes: make([]PatternDetectorProfileChangeClass, 0, 3)}
		switch {
		case !beforeExists:
			change.Classes = append(change.Classes, PatternDetectorAdded)
		case !afterExists:
			change.Classes = append(change.Classes, PatternDetectorRemoved)
		default:
			if beforeDigest.AlgorithmSHA256 != afterDigest.AlgorithmSHA256 {
				change.Classes = append(change.Classes, PatternDetectorAlgorithmDigestChanged)
			}
			if beforeDigest.BehaviorSHA256 != afterDigest.BehaviorSHA256 {
				change.Classes = append(change.Classes, PatternDetectorBehaviorDigestChanged)
			}
			if beforeDigest.ProfileSHA256 != afterDigest.ProfileSHA256 {
				change.Classes = append(change.Classes, PatternDetectorSemanticDigestChanged)
			}
			if len(change.Classes) == 0 {
				change.Classes = append(change.Classes, PatternDetectorLayeredDigestsUnchanged)
			}
		}
		result.Changes = append(result.Changes, change)
	}
	return result
}

func indexPatternDetectorProfileDigests(digests []PatternDetectorProfileDigest) (map[string]PatternDetectorProfileDigest, bool) {
	indexed := make(map[string]PatternDetectorProfileDigest, len(digests))
	for _, digest := range digests {
		if digest.RuleID == "" || !validPatternDetectorDigest(digest.AlgorithmSHA256) ||
			!validPatternDetectorDigest(digest.BehaviorSHA256) || !validPatternDetectorDigest(digest.ProfileSHA256) {
			return nil, false
		}
		if _, duplicate := indexed[digest.RuleID]; duplicate {
			return nil, false
		}
		indexed[digest.RuleID] = digest
	}
	return indexed, true
}

func validPatternDetectorDigest(digest string) bool {
	if len(digest) != 64 {
		return false
	}
	for index := range digest {
		if (digest[index] < '0' || digest[index] > '9') && (digest[index] < 'a' || digest[index] > 'f') {
			return false
		}
	}
	return true
}
