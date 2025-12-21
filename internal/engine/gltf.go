package engine

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"path/filepath"

	"github.com/go-gl/gl/v4.1-core/gl"
)

// Basic glTF 2.0 structs (only what we need)

type gltfBuffer struct {
	ByteLength int    `json:"byteLength"`
	URI        string `json:"uri"`
}

type gltfBufferView struct {
	Buffer     int `json:"buffer"`
	ByteOffset int `json:"byteOffset"`
	ByteLength int `json:"byteLength"`
	ByteStride int `json:"byteStride"`
	Target     int `json:"target"`
}

type gltfAccessor struct {
	BufferView    int       `json:"bufferView"`
	ByteOffset    int       `json:"byteOffset"`
	ComponentType int       `json:"componentType"`
	Count         int       `json:"count"`
	Type          string    `json:"type"`
	Max           []float32 `json:"max"`
	Min           []float32 `json:"min"`
}

type gltfPrimitive struct {
	Attributes map[string]int `json:"attributes"`
	Indices    int            `json:"indices"`
	Material   int            `json:"material"`
	Mode       int            `json:"mode"`
}

type gltfMesh struct {
	Name       string          `json:"name"`
	Primitives []gltfPrimitive `json:"primitives"`
}

type gltfRoot struct {
	Buffers     []gltfBuffer     `json:"buffers"`
	BufferViews []gltfBufferView `json:"bufferViews"`
	Accessors   []gltfAccessor   `json:"accessors"`
	Meshes      []gltfMesh       `json:"meshes"`
}

// Helpers

func componentByteSize(typ string, comp int) int {
	var csize int
	switch comp {
	case 5123: // UNSIGNED_SHORT
		csize = 2
	case 5125: // UNSIGNED_INT
		csize = 4
	case 5126: // FLOAT
		csize = 4
	default:
		panic(fmt.Sprintf("unsupported component type: %d", comp))
	}

	switch typ {
	case "SCALAR":
		return csize * 1
	case "VEC2":
		return csize * 2
	case "VEC3":
		return csize * 3
	case "VEC4":
		return csize * 4
	default:
		panic(fmt.Sprintf("unsupported accessor type: %s", typ))
	}
}

func bytesToFloat32(b []byte) float32 {
	return math.Float32frombits(
		uint32(b[0]) |
			uint32(b[1])<<8 |
			uint32(b[2])<<16 |
			uint32(b[3])<<24)
}

