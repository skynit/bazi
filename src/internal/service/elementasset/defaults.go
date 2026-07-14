package elementasset

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"bazi/internal/model"
)

// defaultAssetsJSON is generated from vue/public/element-assets/manifest.json.
//
//go:embed defaults.json
var defaultAssetsJSON []byte

// DefaultAssets returns the packaged five-element visual library. Keeping the
// metadata in JSON allows the image-generation workflow to refresh the library
// without duplicating forty records in Go source code.
func DefaultAssets() []model.ElementAsset {
	var manifest struct {
		Assets []model.ElementAsset `json:"assets"`
	}
	if err := json.Unmarshal(defaultAssetsJSON, &manifest); err != nil {
		panic(fmt.Sprintf("decode embedded element asset manifest: %v", err))
	}
	return manifest.Assets
}
