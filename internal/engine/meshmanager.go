// internal/engine/meshmanager.go
package engine

import (
	"math"

	"github.com/go-gl/gl/v4.1-core/gl"
)

// internal/engine/meshmanager.go (add fields)
type MeshManager struct {
	vaos   map[string]uint32
	counts map[string]int32

	// new: track buffers so we can delete them
	vbos map[string]uint32
	ebos map[string]uint32
}

func NewMeshManager() *MeshManager {
	return &MeshManager{
		vaos:   make(map[string]uint32),
		counts: make(map[string]int32),
		vbos:   make(map[string]uint32),
		ebos:   make(map[string]uint32),
	}
}

func (mm *MeshManager) GetCount(MeshID string) int32 {
	return mm.counts[MeshID]
}

// Triangle with only positions (layout location 0)
func (mm *MeshManager) RegisterTriangle(id string) {
	// 3 vertices
	vertices := []float32{
		0.0, 0.5, 0.0,
		-0.5, -0.5, 0.0,
		0.5, -0.5, 0.0,
	}
	// Indices for one triangle
	indices := []uint32{0, 1, 2}

	var vao, vbo, ebo uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)
	gl.GenBuffers(1, &ebo)

	gl.BindVertexArray(vao)

	// Vertex buffer
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// Element buffer
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	// Vertex attribute
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, gl.PtrOffset(0))

	gl.BindVertexArray(0)

	mm.vaos[id] = vao
	mm.vbos[id] = vbo
	mm.ebos[id] = ebo
	mm.counts[id] = int32(len(indices))
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

	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, gl.PtrOffset(0))

	gl.BindVertexArray(0)

	mm.vaos[id] = vao
	mm.vbos[id] = vbo
	mm.ebos[id] = ebo
	mm.counts[id] = int32(len(indices))
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

	var vao, vbo, ebo uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)
	gl.GenBuffers(1, &ebo)

	gl.BindVertexArray(vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	stride := int32(8 * 4) // 8 floats per vertex

	// position
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, stride, gl.PtrOffset(0))

	// normal
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, stride, gl.PtrOffset(3*4))

	// uv
	gl.EnableVertexAttribArray(2)
	gl.VertexAttribPointer(2, 2, gl.FLOAT, false, stride, gl.PtrOffset(6*4))

	gl.BindVertexArray(0)

	mm.vaos[id] = vao
	mm.vbos[id] = vbo
	mm.ebos[id] = ebo
	mm.counts[id] = int32(len(indices))
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

	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(3*4))
	gl.BindVertexArray(0)

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

	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, gl.PtrOffset(0))

	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(3*4))
	gl.BindVertexArray(0)

	mm.vaos[id] = vao
	mm.vbos[id] = vbo
	mm.ebos[id] = ebo
	mm.counts[id] = int32(len(indices))
}

func (mm *MeshManager) RegisterSphere(id string, slices, stacks int) {
	var vertices []float32
	var indices []uint32

	// Generate vertices with normals
	for i := 0; i <= stacks; i++ {
		phi := float32(i) * (3.14159 / float32(stacks)) // latitude
		for j := 0; j <= slices; j++ {
			theta := float32(j) * (2.0 * 3.14159 / float32(slices)) // longitude

			x := float32(math.Sin(float64(phi)) * math.Cos(float64(theta)))
			y := float32(math.Cos(float64(phi)))
			z := float32(math.Sin(float64(phi)) * math.Sin(float64(theta)))

			// Position (radius 0.5) and normal (unit vector)
			vertices = append(vertices,
				x*0.5, y*0.5, z*0.5, // position
				x, y, z) // normal
		}
	}

	// Generate indices
	for i := 0; i < stacks; i++ {
		for j := 0; j < slices; j++ {
			first := uint32(i*(slices+1) + j)
			second := first + uint32(slices+1)

			indices = append(indices, first, second, first+1)
			indices = append(indices, second, second+1, first+1)
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

	// Position attribute
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(0))

	// Normal attribute
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(3*4))

	gl.BindVertexArray(0)

	mm.vaos[id] = vao
	mm.vbos[id] = vbo
	mm.ebos[id] = ebo
	mm.counts[id] = int32(len(indices))
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
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, gl.PtrOffset(0))

	gl.BindVertexArray(0)

	mm.vaos[id] = vao
	mm.vbos[id] = vbo
	mm.ebos[id] = ebo
	mm.counts[id] = int32(len(indices))
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

	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, gl.PtrOffset(0))

	gl.BindVertexArray(0)

	mm.vaos[id] = vao
	mm.vbos[id] = vbo
	mm.ebos[id] = ebo
	mm.counts[id] = int32(len(indices))
}
