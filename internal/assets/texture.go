package assets

import "go-engine/Go-Cordance/internal/engine"

// TextureData holds runtime GPU info for a texture.
type TextureData struct {
	GLID uint32
}

// ImportTexture loads a texture via engine.LoadTexture and registers it as an asset.
// It returns the AssetID and the raw GL texture id for convenience.
func ImportTexture(path string) (AssetID, uint32, error) {
	texGL, err := engine.LoadTexture(path)
	if err != nil {
		return 0, 0, err
	}

	// Wrap GL id in TextureData so the registry stores structured data.
	data := TextureData{GLID: texGL}
	id := Register(AssetTexture, path, data)
	return id, texGL, nil
}
