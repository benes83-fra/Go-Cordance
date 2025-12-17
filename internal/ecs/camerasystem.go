package ecs

import (
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

type CameraSystem struct {
	View       mgl32.Mat4
	Projection mgl32.Mat4
	window     *glfw.Window
}

func NewCameraSystem(window *glfw.Window) *CameraSystem {
	return &CameraSystem{window: window}
}

func (cs *CameraSystem) Update(dt float32, entities []*Entity) {
	_ = dt
	w, h := cs.window.GetSize()
	aspect := float32(w) / float32(h)

	for _, e := range entities {
		for _, c := range e.Components {
			if cam, ok := c.(*Camera); ok && cam.Active {
				cs.View = mgl32.LookAtV(
					mgl32.Vec3{cam.Position[0], cam.Position[1], cam.Position[2]},
					mgl32.Vec3{cam.Target[0], cam.Target[1], cam.Target[2]},
					mgl32.Vec3{cam.Up[0], cam.Up[1], cam.Up[2]},
				)
				cs.Projection = mgl32.Perspective(
					mgl32.DegToRad(cam.Fov),
					aspect,
					cam.Near,
					cam.Far,
				)
				return
			}
		}
	}
}
