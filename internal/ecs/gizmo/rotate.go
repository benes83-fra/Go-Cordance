package gizmo

import (
	"go-engine/Go-Cordance/internal/ecs"
	"math"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

// --- HOVER: rotation rings ---
func (gs *GizmoRenderSystem) rotationHover(
	origin, dir, entityPos mgl32.Vec3,
	gizmoScale float32,
	localX, localY, localZ mgl32.Vec3,
	closest float32,
) float32 {
	ringRadius := gizmoScale * 1.2
	ringThickness := gizmoScale * 0.15

	rotationAxes := []struct {
		name   string
		normal mgl32.Vec3
	}{
		{"rx", localX},
		{"ry", localY},
		{"rz", localZ},
	}

	for _, r := range rotationAxes {
		hit, dist := RayCircleIntersect(origin, dir, entityPos, r.normal, ringRadius, ringThickness)
		if hit && dist < closest {
			closest = dist
			gs.HoverAxis = r.name
		}
	}

	// --- Free rotate ring (screen-space) ---
	{
		// ring normal faces the camera
		camForward := gs.CameraSystem.Forward()

		// slightly larger radius than axis rings so it sits outside them
		freeRadius := gizmoScale * 1.4
		freeThickness := ringThickness

		hit, dist := RayCircleIntersect(origin, dir, entityPos, camForward, freeRadius, freeThickness)
		if hit && dist < closest {
			closest = dist
			gs.HoverAxis = "rfree"
		}
	}

	return closest
}

// --- DRAG: rotation ---
func (gs *GizmoRenderSystem) rotationDrag(
	t *ecs.Transform,
	origin, dir, entityPos, localX, localY, localZ mgl32.Vec3,
) {
	// --- Free rotate (screen-space) ---
	if gs.ActiveAxis == "rfree" {
		// project start and current rays onto camera-facing plane at gizmoOrigin
		camForward := gs.CameraSystem.Forward()
		hit0, t0 := RayPlaneIntersection(gs.dragStartRayOrigin, gs.dragStartRayDir, gs.dragStartEntityPos, camForward)
		hit1, t1 := RayPlaneIntersection(origin, dir, gs.dragStartEntityPos, camForward)
		if !hit0 || !hit1 {
			return
		}
		p0 := gs.dragStartRayOrigin.Add(gs.dragStartRayDir.Mul(t0))
		p1 := origin.Add(dir.Mul(t1))

		v0 := p0.Sub(gs.dragStartEntityPos).Normalize()
		v1 := p1.Sub(gs.dragStartEntityPos).Normalize()

		angle := float32(math.Acos(float64(ClampFloat32(v0.Dot(v1), -1, 1))))
		cross := v0.Cross(v1)
		if cross.Dot(camForward) < 0 {
			angle = -angle
		}

		// snapping
		if gs.CameraSystem.Window().GetKey(glfw.KeyLeftControl) == glfw.Press ||
			gs.CameraSystem.Window().GetKey(glfw.KeyRightControl) == glfw.Press {
			snapRad := degToRad(RotationSnapDegrees)
			angle = float32(math.Round(float64(angle/snapRad))) * snapRad
		}

		angle *= RotationSensitivity

		// apply rotation to selection around gizmoOrigin using camForward axis
		q := mgl32.QuatRotate(angle, camForward)
		if len(gs.SelectionIDs) > 1 && gs.World != nil {
			for _, id := range gs.SelectionIDs {
				if e := gs.World.FindByID(id); e != nil {
					if tr := ecs.GetTransform(e); tr != nil {
						// rotate position around pivot
						pos := mgl32.Vec3{tr.Position[0], tr.Position[1], tr.Position[2]}
						rel := pos.Sub(gs.dragStartEntityPos)
						newPos := gs.dragStartEntityPos.Add(q.Rotate(rel))
						tr.Position[0], tr.Position[1], tr.Position[2] = newPos.X(), newPos.Y(), newPos.Z()

						// rotate orientation
						curQ := mgl32.Quat{W: tr.Rotation[0], V: mgl32.Vec3{tr.Rotation[1], tr.Rotation[2], tr.Rotation[3]}}
						newQ := q.Mul(curQ).Normalize()
						tr.Rotation[0] = newQ.W
						tr.Rotation[1] = newQ.V[0]
						tr.Rotation[2] = newQ.V[1]
						tr.Rotation[3] = newQ.V[2]
					}
				}
			}
		} else {
			// single entity: existing behavior (rotate t)
			current := mgl32.Quat{W: t.Rotation[0], V: mgl32.Vec3{t.Rotation[1], t.Rotation[2], t.Rotation[3]}}
			newQ := q.Mul(current).Normalize()
			t.Rotation[0] = newQ.W
			t.Rotation[1] = newQ.V[0]
			t.Rotation[2] = newQ.V[1]
			t.Rotation[3] = newQ.V[2]
		}
		return
	}

	// --- Axis rotations (existing logic) ---
	if gs.ActiveAxis != "rx" && gs.ActiveAxis != "ry" && gs.ActiveAxis != "rz" {
		return
	}

	var axis mgl32.Vec3
	switch gs.ActiveAxis {
	case "rx":
		axis = localX
	case "ry":
		axis = localY
	case "rz":
		axis = localZ
	}

	hit0, t0 := RayPlaneIntersection(gs.dragStartRayOrigin, gs.dragStartRayDir, gs.dragStartEntityPos, axis)
	hit1, t1 := RayPlaneIntersection(origin, dir, gs.dragStartEntityPos, axis)
	if !hit0 || !hit1 {
		return
	}

	p0 := gs.dragStartRayOrigin.Add(gs.dragStartRayDir.Mul(t0))
	p1 := origin.Add(dir.Mul(t1))

	v0 := p0.Sub(entityPos).Normalize()
	v1 := p1.Sub(entityPos).Normalize()

	angle := float32(math.Acos(float64(v0.Dot(v1))))

	cross := v0.Cross(v1)
	if cross.Dot(axis) < 0 {
		angle = -angle
	}

	// snapping
	if gs.CameraSystem.Window().GetKey(glfw.KeyLeftControl) == glfw.Press ||
		gs.CameraSystem.Window().GetKey(glfw.KeyRightControl) == glfw.Press {

		snapRad := degToRad(RotationSnapDegrees)
		angle = float32(math.Round(float64(angle/snapRad))) * snapRad
	}

	// sensitivity
	angle *= RotationSensitivity

	q := mgl32.QuatRotate(angle, axis)
	current := mgl32.Quat{
		W: t.Rotation[0],
		V: mgl32.Vec3{t.Rotation[1], t.Rotation[2], t.Rotation[3]},
	}
	newQ := q.Mul(current).Normalize()

	t.Rotation[0] = newQ.W
	t.Rotation[1] = newQ.V[0]
	t.Rotation[2] = newQ.V[1]
	t.Rotation[3] = newQ.V[2]
}

// --- RENDER: rotation rings ---
func (gs *GizmoRenderSystem) renderRotationRings(
	entityPos mgl32.Vec3,
	gizmoScale float32,
	view, proj mgl32.Mat4,
	localX, localY, localZ mgl32.Vec3,
) {
	vao := gs.MeshManager.GetVAO("gizmo_circle")
	count := gs.MeshManager.GetCount("gizmo_circle")
	if vao == 0 || count == 0 {
		return
	}

	rings := []struct {
		name  string
		axis  mgl32.Vec3
		color [4]float32
		rot   mgl32.Mat4
	}{
		{"rx", localX, [4]float32{1, 0, 0, 1}, rotationFromAxis(localX)},
		{"ry", localY, [4]float32{0, 1, 0, 1}, rotationFromAxis(localY)},
		{"rz", localZ, [4]float32{0, 0, 1, 1}, rotationFromAxis(localZ)},
	}

	for _, r := range rings {
		model := mgl32.Translate3D(entityPos.X(), entityPos.Y(), entityPos.Z())
		model = model.Mul4(r.rot)
		model = model.Mul4(mgl32.Scale3D(gizmoScale*1.2, gizmoScale*1.2, gizmoScale*1.2))

		col := r.color
		if gs.ActiveAxis == r.name {
			col = [4]float32{1, 1, 0, 1}
		} else if gs.HoverAxis == r.name {
			col = [4]float32{1, 1, 1, 1}
		}

		gl.UniformMatrix4fv(gs.Renderer.LocModel, 1, false, &model[0])
		gl.UniformMatrix4fv(gs.Renderer.LocView, 1, false, &view[0])
		gl.UniformMatrix4fv(gs.Renderer.LocProj, 1, false, &proj[0])
		gl.Uniform4fv(gs.Renderer.LocColor, 1, &col[0])

		gl.BindVertexArray(vao)
		gl.DrawElements(gl.LINES, count, gl.UNSIGNED_INT, gl.PtrOffset(0))
		gl.BindVertexArray(0)
	}
	// --- Free rotate ring ---
	{
		camForward := gs.CameraSystem.Forward()
		rot := rotationFromAxis(camForward)

		model := mgl32.Translate3D(entityPos.X(), entityPos.Y(), entityPos.Z())
		model = model.Mul4(rot)
		model = model.Mul4(mgl32.Scale3D(gizmoScale*1.4, gizmoScale*1.4, gizmoScale*1.4))

		col := [4]float32{1, 1, 1, 1} // white
		if gs.ActiveAxis == "rfree" {
			col = [4]float32{1, 1, 0, 1} // active = yellow
		} else if gs.HoverAxis == "rfree" {
			col = [4]float32{1, 1, 1, 1} // hover = bright white
		}

		gl.UniformMatrix4fv(gs.Renderer.LocModel, 1, false, &model[0])
		gl.UniformMatrix4fv(gs.Renderer.LocView, 1, false, &view[0])
		gl.UniformMatrix4fv(gs.Renderer.LocProj, 1, false, &proj[0])
		gl.Uniform4fv(gs.Renderer.LocColor, 1, &col[0])

		gl.BindVertexArray(vao)
		gl.DrawElements(gl.LINES, count, gl.UNSIGNED_INT, gl.PtrOffset(0))
		gl.BindVertexArray(0)
	}

}
