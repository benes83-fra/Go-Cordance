package ecs

import (
	"github.com/go-gl/mathgl/mgl32"
)

// CameraSystem computes view/projection matrices for the active camera.
type CameraSystem struct {
	View       mgl32.Mat4
	Projection mgl32.Mat4
}

// NewCameraSystem creates a new camera system.
func NewCameraSystem() *CameraSystem {
	return &CameraSystem{}
}

func (cs *CameraSystem) Update(dt float32, entities []*Entity) {
	_ = dt
	for _, e := range entities {
		for _, c := range e.Components {
			if cam, ok := c.(*Camera); ok && cam.Active {
				// Compute view matrix
				cs.View = mgl32.LookAtV(
					mgl32.Vec3{cam.Position[0], cam.Position[1], cam.Position[2]},
					mgl32.Vec3{cam.Target[0], cam.Target[1], cam.Target[2]},
					mgl32.Vec3{cam.Up[0], cam.Up[1], cam.Up[2]},
				)
				// Compute projection matrix
				aspect := float32(800.0 / 600.0) // TODO: use actual window size
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
