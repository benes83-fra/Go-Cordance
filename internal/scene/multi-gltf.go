package scene

import (
	"go-engine/Go-Cordance/internal/assets"
	"go-engine/Go-Cordance/internal/ecs"
	"go-engine/Go-Cordance/internal/engine"
	"path/filepath"
	"strings"
)

func (sc *Scene) LoadGLTFMulti(path string) (*ecs.Entity, error) {
	_, meshIDs, err := assets.ImportGLTFMulti(path, engine.GlobalMeshManager)
	if err != nil {
		return nil, err
	}

	trs, err := engine.ExtractGLTFMeshTRS(path)
	if err != nil {
		return nil, err
	}

	mats, err := engine.LoadGLTFMaterialsMulti(path)
	if err != nil {
		return nil, err
	}

	matByMesh := map[string]engine.LoadedMeshMaterial{}
	for _, m := range mats {
		matByMesh[m.MeshID] = m
	}

	root := SpawnMultiMesh(sc, meshIDs, nil, trs)
	name := filepath.Base(path)                         // "sofa.gltf"
	name = strings.TrimSuffix(name, filepath.Ext(name)) // "sofa"
	root.AddComponent(ecs.NewName(name))
	children := root.GetComponent((*ecs.Children)(nil)).(*ecs.Children)

	for _, child := range children.Entities {
		mesh := child.GetComponent((*ecs.Mesh)(nil)).(*ecs.Mesh)

		info, ok := matByMesh[mesh.ID]
		if !ok {
			continue
		}

		mat := child.GetComponent((*ecs.Material)(nil))
		if mat == nil {
			m := ecs.NewMaterial(info.BaseColor)
			child.AddComponent(m)
			mat = m
		}

		m := mat.(*ecs.Material)
		m.BaseColor = info.BaseColor

		if info.DiffuseTexturePath != "" {
			assetID, glID, err := assets.ImportTexture(info.DiffuseTexturePath)
			if err == nil {
				m.UseTexture = true
				m.TextureID = glID
				m.TextureAsset = assetID
			}
		}

		if info.NormalTexturePath != "" {
			assetID, glID, err := assets.ImportTexture(info.NormalTexturePath)
			if err == nil {
				m.UseNormal = true
				m.NormalID = glID
				m.NormalAsset = assetID
			}
		}
	}

	return root, nil
}
