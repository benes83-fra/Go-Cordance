package ecs

// Material holds surface properties for lighting/shading.
type Material struct {
	BaseColor [4]float32 // RGBA
	Ambient   float32    // ambient multiplier
	Diffuse   float32    // diffuse multiplier
	Specular  float32    // specular multiplier
	Shininess float32    // specular exponent
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
