package ecs

type EditorInspectable interface {
	EditorName() string
	EditorFields() map[string]any
	SetEditorField(name string, value any)
}
