package ecs

type MultiMesh struct {
	Meshes []string
}

func NewMultiMesh(meshes []string) *MultiMesh {
	return &MultiMesh{Meshes: meshes}
}

func (mm *MultiMesh) Update(dt float32) {
	_ = dt
}
func (mm *MultiMesh) EditorName() string { return "MultiMesh" }

func (mm *MultiMesh) EditorFields() map[string]any {
	return map[string]any{
		"Meshes": mm.Meshes,
	}
}

func (mm *MultiMesh) SetEditorField(name string, value any) {
	switch name {
	case "Meshes":
		mm.Meshes = value.([]string)
	}
}
