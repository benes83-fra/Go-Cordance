package gltf

import (
	"fmt"
	"go-engine/Go-Cordance/internal/ecs"
	"go-engine/Go-Cordance/internal/engine"
	"log"
	"math"
)

// ExtractGLTFSkins builds Skin components keyed by meshID ("MeshName/primitiveIndex").
func ExtractGLTFSkins(path string) (map[string]*ecs.Skin, error) {
	g, buffers, err := engine.LoadGLTFOrGLB(path)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*ecs.Skin)

	// Pre-decode all skins' inverse bind matrices
	type skinData struct {
		joints   []int
		ibm      [][16]float32
		skeleton int
	}
	skins := make([]skinData, len(g.Skins))
	nodeWorlds := computeNodeWorlds(g)
	for si, s := range g.Skins {
		data := skinData{
			joints:   append([]int(nil), s.Joints...),
			skeleton: s.Skeleton,
		}

		if s.InverseBindMatrices >= 0 {
			acc, err := engine.GetAccessor(g, buffers, s.InverseBindMatrices)
			if err != nil {
				return nil, fmt.Errorf("skin %d inverseBindMatrices: %w", si, err)
			}
			// basic type checks
			if acc.Acc.Type != "MAT4" || acc.Acc.ComponentType != 5126 {
				return nil, fmt.Errorf("skin %d inverseBindMatrices: expected FLOAT MAT4, got type=%s comp=%d", si, acc.Acc.Type, acc.Acc.ComponentType)
			}

			// stride and count sanity
			expectedStride := 16 * 4
			if acc.Stride != 0 && acc.Stride != expectedStride {
				log.Printf("Warning: inverseBindMatrices stride=%d (expected %d). Will still attempt to read using stride.", acc.Stride, expectedStride)
			}
			if acc.Acc.Count != len(s.Joints) {
				log.Printf("Warning: inverseBindMatrices count=%d but skin.joints=%d", acc.Acc.Count, len(s.Joints))
			}

			log.Printf("IBM accessor: count=%d stride=%d base=%d bufLen=%d", acc.Acc.Count, acc.Stride, acc.Base, len(acc.Buf))

			// bounds check
			if acc.Base < 0 || acc.Base >= len(acc.Buf) {
				return nil, fmt.Errorf("inverseBindMatrices accessor base out of range: base=%d bufLen=%d", acc.Base, len(acc.Buf))
			}
			if acc.Stride == 0 {
				acc.Stride = expectedStride
			}

			if acc.Base+acc.Acc.Count*acc.Stride > len(acc.Buf) {
				log.Printf("Warning: accessor claims %d matrices but buffer length may be insufficient (end=%d bufLen=%d). Truncating to available matrices.",
					acc.Acc.Count, acc.Base+acc.Acc.Count*acc.Stride, len(acc.Buf))
				// clamp count to available
				maxCount := (len(acc.Buf) - acc.Base) / acc.Stride
				if maxCount < 0 {
					maxCount = 0
				}
				acc.Acc.Count = maxCount
			}

			for i := 0; i < acc.Acc.Count; i++ {
				off := acc.Base + i*acc.Stride
				var m [16]float32
				for c := 0; c < 16; c++ {
					m[c] = engine.BytesToFloat32(acc.Buf[off+4*c:])
				}
				// Trust the glTF data. Do NOT transpose. Do NOT recompute.
				data.ibm = append(data.ibm, m)
			}
		}

		skins[si] = data
	}

	// Map skins to meshIDs via nodes that reference them
	for nodeIndex, n := range g.Nodes {
		if n.Skin < 0 || n.Skin >= len(g.Skins) {
			continue
		}
		if n.Mesh < 0 || n.Mesh >= len(g.Meshes) {
			continue
		}

		mesh := g.Meshes[n.Mesh]
		meshName := mesh.Name
		if meshName == "" {
			meshName = fmt.Sprintf("mesh_%d", n.Mesh)
		}

		sd := skins[n.Skin]

		// world of the mesh node in bind pose
		meshWorld := nodeWorlds[nodeIndex]
		_ = meshWorld
		// use the skin's IBM as established above (accessor + possible invWorld replacement)
		ibm := sd.ibm
		if len(ibm) != len(sd.joints) {
			log.Printf("Skin %d: joints=%d, ibm=%d (mismatch)", n.Skin, len(sd.joints), len(ibm))
		}

		skinComp := ecs.NewSkin(sd.joints, ibm, sd.skeleton)

		for pi := range mesh.Primitives {
			meshID := fmt.Sprintf("%s/%d", meshName, pi)
			result[meshID] = skinComp
		}
	}

	return result, nil
}

