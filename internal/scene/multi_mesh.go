package scene

import (
	"go-engine/Go-Cordance/internal/ecs"
)

func SpawnMultiMesh(
	sc *Scene,
	meshIDs []string,
	materials map[string]*ecs.Material,
) *ecs.Entity {

	root := sc.AddEntity()
	root.AddComponent(&ecs.Transform{
		Position: [3]float32{1, 0, -3},
		Rotation: [4]float32{1, 0, 0, 0}, // identity quat
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
		child.AddComponent(&ecs.Transform{
			Position: [3]float32{1, 0, -3},
			Rotation: [4]float32{1, 0, 0, 0},
			Scale:    [3]float32{1, 1, 1},
		})

		child.AddComponent(&ecs.Mesh{
			ID:       meshID,
			MeshName: meshID,
		})

		if mat, ok := materials[meshID]; ok {
			child.AddComponent(mat)
		}

		child.AddComponent(ecs.NewParent(root))
		children.AddChild(child)
	}

	return root
}
