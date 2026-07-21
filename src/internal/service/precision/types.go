package precision

type CaseTier string

const (
	TierUnclassified CaseTier = "unclassified"
	TierGold         CaseTier = "gold"
	TierSilver       CaseTier = "silver"
	TierBronze       CaseTier = "bronze"
	TierFeedback     CaseTier = "feedback"
)

// CaseMetadata is the machine-readable schema for future precision fixtures.
// Existing testdata files are read compatibly and reported with metadata
// warnings when these fields are absent.
type CaseMetadata struct {
	Tier             CaseTier `json:"tier"`
	SourceName       string   `json:"source_name"`
	SourceURL        string   `json:"source_url"`
	License          string   `json:"license"`
	SourceHash       string   `json:"source_hash"`
	Confidence       float64  `json:"confidence"`
	ReviewStatus     string   `json:"review_status"`
	Reviewers        []string `json:"reviewers"`
	Purpose          string   `json:"purpose"`
	QuarantineReason string   `json:"quarantine_reason,omitempty"`
}

type Report struct {
	Version           string          `json:"version"`
	ComparatorVersion string          `json:"comparator_version"`
	ComparatorHash    string          `json:"comparator_registry_hash"`
	BaselineKind      string          `json:"baseline_kind"`
	PublicationStatus string          `json:"publication_status"`
	GeneratedAt       string          `json:"generated_at"`
	Modules           []ModuleReport  `json:"modules"`
	External          []ExternalProbe `json:"external"`
	Warnings          []string        `json:"warnings"`
	ReleaseBlockers   []string        `json:"release_blockers"`
	TotalCases        int             `json:"total_cases"`
	EvaluatedCases    int             `json:"evaluated_cases"`
	NonAssertiveCases int             `json:"non_assertive_cases"`
	SkippedCases      int             `json:"skipped_cases"`
	QuarantinedCases  int             `json:"quarantined_cases"`
	UnsupportedChecks int             `json:"unsupported_checks"`
	DiagnosticChecks  int             `json:"diagnostic_checks"`
	DiagnosticPassed  int             `json:"diagnostic_passed"`
	DiagnosticFailed  int             `json:"diagnostic_failed"`
	PublishableCases  int             `json:"publishable_cases"`
	PublishableChecks int             `json:"publishable_checks"`
	PublishablePassed int             `json:"publishable_passed"`
	PublishableFailed int             `json:"publishable_failed"`
}

type ModuleReport struct {
	Name              string                  `json:"name"`
	Path              string                  `json:"path"`
	ProfileID         string                  `json:"profile_id,omitempty"`
	EngineVersion     string                  `json:"engine_version,omitempty"`
	RuleVersion       string                  `json:"rule_version,omitempty"`
	RuleSchool        string                  `json:"rule_school,omitempty"`
	Cases             int                     `json:"cases"`
	EvaluatedCases    int                     `json:"evaluated_cases"`
	NonAssertiveCases int                     `json:"non_assertive_cases"`
	SkippedCases      int                     `json:"skipped_cases"`
	QuarantinedCases  int                     `json:"quarantined_cases"`
	UnsupportedChecks int                     `json:"unsupported_checks"`
	MissingMetadata   int                     `json:"missing_metadata"`
	DiagnosticChecks  int                     `json:"diagnostic_checks"`
	DiagnosticPassed  int                     `json:"diagnostic_passed"`
	DiagnosticFailed  int                     `json:"diagnostic_failed"`
	PublishableCases  int                     `json:"publishable_cases"`
	PublishableChecks int                     `json:"publishable_checks"`
	PublishablePassed int                     `json:"publishable_passed"`
	PublishableFailed int                     `json:"publishable_failed"`
	TierBreakdown     map[CaseTier]TierReport `json:"tier_breakdown"`
	SkipReasons       map[string]int          `json:"skip_reasons,omitempty"`
	DuplicateCaseIDs  []string                `json:"duplicate_case_ids,omitempty"`
	BoundaryStatus    string                  `json:"boundary_status,omitempty"`
	Failures          []CheckResult           `json:"failures,omitempty"`
	Warnings          []string                `json:"warnings,omitempty"`
}

type TierReport struct {
	Cases             int `json:"cases"`
	EvaluatedCases    int `json:"evaluated_cases"`
	NonAssertiveCases int `json:"non_assertive_cases"`
	SkippedCases      int `json:"skipped_cases"`
	QuarantinedCases  int `json:"quarantined_cases"`
	DiagnosticChecks  int `json:"diagnostic_checks"`
	DiagnosticPassed  int `json:"diagnostic_passed"`
	DiagnosticFailed  int `json:"diagnostic_failed"`
	PublishableCases  int `json:"publishable_cases"`
	PublishableChecks int `json:"publishable_checks"`
	PublishablePassed int `json:"publishable_passed"`
	PublishableFailed int `json:"publishable_failed"`
}

type CheckResult struct {
	CaseID string `json:"case_id"`
	Field  string `json:"field"`
	Want   string `json:"want"`
	Got    string `json:"got"`
	Status string `json:"status"`
	Note   string `json:"note,omitempty"`
}

type ExternalProbe struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Path   string `json:"path,omitempty"`
	Note   string `json:"note,omitempty"`
}
