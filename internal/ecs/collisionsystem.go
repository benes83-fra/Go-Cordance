package ecs

import (
	"math"
)

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

// AABB collider (Axis-Aligned Bounding Box)
type ColliderAABB struct {
	HalfExtents [3]float32 // half-widths in x,y,z
}

func NewColliderAABB(halfExtents [3]float32) *ColliderAABB {
	return &ColliderAABB{HalfExtents: halfExtents}
}

func (c *ColliderAABB) ColliderType() string { return "AABB" }

func (c *ColliderPlane) ColliderType() string { return "plane" }

func (c *ColliderSphere) Update(dt float32) {
	_ = dt
}
func (c *ColliderPlane) Update(dt float32) {
	_ = dt
}

func (c *ColliderAABB) Update(dt float32) {
	_ = dt
}

// CollisionSystem checks for overlaps between colliders and resolves them.
type CollisionSystem struct{}

func NewCollisionSystem() *CollisionSystem {
	return &CollisionSystem{}
}

func (cs *CollisionSystem) Update(dt float32, entities []*Entity) {
	// Sphere–plane collisions
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
		if t != nil && rb != nil && sphere != nil {
			for _, other := range entities {
				var plane *ColliderPlane
				for _, c := range other.Components {
					if p, ok := c.(*ColliderPlane); ok {
						plane = p
					}
				}
				if plane != nil && t.Position[1]-sphere.Radius < plane.Y {
					t.Position[1] = plane.Y + sphere.Radius
					rb.Vel[1] = -rb.Vel[1] * 0.5
					rb.Vel[0] *= 0.9
					rb.Vel[2] *= 0.9
				}
			}
		}
	}

	// Sphere–sphere collisions (already added earlier)
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
			dx := tB.Position[0] - tA.Position[0]
			dy := tB.Position[1] - tA.Position[1]
			dz := tB.Position[2] - tA.Position[2]
			distSq := dx*dx + dy*dy + dz*dz
			radiusSum := sA.Radius + sB.Radius
			if distSq < radiusSum*radiusSum {
				dist := float32(math.Sqrt(float64(distSq)))
				if dist == 0 {
					dist = 0.0001
				}
				nx, ny, nz := dx/dist, dy/dist, dz/dist
				penetration := radiusSum - dist
				tA.Position[0] -= nx * penetration * 0.5
				tA.Position[1] -= ny * penetration * 0.5
				tA.Position[2] -= nz * penetration * 0.5
				tB.Position[0] += nx * penetration * 0.5
				tB.Position[1] += ny * penetration * 0.5
				tB.Position[2] += nz * penetration * 0.5
				va := rbA.Vel[0]*nx + rbA.Vel[1]*ny + rbA.Vel[2]*nz
				vb := rbB.Vel[0]*nx + rbB.Vel[1]*ny + rbB.Vel[2]*nz
				rbA.Vel[0] += (vb - va) * nx
				rbA.Vel[1] += (vb - va) * ny
				rbA.Vel[2] += (vb - va) * nz
				rbB.Vel[0] += (va - vb) * nx
				rbB.Vel[1] += (va - vb) * ny
				rbB.Vel[2] += (va - vb) * nz
			}
		}
	}

	// AABB–plane collisions
	for _, e := range entities {
		var t *Transform
		var rb *RigidBody
		var box *ColliderAABB
		for _, c := range e.Components {
			switch comp := c.(type) {
			case *Transform:
				t = comp
			case *RigidBody:
				rb = comp
			case *ColliderAABB:
				box = comp
			}
		}
		if t != nil && rb != nil && box != nil {
			for _, other := range entities {
				var plane *ColliderPlane
				for _, c := range other.Components {
					if p, ok := c.(*ColliderPlane); ok {
						plane = p
					}
				}
				if plane != nil {
					// Check bottom of box against plane
					if t.Position[1]-box.HalfExtents[1] < plane.Y {
						t.Position[1] = plane.Y + box.HalfExtents[1]
						rb.Vel[1] = -rb.Vel[1] * 0.5
						rb.Vel[0] *= 0.8
						rb.Vel[2] *= 0.8
					}
				}
			}
		}
	}
	// AABB–AABB collisions
	for i := 0; i < len(entities); i++ {
		var tA *Transform
		var rbA *RigidBody
		var boxA *ColliderAABB
		for _, c := range entities[i].Components {
			switch comp := c.(type) {
			case *Transform:
				tA = comp
			case *RigidBody:
				rbA = comp
			case *ColliderAABB:
				boxA = comp
			}
		}
		if tA == nil || rbA == nil || boxA == nil {
			continue
		}

		for j := i + 1; j < len(entities); j++ {
			var tB *Transform
			var rbB *RigidBody
			var boxB *ColliderAABB
			for _, c := range entities[j].Components {
				switch comp := c.(type) {
				case *Transform:
					tB = comp
				case *RigidBody:
					rbB = comp
				case *ColliderAABB:
					boxB = comp
				}
			}
			if tB == nil || rbB == nil || boxB == nil {
				continue
			}

			// Compute min/max for each box
			minA := [3]float32{
				tA.Position[0] - boxA.HalfExtents[0],
				tA.Position[1] - boxA.HalfExtents[1],
				tA.Position[2] - boxA.HalfExtents[2],
			}
			maxA := [3]float32{
				tA.Position[0] + boxA.HalfExtents[0],
				tA.Position[1] + boxA.HalfExtents[1],
				tA.Position[2] + boxA.HalfExtents[2],
			}
			minB := [3]float32{
				tB.Position[0] - boxB.HalfExtents[0],
				tB.Position[1] - boxB.HalfExtents[1],
				tB.Position[2] - boxB.HalfExtents[2],
			}
			maxB := [3]float32{
				tB.Position[0] + boxB.HalfExtents[0],
				tB.Position[1] + boxB.HalfExtents[1],
				tB.Position[2] + boxB.HalfExtents[2],
			}

			// Check overlap
			overlapX := minA[0] <= maxB[0] && maxA[0] >= minB[0]
			overlapY := minA[1] <= maxB[1] && maxA[1] >= minB[1]
			overlapZ := minA[2] <= maxB[2] && maxA[2] >= minB[2]

			if overlapX && overlapY && overlapZ {
				// Compute penetration depths
				penX := min(maxA[0]-minB[0], maxB[0]-minA[0])
				penY := min(maxA[1]-minB[1], maxB[1]-minA[1])
				penZ := min(maxA[2]-minB[2], maxB[2]-minA[2])

				// Resolve along smallest axis
				if penX < penY && penX < penZ {
					// Separate along X
					if tA.Position[0] < tB.Position[0] {
						tA.Position[0] -= penX / 2
						tB.Position[0] += penX / 2
					} else {
						tA.Position[0] += penX / 2
						tB.Position[0] -= penX / 2
					}
					rbA.Vel[0] = -rbA.Vel[0] * 0.5
					rbB.Vel[0] = -rbB.Vel[0] * 0.5
				} else if penY < penZ {
					// Separate along Y
					if tA.Position[1] < tB.Position[1] {
						tA.Position[1] -= penY / 2
						tB.Position[1] += penY / 2
					} else {
						tA.Position[1] += penY / 2
						tB.Position[1] -= penY / 2
					}
					rbA.Vel[1] = -rbA.Vel[1] * 0.5
					rbB.Vel[1] = -rbB.Vel[1] * 0.5
				} else {
					// Separate along Z
					if tA.Position[2] < tB.Position[2] {
						tA.Position[2] -= penZ / 2
						tB.Position[2] += penZ / 2
					} else {
						tA.Position[2] += penZ / 2
						tB.Position[2] -= penZ / 2
					}
					rbA.Vel[2] = -rbA.Vel[2] * 0.5
					rbB.Vel[2] = -rbB.Vel[2] * 0.5
				}
			}
		}
	}
	// Sphere–AABB collisions
	for _, e := range entities {
		var tS *Transform
		var rbS *RigidBody
		var sphere *ColliderSphere
		for _, c := range e.Components {
			switch comp := c.(type) {
			case *Transform:
				tS = comp
			case *RigidBody:
				rbS = comp
			case *ColliderSphere:
				sphere = comp
			}
		}
		if tS == nil || rbS == nil || sphere == nil {
			continue
		}

		for _, other := range entities {
			var tB *Transform
			//var rbB *RigidBody
			var box *ColliderAABB
			for _, c := range other.Components {
				switch comp := c.(type) {
				case *Transform:
					tB = comp
					//	case *RigidBody:
					//			rbB = comp
				case *ColliderAABB:
					box = comp
				}
			}
			if tB == nil || box == nil {
				continue
			}

			// Find closest point on AABB to sphere center
			closest := [3]float32{
				clamp(tS.Position[0], tB.Position[0]-box.HalfExtents[0], tB.Position[0]+box.HalfExtents[0]),
				clamp(tS.Position[1], tB.Position[1]-box.HalfExtents[1], tB.Position[1]+box.HalfExtents[1]),
				clamp(tS.Position[2], tB.Position[2]-box.HalfExtents[2], tB.Position[2]+box.HalfExtents[2]),
			}

			dx := tS.Position[0] - closest[0]
			dy := tS.Position[1] - closest[1]
			dz := tS.Position[2] - closest[2]
			distSq := dx*dx + dy*dy + dz*dz

			if distSq < sphere.Radius*sphere.Radius {
				dist := float32(math.Sqrt(float64(distSq)))
				if dist == 0 {
					dist = 0.0001
				}
				nx, ny, nz := dx/dist, dy/dist, dz/dist
				penetration := sphere.Radius - dist

				// Push sphere out of box
				tS.Position[0] += nx * penetration
				tS.Position[1] += ny * penetration
				tS.Position[2] += nz * penetration

				// Reflect sphere Vel
				dot := rbS.Vel[0]*nx + rbS.Vel[1]*ny + rbS.Vel[2]*nz
				rbS.Vel[0] -= 2 * dot * nx
				rbS.Vel[1] -= 2 * dot * ny
				rbS.Vel[2] -= 2 * dot * nz

				// Damping
				rbS.Vel[0] *= 0.5
				rbS.Vel[1] *= 0.5
				rbS.Vel[2] *= 0.5

			}

		}
	}
}

func min(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func clamp(val, min, max float32) float32 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}
