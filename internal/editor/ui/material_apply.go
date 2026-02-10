package ui

import (
	"go-engine/Go-Cordance/internal/assets"
	"go-engine/Go-Cordance/internal/ecs"
	"go-engine/Go-Cordance/internal/editor/state"
)

func ConvertMaterialAssetToECS(av state.AssetView) *ecs.Material {
	mat := &ecs.Material{}

	params, _ := av.MaterialData["params"].(map[string]any)
	textures, _ := av.MaterialData["textures"].(map[string]any)

	// baseColor
	if bc, ok := params["baseColor"].([]any); ok && len(bc) == 4 {
		mat.BaseColor[0] = float32(bc[0].(float64))
		mat.BaseColor[1] = float32(bc[1].(float64))
		mat.BaseColor[2] = float32(bc[2].(float64))
		mat.BaseColor[3] = float32(bc[3].(float64))
	}

	// albedo texture
	if texPath, ok := textures["albedo"].(string); ok && texPath != "" {
		if texAsset := assets.FindAssetByPath(texPath); texAsset != nil {
			mat.UseTexture = true
			mat.TextureAsset = texAsset.ID
			mat.TextureID = assets.ResolveTextureGLID(texAsset.ID)
		}
	}

	return mat
}
