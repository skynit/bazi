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

func (s *DBFeedbackStore) SummaryByChartID(userID, chartID uint) ([]model.FeedbackSummaryItem, int64, error) {
	type row struct {
		Rating string
		Count  int64
	}
	var rows []row
	query := s.db.Model(&model.FortuneFeedback{}).
		Select("rating, COUNT(*) AS count").
		Where("user_id = ? AND chart_id = ?", userID, chartID).
		Group("rating").
		Order("rating ASC")
	if err := query.Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	items := make([]model.FeedbackSummaryItem, 0, len(rows))
	var total int64
	for _, r := range rows {
		items = append(items, model.FeedbackSummaryItem{Rating: r.Rating, Count: r.Count})
		total += r.Count
	}
	return items, total, nil
}
