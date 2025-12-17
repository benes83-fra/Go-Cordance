package ecs

// TorqueSystem applies a constant torque vector to all entities
// with AngularVelocity (and optionally AngularAcceleration).
type TorqueSystem struct {
	Torque [3]float32
}

// NewTorqueSystem creates a system with a constant torque vector.
func NewTorqueSystem(tx, ty, tz float32) *TorqueSystem {
	return &TorqueSystem{Torque: [3]float32{tx, ty, tz}}
}

// Update applies torque as angular acceleration to entities.
func (ts *TorqueSystem) Update(dt float32, entities []*Entity) {
	for _, e := range entities {
		var av *AngularVelocity
		var aa *AngularAcceleration
		for _, c := range e.Components {
			switch comp := c.(type) {
			case *AngularVelocity:
				av = comp
			case *AngularAcceleration:
				aa = comp
			}
		}
		if av != nil {
			// If AngularAcceleration exists, add torque to it
			if aa != nil {
				aa.Acc[0] += ts.Torque[0]
				aa.Acc[1] += ts.Torque[1]
				aa.Acc[2] += ts.Torque[2]
			} else {
				// Otherwise directly modify angular velocity
				av.Vel[0] += ts.Torque[0] * dt
				av.Vel[1] += ts.Torque[1] * dt
				av.Vel[2] += ts.Torque[2] * dt
			}
		}
	}
}
