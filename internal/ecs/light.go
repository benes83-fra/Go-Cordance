package ecs

type LightType int

const (
	LightDirectional LightType = iota
	LightPoint
	LightSpot
)

type LightComponent struct {
	Type      LightType
	Color     [3]float32
	Intensity float32
	Range     float32 // for point/spot
	Angle     float32 // for spot
	version   uint64
}

func NewLightComponent() *LightComponent {
	return &LightComponent{
		Type:      LightDirectional,
		Color:     [3]float32{1, 1, 1},
		Intensity: 1.0,
		Range:     10.0,
		Angle:     30.0,
	}
}

func (l *LightComponent) Update(dt float32) { _ = dt }

func (l *LightComponent) EditorName() string { return "Light" }

func (l *LightComponent) EditorFields() map[string]any {
	return map[string]any{
		"Type":      int(l.Type),
		"Color":     [3]float32{l.Color[0], l.Color[1], l.Color[2]},
		"Intensity": l.Intensity,
		"Range":     l.Range,
		"Angle":     l.Angle,
	}
}

func (l *LightComponent) SetEditorField(name string, value any) {
	switch name {
	case "Type":
		l.Type = LightType(toInt(value))
	case "Color":
		l.Color = toVec3(value)
	case "Intensity":
		l.Intensity = toFloat32(value)
	case "Range":
		l.Range = toFloat32(value)
	case "Angle":
		l.Angle = toFloat32(value)
	}
	l.version++

}

func (l *LightComponent) Version() uint64 { return l.version }
