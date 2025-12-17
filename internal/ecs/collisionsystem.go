package ecs

import "math"

// Collider is a component with a simple bounding sphere.
type Collider struct {
	Radius float32
}

// NewCollider creates a collider with given radius.
func NewCollider(r float32) *Collider {
	return &Collider{Radius: r}
}

func (c *Collider) Update(dt float32) {
	_ = dt
}

// CollisionSystem checks for overlaps between colliders and resolves them.
type CollisionSystem struct{}

func NewCollisionSystem() *CollisionSystem {
	return &CollisionSystem{}
}

func (cs *CollisionSystem) Update(dt float32, entities []*Entity) {
	for i := 0; i < len(entities); i++ {
		for j := i + 1; j < len(entities); j++ {
			var t1, t2 *Transform
			var c1, c2 *Collider
			var rb1, rb2 *RigidBody

			for _, comp := range entities[i].Components {
				switch v := comp.(type) {
				case *Transform:
					t1 = v
				case *Collider:
					c1 = v
				case *RigidBody:
					rb1 = v
				}
			}
			for _, comp := range entities[j].Components {
				switch v := comp.(type) {
				case *Transform:
					t2 = v
				case *Collider:
					c2 = v
				case *RigidBody:
					rb2 = v
				}
			}

			if t1 != nil && t2 != nil && c1 != nil && c2 != nil {
				dx := t2.Position[0] - t1.Position[0]
				dy := t2.Position[1] - t1.Position[1]
				dz := t2.Position[2] - t1.Position[2]
				distSq := dx*dx + dy*dy + dz*dz
				rSum := c1.Radius + c2.Radius

				if distSq < rSum*rSum {
					// Simple response: separate along vector and invert velocity
					dist := float32(math.Sqrt(float64(distSq)))
					if dist == 0 {
						dist = 0.001
					}
					overlap := rSum - dist
					nx, ny, nz := dx/dist, dy/dist, dz/dist

					// Push entities apart
					t1.Position[0] -= nx * overlap / 2
					t1.Position[1] -= ny * overlap / 2
					t1.Position[2] -= nz * overlap / 2
					t2.Position[0] += nx * overlap / 2
					t2.Position[1] += ny * overlap / 2
					t2.Position[2] += nz * overlap / 2

					// Invert velocities for a simple bounce
					if rb1 != nil {
						rb1.Vel[0] = -rb1.Vel[0]
						rb1.Vel[1] = -rb1.Vel[1]
						rb1.Vel[2] = -rb1.Vel[2]
					}
					if rb2 != nil {
						rb2.Vel[0] = -rb2.Vel[0]
						rb2.Vel[1] = -rb2.Vel[1]
						rb2.Vel[2] = -rb2.Vel[2]
					}
				}
			}
		}
	}
}
