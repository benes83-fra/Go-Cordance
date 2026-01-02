package gizmo

import (
	"go-engine/Go-Cordance/internal/ecs"
	"go-engine/Go-Cordance/internal/ecs/gizmo/bridge"
	"math"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

// --- HOVER: axis move X/Y/Z ---
func (gs *GizmoRenderSystem) axisHover(
	origin, dir, entityPos mgl32.Vec3,
	gizmoScale float32,
	closest float32,
) float32 {
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

	for _, a := range axes {
		a0 := entityPos
		a1 := entityPos.Add(a.axis.Mul(axisLength))

		hit, d := RayCapsuleIntersect(origin, dir, a0, a1, pickRadius)
		if hit && d < closest {
			closest = d
			gs.HoverAxis = a.name
		}
	}

	return closest
}

// --- DRAG: axis move X/Y/Z ---
func (gs *GizmoRenderSystem) axisDrag(t *ecs.Transform, origin, dir mgl32.Vec3) {
	if gs.ActiveAxis != "x" && gs.ActiveAxis != "y" && gs.ActiveAxis != "z" {
		return
	}

	var axis mgl32.Vec3
	switch gs.ActiveAxis {
	case "x":
		axis = mgl32.Vec3{1, 0, 0}
	case "y":
		axis = mgl32.Vec3{0, 1, 0}
	case "z":
		axis = mgl32.Vec3{0, 0, 1}
	}

	// compute delta as before
	t0 := projectRayOntoAxis(gs.dragStartRayOrigin, gs.dragStartRayDir, gs.dragStartEntityPos, axis)
	t1 := projectRayOntoAxis(origin, dir, gs.dragStartEntityPos, axis)
	delta := t1 - t0

	// snapping with CTRL (unchanged)
	if gs.CameraSystem.Window().GetKey(glfw.KeyLeftControl) == glfw.Press ||
		gs.CameraSystem.Window().GetKey(glfw.KeyRightControl) == glfw.Press {
		snapped := float32(math.Round(float64(delta/SnapIncrement))) * SnapIncrement
		delta = snapped
	}

	// If multiple selected, move each entity by axis*delta
	if len(gs.SelectionIDs) > 1 && gs.World != nil {
		for _, id := range gs.SelectionIDs {
			if e := gs.World.FindByID(id); e != nil {
				if tr := ecs.GetTransform(e); tr != nil {
					pos := mgl32.Vec3{tr.Position[0], tr.Position[1], tr.Position[2]}
					newPos := pos.Add(axis.Mul(delta))
					tr.Position[0], tr.Position[1], tr.Position[2] = newPos.X(), newPos.Y(), newPos.Z()

					if bridge.SendTransformToEditor != nil {
						bridge.SendTransformToEditor(
							id,
							tr.Position,
							tr.Rotation,
							tr.Scale,
						)
					}
				}
			}
		}
		return
	}

	// single entity fallback (existing behavior)
	newPos := gs.dragStartEntityPos.Add(axis.Mul(delta))
	t.Position[0] = newPos.X()
	t.Position[1] = newPos.Y()
	t.Position[2] = newPos.Z()

}

// --- RENDER: axis arrows ---
func (gs *GizmoRenderSystem) renderAxes(
	t *ecs.Transform,
	gizmoScale float32,
	view, proj mgl32.Mat4,
) {
	axes := []struct {
		name string
		axis mgl32.Vec3
	}{
		{"x", mgl32.Vec3{1, 0, 0}},
		{"y", mgl32.Vec3{0, 1, 0}},
		{"z", mgl32.Vec3{0, 0, 1}},
	}

	for _, a := range axes {
		model := mgl32.Translate3D(t.Position[0], t.Position[1], t.Position[2])

		// rotate arrow mesh from +Z to axis direction
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

		// base axis colors
		var col [4]float32
		switch a.name {
		case "x":
			col = [4]float32{1, 0, 0, 1}
		case "y":
			col = [4]float32{0, 1, 0, 1}
		case "z":
			col = [4]float32{0, 0, 1, 1}
		}

		// dim when plane hovered
		if gs.HoverAxis == "xy" || gs.HoverAxis == "xz" || gs.HoverAxis == "yz" {
			col = [4]float32{col[0] * 0.6, col[1] * 0.6, col[2] * 0.6, 1}
		}

		// highlight
		if gs.ActiveAxis == a.name {
			col = [4]float32{1, 1, 0, 1}
		} else if gs.HoverAxis == a.name {
			col = [4]float32{1, 1, 1, 1}
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
