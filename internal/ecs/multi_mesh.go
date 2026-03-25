package ecs

type MultiMesh struct {
	Meshes []string
	//Materials map[string]*Material
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
		//"Materials": mm.Materials,
	}
}

func (mm *MultiMesh) SetEditorField(name string, value any) {
	switch name {
	case "Meshes":
		if value == nil {
			mm.Meshes = nil
		} else if meshes, ok := value.([]string); ok {
			mm.Meshes = meshes

		}

	case "Materials":
		//mm.Materials = value.(map[string]*Material)
	}
}
