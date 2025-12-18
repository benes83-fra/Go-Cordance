package ecs

import "math"

// Collider is a component with a simple bounding sphere.
type Collider interface {
	ColliderType() string
}

// --- Sphere Collider ---
type ColliderSphere struct {
	Radius float32
}

func NewColliderSphere(radius float32) *ColliderSphere {
	return &ColliderSphere{Radius: radius}
}

func (c *ColliderSphere) ColliderType() string { return "sphere" }

// --- Plane Collider ---
type ColliderPlane struct {
	Y float32 // horizontal plane at this Y level
}

func NewColliderPlane(y float32) *ColliderPlane {
	return &ColliderPlane{Y: y}
}

func (c *ColliderPlane) ColliderType() string { return "plane" }

func (c *ColliderSphere) Update(dt float32) {
	_ = dt
}
func (c *ColliderPlane) Update(dt float32) {
	_ = dt
}

// CollisionSystem checks for overlaps between colliders and resolves them.
type CollisionSystem struct{}

func NewCollisionSystem() *CollisionSystem {
	return &CollisionSystem{}
}

func (cs *CollisionSystem) Update(dt float32, entities []*Entity) {
	// First handle sphere–plane collisions
	for _, e := range entities {
		var t *Transform
		var rb *RigidBody
		var sphere *ColliderSphere

		for _, c := range e.Components {
			switch comp := c.(type) {
			case *Transform:
				t = comp
			case *RigidBody:
				rb = comp
			case *ColliderSphere:
				sphere = comp
			}
		}
		if t == nil || rb == nil || sphere == nil {
			continue
		}

		// Check against all planes
		for _, other := range entities {
			var plane *ColliderPlane
			for _, c := range other.Components {
				if p, ok := c.(*ColliderPlane); ok {
					plane = p
				}
			}
			if plane == nil {
				continue
			}

			if t.Position[1]-sphere.Radius < plane.Y {
				t.Position[1] = plane.Y + sphere.Radius
				rb.Vel[1] = -rb.Vel[1] * 0.5
				rb.Vel[0] *= 0.9
				rb.Vel[2] *= 0.9
			}
		}
	}

	// Now handle sphere–sphere collisions
	for i := 0; i < len(entities); i++ {
		var tA *Transform
		var rbA *RigidBody
		var sA *ColliderSphere
		for _, c := range entities[i].Components {
			switch comp := c.(type) {
			case *Transform:
				tA = comp
			case *RigidBody:
				rbA = comp
			case *ColliderSphere:
				sA = comp
			}
		}
		if tA == nil || rbA == nil || sA == nil {
			continue
		}

		for j := i + 1; j < len(entities); j++ {
			var tB *Transform
			var rbB *RigidBody
			var sB *ColliderSphere
			for _, c := range entities[j].Components {
				switch comp := c.(type) {
				case *Transform:
					tB = comp
				case *RigidBody:
					rbB = comp
				case *ColliderSphere:
					sB = comp
				}
			}
			if tB == nil || rbB == nil || sB == nil {
				continue
			}

			// Distance between centers
			dx := tB.Position[0] - tA.Position[0]
			dy := tB.Position[1] - tA.Position[1]
			dz := tB.Position[2] - tA.Position[2]
			distSq := dx*dx + dy*dy + dz*dz
			radiusSum := sA.Radius + sB.Radius

			if distSq < radiusSum*radiusSum {
				dist := float32(math.Sqrt(float64(distSq)))
				if dist == 0 {
					dist = 0.0001 // avoid divide by zero
				}

				// Normal vector
				nx := dx / dist
				ny := dy / dist
				nz := dz / dist

				// Penetration depth
				penetration := radiusSum - dist

				// Push spheres apart equally
				tA.Position[0] -= nx * penetration * 0.5
				tA.Position[1] -= ny * penetration * 0.5
				tA.Position[2] -= nz * penetration * 0.5

				tB.Position[0] += nx * penetration * 0.5
				tB.Position[1] += ny * penetration * 0.5
				tB.Position[2] += nz * penetration * 0.5

				// Reflect velocities along collision normal (basic elastic response)
				va := rbA.Vel[0]*nx + rbA.Vel[1]*ny + rbA.Vel[2]*nz
				vb := rbB.Vel[0]*nx + rbB.Vel[1]*ny + rbB.Vel[2]*nz

				// Swap Vel components along normal
				rbA.Vel[0] += (vb - va) * nx
				rbA.Vel[1] += (vb - va) * ny
				rbA.Vel[2] += (vb - va) * nz

				rbB.Vel[0] += (va - vb) * nx
				rbB.Vel[1] += (va - vb) * ny
				rbB.Vel[2] += (va - vb) * nz
			}
		}
	}
}

/*
func (cs *CollisionSystem) Update(dt float32, entities []*Entity) {
	for i := 0; i < len(entities); i++ {
		for j := i + 1; j < len(entities); j++ {
			var t1, t2 *Transform
			var s1, s2 *ColliderSphere
			var p1, p2 *ColliderPlane
			var rb1, rb2 *RigidBody

			for _, comp := range entities[i].Components {
				switch v := comp.(type) {
				case *Transform:
					t1 = v
				case *ColliderSphere:
					s1 = v
				case *ColliderPlane:
					p2 = v
				case *RigidBody:
					rb1 = v
				}
			}
			for _, comp := range entities[j].Components {
				switch v := comp.(type) {
				case *Transform:
					t2 = v
				case *ColliderSphere:
					s2 = v
				case *ColliderPlane:
					p2 = v
				case *RigidBody:
					rb2 = v
				}
			}

			if t1 != nil && t2 != nil && s1 != nil && s2 != nil {
				dx := t2.Position[0] - t1.Position[0]
				dy := t2.Position[1] - t1.Position[1]
				dz := t2.Position[2] - t1.Position[2]
				distSq := dx*dx + dy*dy + dz*dz
				rSum := s1.Radius + s2.Radius

				if distSq < rSum*rSum {
					// Simple response: separate along vector and invert Vel
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
*/
