package ecs

import (
	"go-engine/Go-Cordance/internal/engine"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

const SnapIncrement = 0.25
const RotationSensitivity = 0.4  // lower = slower rotation
const RotationSnapDegrees = 15.0 // choose 5, 10, 15, 30, etc.

type GizmoRenderSystem struct {
	Renderer     *engine.DebugRenderer
	MeshManager  *engine.MeshManager
	CameraSystem *CameraSystem
	Enabled      bool

	HoverAxis     string
	ActiveAxis    string
	IsDragging    bool
	LocalRotation bool

	dragStartRayOrigin mgl32.Vec3
	dragStartRayDir    mgl32.Vec3
	dragStartEntityPos mgl32.Vec3
}

func NewGizmoRenderSystem(r *engine.DebugRenderer, mm *engine.MeshManager, cs *CameraSystem) *GizmoRenderSystem {
	return &GizmoRenderSystem{
		Renderer:      r,
		MeshManager:   mm,
		CameraSystem:  cs,
		Enabled:       true,
		LocalRotation: false,
	}
}

func (gs *GizmoRenderSystem) SetCameraSystem(cs *CameraSystem) { gs.CameraSystem = cs }

func (gs *GizmoRenderSystem) Update(_ float32, _ []*Entity, selected *Entity) {
	if !gs.Enabled || gs.CameraSystem == nil || selected == nil {
		return
	}

	t := getTransform(selected)
	if t == nil {
		return
	}

	// compute shared values
	_, entityPos, gizmoScale := gs.computeFrameValues(t)
	origin, dir := RayFromMouse(gs.CameraSystem.Window(), gs.CameraSystem)
	localX, localY, localZ := computeLocalAxes(gs.LocalRotation, t)

	gs.HoverAxis = ""
	closest := float32(1e9)

	// --- HOVER PHASE ---
	closest = gs.axisHover(origin, dir, entityPos, gizmoScale, closest)
	closest = gs.planeHover(origin, dir, entityPos, gizmoScale, closest)
	closest = gs.rotationHover(origin, dir, entityPos, gizmoScale, localX, localY, localZ, closest)
	closest = gs.scaleHover(origin, dir, entityPos, gizmoScale, closest)

	// --- DRAG START / END ---
	gs.handleDragStart(origin, dir, entityPos)
	gs.handleDragEnd()

	// --- DRAGGING ---
	if gs.IsDragging {
		gs.axisDrag(t, origin, dir)
		gs.planeDrag(t, origin, dir)
		gs.rotationDrag(t, origin, dir, entityPos, localX, localY, localZ)
		gs.scaleDrag(t, origin, dir)
	}

	// --- RENDER ---
	view := gs.CameraSystem.View
	proj := gs.CameraSystem.Projection

	gs.renderAxes(t, gizmoScale, view, proj)
	gs.renderPlaneHandles(entityPos, gizmoScale, view, proj)
	gs.renderRotationRings(entityPos, gizmoScale, view, proj, localX, localY, localZ)
	gs.renderScaleHandles(entityPos, gizmoScale, view, proj)
}

// ...

func (gs *GizmoRenderSystem) computeFrameValues(t *Transform) (camPos, entityPos mgl32.Vec3, gizmoScale float32) {
	camPos = mgl32.Vec3{
		gs.CameraSystem.Position[0],
		gs.CameraSystem.Position[1],
		gs.CameraSystem.Position[2],
	}
	entityPos = mgl32.Vec3{t.Position[0], t.Position[1], t.Position[2]}
	dist := camPos.Sub(entityPos).Len()
	gizmoScale = float32(dist * 0.08)
	return
}

func (gs *GizmoRenderSystem) handleDragStart(origin, dir, entityPos mgl32.Vec3) {
	if !gs.IsDragging &&
		gs.HoverAxis != "" &&
		gs.CameraSystem.Window().GetMouseButton(glfw.MouseButtonLeft) == glfw.Press {

		gs.ActiveAxis = gs.HoverAxis
		gs.IsDragging = true
		gs.dragStartRayOrigin = origin
		gs.dragStartRayDir = dir
		gs.dragStartEntityPos = entityPos
	}
}

func (gs *GizmoRenderSystem) handleDragEnd() {
	if gs.IsDragging &&
		gs.CameraSystem.Window().GetMouseButton(glfw.MouseButtonLeft) == glfw.Release {

		gs.IsDragging = false
		gs.ActiveAxis = ""
	}
}
