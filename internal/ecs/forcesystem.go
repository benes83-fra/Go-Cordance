package ecs

// ForceSystem applies a constant or dynamic force to all rigid bodies.
type ForceSystem struct {
	Force [3]float32
}

// NewForceSystem creates a system with a constant force vector.
func NewForceSystem(fx, fy, fz float32) *ForceSystem {
	return &ForceSystem{Force: [3]float32{fx, fy, fz}}
}

// Update applies the force to all entities with a RigidBody.
func (fs *ForceSystem) Update(dt float32, entities []*Entity) {
	for _, e := range entities {
		for _, c := range e.Components {
			if rb, ok := c.(*RigidBody); ok {
				rb.ApplyForce(fs.Force[0], fs.Force[1], fs.Force[2])
			}
		}
	}
}
