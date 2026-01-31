// internal/engine/meshmanager.go
package engine

import (
	"log"
	"math"

	"github.com/go-gl/gl/v4.1-core/gl"
)

// internal/engine/meshmanager.go (add fields)
type MeshManager struct {
	vaos   map[string]uint32
	counts map[string]int32

	// new: track buffers so we can delete them
	// new: track buffers so we can delete them
	vbos map[string]uint32
	ebos map[string]uint32

	// new bookkeeping
	indexTypes   map[string]uint32 // gl.UNSIGNED_INT or gl.UNSIGNED_SHORT
	vertexCounts map[string]int32  // number of vertices (for DrawArrays fallback if needed)
	layoutType   map[string]int    // 8 or 12

}

func NewMeshManager() *MeshManager {
	return &MeshManager{
		vaos:         make(map[string]uint32),
		counts:       make(map[string]int32),
		vbos:         make(map[string]uint32),
		ebos:         make(map[string]uint32),
		indexTypes:   make(map[string]uint32),
		vertexCounts: make(map[string]int32),
		layoutType:   make(map[string]int),
	}
}

func (mm *MeshManager) HasTangents(id string) bool {
	return mm.layoutType[id] == 12
}

func (mm *MeshManager) GetCount(MeshID string) int32 {
	return mm.counts[MeshID]
}

// Triangle with only positions (layout location 0)
func (mm *MeshManager) RegisterTriangle(id string) {
	vertices := []float32{
		// pos           normal        uv
		0.0, 0.5, 0.0, 0, 0, 1, 0.5, 1,
		-0.5, -0.5, 0.0, 0, 0, 1, 0, 0,
		0.5, -0.5, 0.0, 0, 0, 1, 1, 0,
	}

	indices := []uint32{0, 1, 2}

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

	gl.VertexAttribPointerWithOffset(0, 3, gl.FLOAT, false, stride, 0)
	gl.EnableVertexAttribArray(0)

	gl.VertexAttribPointerWithOffset(1, 3, gl.FLOAT, false, stride, 3*4)
	gl.EnableVertexAttribArray(1)

	gl.VertexAttribPointerWithOffset(2, 2, gl.FLOAT, false, stride, 6*4)
	gl.EnableVertexAttribArray(2)

	gl.BindVertexArray(0)

	mm.vaos[id] = vao
	mm.vbos[id] = vbo
	mm.ebos[id] = ebo
	mm.counts[id] = int32(len(indices))
	mm.indexTypes[id] = gl.UNSIGNED_INT
	mm.vertexCounts[id] = 3
}

func (mm *MeshManager) RegisterLine(id string) {
	// Two vertices: origin and unit Z
	vertices := []float32{
		0, 0, 0,
		0, 0, 1,
	}
	indices := []uint32{0, 1}

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	var vbo, ebo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.GenBuffers(1, &ebo)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	gl.VertexAttribPointerWithOffset(0, 3, gl.FLOAT, false, 3*4, 0)
	gl.EnableVertexAttribArray(0)

	gl.BindVertexArray(0)
	mm.indexTypes[id] = gl.UNSIGNED_INT
	mm.vertexCounts[id] = int32(len(vertices) / 3) // 3 floats per vertex
	mm.counts[id] = int32(len(indices))
	mm.vbos[id] = vbo
	mm.ebos[id] = ebo
	mm.vaos[id] = vao
	mm.layoutType[id] = 3

	// verify EBO size: 4 bytes per uint32 index
	mm.verifyEBOSize(ebo, int32(len(indices)*4), id)

}