// --- helper diagnostics (place near top of file or in a utils file) ---
func determinant(m [16]float32) float32 {
	// compute determinant of 4x4 (column-major)
	a := m
	// convert to row-major for formula
	r := [16]float32{
		a[0], a[4], a[8], a[12],
		a[1], a[5], a[9], a[13],
		a[2], a[6], a[10], a[14],
		a[3], a[7], a[11], a[15],
	}
	return r[0]*(r[5]*(r[10]*r[15]-r[11]*r[14])-
		r[6]*(r[9]*r[15]-r[11]*r[13])+
		r[7]*(r[9]*r[14]-r[10]*r[13])) -
		r[1]*(r[4]*(r[10]*r[15]-r[11]*r[14])-
			r[6]*(r[8]*r[15]-r[11]*r[12])+
			r[7]*(r[8]*r[14]-r[10]*r[12])) +
		r[2]*(r[4]*(r[9]*r[15]-r[11]*r[13])-
			r[5]*(r[8]*r[15]-r[11]*r[12])+
			r[7]*(r[8]*r[13]-r[9]*r[12])) -
		r[3]*(r[4]*(r[9]*r[14]-r[10]*r[13])-
			r[5]*(r[8]*r[14]-r[10]*r[12])+
			r[6]*(r[8]*r[13]-r[9]*r[12]))
}

func rotationOrthogonalityError(m [16]float32) float32 {
	// measure how far the upper-left 3x3 is from orthonormal (R^T R = I)
	var err float32
	for r := 0; r < 3; r++ {
		for c := 0; c < 3; c++ {
			var s float32
			for k := 0; k < 3; k++ {
				// column-major: element(row,col) = m[col*4 + row]
				s += m[k*4+r] * m[c*4+k]
			}
			if r == c {
				err += float32(math.Abs(float64(s - 1.0)))
			} else {
				err += float32(math.Abs(float64(s)))
			}
		}
	}
	return err
}
func diffMat(a, b [16]float32) [16]float32 {
	var out [16]float32
	for i := 0; i < 16; i++ {
		out[i] = a[i] - b[i]
	}
	return out
}

