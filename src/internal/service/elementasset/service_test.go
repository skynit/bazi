package elementasset

import (
	"reflect"
	"testing"

	"bazi/internal/model"
)

type memoryRepo struct {
	assets []model.ElementAsset
	err    error
}

func (r memoryRepo) ListActive() ([]model.ElementAsset, error) {
	return append([]model.ElementAsset(nil), r.assets...), r.err
}

func TestSelectIsStableAndRelaxesScene(t *testing.T) {
	repo := memoryRepo{assets: []model.ElementAsset{
		{ID: 1, Key: "wood-object-a", Element: "木", Scene: "object", Orientation: "square", Season: "all", TimePeriod: "all", Weight: 100, Status: "active"},
		{ID: 2, Key: "wood-object-b", Element: "木", Scene: "object", Orientation: "square", Season: "all", TimePeriod: "all", Weight: 100, Status: "active"},
	}}
	svc := New(repo)
	query := Query{Element: "木", Scene: "hero", Orientation: "landscape", Seed: "chart:date:hero", Limit: 2}

	first, err := svc.Select(query)
	if err != nil {
		t.Fatal(err)
	}
	second, err := svc.Select(query)
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != 2 {
		t.Fatalf("expected scene/orientation fallback to return 2 assets, got %d", len(first))
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("stable seed returned different results: %#v != %#v", first, second)
	}
}

func TestBuildBlessingSetReturnsDiverseBundle(t *testing.T) {
	assets := []model.ElementAsset{
		{ID: 1, Key: "wood-hero", Element: "木", Scene: "hero", Orientation: "landscape", Season: "all", TimePeriod: "all", Weight: 100, Status: "active"},
		{ID: 2, Key: "wood-object", Element: "木", Scene: "object", Orientation: "square", Season: "all", TimePeriod: "all", Weight: 100, Status: "active"},
		{ID: 3, Key: "water-object", Element: "水", Scene: "object", Orientation: "square", Season: "all", TimePeriod: "all", Weight: 100, Status: "active"},
		{ID: 4, Key: "metal-object", Element: "金", Scene: "object", Orientation: "square", Season: "all", TimePeriod: "all", Weight: 100, Status: "active"},
	}
	svc := New(memoryRepo{assets: assets})
	set, err := svc.BuildBlessingSet(12, "2026-07-12", "木", "水", "金", []string{"木", "水"})
	if err != nil {
		t.Fatal(err)
	}
	if set.Hero == nil || set.Hero.Key != "wood-hero" {
		t.Fatalf("unexpected hero: %#v", set.Hero)
	}
	if len(set.Ritual) != 3 {
		t.Fatalf("expected three ritual assets, got %d", len(set.Ritual))
	}
	if len(set.Actions) != 2 {
		t.Fatalf("expected two action assets, got %d", len(set.Actions))
	}
	if len(set.Gallery) != len(assets) {
		t.Fatalf("expected full gallery, got %d", len(set.Gallery))
	}
}
