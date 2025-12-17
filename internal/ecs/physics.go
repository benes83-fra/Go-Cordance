package ecs

// Velocity is a simple component that holds linear velocity in 3D.
type Velocity struct {
	V [3]float32
}

// Update is a no-op; physics integration is handled by PhysicsSystem.
func (v *Velocity) Update(dt float32) {
	_ = dt
}

// PhysicsSystem integrates Velocity into Transform each frame.
// Later you can expand this to include acceleration, forces, collisions, etc.
type PhysicsSystem struct{}

// NewPhysicsSystem creates a new physics system.
func NewPhysicsSystem() *PhysicsSystem {
	return &PhysicsSystem{}
}

// Update iterates entities and applies velocity to their Transform.
func (ps *PhysicsSystem) Update(dt float32, entities []*Entity) {
	for _, e := range entities {
		var t *Transform
		var rb *RigidBody
		var acc *Acceleration
		var damp *Damping
		var av *AngularVelocity
		var aa *AngularAcceleration
		var ad *AngularDamping
		var am *AngularMass

		for _, c := range e.Components {
			switch comp := c.(type) {
			case *Transform:
				t = comp
			case *RigidBody:
				rb = comp
			case *Acceleration:
				acc = comp
			case *Damping:
				damp = comp
			case *AngularVelocity:
				av = comp
			case *AngularAcceleration:
				aa = comp

			case *AngularDamping:
				ad = comp
			case *AngularMass:
				am = comp

			}
		}

		// Linear integration
		if t != nil && rb != nil && rb.Mass > 0 {
			ax := rb.Force[0] / rb.Mass
			ay := rb.Force[1] / rb.Mass
			az := rb.Force[2] / rb.Mass
			if acc != nil {
				ax += acc.A[0]
				ay += acc.A[1]
				az += acc.A[2]
			}

			rb.Vel[0] += ax * dt
			rb.Vel[1] += ay * dt
			rb.Vel[2] += az * dt

			if damp != nil {
				rb.Vel[0] *= damp.Factor
				rb.Vel[1] *= damp.Factor
				rb.Vel[2] *= damp.Factor
			}

			t.Position[0] += rb.Vel[0] * dt
			t.Position[1] += rb.Vel[1] * dt
			t.Position[2] += rb.Vel[2] * dt

			rb.ClearForce()
		}

		// Angular integration

		// Apply angular acceleration if present
		if t != nil && av != nil {
			// Apply angular acceleration if present
			if aa != nil {
				av.Vel[0] += aa.Acc[0] * dt
				av.Vel[1] += aa.Acc[1] * dt
				av.Vel[2] += aa.Acc[2] * dt
			}
			if am != nil {
				if aa != nil {
					if am.Inertia[0] > 0 {
						av.Vel[0] += (aa.Acc[0] / am.Inertia[0]) * dt
					}
					if am.Inertia[1] > 0 {
						av.Vel[1] += (aa.Acc[1] / am.Inertia[1]) * dt
					}
					if am.Inertia[2] > 0 {
						av.Vel[2] += (aa.Acc[2] / am.Inertia[2]) * dt
					}
				}
			}
			// Apply angular damping if present
			if ad != nil {
				av.Vel[0] *= ad.Factor
				av.Vel[1] *= ad.Factor
				av.Vel[2] *= ad.Factor
			}
			// Integrate rotation
			t.Rotation[0] += av.Vel[0] * dt
			t.Rotation[1] += av.Vel[1] * dt
			t.Rotation[2] += av.Vel[2] * dt
		}

	}
}
