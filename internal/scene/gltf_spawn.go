package scene

import (
	"go-engine/Go-Cordance/internal/ecs"
	"go-engine/Go-Cordance/internal/engine"
)

func SpawnGLTF(scene *Scene, mm *engine.MeshManager, id, path string) (*ecs.Entity, error) {
	// engine handles geometry
	_, err := mm.RegisterGLTF(id, path)
	if err != nil {
		return nil, err
	}

	// engine provides material metadata
	mats, err := engine.LoadGLTFMaterials(id, path)
	if err != nil {
		return nil, err
	}
	matInfo := mats[0]

	// ECS creates entity
	ent := scene.AddEntity()
	ent.AddComponent(ecs.NewMesh(id))

	// ECS creates Material
	ent.AddComponent(ecs.NewMaterial(matInfo.BaseColor))

	// ECS loads textures using engine
	if matInfo.DiffuseTexturePath != "" {
		texID, _ := engine.LoadTexture(matInfo.DiffuseTexturePath)
		ent.AddComponent(ecs.NewDiffuseTexture(texID))
	}

	if matInfo.NormalTexturePath != "" {
		texID, _ := engine.LoadTexture(matInfo.NormalTexturePath)
		ent.AddComponent(ecs.NewNormalMap(texID))
	}

	ent.AddComponent(ecs.NewTransform([3]float32{0, 0, 0}))

	return ent, nil
}
