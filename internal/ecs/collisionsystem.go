package ecs

import (
	"math"
)

type Contact struct {
	A, B     *RigidBody
	Normal   [3]float32
	Friction float32
	Lifetime int
}

// Collider is a component with a simple bounding sphere.
type Collider interface {
	ColliderType() string
}

// --- Sphere Collider ---
type ColliderSphere struct {
	Radius      float32
	Layer       int
	Mask        uint32
	Restitution float32
	Friction    float32
}

func NewColliderSphere(radius float32) *ColliderSphere {
	return &ColliderSphere{
		Radius:      radius,
		Mask:        0xFFFFFFFF,
		Restitution: 0.5,
		Friction:    0.8,
	}
}

func (c *ColliderSphere) ColliderType() string { return "sphere" }

// --- Plane Collider ---
type ColliderPlane struct {
	Y           float32 // horizontal plane at this Y level
	Layer       int
	Mask        uint32
	Restitution float32
	Friction    float32
}

func NewColliderPlane(y float32) *ColliderPlane {
	return &ColliderPlane{
		Y:           y,
		Mask:        0xFFFFFFFF,
		Restitution: 0.5,
		Friction:    0.8,
	}
}

// AABB collider (Axis-Aligned Bounding Box)
type ColliderAABB struct {
	HalfExtents [3]float32 // half-widths in x,y,z
	Layer       int
	Mask        uint32
	Restitution float32
	Friction    float32
}

