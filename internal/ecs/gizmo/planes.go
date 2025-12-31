package gizmo

import (
	"go-engine/Go-Cordance/internal/ecs"
	"math"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

// --- HOVER: plane XY / XZ / YZ ---
func (gs *GizmoRenderSystem) planeHover(
	origin, dir, entityPos mgl32.Vec3,
	gizmoScale float32,
	closest float32,
) float32 {
	planes := []struct {
		name   string
		normal mgl32.Vec3
	}{
		{"xy", mgl32.Vec3{0, 0, 1}},
		{"xz", mgl32.Vec3{0, 1, 0}},
		{"yz", mgl32.Vec3{1, 0, 0}},
	}

	for _, p := range planes {
		hit, tplane := RayPlaneIntersection(origin, dir, entityPos, p.normal)
		if !hit {
			continue
		}

		point := origin.Add(dir.Mul(tplane))
		local := point.Sub(entityPos)
		half := gizmoScale * 0.5

		inside := false
		switch p.name {
		case "xy":
			if float32(math.Abs(float64(local.X()))) <= half &&
				float32(math.Abs(float64(local.Y()))) <= half {
				inside = true
			}
		case "xz":
			if float32(math.Abs(float64(local.X()))) <= half &&
				float32(math.Abs(float64(local.Z()))) <= half {
				inside = true
			}
		case "yz":
			if float32(math.Abs(float64(local.Y()))) <= half &&
				float32(math.Abs(float64(local.Z()))) <= half {
				inside = true
			}
		}

		if inside && tplane < closest {
			closest = tplane
			gs.HoverAxis = p.name
		}
	}

	return closest
}

// --- DRAG: plane move XY / XZ / YZ ---
func (gs *GizmoRenderSystem) planeDrag(t *ecs.Transform, origin, dir mgl32.Vec3) {
	if gs.ActiveAxis != "xy" && gs.ActiveAxis != "xz" && gs.ActiveAxis != "yz" {
		return
	}

	var planeNormal mgl32.Vec3
	switch gs.ActiveAxis {
	case "xy":
		planeNormal = mgl32.Vec3{0, 0, 1}
	case "xz":
		planeNormal = mgl32.Vec3{0, 1, 0}
	case "yz":
		planeNormal = mgl32.Vec3{1, 0, 0}
	}

	hit, tplane := RayPlaneIntersection(origin, dir, gs.dragStartEntityPos, planeNormal)
	if !hit {
		return
	}

	point := origin.Add(dir.Mul(tplane))
	delta := point.Sub(gs.dragStartEntityPos)

	if gs.CameraSystem.Window().GetKey(glfw.KeyLeftControl) == glfw.Press ||
		gs.CameraSystem.Window().GetKey(glfw.KeyRightControl) == glfw.Press {

		delta[0] = float32(math.Round(float64(delta[0]/SnapIncrement))) * SnapIncrement
		delta[1] = float32(math.Round(float64(delta[1]/SnapIncrement))) * SnapIncrement
		delta[2] = float32(math.Round(float64(delta[2]/SnapIncrement))) * SnapIncrement
	}

	newPos := gs.dragStartEntityPos.Add(delta)
	t.Position[0] = newPos.X()
	t.Position[1] = newPos.Y()
	t.Position[2] = newPos.Z()
}

// --- RENDER: plane squares ---
func (gs *GizmoRenderSystem) renderPlaneHandles(
	entityPos mgl32.Vec3,
	gizmoScale float32,
	view, proj mgl32.Mat4,
) {
	vaoPlane := gs.MeshManager.GetVAO("gizmo_plane")
	countPlane := gs.MeshManager.GetCount("gizmo_plane")
	if vaoPlane == 0 || countPlane == 0 {
		return
	}

	planesRender := []struct {
		name  string
		color [4]float32
		rot   mgl32.Mat4
	}{
		{"xy", [4]float32{1, 1, 0, 0.35}, mgl32.Ident4()},
		{"xz", [4]float32{0, 1, 1, 0.35}, mgl32.HomogRotate3DX(float32(-math.Pi / 2))},
		{"yz", [4]float32{1, 0, 1, 0.35}, mgl32.HomogRotate3DY(float32(math.Pi / 2))},
	}

	for _, p := range planesRender {
		model := mgl32.Translate3D(entityPos.X(), entityPos.Y(), entityPos.Z())
		model = model.Mul4(p.rot)
		model = model.Mul4(mgl32.Scale3D(gizmoScale, gizmoScale, gizmoScale))

		col := p.color
		if gs.ActiveAxis == p.name {
			col = [4]float32{1, 1, 0, 0.8}
		} else if gs.HoverAxis == p.name {
			col = [4]float32{1, 1, 1, 0.8}
		}

		gl.UniformMatrix4fv(gs.Renderer.LocModel, 1, false, &model[0])
		gl.UniformMatrix4fv(gs.Renderer.LocView, 1, false, &view[0])
		gl.UniformMatrix4fv(gs.Renderer.LocProj, 1, false, &proj[0])
		gl.Uniform4fv(gs.Renderer.LocColor, 1, &col[0])

		gl.Enable(gl.BLEND)
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

		gl.BindVertexArray(vaoPlane)
		gl.DrawElements(gl.TRIANGLES, countPlane, gl.UNSIGNED_INT, gl.PtrOffset(0))
		gl.BindVertexArray(0)

		gl.Disable(gl.BLEND)
	}
}
