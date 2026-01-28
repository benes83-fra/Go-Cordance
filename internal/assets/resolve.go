// internal/assets/resolve.go
package assets

func ResolveTextureGLID(id AssetID) uint32 {
	a := Get(id)
	if a == nil {
		return 0
	}
	if tex, ok := a.Data.(TextureData); ok {
		return tex.GLID
	}
	return 0
}
