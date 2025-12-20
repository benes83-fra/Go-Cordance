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

func computeTangents(vertices []float32, indices []uint32) {
	const stride = 12
	vertCount := len(vertices) / stride
	if vertCount == 0 {
		return
	}

	// accumulators for tangent and bitangent
	tanX := make([]float32, vertCount)
	tanY := make([]float32, vertCount)
	tanZ := make([]float32, vertCount)

	bitX := make([]float32, vertCount)
	bitY := make([]float32, vertCount)
	bitZ := make([]float32, vertCount)

	// iterate triangles
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

		// uvs
		u0 := vertices[i0*stride+6]
		v0 := vertices[i0*stride+7]
		u1 := vertices[i1*stride+6]
		v1 := vertices[i1*stride+7]
		u2 := vertices[i2*stride+6]
		v2 := vertices[i2*stride+7]

		// edges
		ex1 := p1x - p0x
		ey1 := p1y - p0y
		ez1 := p1z - p0z

		ex2 := p2x - p0x
		ey2 := p2y - p0y
		ez2 := p2z - p0z

		// uv deltas
		du1 := u1 - u0
		dv1 := v1 - v0
		du2 := u2 - u0
		dv2 := v2 - v0

		denom := du1*dv2 - du2*dv1
		var r float32 = 0.0
		if denom != 0.0 {
			r = 1.0 / denom
		}

		// tangent
		tx := (dv2*ex1 - dv1*ex2) * r
		ty := (dv2*ey1 - dv1*ey2) * r
		tz := (dv2*ez1 - dv1*ez2) * r

		// bitangent
		bx := (-du2*ex1 + du1*ex2) * r
		by := (-du2*ey1 + du1*ey2) * r
		bz := (-du2*ez1 + du1*ez2) * r

		// accumulate
		tanX[i0] += tx
		tanY[i0] += ty
		tanZ[i0] += tz
		tanX[i1] += tx
		tanY[i1] += ty
		tanZ[i1] += tz
		tanX[i2] += tx
		tanY[i2] += ty
		tanZ[i2] += tz

		bitX[i0] += bx
		bitY[i0] += by
		bitZ[i0] += bz
		bitX[i1] += bx
		bitY[i1] += by
		bitZ[i1] += bz
		bitX[i2] += bx
		bitY[i2] += by
		bitZ[i2] += bz
	}

	// orthonormalize tangent against normal and compute handedness
	for i := 0; i < vertCount; i++ {
		// normalize accumulated tangent
		tx := tanX[i]
		ty := tanY[i]
		tz := tanZ[i]
		mag := tx*tx + ty*ty + tz*tz
		if mag > 0.0 {
			inv := float32(1.0 / math.Sqrt(float64(mag)))
			tx *= inv
			ty *= inv
			tz *= inv
		} else {
			tx, ty, tz = 1.0, 0.0, 0.0
		}
		// same for bitangent
		bx := bitX[i]
		by := bitY[i]
		bz := bitZ[i]
		magb := bx*bx + by*by + bz*bz
		if magb > 0.0 {
			invb := float32(1.0 / math.Sqrt(float64(magb)))
			bx *= invb
			by *= invb
			bz *= invb
		} else {
			bx, by, bz = 0.0, 1.0, 0.0
		}

		// Gram-Schmidt tangent
		nx := vertices[i*stride+3]
		ny := vertices[i*stride+4]
		nz := vertices[i*stride+5]

		// t = normalize(t - n * dot(n, t))
		dotNT := nx*tx + ny*ty + nz*tz
		tx = tx - nx*dotNT
		ty = ty - ny*dotNT
		tz = tz - nz*dotNT
		magt := tx*tx + ty*ty + tz*tz
		if magt > 0.0 {
			invt := float32(1.0 / math.Sqrt(float64(magt)))
			tx *= invt
			ty *= invt
			tz *= invt
		} else {
			tx, ty, tz = 1.0, 0.0, 0.0
		}

		// handedness
		cx := ny*tz - nz*ty
		cy := nz*tx - nx*tz
		cz := nx*ty - ny*tx
		handed := cx*bx + cy*by + cz*bz
		w := float32(1.0)
		if handed < 0.0 {
			w = -1.0
		}

		vertices[i*stride+8] = tx
		vertices[i*stride+9] = ty
		vertices[i*stride+10] = tz
		vertices[i*stride+11] = w
	}
}