func (mm *MeshManager) RegisterCube(id string) {
	// 24 vertices: 6 faces * 4 verts, each with position+normal
	vertices := []float32{
		// pos           normal        uv
		// Front (+Z)
		-0.5, -0.5, 0.5, 0, 0, 1, 0, 0,
		0.5, -0.5, 0.5, 0, 0, 1, 1, 0,
		0.5, 0.5, 0.5, 0, 0, 1, 1, 1,
		-0.5, 0.5, 0.5, 0, 0, 1, 0, 1,

		// Back (-Z)
		0.5, -0.5, -0.5, 0, 0, -1, 0, 0,
		-0.5, -0.5, -0.5, 0, 0, -1, 1, 0,
		-0.5, 0.5, -0.5, 0, 0, -1, 1, 1,
		0.5, 0.5, -0.5, 0, 0, -1, 0, 1,

		// Left (-X)
		-0.5, -0.5, -0.5, -1, 0, 0, 0, 0,
		-0.5, -0.5, 0.5, -1, 0, 0, 1, 0,
		-0.5, 0.5, 0.5, -1, 0, 0, 1, 1,
		-0.5, 0.5, -0.5, -1, 0, 0, 0, 1,

		// Right (+X)
		0.5, -0.5, 0.5, 1, 0, 0, 0, 0,
		0.5, -0.5, -0.5, 1, 0, 0, 1, 0,
		0.5, 0.5, -0.5, 1, 0, 0, 1, 1,
		0.5, 0.5, 0.5, 1, 0, 0, 0, 1,

		// Top (+Y)
		-0.5, 0.5, 0.5, 0, 1, 0, 0, 0,
		0.5, 0.5, 0.5, 0, 1, 0, 1, 0,
		0.5, 0.5, -0.5, 0, 1, 0, 1, 1,
		-0.5, 0.5, -0.5, 0, 1, 0, 0, 1,

		// Bottom (-Y)
		-0.5, -0.5, -0.5, 0, -1, 0, 0, 0,
		0.5, -0.5, -0.5, 0, -1, 0, 1, 0,
		0.5, -0.5, 0.5, 0, -1, 0, 1, 1,
		-0.5, -0.5, 0.5, 0, -1, 0, 0, 1,
	}

	indices := []uint32{
		// Front
		0, 1, 2, 2, 3, 0,
		// Back
		4, 5, 6, 6, 7, 4,
		// Left
		8, 9, 10, 10, 11, 8,
		// Right
		12, 13, 14, 14, 15, 12,
		// Top
		16, 17, 18, 18, 19, 16,
		// Bottom
		20, 21, 22, 22, 23, 20,
	}
	vertexCount := int32(len(vertices) / 8)
	vertices12 := make([]float32, vertexCount*12)
	for i := 0; i < int(vertexCount); i++ {
		copy(vertices12[i*12:], vertices[i*8:i*8+8]) // tangent(3) + w(1) will be filled in next
	}
	computeTangents(vertices12, indices)

	var vao, vbo, ebo uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)
	gl.GenBuffers(1, &ebo)

	gl.BindVertexArray(vao)
	// validate indices: ensure max index < vertexCount
	// 8 floats per vertex for cube
	var maxIdx uint32 = 0
	for _, idx := range indices {
		if idx > maxIdx {
			maxIdx = idx
		}
	}
	if int(maxIdx) >= int(vertexCount) {
		log.Printf("ERROR: RegisterCube(%s) maxIndex=%d >= vertexCount=%d", id, maxIdx, vertexCount)
		return
	}

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices12)*4, gl.Ptr(vertices12), gl.STATIC_DRAW)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	stride := int32(12 * 4) // 8 floats per vertex
	/*if stride == 12 {
		computeTangents(vertices, indices)
	}*/
	// position
	// set attribute pointers using offset overloads (safe under Go 1.14+)
	gl.VertexAttribPointerWithOffset(0, 3, gl.FLOAT, false, stride, 0)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointerWithOffset(1, 3, gl.FLOAT, false, stride, 3*4)
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointerWithOffset(2, 2, gl.FLOAT, false, stride, 6*4)
	gl.EnableVertexAttribArray(2)
	gl.VertexAttribPointerWithOffset(3, 4, gl.FLOAT, false, stride, 8*4)
	gl.EnableVertexAttribArray(3)

	gl.BindVertexArray(0)
	// now query the VAO state while still bound
	// POST-REG full diagnostic - place while VAO still bound, right after setting pointers
	// DIAG: place immediately after each VertexAttribPointerWithOffset call (while VAO still bound)

	mm.indexTypes[id] = gl.UNSIGNED_INT
	mm.vertexCounts[id] = int32(len(vertices) / 12)
	mm.counts[id] = int32(len(indices))
	mm.vbos[id] = vbo
	mm.ebos[id] = ebo
	mm.vaos[id] = vao
	mm.layoutType[id] = 12

	mm.verifyEBOSize(ebo, int32(len(indices)*4), id)

}

