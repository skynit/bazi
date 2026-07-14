package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	ElementAssetStatusActive   = "active"
	ElementAssetStatusInactive = "inactive"
)

// ElementAsset stores searchable metadata for one five-element visual asset.
// The image itself is served as a static file or from object storage; the DB only
// stores its URL and presentation metadata.
type ElementAsset struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	CreatedAt        time.Time      `json:"-"`
	UpdatedAt        time.Time      `json:"-"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
	Key              string         `gorm:"type:varchar(128);uniqueIndex;not null" json:"key"`
	Name             string         `gorm:"type:varchar(128);not null" json:"name"`
	Element          string         `gorm:"type:varchar(8);index;not null" json:"element"`
	SecondaryElement string         `gorm:"type:varchar(8);index" json:"secondary_element,omitempty"`
	URL              string         `gorm:"type:varchar(512);not null" json:"url"`
	ThumbnailURL     string         `gorm:"type:varchar(512)" json:"thumbnail_url,omitempty"`
	Scene            string         `gorm:"type:varchar(32);index;not null;default:general" json:"scene"`
	Orientation      string         `gorm:"type:varchar(16);index;not null;default:landscape" json:"orientation"`
	Tone             string         `gorm:"type:varchar(32);index;not null;default:balanced" json:"tone"`
	Style            string         `gorm:"type:varchar(32);index;not null;default:chinese" json:"style"`
	Season           string         `gorm:"type:varchar(16);index;not null;default:all" json:"season"`
	TimePeriod       string         `gorm:"type:varchar(16);index;not null;default:all" json:"time_period"`
	ObjectLabel      string         `gorm:"type:varchar(64)" json:"object_label,omitempty"`
	Description      string         `gorm:"type:varchar(512)" json:"description,omitempty"`
	AltText          string         `gorm:"type:varchar(256);not null" json:"alt_text"`
	DominantColor    string         `gorm:"type:varchar(16)" json:"dominant_color,omitempty"`
	AccentColor      string         `gorm:"type:varchar(16)" json:"accent_color,omitempty"`
	FocalX           float64        `gorm:"not null;default:0.5" json:"focal_x"`
	FocalY           float64        `gorm:"not null;default:0.5" json:"focal_y"`
	Width            int            `gorm:"not null;default:0" json:"width"`
	Height           int            `gorm:"not null;default:0" json:"height"`
	Weight           int            `gorm:"not null;default:100" json:"weight"`
	SortOrder        int            `gorm:"not null;default:0" json:"sort_order"`
	Status           string         `gorm:"type:varchar(16);index;not null;default:active" json:"status"`
}

// BlessingAssetSet is the visual material bundle selected for one blessing page.
type BlessingAssetSet struct {
	Hero    *ElementAsset  `json:"hero,omitempty"`
	Ritual  []ElementAsset `json:"ritual"`
	Actions []ElementAsset `json:"actions"`
	Gallery []ElementAsset `json:"gallery"`
}
