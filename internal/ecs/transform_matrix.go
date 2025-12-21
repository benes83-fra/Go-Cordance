package ecs

import "math"

// NewTransformFromMatrix creates a Transform from a 4x4 matrix.
func NewTransformFromMatrix(m [16]float32) *Transform {
	t := &Transform{}
	t.SetFromMatrix(m)
	return t
}

// SetFromMatrix decomposes a TRS matrix into position, rotation, scale.
func (t *Transform) SetFromMatrix(m [16]float32) {
	// Extract translation
	t.Position = [3]float32{
		m[12],
		m[13],
		m[14],
	}

	// Extract scale
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

	// Convert rotation matrix → quaternion
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
}
