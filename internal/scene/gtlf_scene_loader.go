package scene

import (
	"fmt"

	"go-engine/Go-Cordance/internal/ecs"
	"go-engine/Go-Cordance/internal/engine"
)

// SpawnGLTFScene loads the full glTF scene graph (nodes, transforms, meshes)
// and creates entities with Parent/Children + Transform + Mesh + Material.
func (s *Scene) SpawnGLTFScene(mm *engine.MeshManager, path string) ([]*ecs.Entity, error) {
	// 1. Load geometry for all meshes/primitives
	_, err := mm.RegisterGLTFMulti(path)
	if err != nil {
		return nil, fmt.Errorf("geometry load failed: %w", err)
	}

	// 2. Load material metadata for all primitives
	mats, err := engine.LoadGLTFMaterialsMulti(path)
	if err != nil {
		return nil, fmt.Errorf("material load failed: %w", err)
	}

	// Build quick lookup: meshID -> material info
	matByMeshID := make(map[string]engine.LoadedMeshMaterial)
	for _, m := range mats {
		matByMeshID[m.MeshID] = m
	}

	// 3. Load full glTF root (for nodes, scenes, meshes)
	root, err := engine.LoadGLTFRoot(path)
	if err != nil {
		return nil, fmt.Errorf("gltf root load failed: %w", err)
	}

	var entities []*ecs.Entity

	// Recursive node spawner
	var spawnNode func(nodeIndex int, parent *ecs.Entity)

	spawnNode = func(nodeIndex int, parent *ecs.Entity) {
		n := root.Nodes[nodeIndex]

		// --- Node entity (holds the transform + hierarchy) ---
		nodeEnt := s.AddEntity()

		// Local transform from glTF node TRS/matrix
		M := engine.ComposeNodeTransform(n)
		nodeEnt.AddComponent(ecs.NewTransformFromMatrix(M))

		// Parent/Children wiring
		if parent != nil {
			nodeEnt.AddComponent(ecs.NewParent(parent))

			// ensure parent has Children
			if chComp := parent.GetComponent((*ecs.Children)(nil)); chComp != nil {
				if children, ok := chComp.(*ecs.Children); ok {
					children.AddChild(nodeEnt)
				}
			} else {
				children := ecs.NewChildren()
				children.AddChild(nodeEnt)
				parent.AddComponent(children)
			}
		}

		entities = append(entities, nodeEnt)

		// --- Mesh entities (one per primitive under this node's mesh) ---
		if n.Mesh >= 0 && n.Mesh < len(root.Meshes) {
			mesh := root.Meshes[n.Mesh]
			meshName := mesh.Name
			if meshName == "" {
				meshName = fmt.Sprintf("mesh_%d", n.Mesh)
			}

			for pi := range mesh.Primitives {
				meshID := fmt.Sprintf("%s/%d", meshName, pi)

				meshEnt := s.AddEntity()

				// Transform: local identity, parented to nodeEnt
				meshEnt.AddComponent(ecs.NewTransform([3]float32{0, 0, 0}))
				meshEnt.AddComponent(ecs.NewParent(nodeEnt))

				// add meshEnt to nodeEnt's Children
				if chComp := nodeEnt.GetComponent((*ecs.Children)(nil)); chComp != nil {
					if children, ok := chComp.(*ecs.Children); ok {
						children.AddChild(meshEnt)
					}
				} else {
					children := ecs.NewChildren()
					children.AddChild(meshEnt)
					nodeEnt.AddComponent(children)
				}

				// Mesh component (matches MeshManager IDs)
				meshEnt.AddComponent(ecs.NewMesh(meshID))

				// Material / textures if present
				if info, ok := matByMeshID[meshID]; ok {
					mat := ecs.NewMaterial(info.BaseColor)
					meshEnt.AddComponent(mat)

					if info.DiffuseTexturePath != "" {
						texID, _ := engine.LoadTexture(info.DiffuseTexturePath)
						meshEnt.AddComponent(ecs.NewDiffuseTexture(texID))
					}
					if info.NormalTexturePath != "" {
						texID, _ := engine.LoadTexture(info.NormalTexturePath)
						meshEnt.AddComponent(ecs.NewNormalMap(texID))
					}
				}

				entities = append(entities, meshEnt)
			}
		}

		// --- Recurse into children nodes ---
		for _, child := range n.Children {
			spawnNode(child, nodeEnt)
		}
	}

	// 4. Start from default scene root nodes
	sceneIndex := root.Scene
	if sceneIndex < 0 || sceneIndex >= len(root.Scenes) {
		// fall back to scene 0 if invalid
		sceneIndex = 0
	}
	for _, nodeIndex := range root.Scenes[sceneIndex].Nodes {
		spawnNode(nodeIndex, nil)
	}

	return entities, nil
}
