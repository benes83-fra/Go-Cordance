package ecs

// Acceleration is a component that stores linear acceleration in 3D.
// It can be used directly (e.g., constant gravity) or updated by systems.
type Acceleration struct {
	A [3]float32
}

// NewAcceleration creates an acceleration component with given vector.
func NewAcceleration(ax, ay, az float32) *Acceleration {
	return &Acceleration{A: [3]float32{ax, ay, az}}
}

// Update is a no-op; integration is handled by PhysicsSystem.
func (a *Acceleration) Update(dt float32) {
	_ = dt
}
