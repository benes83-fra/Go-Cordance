package engine

import "github.com/go-gl/gl/v4.1-core/gl"

// MeshManager stores VAOs/VBOs keyed by string IDs.
type MeshManager struct {
	meshes map[string]uint32
}

func NewMeshManager() *MeshManager {
	return &MeshManager{meshes: make(map[string]uint32)}
}

func (mm *MeshManager) RegisterTriangle(id string) {
	vertices := []float32{
		0, 0.5, 0, 1, 0, 0,
		-0.5, -0.5, 0, 0, 1, 0,
		0.5, -0.5, 0, 0, 0, 1,
	}
	var vao, vbo uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)
	// attribute setup omitted for brevity
	mm.meshes[id] = vao
}

func (mm *MeshManager) GetVAO(id string) uint32 {
	return mm.meshes[id]
}
