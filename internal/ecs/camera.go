package ecs

// Camera component holds parameters for view/projection.
type Camera struct {
	Position [3]float32
	Target   [3]float32
	Up       [3]float32
	Fov      float32 // field of view in degrees
	Near     float32
	Far      float32
	Active   bool // mark one camera as active
}

func NewCamera() *Camera {
	return &Camera{
		Position: [3]float32{0, 0, 3},
		Target:   [3]float32{0, 0, 0},
		Up:       [3]float32{0, 1, 0},
		Fov:      60,
		Near:     0.1,
		Far:      100,
		Active:   true,
	}
}

func (c *Camera) Update(dt float32) {
	_ = dt
}
