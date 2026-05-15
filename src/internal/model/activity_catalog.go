package model

import "gorm.io/gorm"

// ActivityCatalog stores a catalog entry for auspicious/inauspicious activities.
type ActivityCatalog struct {
	gorm.Model
	Category string `gorm:"type:varchar(4);not null;index" json:"category"` // 宜 or 忌
	Name     string `gorm:"type:varchar(64);not null" json:"name"`
}
