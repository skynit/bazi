package elementasset

import "testing"

func TestDefaultAssetsLoadsCompleteFiveElementLibrary(t *testing.T) {
	assets := DefaultAssets()
	if len(assets) != 40 {
		t.Fatalf("expected 40 default assets, got %d", len(assets))
	}

	elementCounts := map[string]int{}
	sceneCounts := map[string]int{}
	orientationCounts := map[string]int{}
	keys := map[string]struct{}{}
	urls := map[string]struct{}{}
	for _, asset := range assets {
		if asset.Key == "" || asset.Name == "" || asset.URL == "" || asset.AltText == "" {
			t.Fatalf("asset has missing required metadata: %+v", asset)
		}
		if asset.Width <= 0 || asset.Height <= 0 {
			t.Fatalf("asset %s has invalid dimensions %dx%d", asset.Key, asset.Width, asset.Height)
		}
		if asset.Status != "active" {
			t.Fatalf("asset %s should be active, got %q", asset.Key, asset.Status)
		}
		if _, exists := keys[asset.Key]; exists {
			t.Fatalf("duplicate asset key %q", asset.Key)
		}
		if _, exists := urls[asset.URL]; exists {
			t.Fatalf("duplicate asset URL %q", asset.URL)
		}
		keys[asset.Key] = struct{}{}
		urls[asset.URL] = struct{}{}
		elementCounts[asset.Element]++
		sceneCounts[asset.Scene]++
		orientationCounts[asset.Orientation]++
	}

	for _, element := range []string{"木", "火", "土", "金", "水"} {
		if elementCounts[element] != 8 {
			t.Errorf("expected 8 %s assets, got %d", element, elementCounts[element])
		}
	}
	if sceneCounts["hero"] != 15 || sceneCounts["object"] != 15 || sceneCounts["general"] != 10 {
		t.Errorf("unexpected scene distribution: %#v", sceneCounts)
	}
	if orientationCounts["panorama"] != 5 {
		t.Errorf("expected 5 panorama assets, got %#v", orientationCounts)
	}
}

func TestDefaultAssetsBuildBlessingSetForEveryElement(t *testing.T) {
	assets := DefaultAssets()
	for index := range assets {
		assets[index].ID = uint(index + 1)
	}
	svc := New(memoryRepo{assets: assets})
	elements := []string{"木", "火", "土", "金", "水"}
	for index, primary := range elements {
		secondary := elements[(index+1)%len(elements)]
		avoid := elements[(index+2)%len(elements)]
		set, err := svc.BuildBlessingSet(88, "2026-07-12", primary, secondary, avoid, []string{primary, secondary})
		if err != nil {
			t.Fatalf("build %s blessing set: %v", primary, err)
		}
		if set.Hero == nil || set.Hero.Element != primary || set.Hero.Scene != "hero" {
			t.Fatalf("unexpected %s hero: %+v", primary, set.Hero)
		}
		if len(set.Ritual) != 3 || len(set.Actions) != 2 || len(set.Gallery) != 10 {
			t.Fatalf("unexpected %s bundle sizes: ritual=%d actions=%d gallery=%d", primary, len(set.Ritual), len(set.Actions), len(set.Gallery))
		}
	}
}