func (mm *MeshManager) RegisterWireCube(id string) {
	// same vertices as cube, but indices for edges
	vertices := []float32{
		-0.5, -0.5, 0.5,
		0.5, -0.5, 0.5,
		0.5, 0.5, 0.5,
		-0.5, 0.5, 0.5,
		-0.5, -0.5, -0.5,
		0.5, -0.5, -0.5,
		0.5, 0.5, -0.5,
		-0.5, 0.5, -0.5,
	}
	indices := []uint32{
		0, 1, 1, 2, 2, 3, 3, 0, // front
		4, 5, 5, 6, 6, 7, 7, 4, // back
		0, 4, 1, 5, 2, 6, 3, 7, // sides
	}
	// same VAO/VBO/EBO setup as before
	// store in mm.vaos[id], mm.counts[id]
	var vao, vbo, ebo uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)
	gl.GenBuffers(1, &ebo)

	gl.BindVertexArray(vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	gl.VertexAttribPointerWithOffset(0, 3, gl.FLOAT, false, 3*4, 0)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointerWithOffset(1, 3, gl.FLOAT, false, 6*4, 3*4)
	gl.EnableVertexAttribArray(1)

	gl.BindVertexArray(0)
	mm.layoutType[id] = 8

	mm.vaos[id] = vao
	mm.vbos[id] = vbo
	mm.ebos[id] = ebo
	mm.counts[id] = int32(len(indices))
}
func (mm *MeshManager) RegisterWireSphere(id string, slices, stacks int) {
	var vertices []float32
	var indices []uint32

	// Generate vertices
	for i := 0; i <= stacks; i++ {
		phi := float32(i) * (3.14159 / float32(stacks)) // latitude
		for j := 0; j <= slices; j++ {
			theta := float32(j) * (2.0 * 3.14159 / float32(slices)) // longitude
			x := float32(math.Sin(float64(phi)) * math.Cos(float64(theta)))
			y := float32(math.Cos(float64(phi)))
			z := float32(math.Sin(float64(phi)) * math.Sin(float64(theta)))
			vertices = append(vertices, x*0.5, y*0.5, z*0.5) // radius 0.5
		}
	}

	// Generate line indices (wireframe grid)
	for i := 0; i < stacks; i++ {
		for j := 0; j < slices; j++ {
			first := uint32(i*(slices+1) + j)
			second := first + uint32(slices+1)
			// vertical lines
			indices = append(indices, first, second)
			// horizontal lines
			indices = append(indices, first, first+1)
		}
	}

	var vao, vbo, ebo uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)
	gl.GenBuffers(1, &ebo)

	gl.BindVertexArray(vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	gl.VertexAttribPointerWithOffset(0, 3, gl.FLOAT, false, 3*4, 0)
	gl.EnableVertexAttribArray(0)

	gl.VertexAttribPointerWithOffset(1, 3, gl.FLOAT, false, 6*4, 3*4)
	gl.EnableVertexAttribArray(1)

	gl.BindVertexArray(0)

	mm.indexTypes[id] = gl.UNSIGNED_INT
	mm.vertexCounts[id] = int32(len(vertices) / 3)
	mm.counts[id] = int32(len(indices))
	mm.vbos[id] = vbo
	mm.ebos[id] = ebo
	mm.vaos[id] = vao
	mm.layoutType[id] = 3

	mm.verifyEBOSize(ebo, int32(len(indices)*4), id)

}