func NewColliderAABB(halfExtents [3]float32) *ColliderAABB {
	return &ColliderAABB{
		HalfExtents: halfExtents,
		Mask:        0xFFFFFFFF,
		Restitution: 0.5,
		Friction:    0.8,
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
	spheres  []sphereBody
	boxes    []boxBody
	planes   []planeBody
	contacts []Contact
}

func NewCollisionSystem() *CollisionSystem {
	return &CollisionSystem{
		contacts: make([]Contact, 0, 128),
	}
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
	// Decay old contacts
	for i := 0; i < len(cs.contacts); {
		cs.contacts[i].Lifetime--
		if cs.contacts[i].Lifetime <= 0 {
			cs.contacts[i] = cs.contacts[len(cs.contacts)-1]
			cs.contacts = cs.contacts[:len(cs.contacts)-1]
		} else {
			i++
		}
	}

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
	//cs.applyFriction()
	cs.solveContacts()
}

func (cs *CollisionSystem) handleSpherePlane() {
	for _, s := range cs.spheres {
		for _, p := range cs.planes {
			if !canCollide(s.c.Layer, s.c.Mask, p.c.Layer, p.c.Mask) {
				continue
			}
			restitution := min(s.c.Restitution, p.c.Restitution)
			friction := min(s.c.Friction, p.c.Friction)
			if s.t.Position[1]-s.c.Radius < p.c.Y {
				s.t.Position[1] = p.c.Y + s.c.Radius
				s.r.Vel[1] = -s.r.Vel[1] * restitution
				// s.r.Vel[0] *= friction
				// s.r.Vel[2] *= friction
				cs.addContact(s.r, nil, [3]float32{0, 1, 0}, friction)

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
			restitution := min(a.c.Restitution, b.c.Restitution)
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
				vrel := vb - va
				if vrel < 0 { // only resolve if approaching

					impulse := -(1 + restitution) * vrel * 0.5 // equal mass

					a.r.Vel[0] += impulse * nx
					a.r.Vel[1] += impulse * ny
					a.r.Vel[2] += impulse * nz

					b.r.Vel[0] -= impulse * nx
					b.r.Vel[1] -= impulse * ny
					b.r.Vel[2] -= impulse * nz
				}
				cs.addContact(a.r, b.r, [3]float32{nx, ny, nz}, min(a.c.Friction, b.c.Friction))

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
			restitution := min(b.c.Restitution, p.c.Restitution)
			friction := min(b.c.Friction, p.c.Friction)
			if b.t.Position[1]-b.c.HalfExtents[1] < p.c.Y {
				b.t.Position[1] = p.c.Y + b.c.HalfExtents[1]
				b.r.Vel[1] = -b.r.Vel[1] * restitution
				// b.r.Vel[0] *= friction
				// b.r.Vel[2] *= friction
				cs.addContact(b.r, nil, [3]float32{0, 1, 0}, friction)

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
			restitution := min(a.c.Restitution, b.c.Restitution)
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
				var n [3]float32
				if penX < penY && penX < penZ {
					if a.t.Position[0] < b.t.Position[0] {
						a.t.Position[0] -= penX / 2
						b.t.Position[0] += penX / 2
						n = [3]float32{-1, 0, 0}
					} else {
						a.t.Position[0] += penX / 2
						b.t.Position[0] -= penX / 2
						n = [3]float32{1, 0, 0}
					}
					a.r.Vel[0] = -a.r.Vel[0] * restitution
					b.r.Vel[0] = -b.r.Vel[0] * restitution
				} else if penY < penZ {
					if a.t.Position[1] < b.t.Position[1] {
						a.t.Position[1] -= penY / 2
						b.t.Position[1] += penY / 2
						n = [3]float32{0, -1, 0}
					} else {
						a.t.Position[1] += penY / 2
						b.t.Position[1] -= penY / 2
						n = [3]float32{0, 1, 0}
					}
					a.r.Vel[1] = -a.r.Vel[1] * restitution
					b.r.Vel[1] = -b.r.Vel[1] * restitution
				} else {
					if a.t.Position[2] < b.t.Position[2] {
						a.t.Position[2] -= penZ / 2
						b.t.Position[2] += penZ / 2
						n = [3]float32{0, 0, -1}
					} else {
						a.t.Position[2] += penZ / 2
						b.t.Position[2] -= penZ / 2
						n = [3]float32{0, 0, 1}
					}
					a.r.Vel[2] = -a.r.Vel[2] * restitution
					b.r.Vel[2] = -b.r.Vel[2] * restitution
				}
				cs.addContact(a.r, b.r, n, min(a.c.Friction, b.c.Friction))

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
			restitution := min(s.c.Restitution, b.c.Restitution)
			friction := min(s.c.Friction, b.c.Friction)
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

				s.r.Vel[0] *= restitution
				s.r.Vel[1] *= restitution
				s.r.Vel[2] *= restitution

				// tangential friction
				// s.r.Vel[0] *= friction
				// s.r.Vel[1] *= friction
				// s.r.Vel[2] *= friction
				cs.addContact(s.r, b.r, [3]float32{nx, ny, nz}, friction)

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
		"Restitution":  c.Restitution,
		"Friction":     c.Friction,
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
	case "Restitution":
		c.Restitution = toFloat32(value)
	case "Friction":
		c.Friction = toFloat32(value)

	}
}
func (c *ColliderPlane) EditorName() string { return "ColliderPlane" }

func (c *ColliderPlane) EditorFields() map[string]any {
	return map[string]any{
		"Y":           c.Y,
		"Layer":       c.Layer,
		"Mask":        c.Mask,
		"Restitution": c.Restitution,
		"Friction":    c.Friction,
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
	case "Restitution":
		c.Restitution = toFloat32(value)
	case "Friction":
		c.Friction = toFloat32(value)
	}

}

func (c *ColliderSphere) EditorName() string { return "ColliderSphere" }

func (c *ColliderSphere) EditorFields() map[string]any {
	return map[string]any{
		"Radius":      c.Radius,
		"Layer":       c.Layer,
		"Mask":        c.Mask,
		"Restitution": c.Restitution,
		"Friction":    c.Friction,
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
	case "Restitution":
		c.Restitution = toFloat32(value)
	case "Friction":
		c.Friction = toFloat32(value)
	}

}

func sameLayer(a, b int) bool {
	return a == b
}

func canCollide(aLayer int, aMask uint32, bLayer int, bMask uint32) bool {
	return (aMask&(1<<bLayer)) != 0 && (bMask&(1<<aLayer)) != 0
}

func (cs *CollisionSystem) applyFriction() {
	// Sphere–plane friction
	for _, s := range cs.spheres {
		for _, p := range cs.planes {
			if !canCollide(s.c.Layer, s.c.Mask, p.c.Layer, p.c.Mask) {
				continue
			}
			if s.t.Position[1]-s.c.Radius < p.c.Y {
				n := [3]float32{0, 1, 0} // plane normal
				friction := min(s.c.Friction, p.c.Friction)

				// decompose velocity
				vn := dot3(s.r.Vel, n)
				v_n := mul3(n, vn)
				v_t := sub3(s.r.Vel, v_n)

				// apply friction
				v_t = mul3(v_t, friction)

				// recombine
				s.r.Vel = add3(v_n, v_t)
			}
		}
	}

	// Sphere–sphere friction
	for i := 0; i < len(cs.spheres); i++ {
		a := &cs.spheres[i]
		for j := i + 1; j < len(cs.spheres); j++ {
			b := &cs.spheres[j]
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
				n := [3]float32{dx / dist, dy / dist, dz / dist}
				friction := min(a.c.Friction, b.c.Friction)

				// apply friction to A
				vnA := dot3(a.r.Vel, n)
				v_nA := mul3(n, vnA)
				v_tA := sub3(a.r.Vel, v_nA)
				v_tA = mul3(v_tA, friction)
				a.r.Vel = add3(v_nA, v_tA)

				// apply friction to B
				vnB := dot3(b.r.Vel, n)
				v_nB := mul3(n, vnB)
				v_tB := sub3(b.r.Vel, v_nB)
				v_tB = mul3(v_tB, friction)
				b.r.Vel = add3(v_nB, v_tB)
			}
		}
	}

	// Sphere–AABB friction
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
				n := [3]float32{dx / dist, dy / dist, dz / dist}
				friction := min(s.c.Friction, b.c.Friction)

				// apply friction to sphere
				vn := dot3(s.r.Vel, n)
				v_n := mul3(n, vn)
				v_t := sub3(s.r.Vel, v_n)
				v_t = mul3(v_t, friction)
				s.r.Vel = add3(v_n, v_t)

				// apply friction to box
				vnB := dot3(b.r.Vel, n)
				v_nB := mul3(n, vnB)
				v_tB := sub3(b.r.Vel, v_nB)
				v_tB = mul3(v_tB, friction)
				b.r.Vel = add3(v_nB, v_tB)
			}
		}
	}

	// Box–plane friction
	for _, b := range cs.boxes {
		for _, p := range cs.planes {
			if !canCollide(b.c.Layer, b.c.Mask, p.c.Layer, p.c.Mask) {
				continue
			}
			if b.t.Position[1]-b.c.HalfExtents[1] < p.c.Y {
				n := [3]float32{0, 1, 0}
				friction := min(b.c.Friction, p.c.Friction)

				vn := dot3(b.r.Vel, n)
				v_n := mul3(n, vn)
				v_t := sub3(b.r.Vel, v_n)
				v_t = mul3(v_t, friction)
				b.r.Vel = add3(v_n, v_t)
			}
		}
	}

	// Box–box friction
	for i := 0; i < len(cs.boxes); i++ {
		a := &cs.boxes[i]
		for j := i + 1; j < len(cs.boxes); j++ {
			b := &cs.boxes[j]
			if !canCollide(a.c.Layer, a.c.Mask, b.c.Layer, b.c.Mask) {
				continue
			}

			// same overlap test as handleBoxBox
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
				// pick the smallest axis as normal
				penX := min(maxA[0]-minB[0], maxB[0]-minA[0])
				penY := min(maxA[1]-minB[1], maxB[1]-minA[1])
				penZ := min(maxA[2]-minB[2], maxB[2]-minA[2])

				var n [3]float32
				if penX < penY && penX < penZ {
					if a.t.Position[0] < b.t.Position[0] {
						n = [3]float32{-1, 0, 0}
					} else {
						n = [3]float32{1, 0, 0}
					}
				} else if penY < penZ {
					if a.t.Position[1] < b.t.Position[1] {
						n = [3]float32{0, -1, 0}
					} else {
						n = [3]float32{0, 1, 0}
					}
				} else {
					if a.t.Position[2] < b.t.Position[2] {
						n = [3]float32{0, 0, -1}
					} else {
						n = [3]float32{0, 0, 1}
					}
				}

				friction := min(a.c.Friction, b.c.Friction)

				// apply friction to A
				vnA := dot3(a.r.Vel, n)
				v_nA := mul3(n, vnA)
				v_tA := sub3(a.r.Vel, v_nA)
				v_tA = mul3(v_tA, friction)
				a.r.Vel = add3(v_nA, v_tA)

				// apply friction to B
				vnB := dot3(b.r.Vel, n)
				v_nB := mul3(n, vnB)
				v_tB := sub3(b.r.Vel, v_nB)
				v_tB = mul3(v_tB, friction)
				b.r.Vel = add3(v_nB, v_tB)
			}
		}
	}
}
func dot3(a, b [3]float32) float32 {
	return a[0]*b[0] + a[1]*b[1] + a[2]*b[2]
}

