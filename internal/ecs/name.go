package ecs

// Name is a simple component that gives an entity a human readable label.
type Name struct {
	Value string
}

func NewName(v string) *Name { return &Name{Value: v} }

func (n *Name) Update(dt float32) {
	_ = dt
}

func (n *Name) EditorName() string { return "Name" }

func (n *Name) EditorFields() map[string]any {
	return map[string]any{
		"Value": n.Value,
		//"Materials": mm.Materials,
	}
}

func (n *Name) SetEditorField(name string, value any) {
	switch name {
	case "Meshes":
		n.Value = value.(string)
	}
}