func (mm *MeshManager) RegisterCube8(id string) {
	vertices := []float32{
		// pos           normal        uv
		-0.5, -0.5, 0.5, 0, 0, 1, 0, 0,
		0.5, -0.5, 0.5, 0, 0, 1, 1, 0,
		0.5, 0.5, 0.5, 0, 0, 1, 1, 1,
		-0.5, 0.5, 0.5, 0, 0, 1, 0, 1,

		0.5, -0.5, -0.5, 0, 0, -1, 0, 0,
		-0.5, -0.5, -0.5, 0, 0, -1, 1, 0,
		-0.5, 0.5, -0.5, 0, 0, -1, 1, 1,
		0.5, 0.5, -0.5, 0, 0, -1, 0, 1,

		-0.5, -0.5, -0.5, -1, 0, 0, 0, 0,
		-0.5, -0.5, 0.5, -1, 0, 0, 1, 0,
		-0.5, 0.5, 0.5, -1, 0, 0, 1, 1,
		-0.5, 0.5, -0.5, -1, 0, 0, 0, 1,

		0.5, -0.5, 0.5, 1, 0, 0, 0, 0,
		0.5, -0.5, -0.5, 1, 0, 0, 1, 0,
		0.5, 0.5, -0.5, 1, 0, 0, 1, 1,
		0.5, 0.5, 0.5, 1, 0, 0, 0, 1,

		-0.5, 0.5, 0.5, 0, 1, 0, 0, 0,
		0.5, 0.5, 0.5, 0, 1, 0, 1, 0,
		0.5, 0.5, -0.5, 0, 1, 0, 1, 1,
		-0.5, 0.5, -0.5, 0, 1, 0, 0, 1,

		-0.5, -0.5, -0.5, 0, -1, 0, 0, 0,
		0.5, -0.5, -0.5, 0, -1, 0, 1, 0,
		0.5, -0.5, 0.5, 0, -1, 0, 1, 1,
		-0.5, -0.5, 0.5, 0, -1, 0, 0, 1,
	}

	indices := []uint32{
		0, 1, 2, 2, 3, 0,
		4, 5, 6, 6, 7, 4,
		8, 9, 10, 10, 11, 8,
		12, 13, 14, 14, 15, 12,
		16, 17, 18, 18, 19, 16,
		20, 21, 22, 22, 23, 20,
	}

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

	gl.VertexAttribPointerWithOffset(0, 3, gl.FLOAT, false, stride, 0)
	gl.EnableVertexAttribArray(0)

	gl.VertexAttribPointerWithOffset(1, 3, gl.FLOAT, false, stride, 3*4)
	gl.EnableVertexAttribArray(1)

	gl.VertexAttribPointerWithOffset(2, 2, gl.FLOAT, false, stride, 6*4)
	gl.EnableVertexAttribArray(2)

	gl.BindVertexArray(0)

	mm.vaos[id] = vao
	mm.vbos[id] = vbo
	mm.ebos[id] = ebo
	mm.counts[id] = int32(len(indices))
	mm.indexTypes[id] = gl.UNSIGNED_INT
	mm.vertexCounts[id] = int32(len(vertices) / 8)
	mm.layoutType[id] = 8

	mm.verifyEBOSize(ebo, int32(len(indices)*4), id)
}

func (mm *MeshManager) RegisterPlane(id string) {
	// A simple 1×1 plane in XZ, centered at origin
	// pos(3), normal(3), uv(2)
	baseVerts := []float32{
		//   x     y    z     nx ny nz    u   v
		-0.5, 0, -0.5, 0, 1, 0, 0, 0,
		0.5, 0, -0.5, 0, 1, 0, 1, 0,
		0.5, 0, 0.5, 0, 1, 0, 1, 1,
		-0.5, 0, 0.5, 0, 1, 0, 0, 1,
	}

	indices := []uint32{
		0, 1, 2,
		2, 3, 0,
	}

	// Expand to 12 floats per vertex
	vertexCount := len(baseVerts) / 8
	verts12 := make([]float32, vertexCount*12)

	for i := 0; i < vertexCount; i++ {
		copy(verts12[i*12:], baseVerts[i*8:i*8+8])
		// tangent.xyz + w will be filled by computeTangents
	}

	computeTangents(verts12, indices)

	// Upload to GL
	var vao, vbo, ebo uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)
	gl.GenBuffers(1, &ebo)

	gl.BindVertexArray(vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(verts12)*4, gl.Ptr(verts12), gl.STATIC_DRAW)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	stride := int32(12 * 4)

	gl.VertexAttribPointerWithOffset(0, 3, gl.FLOAT, false, stride, 0)
	gl.EnableVertexAttribArray(0)

	gl.VertexAttribPointerWithOffset(1, 3, gl.FLOAT, false, stride, 3*4)
	gl.EnableVertexAttribArray(1)

	gl.VertexAttribPointerWithOffset(2, 2, gl.FLOAT, false, stride, 6*4)
	gl.EnableVertexAttribArray(2)

	gl.VertexAttribPointerWithOffset(3, 4, gl.FLOAT, false, stride, 8*4)
	gl.EnableVertexAttribArray(3)

	gl.BindVertexArray(0)

	mm.vaos[id] = vao
	mm.vbos[id] = vbo
	mm.ebos[id] = ebo
	mm.counts[id] = int32(len(indices))
	mm.indexTypes[id] = gl.UNSIGNED_INT
	mm.vertexCounts[id] = int32(vertexCount)
	mm.layoutType[id] = 12

	mm.verifyEBOSize(ebo, int32(len(indices)*4), id)
}
func (mm *MeshManager) RegisterPlaneDoubleSided(id string) {
	// Base vertices: pos(3), normal(3), uv(2)
	baseVerts := []float32{
		// FRONT FACE (normal +Y)
		-0.5, 0, -0.5, 0, 1, 0, 0, 0,
		0.5, 0, -0.5, 0, 1, 0, 1, 0,
		0.5, 0, 0.5, 0, 1, 0, 1, 1,
		-0.5, 0, 0.5, 0, 1, 0, 0, 1,

		// BACK FACE (normal -Y)
		-0.5, 0, 0.5, 0, -1, 0, 0, 0,
		0.5, 0, 0.5, 0, -1, 0, 1, 0,
		0.5, 0, -0.5, 0, -1, 0, 1, 1,
		-0.5, 0, -0.5, 0, -1, 0, 0, 1,
	}

	indices := []uint32{
		// FRONT
		0, 1, 2,
		2, 3, 0,

		// BACK (winding reversed)
		4, 5, 6,
		6, 7, 4,
	}

	// Expand to 12 floats per vertex
	vertexCount := len(baseVerts) / 8
	verts12 := make([]float32, vertexCount*12)

	for i := 0; i < vertexCount; i++ {
		copy(verts12[i*12:], baseVerts[i*8:i*8+8])
		// tangent.xyz + w will be filled by computeTangents
	}

	computeTangents(verts12, indices)

	// Upload to GL
	var vao, vbo, ebo uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)
	gl.GenBuffers(1, &ebo)

	gl.BindVertexArray(vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(verts12)*4, gl.Ptr(verts12), gl.STATIC_DRAW)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	stride := int32(12 * 4)

	gl.VertexAttribPointerWithOffset(0, 3, gl.FLOAT, false, stride, 0)
	gl.EnableVertexAttribArray(0)

	gl.VertexAttribPointerWithOffset(1, 3, gl.FLOAT, false, stride, 3*4)
	gl.EnableVertexAttribArray(1)

	gl.VertexAttribPointerWithOffset(2, 2, gl.FLOAT, false, stride, 6*4)
	gl.EnableVertexAttribArray(2)

	gl.VertexAttribPointerWithOffset(3, 4, gl.FLOAT, false, stride, 8*4)
	gl.EnableVertexAttribArray(3)

	gl.BindVertexArray(0)

	mm.vaos[id] = vao
	mm.vbos[id] = vbo
	mm.ebos[id] = ebo
	mm.counts[id] = int32(len(indices))
	mm.indexTypes[id] = gl.UNSIGNED_INT
	mm.vertexCounts[id] = int32(vertexCount)
	mm.layoutType[id] = 12

	mm.verifyEBOSize(ebo, int32(len(indices)*4), id)
}

