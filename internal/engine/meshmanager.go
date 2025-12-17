// internal/engine/meshmanager.go
package engine

import "github.com/go-gl/gl/v4.1-core/gl"

type MeshManager struct {
	vaos map[string]uint32
}

func NewMeshManager() *MeshManager {
	return &MeshManager{vaos: make(map[string]uint32)}
}

// Triangle with only positions (layout location 0)
func (mm *MeshManager) RegisterTriangle(id string) {
	vertices := []float32{
		// x,    y,    z
		0.0, 0.5, 0.0,
		-0.5, -0.5, 0.0,
		0.5, -0.5, 0.0,
	}
	var vao, vbo uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(vertices), gl.Ptr(vertices), gl.STATIC_DRAW)

	// position at location 0
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, gl.PtrOffset(0))

	// unbind to avoid accidental state leaks
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)

	mm.vaos[id] = vao
}

func (mm *MeshManager) GetVAO(id string) uint32 { return mm.vaos[id] }
