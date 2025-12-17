package ecs

// AngularMass (Moment of Inertia) represents rotational inertia around each axis.
// For simplicity, we store separate inertia values for X, Y, Z axes.
type AngularMass struct {
	Inertia [3]float32
}

// NewAngularMass creates an AngularMass with given inertia values.
func NewAngularMass(ix, iy, iz float32) *AngularMass {
	return &AngularMass{Inertia: [3]float32{ix, iy, iz}}
}

// Update is a no-op; integration is handled by PhysicsSystem.
func (am *AngularMass) Update(dt float32) {
	_ = dt
}
