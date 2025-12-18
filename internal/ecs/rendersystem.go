package ecs

import (
	"go-engine/Go-Cordance/internal/engine"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type RenderSystem struct {
	Renderer     *engine.Renderer
	MeshManager  *engine.MeshManager
	CameraSystem *CameraSystem
}

func NewRenderSystem(r *engine.Renderer, mm *engine.MeshManager, cs *CameraSystem) *RenderSystem {
	return &RenderSystem{Renderer: r, MeshManager: mm, CameraSystem: cs}
}

func (rs *RenderSystem) Update(_ float32, entities []*Entity) {
	gl.UseProgram(rs.Renderer.Program)
	view := rs.CameraSystem.View
	proj := rs.CameraSystem.Projection
	for _, e := range entities {
		var t *Transform
		var mesh *Mesh
		var mat *Material
		for _, c := range e.Components {
			switch v := c.(type) {
			case *Transform:
				t = v
			case *Mesh:
				mesh = v
			case *Material:
				mat = v
			}
		}
		if t == nil || mesh == nil || mat == nil {
			continue
		}

		// Build MVP
		model := mgl32.Translate3D(t.Position[0], t.Position[1], t.Position[2])

		// Upload matrices
		//locModel := gl.GetUniformLocation(rs.Renderer.Program, gl.Str("model\x00"))
		//locView := gl.GetUniformLocation(rs.Renderer.Program, gl.Str("view\x00"))
		//locProj := gl.GetUniformLocation(rs.Renderer.Program, gl.Str("projection\x00"))
		//locColor := gl.GetUniformLocation(rs.Renderer.Program, gl.Str("baseColor\x00"))
		gl.UniformMatrix4fv(rs.Renderer.LocModel, 1, false, &model[0])
		gl.UniformMatrix4fv(rs.Renderer.LocView, 1, false, &view[0])
		gl.UniformMatrix4fv(rs.Renderer.LocProj, 1, false, &proj[0])

		// Upload material color

		gl.Uniform4fv(rs.Renderer.LocBaseCol, 1, &mat.Color[0])

		// Draw
		vao := rs.MeshManager.GetVAO(mesh.ID)
		gl.BindVertexArray(vao)
		gl.DrawArrays(gl.TRIANGLES, 0, 3)
		gl.BindVertexArray(0)

	}
}
