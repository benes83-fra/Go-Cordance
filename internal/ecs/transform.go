package ecs

import "math"

// Transform is a simple component that holds position/rotation/scale.
// It implements Component so it can be updated if needed (e.g., animations).

type Transform struct {
	Position [3]float32
	Rotation [4]float32 // quaternion (x, y, z, w)
	Scale    [3]float32

	LocalMatrix [16]float32
	WorldMatrix [16]float32

	Dirty bool
}

func NewTransform(pos [3]float32) *Transform {
	t := &Transform{
		Position: pos,
		Rotation: [4]float32{0, 0, 0, 1},
		Scale:    [3]float32{1, 1, 1},
		Dirty:    true,
	}
	t.RecalculateLocal()
	t.WorldMatrix = t.LocalMatrix
	return t
}

// Update is a no-op by default. You can embed or extend Transform to animate it.
func (t *Transform) Update(dt float32) {
	_ = dt
}

// Translate adds a vector to the position.
func (t *Transform) Translate(x, y, z float32) {
	t.Position[0] += x
	t.Position[1] += y
	t.Position[2] += z
}

// Rotate adds Euler rotation (radians).
func (t *Transform) Rotate(rx, ry, rz float32) {
	t.Rotation[0] += rx
	t.Rotation[1] += ry
	t.Rotation[2] += rz
}

// UniformScale multiplies the scale uniformly.
func (t *Transform) UniformScale(s float32) {
	t.Scale[0] *= s
	t.Scale[1] *= s
	t.Scale[2] *= s
}

// SetRotationDegrees sets rotation from degrees (convenience).
func (t *Transform) SetRotationDegrees(dx, dy, dz float32) {
	const degToRad = math.Pi / 180.0
	t.Rotation[0] = dx * degToRad
	t.Rotation[1] = dy * degToRad
	t.Rotation[2] = dz * degToRad
}
func GetTransform(e *Entity) *Transform {
	for _, c := range e.Components {
		if tr, ok := c.(*Transform); ok {
			return tr
		}
	}
	return nil
}
func (t *Transform) Forward() [3]float32 {
	q := t.Rotation
	x, y, z, w := q[0], q[1], q[2], q[3]

	return [3]float32{
		2 * (x*z + w*y),
		2 * (y*z - w*x),
		1 - 2*(x*x+y*y),
	}
}

func (t *Transform) LookAt(target [3]float32, up [3]float32) {
	// Compute forward vector
	fx := target[0] - t.Position[0]
	fy := target[1] - t.Position[1]
	fz := target[2] - t.Position[2]

	// Normalize forward
	fl := float32(math.Sqrt(float64(fx*fx + fy*fy + fz*fz)))
	if fl > 0 {
		fx /= fl
		fy /= fl
		fz /= fl
	}

	// Right = normalize(cross(up, forward))
	rx := up[1]*fz - up[2]*fy
	ry := up[2]*fx - up[0]*fz
	rz := up[0]*fy - up[1]*fx

	rl := float32(math.Sqrt(float64(rx*rx + ry*ry + rz*rz)))
	if rl > 0 {
		rx /= rl
		ry /= rl
		rz /= rl
	}

	// Recompute up = cross(forward, right)
	ux := fy*rz - fz*ry
	uy := fz*rx - fx*rz
	uz := fx*ry - fy*rx

	// Convert rotation matrix to quaternion
	m00, m01, m02 := rx, ux, -fx
	m10, m11, m12 := ry, uy, -fy
	m20, m21, m22 := rz, uz, -fz

	trace := m00 + m11 + m22

	var qx, qy, qz, qw float32

	if trace > 0 {
		s := float32(math.Sqrt(float64(trace+1.0))) * 2
		qw = 0.25 * s
		qx = (m21 - m12) / s
		qy = (m02 - m20) / s
		qz = (m10 - m01) / s
	} else if m00 > m11 && m00 > m22 {
		s := float32(math.Sqrt(float64(1.0+m00-m11-m22))) * 2
		qw = (m21 - m12) / s
		qx = 0.25 * s
		qy = (m01 + m10) / s
		qz = (m02 + m20) / s
	} else if m11 > m22 {
		s := float32(math.Sqrt(float64(1.0+m11-m00-m22))) * 2
		qw = (m02 - m20) / s
		qx = (m01 + m10) / s
		qy = 0.25 * s
		qz = (m12 + m21) / s
	} else {
		s := float32(math.Sqrt(float64(1.0+m22-m00-m11))) * 2
		qw = (m10 - m01) / s
		qx = (m02 + m20) / s
		qy = (m12 + m21) / s
		qz = 0.25 * s
	}

	t.Rotation = [4]float32{qx, qy, qz, qw}
	t.Dirty = true
}
func (t *Transform) Right() [3]float32 {
	f := t.Forward()
	return [3]float32{f[2], 0, -f[0]}
}

