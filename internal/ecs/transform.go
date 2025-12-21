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
