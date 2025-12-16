package ecs

// RigidBody is a physics component with mass, velocity, and accumulated force.
type RigidBody struct {
	Mass  float32
	Vel   [3]float32
	Force [3]float32
}

// NewRigidBody creates a rigid body with given mass.
func NewRigidBody(mass float32) *RigidBody {
	return &RigidBody{Mass: mass}
}

// ApplyForce adds a force vector to the body (accumulated until next update).
func (rb *RigidBody) ApplyForce(fx, fy, fz float32) {
	rb.Force[0] += fx
	rb.Force[1] += fy
	rb.Force[2] += fz
}

// ClearForce resets accumulated forces (called after integration).
func (rb *RigidBody) ClearForce() {
	rb.Force = [3]float32{0, 0, 0}
}

// Update is a no-op; integration is handled by PhysicsSystem.
func (rb *RigidBody) Update(dt float32) {
	_ = dt
}
