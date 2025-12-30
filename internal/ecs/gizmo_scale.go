package ecs

import (
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

func (gs *GizmoRenderSystem) scaleDrag(t *Transform, origin, dir mgl32.Vec3) {
	if gs.ActiveAxis == "" {
		return
	}

	// uniform scale
	if gs.ActiveAxis == "su" {
		delta := dir.Dot(gs.dragStartRayDir)
		scale := float32(1.0) + delta
		if scale < 0.01 {
			scale = 0.01
		}
		t.Scale[0] *= scale
		t.Scale[1] *= scale
		t.Scale[2] *= scale
		return
	}

	// axis scale
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

	t0 := projectRayOntoAxis(gs.dragStartRayOrigin, gs.dragStartRayDir, gs.dragStartEntityPos, axis)
	t1 := projectRayOntoAxis(origin, dir, gs.dragStartEntityPos, axis)
	delta := t1 - t0

	if gs.CameraSystem.Window().GetKey(glfw.KeyLeftControl) == glfw.Press ||
		gs.CameraSystem.Window().GetKey(glfw.KeyRightControl) == glfw.Press {
		snapped := float32(math.Round(float64(delta/SnapIncrement))) * SnapIncrement
		delta = snapped
	}

	scale := float32(1.0) + delta
	if scale < 0.01 {
		scale = 0.01
	}

	switch gs.ActiveAxis {
	case "sx":
		t.Scale[0] *= scale
	case "sy":
		t.Scale[1] *= scale
	case "sz":
		t.Scale[2] *= scale
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