func (mm *MeshManager) RegisterSphere(id string, slices, stacks int) {
	var baseVerts []float32
	var indices []uint32

	// --- Generate base vertices: pos(3), normal(3), uv(2) = 8 floats ---
	for i := 0; i <= stacks; i++ {
		phi := float32(i) * (math.Pi / float32(stacks)) // latitude
		v := float32(i) / float32(stacks)

		for j := 0; j <= slices; j++ {
			theta := float32(j) * (2.0 * math.Pi / float32(slices)) // longitude
			u := float32(j) / float32(slices)

			x := float32(math.Sin(float64(phi)) * math.Cos(float64(theta)))
			y := float32(math.Cos(float64(phi)))
			z := float32(math.Sin(float64(phi)) * math.Sin(float64(theta)))

			// pos(3), normal(3), uv(2)
			baseVerts = append(baseVerts,
				x*0.5, y*0.5, z*0.5, // position
				x, y, z, // normal
				u, v, // uv
			)
		}
	}

	// --- Generate indices ---
	for i := 0; i < stacks; i++ {
		for j := 0; j < slices; j++ {
			first := uint32(i*(slices+1) + j)
			second := first + uint32(slices+1)

			indices = append(indices, first, second, first+1)
			indices = append(indices, second, second+1, first+1)
		}
	}

	// --- Expand to 12 floats per vertex (pos3, normal3, uv2, tangent3, w1) ---
	vertexCount := len(baseVerts) / 8
	verts12 := make([]float32, vertexCount*12)

	for i := 0; i < vertexCount; i++ {
		copy(verts12[i*12:], baseVerts[i*8:i*8+8])
		// tangent.xyz + w will be filled by computeTangents
	}

	// --- Compute tangents ---
	computeTangents(verts12, indices)

	// --- Upload to GL ---
	var vao, vbo, ebo uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)
	gl.GenBuffers(1, &ebo)

	gl.BindVertexArray(vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(verts12)*4, gl.Ptr(verts12), gl.STATIC_DRAW)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	stride := int32(12 * 4)

	gl.VertexAttribPointerWithOffset(0, 3, gl.FLOAT, false, stride, 0)
	gl.EnableVertexAttribArray(0)

	gl.VertexAttribPointerWithOffset(1, 3, gl.FLOAT, false, stride, 3*4)
	gl.EnableVertexAttribArray(1)

	gl.VertexAttribPointerWithOffset(2, 2, gl.FLOAT, false, stride, 6*4)
	gl.EnableVertexAttribArray(2)

	gl.VertexAttribPointerWithOffset(3, 4, gl.FLOAT, false, stride, 8*4)
	gl.EnableVertexAttribArray(3)

	gl.BindVertexArray(0)

	mm.vaos[id] = vao
	mm.vbos[id] = vbo
	mm.ebos[id] = ebo
	mm.counts[id] = int32(len(indices))
	mm.indexTypes[id] = gl.UNSIGNED_INT
	mm.vertexCounts[id] = int32(vertexCount)
	mm.layoutType[id] = 12

	mm.verifyEBOSize(ebo, int32(len(indices)*4), id)
}

