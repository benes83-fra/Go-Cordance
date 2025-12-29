package ecs

import (
	"math"

	"go-engine/Go-Cordance/internal/engine"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type GizmoRenderSystem struct {
	Renderer     *engine.DebugRenderer
	MeshManager  *engine.MeshManager
	CameraSystem *CameraSystem
	Enabled      bool

	HoverAxis  string
	ActiveAxis string
	IsDragging bool
}

func NewGizmoRenderSystem(r *engine.DebugRenderer, mm *engine.MeshManager, cs *CameraSystem) *GizmoRenderSystem {
	return &GizmoRenderSystem{
		Renderer:     r,
		MeshManager:  mm,
		CameraSystem: cs,
		Enabled:      true,
	}
}

func (gs *GizmoRenderSystem) SetCameraSystem(cs *CameraSystem) { gs.CameraSystem = cs }

func (gs *GizmoRenderSystem) Update(_ float32, _ []*Entity, selected *Entity) {
	if !gs.Enabled || gs.CameraSystem == nil || selected == nil {
		return
	}

	var t *Transform
	for _, c := range selected.Components {
		if tr, ok := c.(*Transform); ok {
			t = tr
			break
		}
	}
	if t == nil {
		return
	}

	gl.UseProgram(gs.Renderer.Program)

	view := gs.CameraSystem.View
	proj := gs.CameraSystem.Projection

	//camera position to mgl32.Vec3 (works if Position is [3]float32 or []float32)
	camPos := mgl32.Vec3{gs.CameraSystem.Position[0], gs.CameraSystem.Position[1], gs.CameraSystem.Position[2]}
	entityPos := mgl32.Vec3{t.Position[0], t.Position[1], t.Position[2]}
	dist := camPos.Sub(entityPos).Len()
	gizmoScale := float32(dist * 0.08)
	// Hover detection
	origin, dir := RayFromMouse(gs.CameraSystem.Window(), gs.CameraSystem)

	gs.HoverAxis = "" // reset

	axisLength := gizmoScale * 1.0
	pickRadius := gizmoScale * 0.2

	axes := []struct {
		name string
		axis mgl32.Vec3
	}{
		{"x", mgl32.Vec3{1, 0, 0}},
		{"y", mgl32.Vec3{0, 1, 0}},
		{"z", mgl32.Vec3{0, 0, 1}},
	}

	closest := float32(math.MaxFloat32)

	for _, a := range axes {
		a0 := entityPos
		a1 := entityPos.Add(a.axis.Mul(axisLength))

		hit, dist := RayCapsuleIntersect(origin, dir, a0, a1, pickRadius)
		if hit && dist < closest {
			closest = dist
			gs.HoverAxis = a.name
		}
	}

	for _, a := range axes {
		model := mgl32.Translate3D(t.Position[0], t.Position[1], t.Position[2])

		z := mgl32.Vec3{0, 0, 1}
		dot := z.Dot(a.axis)
		if dot < 0.9999 {
			axis := z.Cross(a.axis)
			if axis.Len() < 1e-6 {
				if dot < 0 {
					model = model.Mul4(mgl32.HomogRotate3D(float32(math.Pi), mgl32.Vec3{1, 0, 0}))
				}
			} else {
				angle := float32(math.Acos(float64(dot)))
				model = model.Mul4(mgl32.HomogRotate3D(angle, axis.Normalize()))
			}
		}

		model = model.Mul4(mgl32.Scale3D(gizmoScale, gizmoScale, gizmoScale))

		var col [4]float32
		// base axis colors
		switch a.name {
		case "x":
			col = [4]float32{1, 0, 0, 1}
		case "y":
			col = [4]float32{0, 1, 0, 1}
		case "z":
			col = [4]float32{0, 0, 1, 1}
		}

		// highlight
		if gs.ActiveAxis == a.name {
			col = [4]float32{1, 1, 0, 1} // active = yellow
		} else if gs.HoverAxis == a.name {
			col = [4]float32{1, 1, 1, 1} // hover = white
		}

		gl.UniformMatrix4fv(gs.Renderer.LocModel, 1, false, &model[0])
		gl.UniformMatrix4fv(gs.Renderer.LocView, 1, false, &view[0])
		gl.UniformMatrix4fv(gs.Renderer.LocProj, 1, false, &proj[0])
		gl.Uniform4fv(gs.Renderer.LocColor, 1, &col[0])

		vao := gs.MeshManager.GetVAO("gizmo_arrow")
		count := gs.MeshManager.GetCount("gizmo_arrow")
		if vao == 0 || count == 0 {
			continue
		}

		gl.BindVertexArray(vao)
		gl.DrawElements(gl.TRIANGLES, count, gl.UNSIGNED_INT, gl.PtrOffset(0))
		gl.BindVertexArray(0)
	}
}
