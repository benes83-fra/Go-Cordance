package ecs

// Mesh is a component that references a GPU mesh by ID.
type Mesh struct {
	ID       string // key into MeshManager
	MeshName string
	Joints   [][4]uint16
	Weights  [][4]float32
}

func NewMesh(id string) *Mesh {
	return &Mesh{ID: id,
		Joints:  nil,
		Weights: nil}
}
func (m *Mesh) Update(dt float32) { _ = dt }

func (m *Mesh) EditorName() string { return "Mesh" }

func (m *Mesh) EditorFields() map[string]any {

	return map[string]any{
		"MeshName": m.MeshName,
		"MeshID":   m.ID,
		"Joints":   m.Joints,
		"Weights":  m.Weights,
	}
}

func (m *Mesh) SetEditorField(name string, value any) {
	switch name {
	case "MeshName":
		m.MeshName = value.(string)

	case "MeshID":
		m.ID = value.(string)

	case "Joints":
		if arr, ok := value.([][4]uint16); ok {
			m.Joints = arr
		}

	case "Weights":
		if arr, ok := value.([][4]float32); ok {
			m.Weights = arr
		}
	}
}
