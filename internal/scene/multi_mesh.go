package scene

import (
	"go-engine/Go-Cordance/internal/ecs"
	"go-engine/Go-Cordance/internal/engine"
)

func SpawnMultiMesh(
	sc *Scene,
	meshIDs []string,
	materials map[string]*ecs.Material,
	trs map[string]engine.MeshTRS, // optional, may be nil
) *ecs.Entity {

	// Root entity
	root := sc.AddEntity()
	root.AddComponent(&ecs.Transform{
		Position: [3]float32{0, 0, 0},
		Rotation: [4]float32{1, 0, 0, 0},
		Scale:    [3]float32{1, 1, 1},
	})

	root.AddComponent(ecs.NewMultiMesh(meshIDs))

	mm := ecs.NewMultiMaterial()
	for k, v := range materials {
		mm.Materials[k] = v
	}
	root.AddComponent(mm)

	children := ecs.NewChildren()
	root.AddComponent(children)

	for _, meshID := range meshIDs {
		child := sc.AddEntity()

		// Default TRS
		t := engine.MeshTRS{
			Position: [3]float32{0, 0, 0},
			Rotation: [4]float32{1, 0, 0, 0},
			Scale:    [3]float32{1, 1, 1},
		}

		// Override if provided
		if trs != nil {
			if v, ok := trs[meshID]; ok {
				t = v
			}
		}

		child.AddComponent(&ecs.Transform{
			Position: t.Position,
			Rotation: t.Rotation,
			Scale:    t.Scale,
		})

		child.AddComponent(&ecs.Mesh{
			ID:       meshID,
			MeshName: meshID,
		})

		child.AddComponent(ecs.NewName(meshID))

		if mat, ok := materials[meshID]; ok {
			child.AddComponent(mat)
		}

		child.AddComponent(ecs.NewParent(root))
		children.AddChild(child)
	}

	return root
}
func SpawnMultiMeshSimple(
	sc *Scene,
	meshIDs []string,
	materials map[string]*ecs.Material,
) *ecs.Entity {
	return SpawnMultiMesh(sc, meshIDs, materials, nil)
}