func (mm *MeshManager) GetVAO(id string) uint32 { return mm.vaos[id] }

func (mm *MeshManager) Delete() {
	// delete VAOs
	for _, vao := range mm.vaos {
		gl.DeleteVertexArrays(1, &vao)
	}
	// delete VBOs
	for _, vbo := range mm.vbos {
		gl.DeleteBuffers(1, &vbo)
	}
	// delete EBOs
	for _, ebo := range mm.ebos {
		gl.DeleteBuffers(1, &ebo)
	}
	// clear maps
	mm.vaos = nil
	mm.vbos = nil
	mm.ebos = nil
	mm.counts = nil
}

// RegisterGizmoArrow creates a simple arrow mesh pointing +Z (shaft + cone tip).
func (mm *MeshManager) RegisterGizmoArrow(id string) {
	// Simple low-poly arrow: shaft (two triangles as a thin quad) + cone tip (4 triangles)
	// Layout: position only (location 0) — matches RegisterTriangle/RegisterLine style.
	vertices := []float32{
		// shaft (a thin rectangular prism approximated as two triangles per face, but keep it minimal)
		// We'll use a very simple shaft: two triangles forming a thin quad in X-Y at z=0 and z=0.6 for the shaft end
		-0.02, -0.02, 0.0,
		0.02, -0.02, 0.0,
		0.02, 0.02, 0.0,
		-0.02, 0.02, 0.0,
		-0.02, -0.02, 0.6,
		0.02, -0.02, 0.6,
		0.02, 0.02, 0.6,
		-0.02, 0.02, 0.6,
		// tip vertex (cone tip at z=1.0)
		0.0, 0.0, 1.0,
	}

	// indices: shaft as 12 triangles (two quads per side) is overkill; keep a minimal set:
	indices := []uint32{
		// front face quad (0,1,2,3) -> two triangles
		0, 1, 2, 2, 3, 0,
		// top face connecting to shaft end (3,2,6,7)
		3, 2, 6, 6, 7, 3,
		// bottom face (0,4,5,1)
		0, 4, 5, 5, 1, 0,
		// left face (0,3,7,4)
		0, 3, 7, 7, 4, 0,
		// right face (1,5,6,2)
		1, 5, 6, 6, 2, 1,
		// tip faces: connect shaft end ring (4,5,6,7) to tip vertex 8
		4, 5, 8,
		5, 6, 8,
		6, 7, 8,
		7, 4, 8,
	}

	var vao, vbo, ebo uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)
	gl.GenBuffers(1, &ebo)

	gl.BindVertexArray(vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	// position attribute only (location 0)
	gl.VertexAttribPointerWithOffset(0, 3, gl.FLOAT, false, 3*4, 0)
	gl.EnableVertexAttribArray(0)

	gl.BindVertexArray(0)

	mm.indexTypes[id] = gl.UNSIGNED_INT
	mm.vertexCounts[id] = int32(len(vertices) / 3)
	mm.counts[id] = int32(len(indices))
	mm.vbos[id] = vbo
	mm.ebos[id] = ebo
	mm.vaos[id] = vao

	mm.verifyEBOSize(ebo, int32(len(indices)*4), id)

}
func (mm *MeshManager) RegisterGizmoPlane(id string) {
	// Simple 1×1 square in XY plane, centered at origin
	vertices := []float32{
		-0.5, -0.5, 0,
		0.5, -0.5, 0,
		0.5, 0.5, 0,
		-0.5, 0.5, 0,
	}

	indices := []uint32{
		0, 1, 2,
		2, 3, 0,
	}

	var vao, vbo, ebo uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)
	gl.GenBuffers(1, &ebo)

	gl.BindVertexArray(vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	gl.VertexAttribPointerWithOffset(0, 3, gl.FLOAT, false, 3*4, 0)
	gl.EnableVertexAttribArray(0)

	gl.BindVertexArray(0)

	mm.indexTypes[id] = gl.UNSIGNED_INT
	mm.vertexCounts[id] = int32(len(vertices) / 3)
	mm.counts[id] = int32(len(indices))
	mm.vbos[id] = vbo
	mm.ebos[id] = ebo
	mm.vaos[id] = vao
	mm.layoutType[id] = 3

	mm.verifyEBOSize(ebo, int32(len(indices)*4), id)

}
func (mm *MeshManager) RegisterGizmoCircle(id string, segments int) {
	var vertices []float32
	var indices []uint32

	for i := 0; i < segments; i++ {
		angle := float32(i) * 2 * math.Pi / float32(segments)
		x := float32(math.Cos(float64(angle)))
		y := float32(math.Sin(float64(angle)))
		vertices = append(vertices, x, y, 0)
		indices = append(indices, uint32(i), uint32((i+1)%segments))
	}

	var vao, vbo, ebo uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)
	gl.GenBuffers(1, &ebo)

	gl.BindVertexArray(vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	gl.VertexAttribPointerWithOffset(0, 3, gl.FLOAT, false, 3*4, 0)
	gl.EnableVertexAttribArray(0)

	gl.BindVertexArray(0)

	mm.indexTypes[id] = gl.UNSIGNED_INT
	mm.vertexCounts[id] = int32(len(vertices) / 3)
	mm.counts[id] = int32(len(indices))
	mm.vbos[id] = vbo
	mm.ebos[id] = ebo
	mm.vaos[id] = vao
	mm.layoutType[id] = 3

	mm.verifyEBOSize(ebo, int32(len(indices)*4), id)

}
func (mm *MeshManager) RegisterBillboardQuad(id string) {
	// pos(3), normal(3), uv(2)
	baseVerts := []float32{
		// XY plane facing +Z
		-0.5, -0.5, 0, 0, 0, 1, 0, 0,
		0.5, -0.5, 0, 0, 0, 1, 1, 0,
		0.5, 0.5, 0, 0, 0, 1, 1, 1,
		-0.5, 0.5, 0, 0, 0, 1, 0, 1,
	}

	indices := []uint32{
		0, 1, 2,
		2, 3, 0,
	}

	// Expand to 12 floats per vertex
	vertexCount := len(baseVerts) / 8
	verts12 := make([]float32, vertexCount*12)

	for i := 0; i < vertexCount; i++ {
		copy(verts12[i*12:], baseVerts[i*8:i*8+8])
	}

	computeTangents(verts12, indices)

	// Upload to GL
	var vao, vbo, ebo uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)
	gl.GenBuffers(1, &ebo)

	gl.BindVertexArray(vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(verts12)*4, gl.Ptr(verts12), gl.STATIC_DRAW)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	stride := int32(12 * 4)

	gl.VertexAttribPointerWithOffset(0, 3, gl.FLOAT, false, stride, 0)
	gl.EnableVertexAttribArray(0)

	gl.VertexAttribPointerWithOffset(1, 3, gl.FLOAT, false, stride, 3*4)
	gl.EnableVertexAttribArray(1)

	gl.VertexAttribPointerWithOffset(2, 2, gl.FLOAT, false, stride, 6*4)
	gl.EnableVertexAttribArray(2)

	gl.VertexAttribPointerWithOffset(3, 4, gl.FLOAT, false, stride, 8*4)
	gl.EnableVertexAttribArray(3)

	gl.BindVertexArray(0)

	mm.vaos[id] = vao
	mm.vbos[id] = vbo
	mm.ebos[id] = ebo
	mm.counts[id] = int32(len(indices))
	mm.indexTypes[id] = gl.UNSIGNED_INT
	mm.vertexCounts[id] = int32(vertexCount)
	mm.layoutType[id] = 12

	mm.verifyEBOSize(ebo, int32(len(indices)*4), id)
}