func mul3(a [3]float32, s float32) [3]float32 {
	return [3]float32{a[0] * s, a[1] * s, a[2] * s}
}

func sub3(a, b [3]float32) [3]float32 {
	return [3]float32{a[0] - b[0], a[1] - b[1], a[2] - b[2]}
}

func add3(a, b [3]float32) [3]float32 {
	return [3]float32{a[0] + b[0], a[1] + b[1], a[2] + b[2]}
}

func (cs *CollisionSystem) addContact(a, b *RigidBody, n [3]float32, friction float32) {
	n = normalize3(n)

	for i := range cs.contacts {
		c := &cs.contacts[i]
		if (c.A == a && c.B == b) || (c.A == b && c.B == a) {
			c.Normal = n
			c.Friction = friction
			c.Lifetime++
			return
		}
	}

	cs.contacts = append(cs.contacts, Contact{
		A:        a,
		B:        b,
		Normal:   n,
		Friction: friction,
		Lifetime: 1,
	})
}

func (cs *CollisionSystem) solveContacts() {
	for i := range cs.contacts {
		c := &cs.contacts[i]
		n := c.Normal
		fr := c.Friction

		if c.A != nil {
			applyFrictionToBody(c.A, n, fr)
		}
		if c.B != nil {
			applyFrictionToBody(c.B, n, fr)
		}
	}
}
func applyFrictionToBody(rb *RigidBody, n [3]float32, friction float32) {
	vn := dot3(rb.Vel, n)
	v_n := mul3(n, vn)
	v_t := sub3(rb.Vel, v_n)
	v_t = mul3(v_t, friction)
	rb.Vel = add3(v_n, v_t)
}

func normalize3(v [3]float32) [3]float32 {
	l := float32(math.Sqrt(float64(v[0]*v[0] + v[1]*v[1] + v[2]*v[2])))
	if l == 0 {
		return [3]float32{0, 0, 0}
	}
	return [3]float32{v[0] / l, v[1] / l, v[2] / l}
}
