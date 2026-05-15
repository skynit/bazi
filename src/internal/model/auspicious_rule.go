package model

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// AuspiciousRule stores a HuangLi (黄历) auspicious/inauspicious rule.
type AuspiciousRule struct {
	gorm.Model
	Category  string         `gorm:"type:varchar(16);not null;index" json:"category"` // 宜/忌
	DayStem   string         `gorm:"column:day_stem;type:varchar(8);not null;index" json:"day_stem"`
	DayBranch string         `gorm:"column:day_branch;type:varchar(8);not null;index" json:"day_branch"`
	Content   datatypes.JSON `gorm:"type:json;not null" json:"content"`
}
