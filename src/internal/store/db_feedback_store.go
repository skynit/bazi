package store

import (
	"bazi/internal/model"

	"gorm.io/gorm"
)

// DBFeedbackStore implements fortune feedback persistence using GORM.
type DBFeedbackStore struct {
	db *gorm.DB
}

func NewDBFeedbackStore(db *gorm.DB) *DBFeedbackStore {
	return &DBFeedbackStore{db: db}
}

func (s *DBFeedbackStore) Create(feedback *model.FortuneFeedback) error {
	return s.db.Create(feedback).Error
}

func (s *DBFeedbackStore) SummaryByChartID(userID, chartID uint) ([]model.FeedbackSummaryItem, int64, int64, error) {
	type row struct {
		TargetType    string
		TargetID      string
		Rating        string
		EngineVersion string
		RuleVersion   string
		Count         int64
	}
	var rows []row
	query := s.db.Model(&model.FortuneFeedback{}).
		Select("target_type, target_id, rating, engine_version, rule_version, COUNT(*) AS count").
		Where("user_id = ? AND chart_id = ?", userID, chartID).
		Group("target_type, target_id, rating, engine_version, rule_version").
		Order("target_type ASC, target_id ASC, engine_version ASC, rule_version ASC, rating ASC")
	if err := query.Scan(&rows).Error; err != nil {
		return nil, 0, 0, err
	}
	items := make([]model.FeedbackSummaryItem, 0, len(rows))
	var total int64
	for _, r := range rows {
		items = append(items, model.FeedbackSummaryItem{
			TargetType: r.TargetType, TargetID: r.TargetID, Rating: r.Rating,
			EngineVersion: r.EngineVersion, RuleVersion: r.RuleVersion, Count: r.Count,
		})
		total += r.Count
	}
	var researchEligible int64
	if err := s.db.Model(&model.FortuneFeedback{}).
		Where("user_id = ? AND chart_id = ? AND consent_research = ?", userID, chartID, true).
		Count(&researchEligible).Error; err != nil {
		return nil, 0, 0, err
	}
	return items, total, researchEligible, nil
}
