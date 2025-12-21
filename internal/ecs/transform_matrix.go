package ecs

import "math"

func NewTransformFromMatrix(m [16]float32) *Transform {
	t := &Transform{}
	t.SetFromMatrix(m)
	return t
}

// SetFromMatrix decomposes a TRS matrix into position, rotation, scale
// and sets LocalMatrix = m, WorldMatrix = m.
func (t *Transform) SetFromMatrix(m [16]float32) {
	// Translation
	t.Position = [3]float32{
		m[12],
		m[13],
		m[14],
	}

	// Scale
	sx := float32(math.Sqrt(float64(m[0]*m[0] + m[1]*m[1] + m[2]*m[2])))
	sy := float32(math.Sqrt(float64(m[4]*m[4] + m[5]*m[5] + m[6]*m[6])))
	sz := float32(math.Sqrt(float64(m[8]*m[8] + m[9]*m[9] + m[10]*m[10])))

	t.Scale = [3]float32{sx, sy, sz}

	// Normalize rotation matrix
	r00 := m[0] / sx
	r01 := m[1] / sx
	r02 := m[2] / sx
	r10 := m[4] / sy
	r11 := m[5] / sy
	r12 := m[6] / sy
	r20 := m[8] / sz
	r21 := m[9] / sz
	r22 := m[10] / sz

	trace := r00 + r11 + r22

	var qw, qx, qy, qz float32

	if trace > 0 {
		s := float32(math.Sqrt(float64(trace+1.0))) * 2
		qw = s / 4
		qx = (r21 - r12) / s
		qy = (r02 - r20) / s
		qz = (r10 - r01) / s
	} else if r00 > r11 && r00 > r22 {
		s := float32(math.Sqrt(float64(1.0+r00-r11-r22))) * 2
		qw = (r21 - r12) / s
		qx = s / 4
		qy = (r01 + r10) / s
		qz = (r02 + r20) / s
	} else if r11 > r22 {
		s := float32(math.Sqrt(float64(1.0+r11-r00-r22))) * 2
		qw = (r02 - r20) / s
		qx = (r01 + r10) / s
		qy = s / 4
		qz = (r12 + r21) / s
	} else {
		s := float32(math.Sqrt(float64(1.0+r22-r00-r11))) * 2
		qw = (r10 - r01) / s
		qx = (r02 + r20) / s
		qy = (r12 + r21) / s
		qz = s / 4
	}

	t.Rotation = [4]float32{qx, qy, qz, qw}

	// Store matrices
	t.LocalMatrix = m
	t.WorldMatrix = m
	t.Dirty = false
}

// RecalculateLocal rebuilds LocalMatrix from Position/Rotation/Scale.
func (t *Transform) RecalculateLocal() {
	// Build rotation matrix from quaternion
	qx, qy, qz, qw := t.Rotation[0], t.Rotation[1], t.Rotation[2], t.Rotation[3]
	sx, sy, sz := t.Scale[0], t.Scale[1], t.Scale[2]
	tx, ty, tz := t.Position[0], t.Position[1], t.Position[2]

	xx := qx * qx
	yy := qy * qy
	zz := qz * qz
	xy := qx * qy
	xz := qx * qz
	yz := qy * qz
	wx := qw * qx
	wy := qw * qy
	wz := qw * qz

	m := [16]float32{
		1 - 2*(yy+zz), 2 * (xy - wz), 2 * (xz + wy), 0,
		2 * (xy + wz), 1 - 2*(xx+zz), 2 * (yz - wx), 0,
		2 * (xz - wy), 2 * (yz + wx), 1 - 2*(xx+yy), 0,
		0, 0, 0, 1,
	}

	// Apply scale
	m[0] *= sx
	m[1] *= sx
	m[2] *= sx
	m[4] *= sy
	m[5] *= sy
	m[6] *= sy
	m[8] *= sz
	m[9] *= sz
	m[10] *= sz

	// Apply translation
	m[12] = tx
	m[13] = ty
	m[14] = tz

	t.LocalMatrix = m
	t.Dirty = false
}

func IdentityMatrix() [16]float32 {
	return [16]float32{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

func MulMat4(a, b [16]float32) [16]float32 {
	var r [16]float32
	for row := 0; row < 4; row++ {
		for col := 0; col < 4; col++ {
			r[row*4+col] =
				a[row*4+0]*b[0*4+col] +
					a[row*4+1]*b[1*4+col] +
					a[row*4+2]*b[2*4+col] +
					a[row*4+3]*b[3*4+col]
		}
	}
	return r
}
