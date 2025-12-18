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
}
