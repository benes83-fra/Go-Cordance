package scene

import (
	"fmt"

	"go-engine/Go-Cordance/internal/ecs"
	"go-engine/Go-Cordance/internal/engine"
)

// SpawnGLTFScene loads the full glTF scene graph (nodes, transforms, meshes)
// and creates entities with Parent/Children + Transform.
func (s *Scene) SpawnGLTFScene(mm *engine.MeshManager, path string) ([]*ecs.Entity, error) {
	// 1. Geometry
	if err := mm.RegisterGLTFMulti(path); err != nil {
		return nil, fmt.Errorf("geometry load failed: %w", err)
	}

	// 2. Materials metadata
	mats, err := engine.LoadGLTFMaterialsMulti(path)
	if err != nil {
		return nil, fmt.Errorf("material load failed: %w", err)
	}

	// 3. Full glTF root (for nodes)
	root, err := engine.LoadGLTFRoot(path)
	if err != nil {
		return nil, fmt.Errorf("gltf root load failed: %w", err)
	}

	var entities []*ecs.Entity
	entityByNode := make(map[int]*ecs.Entity)

	// Recursive node spawner
	var spawnNode func(nodeIndex int, parent *ecs.Entity)

	spawnNode = func(nodeIndex int, parent *ecs.Entity) {
		n := root.Nodes[nodeIndex]

		// Create entity
		ent := s.AddEntity()

		// Local transform from node TRS/matrix
		M := engine.ComposeNodeTransform(n)
		ent.AddComponent(ecs.NewTransformFromMatrix(M))

		// Hierarchy: parent
		if parent != nil {
			ent.AddComponent(ecs.NewParent(parent))

			// Ensure parent has Children component
			children, ok := parent.GetComponent((*ecs.Children)(nil)).(*ecs.Children)
			if !ok || children == nil {
				children = ecs.NewChildren()
				parent.AddComponent(children)
			}
			children.AddChild(ent)
		}

		// Mesh & material (first primitive only for now)
		if n.Mesh >= 0 && n.Mesh < len(root.Meshes) {
			mesh := root.Meshes[n.Mesh]
			meshName := mesh.Name
			if meshName == "" {
				meshName = fmt.Sprintf("mesh_%d", n.Mesh)
			}
			meshID := fmt.Sprintf("%s/0", meshName)

			ent.AddComponent(ecs.NewMesh(meshID))

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
					break
				}
			}
		}

		entityByNode[nodeIndex] = ent
		entities = append(entities, ent)

		// Children
		for _, child := range n.Children {
			spawnNode(child, ent)
		}
	}

	// Default scene index
	sceneIndex := root.Scene
	if sceneIndex < 0 || sceneIndex >= len(root.Scenes) {
		return nil, fmt.Errorf("invalid default scene index")
	}

	for _, nodeIndex := range root.Scenes[sceneIndex].Nodes {
		spawnNode(nodeIndex, nil)
	}

	return entities, nil
}