func (mm *MeshManager) RegisterSinglePlane(id string) {
	// Base vertices: pos(3), normal(3), uv(2)
	baseVerts := []float32{
		//   x     y    z     nx ny nz    u   v
		-0.5, 0, -0.5, 0, 1, 0, 0, 0,
		0.5, 0, -0.5, 0, 1, 0, 1, 0,
		0.5, 0, 0.5, 0, 1, 0, 1, 1,
		-0.5, 0, 0.5, 0, 1, 0, 0, 1,
	}

	indices := []uint32{
		0, 1, 2,
		2, 3, 0,
	}

	// Expand to 12 floats per vertex (pos3, normal3, uv2, tangent3, w1)
	vertexCount := len(baseVerts) / 8
	verts12 := make([]float32, vertexCount*12)

	for i := 0; i < vertexCount; i++ {
		copy(verts12[i*12:], baseVerts[i*8:i*8+8])
		// tangent.xyz + w will be filled by computeTangents
	}

	computeTangents(verts12, indices)

	// Upload to GL
	var vao, vbo, ebo uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)
	gl.GenBuffers(1, &ebo)

	gl.BindVertexArray(vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(verts12)*4, gl.Ptr(verts12), gl.STATIC_DRAW)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	stride := int32(12 * 4)

	gl.VertexAttribPointerWithOffset(0, 3, gl.FLOAT, false, stride, 0)
	gl.EnableVertexAttribArray(0)

	gl.VertexAttribPointerWithOffset(1, 3, gl.FLOAT, false, stride, 3*4)
	gl.EnableVertexAttribArray(1)

	gl.VertexAttribPointerWithOffset(2, 2, gl.FLOAT, false, stride, 6*4)
	gl.EnableVertexAttribArray(2)

	gl.VertexAttribPointerWithOffset(3, 4, gl.FLOAT, false, stride, 8*4)
	gl.EnableVertexAttribArray(3)

	gl.BindVertexArray(0)

	mm.vaos[id] = vao
	mm.vbos[id] = vbo
	mm.ebos[id] = ebo
	mm.counts[id] = int32(len(indices))
	mm.indexTypes[id] = gl.UNSIGNED_INT
	mm.vertexCounts[id] = int32(vertexCount)
	mm.layoutType[id] = 12

	mm.verifyEBOSize(ebo, int32(len(indices)*4), id)
}

