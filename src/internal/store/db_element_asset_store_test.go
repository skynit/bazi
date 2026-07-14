package store

import (
	"testing"

	"bazi/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupElementAssetStoreTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.ElementAsset{}); err != nil {
		t.Fatalf("migrate element assets: %v", err)
	}
	return db
}

func TestDBElementAssetStoreUpsertDefaultsSynchronizesPackagedAssets(t *testing.T) {
	db := setupElementAssetStoreTestDB(t)
	store := NewDBElementAssetStore(db)

	seed := []model.ElementAsset{
		{Key: "wood-new", Name: "旧名称", Element: "木", URL: "/element-assets/wood/wood-new.webp", Scene: "general", Orientation: "square", AltText: "旧文案", Status: model.ElementAssetStatusInactive},
		{Key: "wood-legacy", Name: "旧内置素材", Element: "木", URL: "/element-assets/wood/wood-legacy.png", Scene: "hero", Orientation: "landscape", AltText: "旧素材", Status: model.ElementAssetStatusActive},
		{Key: "uploaded", Name: "用户素材", Element: "木", URL: "/uploads/element-assets/wood/uploaded.webp", Scene: "object", Orientation: "square", AltText: "用户上传", Status: model.ElementAssetStatusActive},
	}
	if err := db.Create(&seed).Error; err != nil {
		t.Fatalf("seed assets: %v", err)
	}

	defaults := []model.ElementAsset{
		{Key: "wood-new", Name: "晨林", Element: "木", URL: "/element-assets/wood/wood-new.webp", Scene: "hero", Orientation: "landscape", AltText: "晨光山林", Width: 1536, Height: 864, Weight: 100, Status: model.ElementAssetStatusActive},
		{Key: "fire-new", Name: "暖灯", Element: "火", URL: "/element-assets/fire/fire-new.webp", Scene: "object", Orientation: "square", AltText: "暖色灯火", Width: 1024, Height: 1024, Weight: 100, Status: model.ElementAssetStatusActive},
	}
	if err := store.UpsertDefaults(defaults); err != nil {
		t.Fatalf("sync defaults: %v", err)
	}

	var updated model.ElementAsset
	if err := db.Where("key = ?", "wood-new").First(&updated).Error; err != nil {
		t.Fatalf("load updated asset: %v", err)
	}
	if updated.Name != "晨林" || updated.Scene != "hero" || updated.Status != model.ElementAssetStatusActive || updated.Width != 1536 {
		t.Fatalf("existing default was not refreshed: %+v", updated)
	}

	var legacy model.ElementAsset
	if err := db.Where("key = ?", "wood-legacy").First(&legacy).Error; err != nil {
		t.Fatalf("load legacy asset: %v", err)
	}
	if legacy.Status != model.ElementAssetStatusInactive {
		t.Fatalf("legacy packaged asset should be inactive, got %q", legacy.Status)
	}

	var uploaded model.ElementAsset
	if err := db.Where("key = ?", "uploaded").First(&uploaded).Error; err != nil {
		t.Fatalf("load uploaded asset: %v", err)
	}
	if uploaded.Status != model.ElementAssetStatusActive {
		t.Fatalf("uploaded asset should remain active, got %q", uploaded.Status)
	}

	active, err := store.ListActive()
	if err != nil {
		t.Fatalf("list active: %v", err)
	}
	if len(active) != 3 {
		t.Fatalf("expected two packaged assets plus one upload, got %d", len(active))
	}
}
