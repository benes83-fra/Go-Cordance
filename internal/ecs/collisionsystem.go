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
	Layer  int
	Mask   uint32
}

func NewColliderSphere(radius float32) *ColliderSphere {
	return &ColliderSphere{
		Radius: radius,
		Mask:   0xFFFFFFFF,
	}
}

func (c *ColliderSphere) ColliderType() string { return "sphere" }

// --- Plane Collider ---
type ColliderPlane struct {
	Y     float32 // horizontal plane at this Y level
	Layer int
	Mask  uint32
}

func NewColliderPlane(y float32) *ColliderPlane {
	return &ColliderPlane{
		Y:    y,
		Mask: 0xFFFFFFFF,
	}
}

// AABB collider (Axis-Aligned Bounding Box)
type ColliderAABB struct {
	HalfExtents [3]float32 // half-widths in x,y,z
	Layer       int
	Mask        uint32
}

func NewColliderAABB(halfExtents [3]float32) *ColliderAABB {
	return &ColliderAABB{
		HalfExtents: halfExtents,
		Mask:        0xFFFFFFFF,
	}
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
type CollisionSystem struct {
	spheres []sphereBody
	boxes   []boxBody
	planes  []planeBody
}

func NewCollisionSystem() *CollisionSystem {
	return &CollisionSystem{}
}

// keep these types near CollisionSystem
type sphereBody struct {
	t *Transform
	r *RigidBody
	c *ColliderSphere
}
type boxBody struct {
	t *Transform
	r *RigidBody
	c *ColliderAABB
}
type planeBody struct {
	c *ColliderPlane
}

func (cs *CollisionSystem) Update(dt float32, entities []*Entity) {
	_ = dt

	cs.spheres = cs.spheres[:0]
	cs.boxes = cs.boxes[:0]
	cs.planes = cs.planes[:0]

	for _, e := range entities {
		var t *Transform
		var rb *RigidBody
		var sph *ColliderSphere
		var box *ColliderAABB
		var plane *ColliderPlane

		for _, c := range e.Components {
			switch comp := c.(type) {
			case *Transform:
				t = comp
			case *RigidBody:
				rb = comp
			case *ColliderSphere:
				sph = comp
			case *ColliderAABB:
				box = comp
			case *ColliderPlane:
				plane = comp
			}
		}

		if sph != nil && t != nil && rb != nil {
			cs.spheres = append(cs.spheres, sphereBody{t: t, r: rb, c: sph})
		}
		if box != nil && t != nil && rb != nil {
			cs.boxes = append(cs.boxes, boxBody{t: t, r: rb, c: box})
		}
		if plane != nil {
			cs.planes = append(cs.planes, planeBody{c: plane})
		}
	}

	cs.handleSpherePlane()
	cs.handleSphereSphere()
	cs.handleBoxPlane()
	cs.handleBoxBox()
	cs.handleSphereBox()
}

func (cs *CollisionSystem) handleSpherePlane() {
	for _, s := range cs.spheres {
		for _, p := range cs.planes {
			if !canCollide(s.c.Layer, s.c.Mask, p.c.Layer, p.c.Mask) {
				continue
			}

			if s.t.Position[1]-s.c.Radius < p.c.Y {
				s.t.Position[1] = p.c.Y + s.c.Radius
				s.r.Vel[1] = -s.r.Vel[1] * 0.5
				s.r.Vel[0] *= 0.9
				s.r.Vel[2] *= 0.9
			}
		}
	}
}

func (cs *CollisionSystem) handleSphereSphere() {
	spheres := cs.spheres
	for i := 0; i < len(spheres); i++ {
		a := &spheres[i]
		for j := i + 1; j < len(spheres); j++ {

			b := &spheres[j]
			if !canCollide(a.c.Layer, a.c.Mask, b.c.Layer, b.c.Mask) {
				continue
			}

			dx := b.t.Position[0] - a.t.Position[0]
			dy := b.t.Position[1] - a.t.Position[1]
			dz := b.t.Position[2] - a.t.Position[2]
			distSq := dx*dx + dy*dy + dz*dz
			radiusSum := a.c.Radius + b.c.Radius
			if distSq < radiusSum*radiusSum {
				dist := float32(math.Sqrt(float64(distSq)))
				if dist == 0 {
					dist = 0.0001
				}
				nx, ny, nz := dx/dist, dy/dist, dz/dist
				penetration := radiusSum - dist

				a.t.Position[0] -= nx * penetration * 0.5
				a.t.Position[1] -= ny * penetration * 0.5
				a.t.Position[2] -= nz * penetration * 0.5
				b.t.Position[0] += nx * penetration * 0.5
				b.t.Position[1] += ny * penetration * 0.5
				b.t.Position[2] += nz * penetration * 0.5

				va := a.r.Vel[0]*nx + a.r.Vel[1]*ny + a.r.Vel[2]*nz
				vb := b.r.Vel[0]*nx + b.r.Vel[1]*ny + b.r.Vel[2]*nz
				a.r.Vel[0] += (vb - va) * nx
				a.r.Vel[1] += (vb - va) * ny
				a.r.Vel[2] += (vb - va) * nz
				b.r.Vel[0] += (va - vb) * nx
				b.r.Vel[1] += (va - vb) * ny
				b.r.Vel[2] += (va - vb) * nz
			}
		}
	}
}

func (cs *CollisionSystem) handleBoxPlane() {
	for _, b := range cs.boxes {
		for _, p := range cs.planes {
			if !canCollide(b.c.Layer, b.c.Mask, p.c.Layer, p.c.Mask) {
				continue
			}

			if b.t.Position[1]-b.c.HalfExtents[1] < p.c.Y {
				b.t.Position[1] = p.c.Y + b.c.HalfExtents[1]
				b.r.Vel[1] = -b.r.Vel[1] * 0.5
				b.r.Vel[0] *= 0.8
				b.r.Vel[2] *= 0.8
			}
		}
	}
}

func (cs *CollisionSystem) handleBoxBox() {
	boxes := cs.boxes
	for i := 0; i < len(boxes); i++ {
		a := &boxes[i]
		for j := i + 1; j < len(boxes); j++ {
			b := &boxes[j]
			if !canCollide(a.c.Layer, a.c.Mask, b.c.Layer, b.c.Mask) {
				continue
			}

			minA := [3]float32{
				a.t.Position[0] - a.c.HalfExtents[0],
				a.t.Position[1] - a.c.HalfExtents[1],
				a.t.Position[2] - a.c.HalfExtents[2],
			}
			maxA := [3]float32{
				a.t.Position[0] + a.c.HalfExtents[0],
				a.t.Position[1] + a.c.HalfExtents[1],
				a.t.Position[2] + a.c.HalfExtents[2],
			}
			minB := [3]float32{
				b.t.Position[0] - b.c.HalfExtents[0],
				b.t.Position[1] - b.c.HalfExtents[1],
				b.t.Position[2] - b.c.HalfExtents[2],
			}
			maxB := [3]float32{
				b.t.Position[0] + b.c.HalfExtents[0],
				b.t.Position[1] + b.c.HalfExtents[1],
				b.t.Position[2] + b.c.HalfExtents[2],
			}

			overlapX := minA[0] <= maxB[0] && maxA[0] >= minB[0]
			overlapY := minA[1] <= maxB[1] && maxA[1] >= minB[1]
			overlapZ := minA[2] <= maxB[2] && maxA[2] >= minB[2]

			if overlapX && overlapY && overlapZ {
				penX := min(maxA[0]-minB[0], maxB[0]-minA[0])
				penY := min(maxA[1]-minB[1], maxB[1]-minA[1])
				penZ := min(maxA[2]-minB[2], maxB[2]-minA[2])

				if penX < penY && penX < penZ {
					if a.t.Position[0] < b.t.Position[0] {
						a.t.Position[0] -= penX / 2
						b.t.Position[0] += penX / 2
					} else {
						a.t.Position[0] += penX / 2
						b.t.Position[0] -= penX / 2
					}
					a.r.Vel[0] = -a.r.Vel[0] * 0.5
					b.r.Vel[0] = -b.r.Vel[0] * 0.5
				} else if penY < penZ {
					if a.t.Position[1] < b.t.Position[1] {
						a.t.Position[1] -= penY / 2
						b.t.Position[1] += penY / 2
					} else {
						a.t.Position[1] += penY / 2
						b.t.Position[1] -= penY / 2
					}
					a.r.Vel[1] = -a.r.Vel[1] * 0.5
					b.r.Vel[1] = -b.r.Vel[1] * 0.5
				} else {
					if a.t.Position[2] < b.t.Position[2] {
						a.t.Position[2] -= penZ / 2
						b.t.Position[2] += penZ / 2
					} else {
						a.t.Position[2] += penZ / 2
						b.t.Position[2] -= penZ / 2
					}
					a.r.Vel[2] = -a.r.Vel[2] * 0.5
					b.r.Vel[2] = -b.r.Vel[2] * 0.5
				}
			}
		}
	}
}

func (cs *CollisionSystem) handleSphereBox() {
	for _, s := range cs.spheres {
		for _, b := range cs.boxes {
			if !canCollide(s.c.Layer, s.c.Mask, b.c.Layer, b.c.Mask) {
				continue
			}

			closest := [3]float32{
				clamp(s.t.Position[0], b.t.Position[0]-b.c.HalfExtents[0], b.t.Position[0]+b.c.HalfExtents[0]),
				clamp(s.t.Position[1], b.t.Position[1]-b.c.HalfExtents[1], b.t.Position[1]+b.c.HalfExtents[1]),
				clamp(s.t.Position[2], b.t.Position[2]-b.c.HalfExtents[2], b.t.Position[2]+b.c.HalfExtents[2]),
			}

			dx := s.t.Position[0] - closest[0]
			dy := s.t.Position[1] - closest[1]
			dz := s.t.Position[2] - closest[2]
			distSq := dx*dx + dy*dy + dz*dz

			if distSq < s.c.Radius*s.c.Radius {
				dist := float32(math.Sqrt(float64(distSq)))
				if dist == 0 {
					dist = 0.0001
				}
				nx, ny, nz := dx/dist, dy/dist, dz/dist
				penetration := s.c.Radius - dist

				s.t.Position[0] += nx * penetration
				s.t.Position[1] += ny * penetration
				s.t.Position[2] += nz * penetration

				dot := s.r.Vel[0]*nx + s.r.Vel[1]*ny + s.r.Vel[2]*nz
				s.r.Vel[0] -= 2 * dot * nx
				s.r.Vel[1] -= 2 * dot * ny
				s.r.Vel[2] -= 2 * dot * nz

				s.r.Vel[0] *= 0.5
				s.r.Vel[1] *= 0.5
				s.r.Vel[2] *= 0.5
			}
		}
	}
}

func (c *ColliderAABB) EditorName() string { return "ColliderAABB" }

func (c *ColliderAABB) EditorFields() map[string]any {
	return map[string]any{
		"HalfExtentsX": c.HalfExtents[0],
		"HalfExtentsY": c.HalfExtents[1],
		"HalfExtentsZ": c.HalfExtents[2],
		"Layer":        c.Layer,
		"Mask":         c.Mask,
	}
}

func (c *ColliderAABB) SetEditorField(name string, value any) {
	switch name {
	case "HalfExtentsX":
		c.HalfExtents[0] = toFloat32(value)
	case "HalfExtentsY":
		c.HalfExtents[1] = toFloat32(value)
	case "HalfExtentsZ":
		c.HalfExtents[2] = toFloat32(value)
	case "Layer":
		c.Layer = toInt(value)
	case "Mask":
		c.Mask = uint32(toInt(value))

	}
}
func (c *ColliderPlane) EditorName() string { return "ColliderPlane" }

func (c *ColliderPlane) EditorFields() map[string]any {
	return map[string]any{
		"Y":     c.Y,
		"Layer": c.Layer,
		"Mask":  c.Mask,
	}
}

func (c *ColliderPlane) SetEditorField(name string, value any) {
	switch name {
	case "Y":
		c.Y = toFloat32(value)
	case "Layer":
		c.Layer = toInt(value)
	case "Mask":
		c.Mask = uint32(toInt(value))
	}

}

func (c *ColliderSphere) EditorName() string { return "ColliderSphere" }

func (c *ColliderSphere) EditorFields() map[string]any {
	return map[string]any{
		"Radius": c.Radius,
		"Layer":  c.Layer,
		"Mask":   c.Mask,
	}
}

func (c *ColliderSphere) SetEditorField(name string, value any) {
	switch name {
	case "Radius":
		c.Radius = toFloat32(value)
	case "Layer":
		c.Layer = toInt(value)
	case "Mask":
		c.Mask = uint32(toInt(value))
	}

}

func sameLayer(a, b int) bool {
	return a == b
}

func canCollide(aLayer int, aMask uint32, bLayer int, bMask uint32) bool {
	return (aMask&(1<<bLayer)) != 0 && (bMask&(1<<aLayer)) != 0
}
