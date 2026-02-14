package assets

import (
	"encoding/json"
	"go-engine/Go-Cordance/internal/shaderlang"
	"os"
)

func LoadShader(path string) (AssetID, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}

	var src shaderlang.ShaderSource
	if err := json.Unmarshal(data, &src); err != nil {
		return 0, err
	}

	id := Register(AssetShader, path, src)
	return id, nil
}

type ShaderFile struct {
	Name         string         `json:"name"`
	VertexPath   string         `json:"vertex"`
	FragmentPath string         `json:"fragment"`
	Defines      map[string]any `json:"defines"`
}
