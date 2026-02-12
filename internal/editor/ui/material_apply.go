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

	// Default lighting params
	mat.Ambient = 0.2
	mat.Diffuse = 0.8
	mat.Specular = 0.5
	mat.Shininess = 32.0

	// --- Resolve textures using editor AssetList (NOT game registry) ---

	// Albedo
	if texPath, ok := textures["albedo"].(string); ok && texPath != "" {
		for _, tex := range state.Global.Assets.Textures {
			if tex.Path == texPath {
				mat.UseTexture = true
				mat.TextureAsset = assets.AssetID(tex.ID) // <-- FIX
				// TextureID is resolved by the game later
				break
			}
		}
	}

	// Normal map
	if nPath, ok := textures["normal"].(string); ok && nPath != "" {
		for _, tex := range state.Global.Assets.Textures {
			if tex.Path == nPath {
				mat.UseNormal = true
				mat.NormalAsset = assets.AssetID(tex.ID) // <-- FIX
				break
			}
		}
	}

	return mat
}
