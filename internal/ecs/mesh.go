package ecs

// Mesh is a component that references a GPU mesh by ID.
type Mesh struct {
	ID       string // key into MeshManager
	MeshName string
}

func NewMesh(id string) *Mesh     { return &Mesh{ID: id} }
func (m *Mesh) Update(dt float32) { _ = dt }

func (m *Mesh) EditorName() string { return "Mesh" }

func (m *Mesh) EditorFields() map[string]any {

	return map[string]any{
		"MeshName": m.MeshName,
		"MeshID":   m.ID,
	}
}

func (m *Mesh) SetEditorField(name string, value any) {
	switch name {
	case "MeshName":
		m.MeshName = value.(string)
	case "MeshID":
		m.ID = value.(string)
	}
}
