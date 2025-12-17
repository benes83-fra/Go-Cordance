package ecs

// Material holds shader and texture references.
type Material struct {
	ShaderID  string
	TextureID string
	Color     [4]float32
}

func NewMaterial(shader, texture string, color [4]float32) *Material {
	return &Material{ShaderID: shader, TextureID: texture, Color: color}
}
func (m *Material) Update(dt float32) { _ = dt }
