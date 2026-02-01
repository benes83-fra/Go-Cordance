package assets

import "go-engine/Go-Cordance/internal/engine"

// ImportGLTFMesh loads a single-mesh GLTF and registers it as an asset.
// Data = meshID string used by MeshManager.
func ImportGLTFMesh(meshID, path string, mm *engine.MeshManager) (AssetID, error) {
	if _, err := mm.RegisterGLTF(meshID, path); err != nil {
		return 0, err
	}
	return Register(AssetMesh, path, meshID), nil
}

// ImportGLTFMulti loads a multi-mesh GLTF and registers the root asset.
// Later you can extend this to register each primitive separately.
func ImportGLTFMulti(path string, mm *engine.MeshManager) (AssetID, []string, error) {
	meshIDs, err := mm.RegisterGLTFMulti(path)
	if err != nil {

		return 0, nil, err
	}
	id := Register(AssetMesh, path, meshIDs)
	return id, meshIDs, nil
}
