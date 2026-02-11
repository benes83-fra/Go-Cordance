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

	// BaseColor
	if bc, ok := params["baseColor"].([]any); ok && len(bc) == 4 {
		mat.BaseColor = [4]float32{
			float32(bc[0].(float64)),
			float32(bc[1].(float64)),
			float32(bc[2].(float64)),
			float32(bc[3].(float64)),
		}
	}

	// Default lighting params (your engine expects these)
	mat.Ambient = 0.2
	mat.Diffuse = 0.8
	mat.Specular = 0.5
	mat.Shininess = 32.0

	// Albedo texture
	if texPath, ok := textures["albedo"].(string); ok && texPath != "" {
		if texAsset := assets.FindAssetByPath(texPath); texAsset != nil {
			mat.UseTexture = true
			mat.TextureAsset = texAsset.ID
			mat.TextureID = assets.ResolveTextureGLID(texAsset.ID)
		}
	}

	// Normal map
	if nPath, ok := textures["normal"].(string); ok && nPath != "" {
		if nAsset := assets.FindAssetByPath(nPath); nAsset != nil {
			mat.UseNormal = true
			mat.NormalAsset = nAsset.ID
			mat.NormalID = assets.ResolveTextureGLID(nAsset.ID)
		}
	}

	return mat
}
