package store

import (
	"bazi/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DBElementAssetStore struct {
	db *gorm.DB
}

func NewDBElementAssetStore(db *gorm.DB) *DBElementAssetStore {
	return &DBElementAssetStore{db: db}
}

func (s *DBElementAssetStore) ListActive() ([]model.ElementAsset, error) {
	var assets []model.ElementAsset
	err := s.db.Where("status = ?", model.ElementAssetStatusActive).
		Order("sort_order ASC, id ASC").Find(&assets).Error
	return assets, err
}

func (s *DBElementAssetStore) Create(asset *model.ElementAsset) error {
	return s.db.Create(asset).Error
}

func (s *DBElementAssetStore) UpsertDefaults(assets []model.ElementAsset) error {
	if len(assets) == 0 {
		return nil
	}

	keys := make([]string, 0, len(assets))
	for _, asset := range assets {
		keys = append(keys, asset.Key)
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// Static assets removed from the packaged manifest must stop taking part
		// in selection, while uploaded assets under /uploads remain untouched.
		if err := tx.Model(&model.ElementAsset{}).
			Where("url LIKE ? AND key NOT IN ?", "/element-assets/%", keys).
			Update("status", model.ElementAssetStatusInactive).Error; err != nil {
			return err
		}

		return tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "key"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"name", "element", "secondary_element", "url", "thumbnail_url",
				"scene", "orientation", "tone", "style", "season", "time_period",
				"object_label", "description", "alt_text", "dominant_color", "accent_color",
				"focal_x", "focal_y", "width", "height", "weight", "sort_order", "status",
			}),
		}).Create(&assets).Error
	})
}
