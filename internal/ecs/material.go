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

	DiffuseTexturePath           string
	NormalTexturePath            string
	OcclusionTexturePath         string
	MetallicRoughnessTexturePath string

	TexCoordMap map[string]int        // e.g. "baseColor":0, "occlusion":1
	UVScale     map[string][2]float32 // per-texture uv scale
	UVOffset    map[string][2]float32 // per-texture uv offset

	NormalScale    float32
	SheenColor     [3]float32
	SheenRoughness float32
	SpecularFactor float32
	OcclusionAsset assets.AssetID
	OcclusionID    uint32

	MetallicRoughnessAsset assets.AssetID
	MetallicRoughnessID    uint32
	UseIBL                 bool
	IrradianceTex          uint32
	PrefilteredEnvTex      uint32
	BRDFLUTTex             uint32
	ClearcoatFactor        float32
	ClearcoatRoughness     float32
	ClearcoatTexture       uint32
	ClearcoatRoughTex      uint32
	ClearcoatNormalTex     uint32
	UseClearcoat           bool
	TransmissionFactor     float32
	UseTransmission        bool
	TransmissionTex        uint32 // optional

	// (optional) OcclusionID/MetallicRoughnessID can be zero if not present

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

		// New fields for editor (read-only or editable later)
		"DiffuseTexturePath":           m.DiffuseTexturePath,
		"NormalTexturePath":            m.NormalTexturePath,
		"OcclusionTexturePath":         m.OcclusionTexturePath,
		"MetallicRoughnessTexturePath": m.MetallicRoughnessTexturePath,
		"OcclusionAsset":               m.OcclusionAsset,
		"OcclusionID":                  m.OcclusionID,
		"MetallicRoughnessAsset":       m.MetallicRoughnessAsset,
		"MetallicRoughnessID":          m.MetallicRoughnessID,
		"TexCoordMap":                  m.TexCoordMap,
		"UVScale":                      m.UVScale,
		"UVOffset":                     m.UVOffset,

		"NormalScale":        m.NormalScale,
		"SheenColor":         m.SheenColor,
		"SheenRoughness":     m.SheenRoughness,
		"SpecularFactor":     m.SpecularFactor,
		"UseIBL":             m.UseIBL,
		"IrradianceTex":      m.IrradianceTex,
		"PrefilteredEnvTex":  m.PrefilteredEnvTex,
		"BRDFLUTTex":         m.BRDFLUTTex,
		"ClearcoatFactor":    m.ClearcoatFactor,
		"ClearcoatRoughness": m.ClearcoatRoughness,
		"ClearcoatTexture":   m.ClearcoatTexture,
		"ClearcoatRoughTex":  m.ClearcoatRoughTex,
		"ClearcoatNormalTex": m.ClearcoatNormalTex,
		"UseClearcoat":       m.UseClearcoat,
		"TransmissionFactor": m.TransmissionFactor,
		"UseTransmission":    m.UseTransmission,
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
	case "DiffuseTexturePath":
		m.DiffuseTexturePath = value.(string)
	case "NormalTexturePath":
		m.NormalTexturePath = value.(string)
	case "OcclusionTexturePath":
		m.OcclusionTexturePath = value.(string)
	case "MetallicRoughnessTexturePath":
		m.MetallicRoughnessTexturePath = value.(string)
	case "OcclusionAsset":
		m.OcclusionAsset = assets.AssetID(toInt(value))
	case "OcclusionID":
		m.OcclusionID = uint32(toInt(value))

	case "MetallicRoughnessAsset":
		m.MetallicRoughnessAsset = assets.AssetID(toInt(value))
	case "MetallicRoughnessID":
		m.MetallicRoughnessID = uint32(toInt(value))
	case "NormalScale":
		m.NormalScale = toFloat32(value)
	case "SheenRoughness":
		m.SheenRoughness = toFloat32(value)
	case "SheenColor":
		m.SheenColor = toVec3(value)
	case "SpecularFactor":
		m.SpecularFactor = toFloat32(value)
	case "TexCoordMap":
		if value == nil {
			m.TexCoordMap = nil
		} else if mp, ok := value.(map[string]int); ok {
			m.TexCoordMap = mp
		} else {
			// optional: log.Printf("Material.TexCoordMap: unexpected type %T", value)
		}
	case "UVScale":
		if value == nil {
			m.UVScale = nil
		} else if mp, ok := value.(map[string][2]float32); ok {
			m.UVScale = mp
		}
	case "UVOffset":
		if value == nil {
			m.UVOffset = nil
		} else if mp, ok := value.(map[string][2]float32); ok {
			m.UVOffset = mp
		}

	case "UseIBL":
		m.UseIBL = toBool(value)
	case "IrradianceTex":
		m.IrradianceTex = uint32(toInt(value))
	case "PrefilteredTex":
		m.PrefilteredEnvTex = uint32(toInt(value))
	case "BRFFLUTTex":
		m.BRDFLUTTex = uint32(toInt(value))
	case "ClearcoatFactor":
		m.ClearcoatFactor = toFloat32(value)
	case "ClearcoatRoughness":
		m.ClearcoatRoughness = toFloat32(value)
	case "ClearcoatTexture":
		m.ClearcoatTexture = uint32(toInt(value))
	case "ClearcoatRoughTex":
		m.ClearcoatRoughTex = uint32(toInt(value))
	case "ClearcoatNormalTex":
		m.ClearcoatNormalTex = uint32(toInt(value))
	case "UseClearcoat":
		m.UseClearcoat = toBool(value)
	case "TransmissionFactor":
		m.TransmissionFactor = toFloat32(value)
	case "UseTransmission":
		m.UseTransmission = toBool(value)

	}

	m.Dirty = true
}

// in ecs/material.go or next to selectMaterialShader

func resolveMaterialShader(mat *Material) {
	if mat == nil {
		return
	}
	if mat.ShaderName == "" {
		mat.Shader = nil
		return
	}

	sp, err := engine.GetShaderProgram(mat.ShaderName)
	if err != nil {
		// fallback: no shader, keep default
		mat.Shader = nil
		return
	}
	mat.Shader = sp
}
