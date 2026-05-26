package store

import (
	"bazi/internal/model"

	"gorm.io/gorm"
)

// DBFortuneStore implements fortune history persistence using GORM.
type DBFortuneStore struct {
	db *gorm.DB
}

// NewDBFortuneStore creates a DBFortuneStore.
func NewDBFortuneStore(db *gorm.DB) *DBFortuneStore {
	return &DBFortuneStore{db: db}
}

// Create inserts a new fortune record.
func (s *DBFortuneStore) Create(record *model.FortuneRecord) error {
	return s.db.Create(record).Error
}

// ListByChartID returns fortune records for a chart with pagination.
func (s *DBFortuneStore) ListByChartID(chartID uint, page, pageSize int) ([]model.HistoryResponse, int64, error) {
	var records []model.FortuneRecord
	var total int64

	offset := (page - 1) * pageSize
	query := s.db.Model(&model.FortuneRecord{}).Where("chart_id = ?", chartID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&records).Error; err != nil {
		return nil, 0, err
	}

	// Convert to HistoryResponse
	responses := make([]model.HistoryResponse, len(records))
	for i, r := range records {
		responses[i] = model.HistoryResponse{
			ID:        r.ID,
			ChartID:   r.ChartID,
			QueryDate: r.QueryDate.Format("2006-01-02"),
			Summary:   r.Summary,
			CreatedAt: r.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return responses, total, nil
}

// ListByUserID returns fortune records for a user with pagination.
func (s *DBFortuneStore) ListByUserID(userID uint, page, pageSize int) ([]model.FortuneRecord, int64, error) {
	var records []model.FortuneRecord
	var total int64

	offset := (page - 1) * pageSize
	query := s.db.Model(&model.FortuneRecord{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&records).Error; err != nil {
		return nil, 0, err
	}
	return records, total, nil
}