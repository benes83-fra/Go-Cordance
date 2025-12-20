package engine

import "math"

// computeNormals fills normal components in an interleaved vertex array
// vertices layout: pos(3), normal(3), uv(2) => stride 8 floats
func computeNormals(vertices []float32, indices []uint32) {
	stride := 8
	vertCount := len(vertices) / stride

	// zero normals
	nx := make([]float32, vertCount)
	ny := make([]float32, vertCount)
	nz := make([]float32, vertCount)

	// accumulate face normals
	for i := 0; i < len(indices); i += 3 {
		i0 := int(indices[i+0])
		i1 := int(indices[i+1])
		i2 := int(indices[i+2])

		// positions
		p0x := vertices[i0*stride+0]
		p0y := vertices[i0*stride+1]
		p0z := vertices[i0*stride+2]

		p1x := vertices[i1*stride+0]
		p1y := vertices[i1*stride+1]
		p1z := vertices[i1*stride+2]

		p2x := vertices[i2*stride+0]
		p2y := vertices[i2*stride+1]
		p2z := vertices[i2*stride+2]

		// edges
		ux := p1x - p0x
		uy := p1y - p0y
		uz := p1z - p0z

		vx := p2x - p0x
		vy := p2y - p0y
		vz := p2z - p0z

		// face normal = cross( u, v )
		fx := uy*vz - uz*vy
		fy := uz*vx - ux*vz
		fz := ux*vy - uy*vx

		// accumulate
		nx[i0] += fx
		ny[i0] += fy
		nz[i0] += fz

		nx[i1] += fx
		ny[i1] += fy
		nz[i1] += fz

		nx[i2] += fx
		ny[i2] += fy
		nz[i2] += fz
	}

	// normalize and write back into vertices
	for i := 0; i < vertCount; i++ {
		fx := nx[i]
		fy := ny[i]
		fz := nz[i]
		// normalize
		len := float32(1.0)
		mag := fx*fx + fy*fy + fz*fz
		if mag > 0.0 {
			len = float32(1.0) / float32(math.Sqrt(float64(mag)))
		}
		fx *= len
		fy *= len
		fz *= len

		vertices[i*stride+3] = fx
		vertices[i*stride+4] = fy
		vertices[i*stride+5] = fz
	}
}
