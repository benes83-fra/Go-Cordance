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
			// if i < 5 {
			// 	fmt.Printf("Joint %d final matrix: %v\n", i, skin.JointMatrices[i])
			// }
		}
	}
}
