package scene

import (
	"go-engine/Go-Cordance/internal/ecs"
	"go-engine/Go-Cordance/internal/engine"
)

// MeshTRS is assumed to be:
// type MeshTRS struct {
//     Position [3]float32
//     Rotation [4]float32 // quaternion
//     Scale    [3]float32
// }

func SpawnMultiMesh(
	sc *Scene,
	meshIDs []string,
	materials map[string]*ecs.Material, // per meshID, optional
	trs map[string]engine.MeshTRS, // per meshID, optional, may be nil
) *ecs.Entity {

	// Root entity: placement of the whole multimesh in world space.
	root := sc.AddEntity()
	root.AddComponent(&ecs.Transform{
		Position: [3]float32{0, 0, 0},
		Rotation: [4]float32{1, 0, 0, 0},
		Scale:    [3]float32{1, 1, 1},
	})

	// MultiMesh: just the list of mesh IDs; renderer uses children for TRS.
	root.AddComponent(ecs.NewMultiMesh(meshIDs))

	// Optional "global" multimaterial on the root (e.g. for shared overrides).
	if materials != nil {
		mm := ecs.NewMultiMaterial()
		for k, v := range materials {
			mm.Materials[k] = v
		}
		root.AddComponent(mm)
	}

	// Children container on root.
	children := ecs.NewChildren()
	root.AddComponent(children)

	for _, meshID := range meshIDs {
		child := sc.AddEntity()

		// Default local TRS (relative to root) if no imported TRS is provided.
		t := engine.MeshTRS{
			Position: [3]float32{0, 0, 0},
			Rotation: [4]float32{1, 0, 0, 0},
			Scale:    [3]float32{1, 1, 1},
		}

		// Override from imported per-mesh TRS if available.
		if trs != nil {
			if v, ok := trs[meshID]; ok {
				t = v
			}
		}

		// Local transform of this submesh under the root.
		child.AddComponent(&ecs.Transform{
			Position: t.Position,
			Rotation: t.Rotation,
			Scale:    t.Scale,
		})

		// Mesh component: one submesh of the multimesh.
		child.AddComponent(&ecs.Mesh{
			ID:       meshID,
			MeshName: meshID,
		})

		// Name for editor/debugging.
		child.AddComponent(ecs.NewName(meshID))

		// Optional per-mesh material override.
		if materials != nil {
			if mat, ok := materials[meshID]; ok && mat != nil {
				child.AddComponent(mat)
			}
		}

		// Parenting: child under multimesh root.
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
