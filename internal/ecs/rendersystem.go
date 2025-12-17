package ecs

import (
	"go-engine/Go-Cordance/internal/engine"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type RenderSystem struct {
	Renderer    *engine.Renderer
	MeshManager *engine.MeshManager
}

func NewRenderSystem(r *engine.Renderer, mm *engine.MeshManager) *RenderSystem {
	return &RenderSystem{Renderer: r, MeshManager: mm}
}

func (rs *RenderSystem) Update(dt float32, entities []*Entity) {
	_ = dt
	var camSys *CameraSystem
	// find CameraSystem in scene (you can inject it)
	// assume you pass it in when constructing RenderSystem

	for _, e := range entities {
		var t *Transform
		var mesh *Mesh
		var mat *Material
		for _, c := range e.Components {
			switch comp := c.(type) {
			case *Transform:
				t = comp
			case *Mesh:
				mesh = comp
			case *Material:
				mat = comp
			}
		}
		if t != nil && mesh != nil && mat != nil && camSys != nil {
			vao := rs.MeshManager.GetVAO(mesh.ID)
			gl.UseProgram(rs.Renderer.Program)
			gl.BindVertexArray(vao)

			// Upload matrices (uniform locations assumed)
			model := mgl32.Translate3D(t.Position[0], t.Position[1], t.Position[2])
			mvp := camSys.Projection.Mul4(camSys.View).Mul4(model)
			loc := gl.GetUniformLocation(rs.Renderer.Program, gl.Str("MVP\x00"))
			gl.UniformMatrix4fv(loc, 1, false, &mvp[0])

			gl.DrawArrays(gl.TRIANGLES, 0, 3)
		}
	}
}
