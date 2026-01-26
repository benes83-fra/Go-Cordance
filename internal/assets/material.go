package assets

// Material assets are metadata-only for now.
// Later you can store shader, default textures, import settings, etc.

type MaterialInfo struct {
	BaseColor [4]float32
	Diffuse   string
	Normal    string
}

func RegisterMaterial(path string, info MaterialInfo) AssetID {
	return Register(AssetMaterial, path, info)
}
