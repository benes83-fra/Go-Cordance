package ecs

import (
	"math"

	"go-engine/Go-Cordance/internal/engine"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

const SnapIncrement = 0.25

type GizmoRenderSystem struct {
	Renderer           *engine.DebugRenderer
	MeshManager        *engine.MeshManager
	CameraSystem       *CameraSystem
	Enabled            bool
	dragStartRayOrigin mgl32.Vec3
	dragStartRayDir    mgl32.Vec3
	dragStartEntityPos mgl32.Vec3

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

	// find transform on selected entity
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

	// camera and entity positions
	camPos := mgl32.Vec3{gs.CameraSystem.Position[0], gs.CameraSystem.Position[1], gs.CameraSystem.Position[2]}
	entityPos := mgl32.Vec3{t.Position[0], t.Position[1], t.Position[2]}

	// scale gizmo with distance so it stays readable
	dist := camPos.Sub(entityPos).Len()
	gizmoScale := float32(dist * 0.08)

	// ray from mouse
	origin, dir := RayFromMouse(gs.CameraSystem.Window(), gs.CameraSystem)

	// reset hover
	gs.HoverAxis = ""

	axisLength := gizmoScale * 1.0
	pickRadius := gizmoScale * 0.2

	// axis definitions
	axes := []struct {
		name string
		axis mgl32.Vec3
	}{
		{"x", mgl32.Vec3{1, 0, 0}},
		{"y", mgl32.Vec3{0, 1, 0}},
		{"z", mgl32.Vec3{0, 0, 1}},
	}

	// plane definitions (name and plane normal in world space)
	planes := []struct {
		name   string
		normal mgl32.Vec3
	}{
		{"xy", mgl32.Vec3{0, 0, 1}},
		{"xz", mgl32.Vec3{0, 1, 0}},
		{"yz", mgl32.Vec3{1, 0, 0}},
	}

	// --- Axis hover detection (capsule) ---
	closest := float32(math.MaxFloat32)
	for _, a := range axes {
		a0 := entityPos
		a1 := entityPos.Add(a.axis.Mul(axisLength))

		hit, d := RayCapsuleIntersect(origin, dir, a0, a1, pickRadius)
		if hit && d < closest {
			closest = d
			gs.HoverAxis = a.name
		}
	}
	// --- Render plane handles ---
	vaoPlane := gs.MeshManager.GetVAO("gizmo_plane")
	countPlane := gs.MeshManager.GetCount("gizmo_plane")

	if vaoPlane != 0 && countPlane != 0 {

		planesRender := []struct {
			name  string
			color [4]float32
			rot   mgl32.Mat4
		}{
			// XY plane (facing +Z)
			{"xy", [4]float32{1, 1, 0, 0.35}, mgl32.Ident4()},
			// XZ plane (facing +Y)
			{"xz", [4]float32{0, 1, 1, 0.35}, mgl32.HomogRotate3DX(float32(-math.Pi / 2))},
			// YZ plane (facing +X)
			{"yz", [4]float32{1, 0, 1, 0.35}, mgl32.HomogRotate3DY(float32(math.Pi / 2))},
		}

		for _, p := range planesRender {

			model := mgl32.Translate3D(entityPos.X(), entityPos.Y(), entityPos.Z())
			model = model.Mul4(p.rot)
			model = model.Mul4(mgl32.Scale3D(gizmoScale, gizmoScale, gizmoScale))

			col := p.color

			// highlight
			if gs.ActiveAxis == p.name {
				col = [4]float32{1, 1, 0, 0.8} // active = yellow
			} else if gs.HoverAxis == p.name {
				col = [4]float32{1, 1, 1, 0.8} // hover = white
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

	// --- Plane hover detection (square centered on entity) ---
	// We compute plane intersection t and compare to closest to pick the nearest hit
	for _, p := range planes {
		hit, tplane := RayPlaneIntersection(origin, dir, entityPos, p.normal)
		if !hit {
			continue
		}

		// intersection point in world space
		point := origin.Add(dir.Mul(tplane))
		local := point.Sub(entityPos)
		half := gizmoScale * 0.5

		inside := false
		switch p.name {
		case "xy":
			if float32(math.Abs(float64(local.X()))) <= half && float32(math.Abs(float64(local.Y()))) <= half {
				inside = true
			}
		case "xz":
			if float32(math.Abs(float64(local.X()))) <= half && float32(math.Abs(float64(local.Z()))) <= half {
				inside = true
			}
		case "yz":
			if float32(math.Abs(float64(local.Y()))) <= half && float32(math.Abs(float64(local.Z()))) <= half {
				inside = true
			}
		}

		// choose plane only if it's the closest hit
		if inside && tplane < closest {
			closest = tplane
			gs.HoverAxis = p.name
		}
	}

	// --- Drag start / end handling ---
	// Start drag when LMB pressed over a hover target
	if !gs.IsDragging && gs.HoverAxis != "" && gs.CameraSystem.Window().GetMouseButton(glfw.MouseButtonLeft) == glfw.Press {
		gs.ActiveAxis = gs.HoverAxis
		gs.IsDragging = true
		gs.dragStartRayOrigin = origin
		gs.dragStartRayDir = dir
		gs.dragStartEntityPos = entityPos
	}

	// End drag when LMB released
	if gs.IsDragging && gs.CameraSystem.Window().GetMouseButton(glfw.MouseButtonLeft) == glfw.Release {
		gs.IsDragging = false
		gs.ActiveAxis = ""
	}

	// --- Dragging logic ---
	if gs.IsDragging && gs.ActiveAxis != "" {
		// Axis dragging (X/Y/Z)
		if gs.ActiveAxis == "x" || gs.ActiveAxis == "y" || gs.ActiveAxis == "z" {
			var axis mgl32.Vec3
			switch gs.ActiveAxis {
			case "x":
				axis = mgl32.Vec3{1, 0, 0}
			case "y":
				axis = mgl32.Vec3{0, 1, 0}
			case "z":
				axis = mgl32.Vec3{0, 0, 1}
			}

			t0 := projectRayOntoAxis(gs.dragStartRayOrigin, gs.dragStartRayDir, gs.dragStartEntityPos, axis)
			t1 := projectRayOntoAxis(origin, dir, gs.dragStartEntityPos, axis)
			delta := t1 - t0

			// snapping with CTRL
			if gs.CameraSystem.Window().GetKey(glfw.KeyLeftControl) == glfw.Press ||
				gs.CameraSystem.Window().GetKey(glfw.KeyRightControl) == glfw.Press {
				snapped := float32(math.Round(float64(delta/SnapIncrement))) * SnapIncrement
				delta = snapped
			}

			newPos := gs.dragStartEntityPos.Add(axis.Mul(delta))
			t.Position[0] = newPos.X()
			t.Position[1] = newPos.Y()
			t.Position[2] = newPos.Z()
		}

		// Plane dragging (XY, XZ, YZ)
		if gs.ActiveAxis == "xy" || gs.ActiveAxis == "xz" || gs.ActiveAxis == "yz" {
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
			if hit {
				point := origin.Add(dir.Mul(tplane))
				delta := point.Sub(gs.dragStartEntityPos)

				// snapping per-component if CTRL held
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
		}
	}

	// --- Rendering axes (no plane meshes here) ---
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

		// If a plane is hovered, dim axes slightly to emphasize plane (optional)
		if gs.HoverAxis == "xy" || gs.HoverAxis == "xz" || gs.HoverAxis == "yz" {
			col = [4]float32{col[0] * 0.6, col[1] * 0.6, col[2] * 0.6, 1}
		}

		// highlight logic
		if gs.ActiveAxis == a.name {
			col = [4]float32{1, 1, 0, 1} // active axis = yellow
		} else if gs.HoverAxis == a.name {
			col = [4]float32{1, 1, 1, 1} // hover axis = white
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

	// Optionally: draw a simple plane indicator if you later add a mesh for planes.
	// For now plane highlighting is handled by dimming axes and using HoverAxis/ActiveAxis state.
}
