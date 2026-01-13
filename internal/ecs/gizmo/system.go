package gizmo

import (
	"go-engine/Go-Cordance/internal/ecs/gizmo/bridge"
	"go-engine/Go-Cordance/internal/editor/state"
	"go-engine/Go-Cordance/internal/editor/undo"
	"go-engine/Go-Cordance/internal/engine"
	"log"
	"math"

	"go-engine/Go-Cordance/internal/ecs"

	"sync"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

const SnapIncrement = 0.25
const RotationSensitivity = 0.4  // lower = slower rotation
const RotationSnapDegrees = 15.0 // choose 5, 10, 15, 30, etc.
var globalId *ecs.Entity

type GizmoRenderSystem struct {
	Renderer      *engine.DebugRenderer
	MeshManager   *engine.MeshManager
	CameraSystem  *ecs.CameraSystem
	Enabled       bool
	Mode          GizmoMode
	HoverAxis     string
	ActiveAxis    string
	IsDragging    bool
	LocalRotation bool
	SelectionIDs  []int64
	PivotMode     state.PivotMode
	Undo          *undo.UndoStack
	World         *ecs.World
	dragBefore    []undo.TransformSnapshot

	dragStartRayOrigin mgl32.Vec3
	dragStartRayDir    mgl32.Vec3
	dragStartEntityPos mgl32.Vec3
	ShowLightGizmos    bool
}

func NewGizmoRenderSystem(r *engine.DebugRenderer, mm *engine.MeshManager, cs *ecs.CameraSystem) *GizmoRenderSystem {
	return &GizmoRenderSystem{
		Renderer:      r,
		MeshManager:   mm,
		CameraSystem:  cs,
		Mode:          GizmoCombined,
		Enabled:       true,
		LocalRotation: false,
		Undo:          undo.NewUndoStack(),
	}
}

func (gs *GizmoRenderSystem) SetCameraSystem(cs *ecs.CameraSystem) { gs.CameraSystem = cs }

func (gs *GizmoRenderSystem) Update(_ float32, _ []*ecs.Entity, selected *ecs.Entity) {
	if !gs.Enabled || gs.CameraSystem == nil || selected == nil {
		return
	}

	t := getTransform(selected)
	if t == nil {
		return
	}

	// compute shared values for the active entity (used for local axes, etc.)
	_, entityPos, gizmoScale := gs.computeFrameValues(t)
	localX, localY, localZ := computeLocalAxes(gs.LocalRotation, t)

	// compute mouse ray once
	mouseOrigin, mouseDir := RayFromMouse(gs.CameraSystem.Window(), gs.CameraSystem)
	globalId = selected
	// resolve selection entities (IDs -> []*ecs.Entity)
	selection := gs.selectedEntities()
	//log.Printf("gizmo: Update called with selection IDs = %v, resolved entities = %d", gs.SelectionIDs, len(selection))
	// Ensure the active entity (selected) is first in the selection slice so
	// computeGizmoOrigin treats it as the pivot when PivotModePivot.
	if selected != nil && len(selection) > 0 {
		for i, e := range selection {
			if e == selected {
				if i != 0 {
					// move element i to index 0 (preserve order of others)
					copy(selection[1:i+1], selection[0:i]) // shift right
					selection[0] = e
				}
				break
			}
		}
	}

	// compute gizmo origin from selection pivot mode; fall back to active entity position
	// compute gizmo origin from selection pivot mode; fall back to active entity position
	gizmoOrigin := entityPos
	if len(selection) > 0 {
		if gs.PivotMode == state.PivotModePivot {
			// Use the active entity (selected) as pivot to avoid relying on selection ordering
			gizmoOrigin = entityPos

		} else {
			// center mode: compute AABB center of selection
			gizmoOrigin = computeGizmoOrigin(selection, gs.PivotMode)

		}

		// recompute gizmoScale from camera distance to pivot so size follows selection pivot
		camPos := mgl32.Vec3{
			gs.CameraSystem.Position[0],
			gs.CameraSystem.Position[1],
			gs.CameraSystem.Position[2],
		}
		gizmoScale = camPos.Sub(gizmoOrigin).Len() * 0.08
	}

	// use mouseOrigin/mouseDir as origin/dir for hover/drag math
	origin, dir := mouseOrigin, mouseDir

	gs.HoverAxis = ""
	closest := float32(1e9)
	// --- HOVER PHASE ---
	if gs.Mode == GizmoMove || gs.Mode == GizmoCombined {
		closest = gs.axisHover(origin, dir, gizmoOrigin, gizmoScale, closest)
		closest = gs.planeHover(origin, dir, gizmoOrigin, gizmoScale, closest)
	}

	if gs.Mode == GizmoRotate || gs.Mode == GizmoCombined {
		closest = gs.rotationHover(origin, dir, gizmoOrigin, gizmoScale, localX, localY, localZ, closest)
	}

	if gs.Mode == GizmoScale || gs.Mode == GizmoCombined {
		closest = gs.scaleHover(origin, dir, gizmoOrigin, gizmoScale, closest)
	}

	// --- DRAG START / END ---
	gs.handleDragStart(origin, dir, gizmoOrigin)
	gs.handleDragEnd()

	// --- DRAGGING ---
	if gs.IsDragging {
		if gs.Mode == GizmoMove || gs.Mode == GizmoCombined {
			gs.axisDrag(t, origin, dir) // axisDrag still uses t for per-entity math
			gs.planeDrag(t, origin, dir)
		}

		if gs.Mode == GizmoRotate || gs.Mode == GizmoCombined {
			gs.rotationDrag(t, origin, dir, gizmoOrigin, localX, localY, localZ)
		}

		if gs.Mode == GizmoScale || gs.Mode == GizmoCombined {
			gs.scaleDrag(t, origin, dir, gizmoOrigin)
		}
	}

	// --- RENDER ---
	// --- RENDER ---
	view := gs.CameraSystem.View
	proj := gs.CameraSystem.Projection

	if gs.Mode == GizmoMove || gs.Mode == GizmoCombined {
		gs.renderAxes(t, gizmoScale, view, proj)
		gs.renderPlaneHandles(gizmoOrigin, gizmoScale, view, proj)
	}

	if gs.Mode == GizmoRotate || gs.Mode == GizmoCombined {
		gs.renderRotationRings(gizmoOrigin, gizmoScale, view, proj, localX, localY, localZ)
	}

	if gs.Mode == GizmoScale || gs.Mode == GizmoCombined {
		gs.renderScaleHandles(gizmoOrigin, gizmoScale, view, proj)
	}

	// --- LIGHT GIZMOS (optional) ---
	if state.Global.ShowLightGizmos {
		gs.renderSpotlightGizmos()
	}
}

// ...

func (gs *GizmoRenderSystem) computeFrameValues(t *ecs.Transform) (camPos, entityPos mgl32.Vec3, gizmoScale float32) {
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

func (gs *GizmoRenderSystem) handleDragStart(origin, dir, gizmoOrigin mgl32.Vec3) {
	if !gs.IsDragging &&
		gs.HoverAxis != "" &&
		gs.CameraSystem.Window().GetMouseButton(glfw.MouseButtonLeft) == glfw.Press {

		gs.ActiveAxis = gs.HoverAxis
		gs.IsDragging = true
		gs.dragStartRayOrigin = origin
		gs.dragStartRayDir = dir
		// use gizmo pivot for plane/axis math
		gs.dragStartEntityPos = gizmoOrigin
		log.Printf("gizmo: drag start selection IDs = %v", gs.SelectionIDs)
		// capture before snapshots for undo (if world available)
		if gs.World != nil && len(gs.SelectionIDs) > 0 {
			gs.dragBefore = captureSnapshotsByID(gs.World, gs.SelectionIDs)
		} else {
			gs.dragBefore = nil
		}
	}
}

func (gs *GizmoRenderSystem) handleDragEnd() {
	if gs.IsDragging &&
		gs.CameraSystem.Window().GetMouseButton(glfw.MouseButtonLeft) == glfw.Release {

		gs.IsDragging = false

		// --- SEND FINAL TRANSFORM TO EDITOR ---
		if gs.World != nil && len(gs.SelectionIDs) > 0 {
			for _, id := range gs.SelectionIDs {
				if e := gs.World.FindByID(id); e != nil {
					if tr := ecs.GetTransform(e); tr != nil {
						bridge.NotifyEditorOfTransformFinal(
							id,
							tr.Position,
							tr.Rotation,
							tr.Scale,
						)
					}
				}
			}
		}

		// capture after snapshots and push undo command
		if gs.World != nil && len(gs.SelectionIDs) > 0 && len(gs.dragBefore) > 0 {
			after := captureSnapshotsByID(gs.World, gs.SelectionIDs)
			cmd := &undo.TransformCommand{Before: gs.dragBefore, After: after}
			gs.Undo.Push(cmd)
		}

		gs.ActiveAxis = ""
	}
}

func (gs *GizmoRenderSystem) selectedEntities() []*ecs.Entity {
	if gs.World == nil || len(gs.SelectionIDs) == 0 {
		//log.Printf("gizmo: selectedEntities called but World is nil or no SelectionIDs...World=%v, SelectionIDs=%v", gs.World, gs.SelectionIDs)
		return nil
	}
	out := make([]*ecs.Entity, 0, len(gs.SelectionIDs))
	for _, id := range gs.SelectionIDs {
		if e := gs.World.FindByID(id); e != nil {
			out = append(out, e)
		}
	}
	return out
}

func (gs *GizmoRenderSystem) SetWorld(w *ecs.World) { gs.World = w }
func (gs *GizmoRenderSystem) SetSelectionIDs(ids []int64) {
	gs.SelectionIDs = ids
	log.Printf("gizmo from setter: SetSelectionIDs called with IDs = %v", ids)
}
func (gs *GizmoRenderSystem) SetPivotMode(mode state.PivotMode) {
	gs.PivotMode = mode
	log.Printf("gizmo: SetPivotMode called with mode = %v", mode)
}

// RegisterGlobalGizmo registers a global reference to the GizmoRenderSystem.
// Use this only as a small bridge when you don't want to change editor signatures.
var (
	globalGizmo   *GizmoRenderSystem
	globalGizmoMu sync.Mutex
)

func RegisterGlobalGizmo(gs *GizmoRenderSystem) {
	globalGizmoMu.Lock()
	defer globalGizmoMu.Unlock()
	globalGizmo = gs
}
func SetGlobalPivotMode(mode state.PivotMode) {
	globalGizmoMu.Lock()
	defer globalGizmoMu.Unlock()

	if globalGizmo != nil {
		globalGizmo.SetPivotMode(mode)
	}
}

// SetGlobalSelectionIDs updates the registered gizmo's selection IDs.
// Editor can call this when selection changes.
func SetGlobalSelectionIDs(ids []int64) {
	log.Printf("gizmo: SetGlobalSelectionIDs called with IDs = %v", ids)
	if globalGizmo != nil {
		globalGizmo.SetSelectionIDs(ids)
	}
}

func (gs *GizmoRenderSystem) renderSpotlightGizmos() {
	if gs.World == nil || gs.CameraSystem == nil {
		return
	}

	for _, e := range gs.World.Entities {
		lc, ok := e.GetComponent((*ecs.LightComponent)(nil)).(*ecs.LightComponent)
		if !ok || lc.Type != ecs.LightSpot {
			continue
		}

		tr, _ := e.GetComponent((*ecs.Transform)(nil)).(*ecs.Transform)
		if tr == nil {
			continue
		}

		gs.drawSpotlightCone(tr, lc)
	}
}

func (gs *GizmoRenderSystem) drawSpotlightCone(tr *ecs.Transform, lc *ecs.LightComponent) {
	// Position and direction
	pos := mgl32.Vec3{tr.Position[0], tr.Position[1], tr.Position[2]}
	q := mgl32.Quat{
		W: tr.Rotation[0],
		V: mgl32.Vec3{tr.Rotation[1], tr.Rotation[2], tr.Rotation[3]},
	}
	dir := q.Rotate(mgl32.Vec3{0, 0, -1}) // assuming -Z is forward

	// Spotlight parameters
	angleRad := lc.Angle * (math.Pi / 180.0)
	length := lc.Range
	if length <= 0 {
		return
	}

	// Radius of the cone at the end
	radius := float32(math.Tan(float64(angleRad))) * length

	// Center of the cone end
	end := pos.Add(dir.Mul(length))

	// Choose a color (e.g. yellow)
	color := mgl32.Vec3{1, 1, 0}

	const segments = 16
	// Build an orthonormal basis for the circle plane
	up := mgl32.Vec3{0, 1, 0}
	if math.Abs(float64(dir.Dot(up))) > 0.99 {
		up = mgl32.Vec3{1, 0, 0}
	}
	right := dir.Cross(up).Normalize()
	up = right.Cross(dir).Normalize()

	var prev mgl32.Vec3
	for i := 0; i <= segments; i++ {
		a := float32(i) * 2 * math.Pi / segments
		circlePoint := end.
			Add(right.Mul(radius * float32(math.Cos(float64(a))))).
			Add(up.Mul(radius * float32(math.Sin(float64(a)))))

			// lines from origin to rim
		view := gs.CameraSystem.View
		proj := gs.CameraSystem.Projection

		gs.Renderer.DrawLine(pos, circlePoint, color, view, proj)

		if i > 0 {
			gs.Renderer.DrawLine(prev, circlePoint, color, view, proj)
		}

		prev = circlePoint
	}
}
