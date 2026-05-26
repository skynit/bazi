package store

import (
	"bazi/internal/model"

	"gorm.io/gorm"
)

// DBChartStore implements chart persistence using GORM.
type DBChartStore struct {
	db *gorm.DB
}

// NewDBChartStore creates a DBChartStore.
func NewDBChartStore(db *gorm.DB) *DBChartStore {
	return &DBChartStore{db: db}
}

// Create inserts a new birth chart.
func (s *DBChartStore) Create(chart *model.BirthChart) error {
	return s.db.Create(chart).Error
}

// FindByID retrieves a chart by ID.
func (s *DBChartStore) FindByID(id uint) (*model.BirthChart, error) {
	var chart model.BirthChart
	if err := s.db.First(&chart, id).Error; err != nil {
		return nil, err
	}
	return &chart, nil
}

// ListByUser returns charts belonging to a user with pagination.
func (s *DBChartStore) ListByUser(userID uint, page, pageSize int) ([]model.BirthChart, int64, error) {
	var charts []model.BirthChart
	var total int64

	offset := (page - 1) * pageSize
	query := s.db.Model(&model.BirthChart{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&charts).Error; err != nil {
		return nil, 0, err
	}
	return charts, total, nil
}

// Update updates an existing chart.
func (s *DBChartStore) Update(chart *model.BirthChart) error {
	return s.db.Save(chart).Error
}