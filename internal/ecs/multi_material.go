package ecs

type MultiMaterial struct {
	Materials map[string]*Material // key = meshID (e.g. "Teapot/0")
}

func NewMultiMaterial() *MultiMaterial {
	return &MultiMaterial{
		Materials: make(map[string]*Material),
	}
}

func (multmat *MultiMaterial) Update(dt float32) { _ = dt }

func (mm *MultiMaterial) EditorName() string { return "MultiMaterial" }

func (mm *MultiMaterial) EditorFields() map[string]any {
	return map[string]any{
		"Materials": mm.Materials,
	}
}

func (mm *MultiMaterial) SetEditorField(name string, value any) {
	switch name {
	case "Materials":
		mm.Materials = value.(map[string]*Material)
	}
}
