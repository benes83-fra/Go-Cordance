package assets

import "go-engine/Go-Cordance/internal/engine"

// TextureData holds runtime GPU info for a texture.
type TextureData struct {
	GLID uint32
	SRGB bool
}

// ImportTexture loads a texture via engine.LoadTexture and registers it as an asset.
// Default behavior: treat as sRGB (good for base color / albedo).
func ImportTexture(path string) (AssetID, uint32, error) {
	return ImportTextureWithSRGB(path, true)
}

// ImportTextureWithSRGB lets caller choose color space.
// srgb == true -> use sRGB sampling (albedo)
// srgb == false -> use linear sampling (normal, occlusion, metallic/roughness)
func ImportTextureWithSRGB(path string, srgb bool) (AssetID, uint32, error) {
	texGL, err := engine.LoadTextureWithColorSpace(path, srgb) // see note below
	if err != nil {
		return 0, 0, err
	}
	data := TextureData{GLID: texGL}
	id := Register(AssetTexture, path, data)
	return id, texGL, nil
}
