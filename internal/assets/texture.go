package assets

import "go-engine/Go-Cordance/internal/engine"

// ImportTexture loads a texture via engine.LoadTexture and registers it as an asset.
// It does NOT modify ECS.Material.TextureID — you still assign that manually.
func ImportTexture(path string) (AssetID, uint32, error) {
	texGL, err := engine.LoadTexture(path)
	if err != nil {
		return 0, 0, err
	}

	id := Register(AssetTexture, path, texGL)
	return id, texGL, nil
}
