package assets

import (
	"encoding/json"
	"os"
)

// Material assets are metadata-only for now.
// Later you can store shader, default textures, import settings, etc.

type MaterialInfo struct {
	BaseColor [4]float32
	Diffuse   string
	Normal    string
}

func RegisterMaterial(path string, info MaterialInfo) AssetID {
	return Register(AssetMaterial, path, info)
}

type MaterialFile struct {
	Name     string            `json:"name"`
	Shader   string            `json:"shader"`
	Params   map[string]any    `json:"params"`
	Textures map[string]string `json:"textures"`
}

func LoadMaterial(path string) (AssetID, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}

	var mf MaterialFile
	if err := json.Unmarshal(data, &mf); err != nil {
		return 0, err
	}

	// Store the raw struct as Data
	id := Register(AssetMaterial, path, mf)
	return id, nil
}
