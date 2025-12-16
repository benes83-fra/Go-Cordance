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
		for _, c := range e.Components {
			switch comp := c.(type) {
			case *Transform:
				t = comp
			case *RigidBody:
				rb = comp
			case *Acceleration:
				acc = comp
			}
		}
		if t != nil && rb != nil {
			// Compute acceleration from forces if any
			ax := rb.Force[0]/rb.Mass + acc.A[0]
			ay := rb.Force[1]/rb.Mass + acc.A[1]
			az := rb.Force[2]/rb.Mass + acc.A[2]

			// Integrate velocity
			rb.Vel[0] += ax * dt
			rb.Vel[1] += ay * dt
			rb.Vel[2] += az * dt

			// Integrate position
			t.Position[0] += rb.Vel[0] * dt
			t.Position[1] += rb.Vel[1] * dt
			t.Position[2] += rb.Vel[2] * dt

			// Clear forces for next frame
			rb.ClearForce()
		}
	}
}