// RegisterGLTF loads a single-mesh, single-primitive glTF 2.0 file
// and registers an interleaved mesh: pos(3), normal(3), uv(2), tangent(4).
func (mm *MeshManager) RegisterGLTF(id, path string) error {
	baseDir := filepath.Dir(path)

	// Load JSON
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read glTF: %w", err)
	}

	var g gltfRoot
	if err := json.Unmarshal(raw, &g); err != nil {
		return fmt.Errorf("unmarshal glTF: %w", err)
	}

	if len(g.Meshes) == 0 {
		return fmt.Errorf("glTF: no meshes")
	}
	mesh := g.Meshes[0]
	if len(mesh.Primitives) == 0 {
		return fmt.Errorf("glTF: mesh has no primitives")
	}
	prim := mesh.Primitives[0]

	// Load all buffers into memory once
	buffers := make([][]byte, len(g.Buffers))
	for i, b := range g.Buffers {
		binPath := filepath.Join(baseDir, b.URI)
		data, err := ioutil.ReadFile(binPath)
		if err != nil {
			return fmt.Errorf("read buffer %d (%s): %w", i, binPath, err)
		}
		if len(data) < b.ByteLength {
			return fmt.Errorf("buffer %d length mismatch: have %d, expected %d", i, len(data), b.ByteLength)
		}
		buffers[i] = data
	}

	// Helper to get accessor + bufferView + base pointer
	getAccessor := func(idx int) (acc gltfAccessor, bv gltfBufferView, buf []byte, base int, stride int, err error) {
		if idx < 0 || idx >= len(g.Accessors) {
			err = fmt.Errorf("accessor index out of range: %d", idx)
			return
		}
		acc = g.Accessors[idx]
		if acc.BufferView < 0 || acc.BufferView >= len(g.BufferViews) {
			err = fmt.Errorf("bufferView index out of range: %d", acc.BufferView)
			return
		}
		bv = g.BufferViews[acc.BufferView]
		if bv.Buffer < 0 || bv.Buffer >= len(buffers) {
			err = fmt.Errorf("buffer index out of range: %d", bv.Buffer)
			return
		}
		buf = buffers[bv.Buffer]
		elemSize := componentByteSize(acc.Type, acc.ComponentType)
		stride = bv.ByteStride
		if stride == 0 {
			stride = elemSize
		}
		base = bv.ByteOffset + acc.ByteOffset
		end := base + acc.Count*stride
		if end > len(buf) {
			err = fmt.Errorf("accessor %d range out of buffer: end=%d len=%d", idx, end, len(buf))
			return
		}
		return
	}

	// POSITION (required)
	posIndex, ok := prim.Attributes["POSITION"]
	if !ok {
		return fmt.Errorf("glTF: primitive missing POSITION")
	}
	posAcc, _, posBuf, posBase, posStride, err := getAccessor(posIndex)
	if err != nil {
		return err
	}
	if posAcc.Type != "VEC3" || posAcc.ComponentType != 5126 {
		return fmt.Errorf("POSITION must be VEC3 float")
	}
	vertexCount := posAcc.Count

	// NORMAL (required for your lighting)
	norIndex, ok := prim.Attributes["NORMAL"]
	if !ok {
		return fmt.Errorf("glTF: primitive missing NORMAL")
	}
	norAcc, _, norBuf, norBase, norStride, err := getAccessor(norIndex)
	if err != nil {
		return err
	}
	if norAcc.Type != "VEC3" || norAcc.ComponentType != 5126 {
		return fmt.Errorf("NORMAL must be VEC3 float")
	}
	if norAcc.Count != vertexCount {
		return fmt.Errorf("NORMAL count %d != POSITION count %d", norAcc.Count, vertexCount)
	}

	// TEXCOORD_0 (optional)
	var uvBuf []byte
	var uvBase, uvStride int
	hasUV := false
	if uvIndex, ok := prim.Attributes["TEXCOORD_0"]; ok {
		uvAcc, _, buf, base, stride, err := getAccessor(uvIndex)
		if err != nil {
			return err
		}
		if uvAcc.Type != "VEC2" || uvAcc.ComponentType != 5126 {
			return fmt.Errorf("TEXCOORD_0 must be VEC2 float")
		}
		if uvAcc.Count != vertexCount {
			return fmt.Errorf("TEXCOORD_0 count %d != POSITION count %d", uvAcc.Count, vertexCount)
		}
		uvBuf, uvBase, uvStride = buf, base, stride
		hasUV = true
	}

	// TANGENT (optional)
	var tanBuf []byte
	var tanBase, tanStride int
	hasTan := false
	if tanIndex, ok := prim.Attributes["TANGENT"]; ok {
		tanAcc, _, buf, base, stride, err := getAccessor(tanIndex)
		if err != nil {
			return err
		}
		if tanAcc.Type != "VEC4" || tanAcc.ComponentType != 5126 {
			return fmt.Errorf("TANGENT must be VEC4 float")
		}
		if tanAcc.Count != vertexCount {
			return fmt.Errorf("TANGENT count %d != POSITION count %d", tanAcc.Count, vertexCount)
		}
		tanBuf, tanBase, tanStride = buf, base, stride
		hasTan = true
	}

	// INDICES
	idxAcc, idxBV, idxBuf, idxBase, idxStride, err := getAccessor(prim.Indices)
	if err != nil {
		return err
	}
	if idxAcc.Type != "SCALAR" {
		return fmt.Errorf("indices accessor must be SCALAR")
	}
	if idxAcc.ComponentType != 5123 && idxAcc.ComponentType != 5125 {
		return fmt.Errorf("indices must be UNSIGNED_SHORT (5123) or UNSIGNED_INT (5125)")
	}
	_ = idxBV // currently unused, but kept for completeness

	indexCount := idxAcc.Count
	indices := make([]uint32, indexCount)

	switch idxAcc.ComponentType {
	case 5123: // UNSIGNED_SHORT (2 bytes)
		for i := 0; i < indexCount; i++ {
			off := idxBase + i*idxStride
			b := idxBuf[off : off+2]
			indices[i] = uint32(b[0]) | uint32(b[1])<<8
		}
	case 5125: // UNSIGNED_INT (4 bytes)
		for i := 0; i < indexCount; i++ {
			off := idxBase + i*idxStride
			b := idxBuf[off : off+4]
			indices[i] = uint32(b[0]) |
				uint32(b[1])<<8 |
				uint32(b[2])<<16 |
				uint32(b[3])<<24
		}
	}

	// Build interleaved vertex buffer: pos(3), normal(3), uv(2), tangent(4)
	vertices := make([]float32, 0, vertexCount*12)

	for i := 0; i < vertexCount; i++ {
		// POSITION
		pOff := posBase + i*posStride
		px := bytesToFloat32(posBuf[pOff+0:])
		py := bytesToFloat32(posBuf[pOff+4:])
		pz := bytesToFloat32(posBuf[pOff+8:])

		// NORMAL
		nOff := norBase + i*norStride
		nx := bytesToFloat32(norBuf[nOff+0:])
		ny := bytesToFloat32(norBuf[nOff+4:])
		nz := bytesToFloat32(norBuf[nOff+8:])

		// UV
		var u, v float32
		if hasUV {
			uvOff := uvBase + i*uvStride
			u = bytesToFloat32(uvBuf[uvOff+0:])
			v = bytesToFloat32(uvBuf[uvOff+4:])
		} else {
			u, v = 0, 0
		}

		// TANGENT
		tx, ty, tz, tw := float32(1), float32(0), float32(0), float32(1)
		if hasTan {
			tOff := tanBase + i*tanStride
			tx = bytesToFloat32(tanBuf[tOff+0:])
			ty = bytesToFloat32(tanBuf[tOff+4:])
			tz = bytesToFloat32(tanBuf[tOff+8:])
			tw = bytesToFloat32(tanBuf[tOff+12:])
		}

		vertices = append(vertices,
			px, py, pz,
			nx, ny, nz,
			u, v,
			tx, ty, tz, tw,
		)
	}

	// Upload to GL
	var vao, vbo, ebo uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)
	gl.GenBuffers(1, &ebo)

	gl.BindVertexArray(vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	stride := int32(12 * 4)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, stride, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, stride, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(2)
	gl.VertexAttribPointer(2, 2, gl.FLOAT, false, stride, gl.PtrOffset(6*4))
	gl.EnableVertexAttribArray(3)
	gl.VertexAttribPointer(3, 4, gl.FLOAT, false, stride, gl.PtrOffset(8*4))

	gl.BindVertexArray(0)

	mm.vaos[id] = vao
	mm.vbos[id] = vbo
	mm.ebos[id] = ebo
	mm.counts[id] = int32(len(indices))

	return nil
}
