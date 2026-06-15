package precision

type CaseTier string

const (
	TierGold     CaseTier = "gold"
	TierSilver   CaseTier = "silver"
	TierBronze   CaseTier = "bronze"
	TierFeedback CaseTier = "feedback"
)

// CaseMetadata is the machine-readable schema for future precision fixtures.
// Existing testdata files are read compatibly and reported with metadata
// warnings when these fields are absent.
type CaseMetadata struct {
	Tier         CaseTier `json:"tier"`
	SourceName   string   `json:"source_name"`
	SourceURL    string   `json:"source_url"`
	License      string   `json:"license"`
	SourceHash   string   `json:"source_hash"`
	Confidence   float64  `json:"confidence"`
	ReviewStatus string   `json:"review_status"`
	Reviewers    []string `json:"reviewers"`
}

type Report struct {
	Version       string          `json:"version"`
	GeneratedAt   string          `json:"generated_at"`
	Modules       []ModuleReport  `json:"modules"`
	External      []ExternalProbe `json:"external"`
	Warnings      []string        `json:"warnings"`
	TotalCases    int             `json:"total_cases"`
	TotalChecks   int             `json:"total_checks"`
	PassedChecks  int             `json:"passed_checks"`
	FailedChecks  int             `json:"failed_checks"`
	SkippedChecks int             `json:"skipped_checks"`
}

type ModuleReport struct {
	Name            string        `json:"name"`
	Path            string        `json:"path"`
	Cases           int           `json:"cases"`
	Checks          int           `json:"checks"`
	Passed          int           `json:"passed"`
	Failed          int           `json:"failed"`
	Skipped         int           `json:"skipped"`
	MissingMetadata int           `json:"missing_metadata"`
	BoundaryStatus  string        `json:"boundary_status,omitempty"`
	Failures        []CheckResult `json:"failures,omitempty"`
	Warnings        []string      `json:"warnings,omitempty"`
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
