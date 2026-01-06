package ecs

type LightComponent struct {
	Color     [3]float32
	Intensity float32
}

func NewLightComponent() *LightComponent {
	return &LightComponent{
		Color:     [3]float32{1, 1, 1},
		Intensity: 1.0,
	}
}

func (l *LightComponent) Update(dt float32) { _ = dt }

func (l *LightComponent) EditorName() string { return "Light" }

func (l *LightComponent) EditorFields() map[string]any {
	return map[string]any{
		"Color":     l.Color,
		"Intensity": l.Intensity,
	}
}

func (l *LightComponent) SetEditorField(name string, value any) {
	switch name {
	case "Color":
		l.Color = toVec3(value)
	case "Intensity":
		l.Intensity = toFloat32(value)
	}
}