// by transposing the engine result if necessary.
func computeNodeWorlds(g *engine.GltfRoot) [][16]float32 {
	n := len(g.Nodes)
	worlds := make([][16]float32, n)
	parents := make([]int, n)
	for i := range parents {
		parents[i] = -2
	}

	var markChildren func(parent int)
	markChildren = func(parent int) {
		for _, c := range g.Nodes[parent].Children {
			if c < 0 || c >= n {
				continue
			}
			if parents[c] == -2 {
				parents[c] = parent
				markChildren(c)
			}
		}
	}

	for _, scene := range g.Scenes {
		for _, root := range scene.Nodes {
			if root >= 0 && root < n {
				parents[root] = -1
				markChildren(root)
			}
		}
	}

	for i := 0; i < n; i++ {
		if parents[i] != -2 {
			continue
		}
		found := false
		for p, node := range g.Nodes {
			for _, c := range node.Children {
				if c == i {
					parents[i] = p
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			parents[i] = -1
		}
	}

	var compute func(int) [16]float32
	compute = func(idx int) [16]float32 {
		if worlds[idx] != ([16]float32{}) {
			return worlds[idx]
		}

		local := engine.ComposeNodeTransform(g.Nodes[idx])

		var world [16]float32
		if parents[idx] >= 0 {
			parentWorld := compute(parents[idx])
			world = engine.MulMat4(parentWorld, local) // column-major A*B
		} else {
			world = local
		}

		// NO transpose here
		worlds[idx] = world
		return worlds[idx]
	}

	for i := 0; i < n; i++ {
		_ = compute(i)
	}
	return worlds
}

// mulMat4 multiplies A * B (both column-major) and returns column-major result.
func mulMat4(a, b [16]float32) [16]float32 {
	var out [16]float32
	for r := 0; r < 4; r++ {
		for c := 0; c < 4; c++ {
			var s float32
			for k := 0; k < 4; k++ {
				s += a[k*4+r] * b[c*4+k]
			}
			out[c*4+r] = s
		}
	}
	return out
}

// invertMat4 returns inverse of a 4x4 matrix (column-major). Returns ok=false if non-invertible.
func invertMat4(m [16]float32) (inv [16]float32, ok bool) {
	// Use standard Gauss-Jordan or analytic inverse for 4x4.
	// For brevity use a straightforward adjugate/determinant method.
	// Convert to row-major for easier formula
	r := [16]float32{
		m[0], m[4], m[8], m[12],
		m[1], m[5], m[9], m[13],
		m[2], m[6], m[10], m[14],
		m[3], m[7], m[11], m[15],
	}
	// compute inverse of r (row-major) using standard formula
	// (omitted here for brevity — paste a tested 4x4 inverse implementation)
	// --- BEGIN simple implementation (from common sources) ---
	var invOut [16]float32
	invOut[0] = r[5]*r[10]*r[15] - r[5]*r[11]*r[14] - r[9]*r[6]*r[15] + r[9]*r[7]*r[14] + r[13]*r[6]*r[11] - r[13]*r[7]*r[10]
	invOut[1] = -r[1]*r[10]*r[15] + r[1]*r[11]*r[14] + r[9]*r[2]*r[15] - r[9]*r[3]*r[14] - r[13]*r[2]*r[11] + r[13]*r[3]*r[10]
	invOut[2] = r[1]*r[6]*r[15] - r[1]*r[7]*r[14] - r[5]*r[2]*r[15] + r[5]*r[3]*r[14] + r[13]*r[2]*r[7] - r[13]*r[3]*r[6]
	invOut[3] = -r[1]*r[6]*r[11] + r[1]*r[7]*r[10] + r[5]*r[2]*r[11] - r[5]*r[3]*r[10] - r[9]*r[2]*r[7] + r[9]*r[3]*r[6]
	invOut[4] = -r[4]*r[10]*r[15] + r[4]*r[11]*r[14] + r[8]*r[6]*r[15] - r[8]*r[7]*r[14] - r[12]*r[6]*r[11] + r[12]*r[7]*r[10]
	invOut[5] = r[0]*r[10]*r[15] - r[0]*r[11]*r[14] - r[8]*r[2]*r[15] + r[8]*r[3]*r[14] + r[12]*r[2]*r[11] - r[12]*r[3]*r[10]
	invOut[6] = -r[0]*r[6]*r[15] + r[0]*r[7]*r[14] + r[4]*r[2]*r[15] - r[4]*r[3]*r[14] - r[12]*r[2]*r[7] + r[12]*r[3]*r[6]
	invOut[7] = r[0]*r[6]*r[11] - r[0]*r[7]*r[10] - r[4]*r[2]*r[11] + r[4]*r[3]*r[10] + r[8]*r[2]*r[7] - r[8]*r[3]*r[6]
	invOut[8] = r[4]*r[9]*r[15] - r[4]*r[11]*r[13] - r[8]*r[5]*r[15] + r[8]*r[7]*r[13] + r[12]*r[5]*r[11] - r[12]*r[7]*r[9]
	invOut[9] = -r[0]*r[9]*r[15] + r[0]*r[11]*r[13] + r[8]*r[1]*r[15] - r[8]*r[3]*r[13] - r[12]*r[1]*r[11] + r[12]*r[3]*r[9]
	invOut[10] = r[0]*r[5]*r[15] - r[0]*r[7]*r[13] - r[4]*r[1]*r[15] + r[4]*r[3]*r[13] + r[12]*r[1]*r[7] - r[12]*r[3]*r[5]
	invOut[11] = -r[0]*r[5]*r[11] + r[0]*r[7]*r[9] + r[4]*r[1]*r[11] - r[4]*r[3]*r[9] - r[8]*r[1]*r[7] + r[8]*r[3]*r[5]
	invOut[12] = -r[4]*r[9]*r[14] + r[4]*r[10]*r[13] + r[8]*r[5]*r[14] - r[8]*r[6]*r[13] - r[12]*r[5]*r[10] + r[12]*r[6]*r[9]
	invOut[13] = r[0]*r[9]*r[14] - r[0]*r[10]*r[13] - r[8]*r[1]*r[14] + r[8]*r[2]*r[13] + r[12]*r[1]*r[10] - r[12]*r[2]*r[9]
	invOut[14] = -r[0]*r[5]*r[14] + r[0]*r[6]*r[13] + r[4]*r[1]*r[14] - r[4]*r[2]*r[13] - r[12]*r[1]*r[6] + r[12]*r[2]*r[5]
	invOut[15] = r[0]*r[5]*r[10] - r[0]*r[6]*r[9] - r[4]*r[1]*r[10] + r[4]*r[2]*r[9] + r[8]*r[1]*r[6] - r[8]*r[2]*r[5]

	det := r[0]*invOut[0] + r[1]*invOut[4] + r[2]*invOut[8] + r[3]*invOut[12]
	if det == 0 {
		return inv, false
	}
	invDet := 1.0 / det
	// convert invOut (row-major) back to column-major inv
	for i := 0; i < 16; i++ {
		// invOut is row-major; convert to column-major index
		row := i / 4
		col := i % 4
		// inv column-major index = col*4 + row
		inv[col*4+row] = invOut[i] * float32(invDet)
	}
	return inv, true
	// --- END simple implementation ---
}
func identityError(m [16]float32) float32 {
	var err float32
	for r := 0; r < 4; r++ {
		for c := 0; c < 4; c++ {
			v := m[c*4+r]
			if r == c {
				err += float32(math.Abs(float64(v - 1.0)))
			} else {
				err += float32(math.Abs(float64(v)))
			}
		}
	}
	return err
}
