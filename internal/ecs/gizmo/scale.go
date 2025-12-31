package gizmo

import (
	"go-engine/Go-Cordance/internal/ecs"
	"math"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

func (gs *GizmoRenderSystem) scaleHover(origin, dir, entityPos mgl32.Vec3, gizmoScale, closest float32) float32 {
	//handleSize := gizmoScale * 0.15
	pickRadius := gizmoScale * 0.25

	scaleAxes := []struct {
		name string
		axis mgl32.Vec3
	}{
		{"sx", mgl32.Vec3{1, 0, 0}},
		{"sy", mgl32.Vec3{0, 1, 0}},
		{"sz", mgl32.Vec3{0, 0, 1}},
	}

	for _, s := range scaleAxes {
		handlePos := entityPos.Add(s.axis.Mul(gizmoScale))
		hit, dist := RaySphereIntersect(origin, dir, handlePos, pickRadius)
		if hit && dist < closest {
			closest = dist
			gs.HoverAxis = s.name
		}
	}

	// uniform scale
	hit, dist := RaySphereIntersect(origin, dir, entityPos, pickRadius)
	if hit && dist < closest {
		closest = dist
		gs.HoverAxis = "su"
	}

	return closest
}

// scaleDrag scales either the active entity (t) or the whole selection around gizmoOrigin.
// - t: transform of the active entity (kept for local axes / single-entity updates)
// - origin, dir: current mouse ray
// - gizmoOrigin: pivot to scale around (selection center or active pivot)
func (gs *GizmoRenderSystem) scaleDrag(t *ecs.Transform, origin, dir, gizmoOrigin mgl32.Vec3) {
	if gs.ActiveAxis == "" {
		return
	}

	// uniform scale (screen-space)
	if gs.ActiveAxis == "su" {
		delta := dir.Dot(gs.dragStartRayDir)
		scaleFactor := float32(1.0) + delta
		if scaleFactor < 0.01 {
			scaleFactor = 0.01
		}

		// If multiple selected, apply uniform scale to all selection around gizmoOrigin
		if len(gs.SelectionIDs) > 1 && gs.World != nil {
			for _, id := range gs.SelectionIDs {
				if e := gs.World.FindByID(id); e != nil {
					if tr := ecs.GetTransform(e); tr != nil {
						// position relative to pivot
						pos := mgl32.Vec3{tr.Position[0], tr.Position[1], tr.Position[2]}
						rel := pos.Sub(gizmoOrigin)
						newPos := gizmoOrigin.Add(rel.Mul(scaleFactor))
						tr.Position[0], tr.Position[1], tr.Position[2] = newPos.X(), newPos.Y(), newPos.Z()

						// scale each entity uniformly
						tr.Scale[0] *= scaleFactor
						tr.Scale[1] *= scaleFactor
						tr.Scale[2] *= scaleFactor
					}
				}
			}
		} else {
			// single entity (existing behavior)
			t.Scale[0] *= scaleFactor
			t.Scale[1] *= scaleFactor
			t.Scale[2] *= scaleFactor
		}
		return
	}

	// axis scale (world axes or local axes depending on other flags)
	var axis mgl32.Vec3
	switch gs.ActiveAxis {
	case "sx":
		axis = mgl32.Vec3{1, 0, 0}
	case "sy":
		axis = mgl32.Vec3{0, 1, 0}
	case "sz":
		axis = mgl32.Vec3{0, 0, 1}
	default:
		return
	}

	// Use gizmoOrigin as the axis origin for projection
	t0 := projectRayOntoAxis(gs.dragStartRayOrigin, gs.dragStartRayDir, gizmoOrigin, axis)
	t1 := projectRayOntoAxis(origin, dir, gizmoOrigin, axis)
	delta := t1 - t0

	// snapping
	if gs.CameraSystem.Window().GetKey(glfw.KeyLeftControl) == glfw.Press ||
		gs.CameraSystem.Window().GetKey(glfw.KeyRightControl) == glfw.Press {
		snapped := float32(math.Round(float64(delta/SnapIncrement))) * SnapIncrement
		delta = snapped
	}

	scaleFactor := float32(1.0) + delta
	if scaleFactor < 0.01 {
		scaleFactor = 0.01
	}

	// If multiple selected, scale each entity's position relative to gizmoOrigin
	if len(gs.SelectionIDs) > 1 && gs.World != nil {
		for _, id := range gs.SelectionIDs {
			if e := gs.World.FindByID(id); e != nil {
				if tr := ecs.GetTransform(e); tr != nil {
					// position relative to pivot
					pos := mgl32.Vec3{tr.Position[0], tr.Position[1], tr.Position[2]}
					rel := pos.Sub(gizmoOrigin)

					// scale only along the chosen axis
					var scaledRel mgl32.Vec3
					switch gs.ActiveAxis {
					case "sx":
						scaledRel = mgl32.Vec3{rel.X() * scaleFactor, rel.Y(), rel.Z()}
						tr.Scale[0] *= scaleFactor
					case "sy":
						scaledRel = mgl32.Vec3{rel.X(), rel.Y() * scaleFactor, rel.Z()}
						tr.Scale[1] *= scaleFactor
					case "sz":
						scaledRel = mgl32.Vec3{rel.X(), rel.Y(), rel.Z() * scaleFactor}
						tr.Scale[2] *= scaleFactor
					}

					newPos := gizmoOrigin.Add(scaledRel)
					tr.Position[0], tr.Position[1], tr.Position[2] = newPos.X(), newPos.Y(), newPos.Z()
				}
			}
		}
	} else {
		// single entity: update scale on t (existing behavior)
		switch gs.ActiveAxis {
		case "sx":
			t.Scale[0] *= scaleFactor
		case "sy":
			t.Scale[1] *= scaleFactor
		case "sz":
			t.Scale[2] *= scaleFactor
		}
	}
}

func (gs *GizmoRenderSystem) renderScaleHandles(entityPos mgl32.Vec3, gizmoScale float32, view, proj mgl32.Mat4) {
	vao := gs.MeshManager.GetVAO("cube")
	count := gs.MeshManager.GetCount("cube")
	if vao == 0 || count == 0 {
		return
	}

	handles := []struct {
		name  string
		axis  mgl32.Vec3
		color [4]float32
	}{
		{"sx", mgl32.Vec3{1, 0, 0}, [4]float32{1, 0, 0, 1}},
		{"sy", mgl32.Vec3{0, 1, 0}, [4]float32{0, 1, 0, 1}},
		{"sz", mgl32.Vec3{0, 0, 1}, [4]float32{0, 0, 1, 1}},
	}

	for _, h := range handles {
		pos := entityPos.Add(h.axis.Mul(gizmoScale))

		model := mgl32.Translate3D(pos.X(), pos.Y(), pos.Z())
		model = model.Mul4(mgl32.Scale3D(gizmoScale*0.15, gizmoScale*0.15, gizmoScale*0.15))

		col := h.color
		if gs.ActiveAxis == h.name {
			col = [4]float32{1, 1, 0, 1}
		} else if gs.HoverAxis == h.name {
			col = [4]float32{1, 1, 1, 1}
		}

		gl.UniformMatrix4fv(gs.Renderer.LocModel, 1, false, &model[0])
		gl.UniformMatrix4fv(gs.Renderer.LocView, 1, false, &view[0])
		gl.UniformMatrix4fv(gs.Renderer.LocProj, 1, false, &proj[0])
		gl.Uniform4fv(gs.Renderer.LocColor, 1, &col[0])

		gl.BindVertexArray(vao)
		gl.DrawElements(gl.TRIANGLES, count, gl.UNSIGNED_INT, gl.PtrOffset(0))
		gl.BindVertexArray(0)
	}

	// uniform scale cube
	model := mgl32.Translate3D(entityPos.X(), entityPos.Y(), entityPos.Z())
	model = model.Mul4(mgl32.Scale3D(gizmoScale*0.2, gizmoScale*0.2, gizmoScale*0.2))

	col := [4]float32{1, 1, 1, 1}
	if gs.ActiveAxis == "su" {
		col = [4]float32{1, 1, 0, 1}
	} else if gs.HoverAxis == "su" {
		col = [4]float32{1, 1, 1, 1}
	}

	gl.UniformMatrix4fv(gs.Renderer.LocModel, 1, false, &model[0])
	gl.UniformMatrix4fv(gs.Renderer.LocView, 1, false, &view[0])
	gl.UniformMatrix4fv(gs.Renderer.LocProj, 1, false, &proj[0])
	gl.Uniform4fv(gs.Renderer.LocColor, 1, &col[0])

	gl.BindVertexArray(vao)
	gl.DrawElements(gl.TRIANGLES, count, gl.UNSIGNED_INT, gl.PtrOffset(0))
	gl.BindVertexArray(0)
}
