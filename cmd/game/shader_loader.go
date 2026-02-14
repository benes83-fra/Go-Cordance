package main

import (
	"go-engine/Go-Cordance/internal/assets"
	"go-engine/Go-Cordance/internal/engine"
	"go-engine/Go-Cordance/internal/shaderlang"
)

func LoadAllShaders() error {
	for _, a := range assets.All() {
		if a.Type != assets.AssetShader {
			continue
		}

		src := a.Data.(shaderlang.ShaderSource)

		// Load GLSL text
		vert, err := shaderlang.LoadGLSL(src.VertexPath)
		if err != nil {
			return err
		}
		frag, err := shaderlang.LoadGLSL(src.FragmentPath)
		if err != nil {
			return err
		}

		// Apply defines
		vert = shaderlang.ApplyDefines(vert, src.Defines)
		frag = shaderlang.ApplyDefines(frag, src.Defines)

		// Compile in engine
		prog, err := engine.LoadShaderProgram(src.Name, vert, frag)
		if err != nil {
			return err
		}

		engine.RegisterShaderProgram(src.Name, prog)
	}

	return nil
}
