package ecs

import (
	"go-engine/Go-Cordance/internal/engine"
)

// SkinningSystem currently just exists as a hook for future CPU/GPU skinning.
// It does NOT change any rendering state yet.
type SkinningSystem struct {
	world   *World
	checked bool
}

func NewSkinningSystem(world *World) *SkinningSystem {
	return &SkinningSystem{world: world}
}

func (sys *SkinningSystem) Update(dt float32, ents []*Entity) {
	// put this in SkinningSystem.Update but gated by a one-time flag or key

	// if !sys.checked {
	// 	sys.checked = true
	// 	for _, e := range ents {
	// 		s := e.GetComponent((*Skin)(nil))
	// 		if s == nil {
	// 			continue
	// 		}
	// 		skin := s.(*Skin)
	// 		log.Printf("Skin joints=%d, ibm=%d, jointEntities=%d", len(skin.Joints), len(skin.InverseBindMatrices), len(skin.JointEntities))
	// 		for i := range skin.Joints {
	// 			je := skin.JointEntities[i]
	// 			if je == nil {
	// 				log.Printf("JOINT MAPPING MISSING at index %d", i)
	// 				continue
	// 			}
	// 			tr := je.GetComponent((*Transform)(nil))
	// 			if tr == nil {
	// 				log.Printf("JOINT TRANSFORM MISSING at index %d", i)
	// 				continue
	// 			}
	// 			jointWorld := tr.(*Transform).WorldMatrix
	// 			ibm := skin.InverseBindMatrices[i]

	// 			// product = jointWorld * ibm
	// 			prod := engine.MulMat4(jointWorld, ibm)

	// 			// compute simple error metric: off-diagonal magnitude + |diag-1|
	// 			var err float32
	// 			for r := 0; r < 4; r++ {
	// 				for c := 0; c < 4; c++ {
	// 					v := prod[c*4+r] // column-major indexing used in MulMat4
	// 					if r == c {
	// 						err += float32(math.Abs(float64(v - 1.0)))
	// 					} else {
	// 						err += float32(math.Abs(float64(v)))
	// 					}
	// 				}
	// 			}
	// 			log.Printf("Joint %d product-error=%f", i, err)
	// 			if err > 0.01 {
	// 				log.Printf("Joint %d product != identity (prod=%v)", i, prod)
	// 				log.Printf("jointWorld[%d]: %v", i, jointWorld)
	// 				log.Printf("ibm[%d]: %v", i, ibm)
	// 			}
	// 		}
	// 	}
	// }

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
			skin.JointMatrices[i] = engine.MulMat4(jointWorld, skin.InverseBindMatrices[i])

		}
	}
}
