package ecs

import (
	"go-engine/Go-Cordance/internal/assets"
	"go-engine/Go-Cordance/internal/engine"
)

// Material holds surface properties for lighting/shading.
type Material struct {
	BaseColor [4]float32 // RGBA
	Ambient   float32
	Diffuse   float32
	Specular  float32
	Shininess float32
	Metallic  float32
	Roughness float32

	Type int

	// --- Existing inspector workflow (kept intact) ---
	UseTexture bool
	TextureID  uint32 // raw GL texture ID (inspector uses this)
	UseNormal  bool
	NormalID   uint32 // raw GL normal map ID

	// --- New asset pipeline fields (optional, non-breaking) ---
	TextureAsset assets.AssetID // future: replace TextureID
	NormalAsset  assets.AssetID // future: replace NormalID
	ShaderName   string
	Shader       *engine.ShaderProgram

	Dirty bool
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
		"BaseColor": m.BaseColor,
		"Ambient":   m.Ambient,
		"Diffuse":   m.Diffuse,
		"Specular":  m.Specular,
		"Shininess": m.Shininess,
		"Metallic":  m.Metallic,
		"Roughness": m.Roughness,
		"Type":      m.Type,

		// Inspector-visible fields (unchanged)
		"UseTexture":   m.UseTexture,
		"UseNormal":    m.UseNormal,
		"TextureID":    m.TextureID,
		"NormalID":     m.NormalID,
		"TextureAsset": m.TextureAsset,
		"NormalAsset":  m.NormalAsset,

		"ShaderName": m.ShaderName,

		// Asset pipeline fields (hidden from inspector for now)
		// They can be exposed later when the editor supports asset picking.
		// "TextureAsset": m.TextureAsset,
		// "NormalAsset":  m.NormalAsset,
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
	case "Metallic":
		m.Metallic = toFloat32(value)
	case "Roughness":
		m.Roughness = toFloat32(value)
	case "Type":
		m.Type = toInt(value)
	// --- Inspector workflow (kept intact) ---
	case "UseTexture":
		m.UseTexture = toBool(value)
	case "UseNormal":
		m.UseNormal = toBool(value)
	case "TextureID":
		m.TextureID = uint32(toInt(value))
	case "NormalID":
		m.NormalID = uint32(toInt(value))
	case "TextureAsset":
		m.TextureAsset = assets.AssetID(toInt(value))
	case "NormalAsset":
		m.NormalAsset = assets.AssetID(toInt(value))
	case "ShaderName":
		m.ShaderName = value.(string)
		if m.ShaderName != "" {
			m.Shader = engine.MustGetShaderProgram(m.ShaderName)
		} else {
			m.Shader = nil
		}

		// --- Asset pipeline fields (future use) ---
		// case "TextureAsset":
		//     m.TextureAsset = assets.AssetID(toInt(value))
		// case "NormalAsset":
		//     m.NormalAsset = assets.AssetID(toInt(value))
	}

	m.Dirty = true
}
