package ecs

import (
	"go-engine/Go-Cordance/internal/assets"
	"go-engine/Go-Cordance/internal/engine"
)

type Shader struct {
	AssetID assets.AssetID
	Program *engine.ShaderProgram
}
