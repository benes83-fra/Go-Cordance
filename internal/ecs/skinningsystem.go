package ecs

import (
	"go-engine/Go-Cordance/internal/engine"
)

// SkinningSystem currently just exists as a hook for future CPU/GPU skinning.
// It does NOT change any rendering state yet.
type SkinningSystem struct {
	world *World
}

func NewSkinningSystem(world *World) *SkinningSystem {
	return &SkinningSystem{world: world}
}

// Update is intentionally a no-op for now. Later we can:
// - walk entities with *Skin (+ maybe *Mesh / *Transform)
// - compute joint matrices
// - stash them on Skin or a separate component for the renderer.
// func (sys *SkinningSystem) Update(dt float32, ents []*Entity) {
// 	for _, e := range ents {
// 		s := e.GetComponent((*Skin)(nil))
// 		if s == nil {
// 			continue
// 		}
// 		skin := s.(*Skin)

// 		// Ensure JointMatrices slice exists
// 		if len(skin.JointMatrices) != len(skin.Joints) {
// 			skin.JointMatrices = make([][16]float32, len(skin.Joints))
// 		}

// 		for i, jointNodeIndex := range skin.Joints {
// 			jointEnt := sys.world.FindByID(int64(jointNodeIndex))
// 			if jointEnt == nil {
// 				continue
// 			}

// 			tr := jointEnt.GetComponent((*Transform)(nil))
// 			if tr == nil {
// 				continue
// 			}

// 			world := tr.(*Transform).WorldMatrix

// 			skin.JointMatrices[i] = engine.MulMat4(world, skin.InverseBindMatrices[i])
// 			if i < 5 {
// 				fmt.Printf("Joint %d final matrix: %v\n", i, skin.JointMatrices[i])
// 			}

// 		}
// 	}
// }

// func (sys *SkinningSystem) Update(dt float32, ents []*Entity) {
// 	for _, e := range ents {
// 		s := e.GetComponent((*Skin)(nil))
// 		if s == nil {
// 			continue
// 		}
// 		skin := s.(*Skin)

// 		// Root (mesh) transform for this skin
// 		rootTr := e.GetComponent((*Transform)(nil))
// 		if rootTr == nil {
// 			continue
// 		}
// 		rootWorld := rootTr.(*Transform).WorldMatrix
// 		rootWorldInv := engine.InverseMat4(rootWorld)

// 		// Ensure JointMatrices slice exists
// 		if len(skin.JointMatrices) != len(skin.Joints) {
// 			skin.JointMatrices = make([][16]float32, len(skin.Joints))
// 		}

// 		for i, jointNodeIndex := range skin.Joints {
// 			jointEnt := sys.world.FindByID(int64(jointNodeIndex))
// 			if jointEnt == nil {
// 				continue
// 			}

// 			tr := jointEnt.GetComponent((*Transform)(nil))
// 			if tr == nil {
// 				continue
// 			}

// 			jointWorld := tr.(*Transform).WorldMatrix

//				// Convert joint world to mesh-local, then apply inverse bind:
//				// jointMatrix = M_mesh^-1 * M_joint * B^-1
//				jointInMesh := engine.MulMat4(rootWorldInv, jointWorld)
//				skin.JointMatrices[i] = engine.MulMat4(jointInMesh, skin.InverseBindMatrices[i])
//			}
//		}
//	}
// func (sys *SkinningSystem) Update(dt float32, ents []*Entity) {
// 	for _, e := range ents {
// 		s := e.GetComponent((*Skin)(nil))
// 		if s == nil {
// 			continue
// 		}
// 		skin := s.(*Skin)

// 		// --- 1) Ensure JointMatrices is sized by *node index*, not joint count ---
// 		maxNodeIndex := -1
// 		for _, nodeIdx := range skin.Joints {
// 			if nodeIdx > maxNodeIndex {
// 				maxNodeIndex = nodeIdx
// 			}
// 		}
// 		if maxNodeIndex < 0 {
// 			continue
// 		}
// 		if len(skin.JointMatrices) <= maxNodeIndex {
// 			skin.JointMatrices = make([][16]float32, maxNodeIndex+1)
// 		}

// 		// --- 2) Fill matrices at index = node index ---
// 		for i, jointNodeIndex := range skin.Joints {
// 			jointEnt := sys.world.FindByID(int64(jointNodeIndex))
// 			if jointEnt == nil {
// 				continue
// 			}

// 			tr := jointEnt.GetComponent((*Transform)(nil))
// 			if tr == nil {
// 				continue
// 			}

// 			jointWorld := tr.(*Transform).WorldMatrix

//				// jointMatrix[nodeIndex] = M_jointWorld * B^-1 (IBM[i])
//				skin.JointMatrices[jointNodeIndex] =
//					engine.MulMat4(jointWorld, skin.InverseBindMatrices[i])
//			}
//		}
//	}
func (sys *SkinningSystem) Update(dt float32, ents []*Entity) {
	for _, e := range ents {
		s := e.GetComponent((*Skin)(nil))
		if s == nil {
			continue
		}
		skin := s.(*Skin)

		// Ensure palette-sized joint matrices
		if len(skin.JointMatrices) != len(skin.Joints) {
			skin.JointMatrices = make([][16]float32, len(skin.Joints))
		}

		for i := range skin.Joints { // i = palette index
			jointEnt := skin.JointEntities[i]
			if jointEnt == nil {
				continue
			}

			tr := jointEnt.GetComponent((*Transform)(nil))
			if tr == nil {
				continue
			}

			jointWorld := tr.(*Transform).WorldMatrix

			// jointMatrix[paletteIndex] = M_jointWorld * B^-1
			skin.JointMatrices[i] =
				engine.MulMat4(jointWorld, skin.InverseBindMatrices[i])
		}
	}
}
