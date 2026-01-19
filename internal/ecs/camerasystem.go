// internal/ecs/camerasystem.go
package ecs

import (
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

type CameraSystem struct {
	View       mgl32.Mat4
	Projection mgl32.Mat4
	window     *glfw.Window
	Position   [3]float32 // NEW
}

func NewCameraSystem(window *glfw.Window) *CameraSystem {
	return &CameraSystem{window: window}
}

func (cs *CameraSystem) Update(_ float32, entities []*Entity) {
	w, h := cs.window.GetSize()
	if h == 0 {
		h = 1
	}
	aspect := float32(w) / float32(h)

	for _, e := range entities {
		if cam, ok := e.GetComponent((*Camera)(nil)).(*Camera); ok {
			cs.View = cam.ViewMatrix()
			cs.Projection = cam.ProjectionMatrix()
			cs.Position = cam.Position
		}
		for _, c := range e.Components {
			if cam, ok := c.(*Camera); ok && cam.Active {
				cs.View = mgl32.LookAtV(
					mgl32.Vec3{cam.Position[0], cam.Position[1], cam.Position[2]},
					mgl32.Vec3{cam.Target[0], cam.Target[1], cam.Target[2]},
					mgl32.Vec3{cam.Up[0], cam.Up[1], cam.Up[2]},
				)

				cs.Projection = mgl32.Perspective(mgl32.DegToRad(cam.Fov), aspect, cam.Near, cam.Far)
				cs.Position = cam.Position // IMPORTANT
				return
			}
		}
	}
	// Fallback to identity (prevents black if no active camera)
	cs.View = mgl32.Ident4()
	cs.Projection = mgl32.Ident4()
}
func (cs *CameraSystem) Window() *glfw.Window {
	return cs.window
}

// Forward returns the camera forward vector (direction camera is looking toward) in world space.
func (cs *CameraSystem) Forward() mgl32.Vec3 {
	// Derive forward from the view matrix.
	// Using the view matrix layout, -Z axis of the camera in world space is:
	f := mgl32.Vec3{-cs.View[8], -cs.View[9], -cs.View[10]}
	return f.Normalize()
}

func (cs *CameraSystem) FocusOn(e *Entity) {
	t, ok := e.GetComponent((*Transform)(nil)).(*Transform)
	if !ok {
		return
	}

	target := mgl32.Vec3{t.Position[0], t.Position[1], t.Position[2]}

	// Move camera back a bit
	offset := mgl32.Vec3{0, 2, 6} // tweak as needed
	cs.Position = [3]float32{
		target.X() + offset.X(),
		target.Y() + offset.Y(),
		target.Z() + offset.Z(),
	}

	// Look at entity
	cs.LookAt(target)
}

func (cs *CameraSystem) LookAt(target mgl32.Vec3) {
	pos := mgl32.Vec3{cs.Position[0], cs.Position[1], cs.Position[2]}
	up := mgl32.Vec3{0, 1, 0}
	cs.View = mgl32.LookAtV(pos, target, up)
}
