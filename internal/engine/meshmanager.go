// internal/engine/meshmanager.go
package engine

import "github.com/go-gl/gl/v4.1-core/gl"

type MeshManager struct {
	vaos   map[string]uint32
	counts map[string]int32
}

func (mm *MeshManager) GetCount(MeshID string) int32 {
	return mm.counts[MeshID]
}
func NewMeshManager() *MeshManager {
	return &MeshManager{
		vaos:   make(map[string]uint32),
		counts: make(map[string]int32),
	}
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
	mm.counts[id] = int32(len(indices))
}

func (mm *MeshManager) RegisterCube(id string) {
	// 8 vertices of a cube (centered at origin, size 1)
	vertices := []float32{
		// Front face
		-0.5, -0.5, 0.5,
		0.5, -0.5, 0.5,
		0.5, 0.5, 0.5,
		-0.5, 0.5, 0.5,
		// Back face
		-0.5, -0.5, -0.5,
		0.5, -0.5, -0.5,
		0.5, 0.5, -0.5,
		-0.5, 0.5, -0.5,
	}

	// Indices for 12 triangles (two per face)
	indices := []uint32{
		0, 1, 2, 2, 3, 0, // front
		4, 5, 6, 6, 7, 4, // back
		0, 4, 7, 7, 3, 0, // left
		1, 5, 6, 6, 2, 1, // right
		3, 2, 6, 6, 7, 3, // top
		0, 1, 5, 5, 4, 0, // bottom
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
	mm.counts[id] = int32(len(indices))
}

func (mm *MeshManager) GetVAO(id string) uint32 { return mm.vaos[id] }