func (t *Transform) Up() [3]float32 {
	q := t.Rotation
	x, y, z, w := q[0], q[1], q[2], q[3]

	return [3]float32{
		2 * (x*y - w*z),
		1 - 2*(x*x+z*z),
		2 * (y*z + w*x),
	}
}

func (t *Transform) SetForward(forward [3]float32, up [3]float32) {
	// Normalize forward
	fx, fy, fz := forward[0], forward[1], forward[2]
	fl := float32(math.Sqrt(float64(fx*fx + fy*fy + fz*fz)))
	if fl > 0 {
		fx /= fl
		fy /= fl
		fz /= fl
	}

	// Right = normalize(cross(up, forward))
	rx := up[1]*fz - up[2]*fy
	ry := up[2]*fx - up[0]*fz
	rz := up[0]*fy - up[1]*fx

	rl := float32(math.Sqrt(float64(rx*rx + ry*ry + rz*rz)))
	if rl > 0 {
		rx /= rl
		ry /= rl
		rz /= rl
	}

	// Recompute up = cross(forward, right)
	ux := fy*rz - fz*ry
	uy := fz*rx - fx*rz
	uz := fx*ry - fy*rx

	// Convert rotation matrix to quaternion
	m00, m01, m02 := rx, ux, -fx
	m10, m11, m12 := ry, uy, -fy
	m20, m21, m22 := rz, uz, -fz

	trace := m00 + m11 + m22

	var qx, qy, qz, qw float32

	if trace > 0 {
		s := float32(math.Sqrt(float64(trace+1.0))) * 2
		qw = 0.25 * s
		qx = (m21 - m12) / s
		qy = (m02 - m20) / s
		qz = (m10 - m01) / s
	} else if m00 > m11 && m00 > m22 {
		s := float32(math.Sqrt(float64(1.0+m00-m11-m22))) * 2
		qw = (m21 - m12) / s
		qx = 0.25 * s
		qy = (m01 + m10) / s
		qz = (m02 + m20) / s
	} else if m11 > m22 {
		s := float32(math.Sqrt(float64(1.0+m11-m00-m22))) * 2
		qw = (m02 - m20) / s
		qx = (m01 + m10) / s
		qy = 0.25 * s
		qz = (m12 + m21) / s
	} else {
		s := float32(math.Sqrt(float64(1.0+m22-m00-m11))) * 2
		qw = (m10 - m01) / s
		qx = (m02 + m20) / s
		qy = (m12 + m21) / s
		qz = 0.25 * s
	}

	t.Rotation = [4]float32{qx, qy, qz, qw}
	t.Dirty = true
}

func (t *Transform) EditorName() string { return "Transform" }

func (t *Transform) EditorFields() map[string]any {
	return map[string]any{
		"Position": t.Position,
		"Rotation": t.Rotation,
		"Scale":    t.Scale,
	}
}

func (t *Transform) SetEditorField(name string, value any) {
	switch name {
	case "Position":
		t.Position = value.([3]float32)
	case "Rotation":
		t.Rotation = value.([4]float32)
	case "Scale":
		t.Scale = value.([3]float32)
	}
	t.Dirty = true
}
