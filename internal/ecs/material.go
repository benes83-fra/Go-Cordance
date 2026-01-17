package ecs

// Material holds surface properties for lighting/shading.
type Material struct {
	BaseColor [4]float32 // RGBA
	Ambient   float32    // ambient multiplier
	Diffuse   float32    // diffuse multiplier
	Specular  float32    // specular multiplier
	Shininess float32    // specular exponent

	UseTexture bool
	TextureID  uint32
	UseNormal  bool
	NormalID   uint32
}

func NewMaterial(color [4]float32) *Material {
	return &Material{
		BaseColor: color,
		Ambient:   0.2,
		Diffuse:   0.8,
		Specular:  0.5,
		Shininess: 32.0,
	}
}

func (m *Material) Update(dt float32) { _ = dt }

func (m *Material) EditorName() string { return "Material" }

func (m *Material) EditorFields() map[string]any {
	return map[string]any{
		"BaseColor":  m.BaseColor,
		"Ambient":    m.Ambient,
		"Diffuse":    m.Diffuse,
		"Specular":   m.Specular,
		"Shininess":  m.Shininess,
		"UseTexture": m.UseTexture,
		"UseNormal":  m.UseNormal,
	}
}

func (m *Material) SetEditorField(name string, value any) {
	switch name {
	case "BaseColor":
		m.BaseColor = toVec4(value)
	case "Ambient":
		m.Ambient = toFloat32(value)
	case "Diffuse":
		m.Diffuse = toFloat32(value)
	case "Specular":
		m.Specular = toFloat32(value)
	case "Shininess":
		m.Shininess = toFloat32(value)
	case "UseTexture":
		m.UseNormal = toBool(value)
	case "UseNormal":
		m.UseNormal = toBool(value)
	}
}
