package ecs

// AngularDamping applies drag to angular velocity.
// Factor is typically in [0,1], where 1 = no damping, 0 = full stop.
type AngularDamping struct {
	Factor float32
}

func NewAngularDamping(factor float32) *AngularDamping {
	return &AngularDamping{Factor: factor}
}

func (ad *AngularDamping) Update(dt float32) {
	_ = dt
}
