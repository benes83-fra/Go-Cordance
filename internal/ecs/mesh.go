package ecs

// Mesh is a component that references a GPU mesh by ID.
type Mesh struct {
	ID string // key into MeshManager
}

func NewMesh(id string) *Mesh     { return &Mesh{ID: id} }
func (m *Mesh) Update(dt float32) { _ = dt }
