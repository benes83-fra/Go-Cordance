// engine/reload.go
package loader

type AssetReloadRequest struct {
	Textures bool
	Meshes   bool
}

var AssetReloadChan = make(chan AssetReloadRequest, 4)
