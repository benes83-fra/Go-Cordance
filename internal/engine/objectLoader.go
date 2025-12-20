package engine

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
)

// parseIndex handles v, v/t, v//n, v/t/n and returns 1-based indices (0 means missing)
func parseIndex(s string) (int, int, int, error) {
	parts := strings.Split(s, "/")
	var vi, ti, ni int
	var err error
	if len(parts) >= 1 && parts[0] != "" {
		vi, err = strconv.Atoi(parts[0])
		if err != nil {
			return 0, 0, 0, err
		}
	}
	if len(parts) >= 2 && parts[1] != "" {
		ti, err = strconv.Atoi(parts[1])
		if err != nil {
			return 0, 0, 0, err
		}
	}
	if len(parts) >= 3 && parts[2] != "" {
		ni, err = strconv.Atoi(parts[2])
		if err != nil {
			return 0, 0, 0, err
		}
	}
	return vi, ti, ni, nil
}

// RegisterOBJ loads a basic OBJ and registers an interleaved mesh (pos, normal, uv).
// Supports negative indices and will generate normals if none are present.
func (mm *MeshManager) RegisterOBJ(id, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	var positions [][]float32
	var normals [][]float32
	var uvs [][]float32

	// temporary face storage to allow negative index resolution
	type faceElem struct{ vi, ti, ni int }
	var faces [][]faceElem

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		switch fields[0] {
		case "v":
			if len(fields) < 4 {
				continue
			}
			x, _ := strconv.ParseFloat(fields[1], 32)
			y, _ := strconv.ParseFloat(fields[2], 32)
			z, _ := strconv.ParseFloat(fields[3], 32)
			positions = append(positions, []float32{float32(x), float32(y), float32(z)})
		case "vt":
			if len(fields) < 3 {
				continue
			}
			u, _ := strconv.ParseFloat(fields[1], 32)
			v, _ := strconv.ParseFloat(fields[2], 32)
			uvs = append(uvs, []float32{float32(u), float32(v)})
		case "vn":
			if len(fields) < 4 {
				continue
			}
			nx, _ := strconv.ParseFloat(fields[1], 32)
			ny, _ := strconv.ParseFloat(fields[2], 32)
			nz, _ := strconv.ParseFloat(fields[3], 32)
			normals = append(normals, []float32{float32(nx), float32(ny), float32(nz)})
		case "f":
			if len(fields) < 4 {
				continue
			}
			// triangulate polygon fan
			elems := fields[1:]
			var face []faceElem
			for _, e := range elems {
				vi, ti, ni, err := parseIndex(e)
				if err != nil {
					return fmt.Errorf("parseIndex error: %v", err)
				}
				face = append(face, faceElem{vi, ti, ni})
			}
			// triangulate
			for i := 1; i < len(face)-1; i++ {
				faces = append(faces, []faceElem{face[0], face[i], face[i+1]})
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	// If normals are missing, compute them after we resolve indices
	needNormals := len(normals) == 0

	// Resolve negative indices and build unique vertex list
	vertMap := make(map[string]uint32)
	var vertices []float32
	var indices []uint32
	var nextIndex uint32 = 0

	// helper to resolve 1-based or negative indices to 0-based
	resolve := func(idx, length int) int {
		if idx == 0 {
			return -1
		}
		if idx > 0 {
			return idx - 1
		}
		// negative index: relative to end
		return length + idx
	}

	// If normals missing, create a placeholder normals slice sized to positions (will fill later)
	if needNormals {
		normals = make([][]float32, len(positions))
		for i := range normals {
			normals[i] = []float32{0, 0, 0}
		}
	}

	// Build vertices and indices
	for _, tri := range faces {
		for _, e := range tri {
			vi := resolve(e.vi, len(positions))
			ti := resolve(e.ti, len(uvs))
			ni := resolve(e.ni, len(normals))

			// clamp missing to -1
			if vi < 0 {
				vi = -1
			}
			if ti < 0 {
				ti = -1
			}
			if ni < 0 {
				ni = -1
			}

			key := fmt.Sprintf("%d/%d/%d", vi, ti, ni)
			idx, ok := vertMap[key]
			if !ok {
				var px, py, pz float32
				if vi >= 0 {
					p := positions[vi]
					px, py, pz = p[0], p[1], p[2]
				}
				var tx, ty float32
				if ti >= 0 {
					t := uvs[ti]
					tx, ty = t[0], t[1]
				}
				var nx, ny, nz float32
				if ni >= 0 {
					n := normals[ni]
					nx, ny, nz = n[0], n[1], n[2]
				}
				// append interleaved vertex: pos(3), normal(3), uv(2)
				vertices = append(vertices, px, py, pz, nx, ny, nz, tx, ty)
				idx = nextIndex
				vertMap[key] = idx
				nextIndex++
			}
			indices = append(indices, idx)
		}
	}

	// If normals were missing, compute them per-vertex now
	if needNormals {
		computeNormals(vertices, indices)
		// after computeNormals, vertices' normal slots are filled
	}

	// create GL buffers
	var vao, vbo, ebo uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)
	gl.GenBuffers(1, &ebo)

	gl.BindVertexArray(vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	stride := int32(8 * 4)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, stride, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, stride, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(2)
	gl.VertexAttribPointer(2, 2, gl.FLOAT, false, stride, gl.PtrOffset(6*4))

	gl.BindVertexArray(0)

	mm.vaos[id] = vao
	mm.vbos[id] = vbo
	mm.ebos[id] = ebo
	mm.counts[id] = int32(len(indices))

	return nil
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
		nx := vertices[i*stride+3]
		ny := vertices[i*stride+4]
		nz := vertices[i*stride+5]

		// Gram-Schmidt orthogonalize tangent with normal
		tx := tanX[i]
		ty := tanY[i]
		tz := tanZ[i]

		// t = normalize( t - n * dot(n, t) )
		dotNT := nx*tx + ny*ty + nz*tz
		tx = tx - nx*dotNT
		ty = ty - ny*dotNT
		tz = tz - nz*dotNT

		// normalize t
		magT := tx*tx + ty*ty + tz*tz
		if magT > 0.0 {
			inv := float32(1.0 / math.Sqrt(float64(magT)))
			tx *= inv
			ty *= inv
			tz *= inv
		} else {
			// fallback tangent if degenerate
			tx, ty, tz = 1.0, 0.0, 0.0
		}

		// compute handedness: w = sign( dot( cross(n, t), bitangent ) )
		// cross(n, t)
		cx := ny*tz - nz*ty
		cy := nz*tx - nx*tz
		cz := nx*ty - ny*tx

		bx := bitX[i]
		by := bitY[i]
		bz := bitZ[i]

		handed := cx*bx + cy*by + cz*bz
		w := float32(1.0)
		if handed < 0.0 {
			w = -1.0
		}

		// write tangent vec4 into vertices (offset 8..11)
		vertices[i*stride+8] = tx
		vertices[i*stride+9] = ty
		vertices[i*stride+10] = tz
		vertices[i*stride+11] = w
	}
}
