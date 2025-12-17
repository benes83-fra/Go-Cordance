package ecs

// Damping is a component that applies velocity damping (drag).
// Factor is typically in [0,1], where 1 = no damping, 0 = full stop.
type Damping struct {
	Factor float32
}

// NewDamping creates a damping component with given factor.
func NewDamping(factor float32) *Damping {
	return &Damping{Factor: factor}
}

// Update is a no-op; integration is handled by PhysicsSystem.
func (d *Damping) Update(dt float32) {
	_ = dt
}
