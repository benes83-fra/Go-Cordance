package ecs

import (
	"math"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

type CameraControllerSystem struct {
	window      *glfw.Window
	lastX       float64
	lastY       float64
	firstRun    bool
	yaw         float32
	pitch       float32
	speed       float32
	sensitivity float32
}

func NewCameraControllerSystem(window *glfw.Window) *CameraControllerSystem {
	return &CameraControllerSystem{
		window:      window,
		speed:       2.5, // movement speed units/sec
		sensitivity: 0.1, // mouse sensitivity
		firstRun:    true,
	}
}

func (cc *CameraControllerSystem) Update(dt float32, entities []*Entity) {
	for _, e := range entities {
		for _, c := range e.Components {
			if cam, ok := c.(*Camera); ok && cam.Active {
				cc.handleKeyboard(dt, cam)
				cc.handleMouse(cam)
			}
		}
	}
}

func (cc *CameraControllerSystem) handleKeyboard(dt float32, cam *Camera) {
	// forward vector
	dir := mgl32.Vec3{
		cam.Target[0] - cam.Position[0],
		cam.Target[1] - cam.Position[1],
		cam.Target[2] - cam.Position[2],
	}.Normalize()

	right := dir.Cross(mgl32.Vec3{cam.Up[0], cam.Up[1], cam.Up[2]}).Normalize()

	if cc.window.GetKey(glfw.KeyW) == glfw.Press {
		cam.Position[0] += dir[0] * cc.speed * dt
		cam.Position[1] += dir[1] * cc.speed * dt
		cam.Position[2] += dir[2] * cc.speed * dt
	}
	if cc.window.GetKey(glfw.KeyS) == glfw.Press {
		cam.Position[0] -= dir[0] * cc.speed * dt
		cam.Position[1] -= dir[1] * cc.speed * dt
		cam.Position[2] -= dir[2] * cc.speed * dt
	}
	if cc.window.GetKey(glfw.KeyA) == glfw.Press {
		cam.Position[0] -= right[0] * cc.speed * dt
		cam.Position[1] -= right[1] * cc.speed * dt
		cam.Position[2] -= right[2] * cc.speed * dt
	}
	if cc.window.GetKey(glfw.KeyD) == glfw.Press {
		cam.Position[0] += right[0] * cc.speed * dt
		cam.Position[1] += right[1] * cc.speed * dt
		cam.Position[2] += right[2] * cc.speed * dt
	}
}

func (cc *CameraControllerSystem) handleMouse(cam *Camera) {
	x, y := cc.window.GetCursorPos()
	if cc.firstRun {
		cc.lastX, cc.lastY = x, y
		cc.firstRun = false
	}
	xoffset := float32(x-cc.lastX) * cc.sensitivity
	yoffset := float32(cc.lastY-y) * cc.sensitivity // reversed
	cc.lastX, cc.lastY = x, y

	cc.yaw += xoffset
	cc.pitch += yoffset
	if cc.pitch > 89 {
		cc.pitch = 89
	}
	if cc.pitch < -89 {
		cc.pitch = -89
	}

	front := mgl32.Vec3{
		float32(math.Cos(float64(mgl32.DegToRad(cc.yaw))) * math.Cos(float64(mgl32.DegToRad(cc.pitch)))),
		float32(math.Sin(float64(mgl32.DegToRad(cc.pitch)))),
		float32(math.Sin(float64(mgl32.DegToRad(cc.yaw))) * math.Cos(float64(mgl32.DegToRad(cc.pitch)))),
	}.Normalize()

	cam.Target[0] = cam.Position[0] + front[0]
	cam.Target[1] = cam.Position[1] + front[1]
	cam.Target[2] = cam.Position[2] + front[2]
}
