package ecs

import (
	"go-engine/Go-Cordance/internal/engine"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type DebugRenderSystem struct {
	Renderer     *engine.DebugRenderer
	MeshManager  *engine.MeshManager
	CameraSystem *CameraSystem
	Enabled      bool
}

func NewDebugRenderSystem(r *engine.DebugRenderer, mm *engine.MeshManager, cs *CameraSystem) *DebugRenderSystem {
	return &DebugRenderSystem{
		Renderer:     r,
		MeshManager:  mm,
		CameraSystem: cs,
		Enabled:      true,
	}
}

// DebugRenderSystem for colliders
func (ds *DebugRenderSystem) Update(_ float32, entities []*Entity) {
	if !ds.Enabled {
		return
	}

	gl.UseProgram(ds.Renderer.Program)
	view := ds.CameraSystem.View
	proj := ds.CameraSystem.Projection

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

		model := mgl32.Translate3D(t.Position[0], t.Position[1], t.Position[2])
		gl.UniformMatrix4fv(ds.Renderer.LocModel, 1, false, &model[0])
		gl.UniformMatrix4fv(ds.Renderer.LocView, 1, false, &view[0])
		gl.UniformMatrix4fv(ds.Renderer.LocProj, 1, false, &proj[0])

		if sphere != nil {
			scale := mgl32.Scale3D(sphere.Radius, sphere.Radius, sphere.Radius)
			model = model.Mul4(scale)
			gl.UniformMatrix4fv(ds.Renderer.LocModel, 1, false, &model[0])

			col := [4]float32{1, 0, 0, 1}
			gl.Uniform4fv(ds.Renderer.LocColor, 1, &col[0])
			vao := ds.MeshManager.GetVAO("wire_sphere")
			gl.BindVertexArray(vao)
			count := ds.MeshManager.GetCount("wire_sphere")
			gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
			gl.DrawElements(gl.LINES, count, gl.UNSIGNED_INT, gl.PtrOffset(0))
			gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
			gl.BindVertexArray(0)
		}

		if box != nil {
			scale := mgl32.Scale3D(box.HalfExtents[0]*2, box.HalfExtents[1]*2, box.HalfExtents[2]*2)
			model = model.Mul4(scale)
			gl.UniformMatrix4fv(ds.Renderer.LocModel, 1, false, &model[0])

			col := [4]float32{0, 1, 1, 1}
			gl.Uniform4fv(ds.Renderer.LocColor, 1, &col[0])
			vao := ds.MeshManager.GetVAO("wire_cube")
			gl.BindVertexArray(vao)
			count := ds.MeshManager.GetCount("wire_cube")
			gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
			gl.DrawElements(gl.LINES, count, gl.UNSIGNED_INT, gl.PtrOffset(0))
			gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
			gl.BindVertexArray(0)
		}
	}
}

/*
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

		model := mgl32.Translate3D(t.Position[0], t.Position[1], t.Position[2])
		gl.UniformMatrix4fv(drs.Renderer.LocModel, 1, false, &model[0])
		gl.UniformMatrix4fv(drs.Renderer.LocView, 1, false, &view[0])
		gl.UniformMatrix4fv(drs.Renderer.LocProj, 1, false, &proj[0])

		if sphere != nil {
			// scale by radius
			scale := mgl32.Scale3D(sphere.Radius, sphere.Radius, sphere.Radius)
			model = model.Mul4(scale)
			gl.UniformMatrix4fv(drs.Renderer.LocModel, 1, false, &model[0])

			col := [4]float32{1, 0, 0, 1}
			gl.Uniform4fv(drs.Renderer.LocColor, 1, &col[0])
			vao := drs.MeshManager.GetVAO("wire_sphere")
			gl.BindVertexArray(vao)
			count := drs.MeshManager.GetCount("wire_sphere")
			gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
			gl.DrawElements(gl.LINES, count, gl.UNSIGNED_INT, gl.PtrOffset(0))
			gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
			gl.BindVertexArray(0)
		}

		if box != nil {
			// scale by half extents
			scale := mgl32.Scale3D(box.HalfExtents[0]*2, box.HalfExtents[1]*2, box.HalfExtents[2]*2)
			model = model.Mul4(scale)
			gl.UniformMatrix4fv(drs.Renderer.LocModel, 1, false, &model[0])

			col := [4]float32{0, 1, 1, 1}
			gl.Uniform4fv(drs.Renderer.LocColor, 1, &col[0])
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
*/
