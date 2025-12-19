package ecs

import (
	"go-engine/Go-Cordance/internal/engine"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type DebugRenderSystem struct {
	Renderer     *engine.Renderer
	MeshManager  *engine.MeshManager
	CameraSystem *CameraSystem
	Enabled      bool
}

func NewDebugRenderSystem(r *engine.Renderer, mm *engine.MeshManager, cs *CameraSystem) *DebugRenderSystem {
	return &DebugRenderSystem{Renderer: r, MeshManager: mm, CameraSystem: cs}
}

func (drs *DebugRenderSystem) Update(_ float32, entities []*Entity) {
	if !drs.Enabled {
		return
	}
	gl.UseProgram(drs.Renderer.Program)
	view := drs.CameraSystem.View
	proj := drs.CameraSystem.Projection

	for _, e := range entities {
		var t *Transform
		var sphere *ColliderSphere
		var box *ColliderAABB

		for _, c := range e.Components {
			switch comp := c.(type) {
			case *Transform:
				t = comp
			case *ColliderSphere:
				sphere = comp
			case *ColliderAABB:
				box = comp
			}
		}

		if t == nil {
			continue
		}

		// Build model matrix
		model := mgl32.Translate3D(t.Position[0], t.Position[1], t.Position[2])

		gl.UniformMatrix4fv(drs.Renderer.LocModel, 1, false, &model[0])
		gl.UniformMatrix4fv(drs.Renderer.LocView, 1, false, &view[0])
		gl.UniformMatrix4fv(drs.Renderer.LocProj, 1, false, &proj[0])

		// Debug color (red for spheres, cyan for boxes)
		if sphere != nil {
			col := [4]float32{1, 0, 0, 1} // red
			gl.Uniform4fv(drs.Renderer.LocBaseCol, 1, &col[0])
			vao := drs.MeshManager.GetVAO("wire_sphere")
			gl.BindVertexArray(vao)
			count := drs.MeshManager.GetCount("wire_sphere")
			gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
			gl.DrawElements(gl.LINES, count, gl.UNSIGNED_INT, gl.PtrOffset(0))
			gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
			gl.BindVertexArray(0)
		}

		if box != nil {
			col := [4]float32{0, 1, 1, 1} // cyan
			gl.Uniform4fv(drs.Renderer.LocBaseCol, 1, &col[0])
			vao := drs.MeshManager.GetVAO("wire_cube")
			gl.BindVertexArray(vao)
			count := drs.MeshManager.GetCount("wire_cube")
			gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
			gl.DrawElements(gl.LINES, count, gl.UNSIGNED_INT, gl.PtrOffset(0))
			gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
			gl.BindVertexArray(0)
		}

	}
}
