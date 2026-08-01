package model

import "gorm.io/gorm"

const (
	FeedbackRatingClear             = "clear"
	FeedbackRatingHelpful           = "helpful"
	FeedbackRatingInsufficientBasis = "insufficient_basis"
	FeedbackRatingTooGeneric        = "too_generic"
	FeedbackRatingConfusing         = "confusing"

	FeedbackSummaryScopeInterpretationQuality = "interpretation_quality_not_outcome_accuracy"
)

// FortuneFeedback stores anonymizable user feedback for interpretation and
// fortune sections. Consent flags decide whether it can enter research/training
// datasets.
type FortuneFeedback struct {
	gorm.Model
	UserID     uint   `gorm:"not null;index" json:"user_id"`
	ChartID    uint   `gorm:"not null;index" json:"chart_id"`
	TargetType string `gorm:"type:varchar(32);not null;default:'section';index" json:"target_type"`
	TargetID   string `gorm:"type:varchar(128);not null;default:'';index" json:"target_id"`
	Rating     string `gorm:"type:varchar(32);not null;index" json:"rating"`
	Tags       string `gorm:"type:text" json:"tags"` // JSON array encoded as text for MySQL/SQLite portability.
	Comment    string `gorm:"type:text" json:"comment"`
	// Legacy event fields remain mapped only for privacy audit and disposal.
	// The public feedback API neither accepts nor populates them.
	EventYear       int    `gorm:"default:0;index" json:"-"`
	EventCategory   string `gorm:"type:varchar(64);not null;default:'';index" json:"-"`
	ConsentResearch bool   `gorm:"not null;default:false" json:"consent_research"`
	ConsentTraining bool   `gorm:"not null;default:false" json:"consent_training"`
	EngineVersion   string `gorm:"type:varchar(64);not null;default:'';index" json:"engine_version"`
	RuleVersion     string `gorm:"type:varchar(64);not null;default:'';index" json:"rule_version"`
}