// --- helpers for index/vertex bookkeeping and EBO verification ---

// SetMeshIndexInfo allows external code (glTF/OBJ importers) to register
// the index type and vertex count for a mesh that was created elsewhere.
func (mm *MeshManager) SetMeshIndexInfo(id string, indexType uint32, vertexCount int32) {
	mm.indexTypes[id] = indexType
	mm.vertexCounts[id] = vertexCount
}

// GetIndexType returns the recorded index type for a mesh (default UNSIGNED_INT).
func (mm *MeshManager) GetIndexType(id string) uint32 {
	if t, ok := mm.indexTypes[id]; ok {
		return t
	}
	return gl.UNSIGNED_INT
}

// GetVertexCount returns the recorded vertex count for a mesh (fallbacks to index count).
func (mm *MeshManager) GetVertexCount(id string) int32 {
	if c, ok := mm.vertexCounts[id]; ok {
		return c
	}
	// fallback: if we only have index count, return that (useful for DrawArrays fallback)
	if c, ok := mm.counts[id]; ok {
		return c
	}
	return 0
}

// verifyEBOSize logs a warning if the element buffer size doesn't match expected bytes.
// bytesPerIndex should be 4 for uint32 indices, 2 for uint16 indices.
func (mm *MeshManager) verifyEBOSize(ebo uint32, expectedBytes int32, meshID string) {
	var eboSize int32
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.GetBufferParameteriv(gl.ELEMENT_ARRAY_BUFFER, gl.BUFFER_SIZE, &eboSize)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, 0)
	if eboSize != expectedBytes {
		// Use Printf/Log as appropriate in your project; keep it lightweight
		// This warns you early if an index upload used a different element size.
		// Do not panic here; just warn so you can inspect the importer.
		// fmt.Printf is avoided to keep imports minimal; use log if already imported.
		// If you prefer log.Printf, replace the next line with log.Printf.
		println("Warning: EBO size mismatch for mesh", meshID, "got", eboSize, "expected", expectedBytes)
	}
}

func (mm *MeshManager) GetEBO(meshID string) uint32 {
	if e, ok := mm.ebos[meshID]; ok {
		return e
	}
	return 0
}

func (mm *MeshManager) GetVBO(meshID string) uint32 {
	return mm.vbos[meshID]
}
