package scene

import (
	"fmt"
	"go-engine/Go-Cordance/internal/ecs"
	"go-engine/Go-Cordance/internal/engine"
)

// SpawnGLTFScene loads the full glTF scene graph (nodes, transforms, meshes).
func (s *Scene) SpawnGLTFScene(mm *engine.MeshManager, path string) ([]*ecs.Entity, error) {
	// Load geometry
	if err := mm.RegisterGLTFMulti(path); err != nil {
		return nil, err
	}

	// Load materials
	mats, err := engine.LoadGLTFMaterialsMulti(path)
	if err != nil {
		return nil, err
	}

	// Load glTF root (for nodes)
	root, err := engine.LoadGLTFRoot(path)
	if err != nil {
		return nil, err
	}

	var entities []*ecs.Entity

	// Recursive function to spawn nodes
	var spawnNode func(nodeIndex int, parent *ecs.Entity)

	spawnNode = func(nodeIndex int, parent *ecs.Entity) {
		n := root.Nodes[nodeIndex]

		// Create entity
		ent := s.AddEntity()

		// Transform
		M := engine.ComposeNodeTransform(n)
		ent.AddComponent(ecs.NewTransformFromMatrix(M))

		// Mesh?
		if n.Mesh >= 0 && n.Mesh < len(root.Meshes) {
			meshName := root.Meshes[n.Mesh].Name
			meshID := fmt.Sprintf("%s/0", meshName) // first primitive

			ent.AddComponent(ecs.NewMesh(meshID))

			// Apply material
			for _, m := range mats {
				if m.MeshID == meshID {
					mat := ecs.NewMaterial(m.BaseColor)
					ent.AddComponent(mat)

					if m.DiffuseTexturePath != "" {
						texID, _ := engine.LoadTexture(m.DiffuseTexturePath)
						ent.AddComponent(ecs.NewDiffuseTexture(texID))
					}
					if m.NormalTexturePath != "" {
						texID, _ := engine.LoadTexture(m.NormalTexturePath)
						ent.AddComponent(ecs.NewNormalMap(texID))
					}
				}
			}
		}

		entities = append(entities, ent)

		// Recurse children
		for _, child := range n.Children {
			spawnNode(child, ent)
		}
	}

	// Start from default scene
	sceneIndex := root.Scene
	if sceneIndex < 0 || sceneIndex >= len(root.Scenes) {
		return nil, fmt.Errorf("invalid default scene index")
	}

	for _, nodeIndex := range root.Scenes[sceneIndex].Nodes {
		spawnNode(nodeIndex, nil)
	}

	return entities, nil
}
