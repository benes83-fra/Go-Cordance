package ecs

import (
	"go-engine/Go-Cordance/internal/engine"
	"math"
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

// func (sys *SkinningSystem) Update(dt float32, ents []*Entity) {
// 	// Optional: keep a very lightweight one-time log, but DO NOT modify IBMs.
// 	if !sys.checked {
// 		sys.checked = true
// 		for _, e := range ents {
// 			s := e.GetComponent((*Skin)(nil))
// 			if s == nil {
// 				continue
// 			}
// 			skin := s.(*Skin)
// 			log.Printf("Skin joints=%d, ibm=%d, jointEntities=%d",
// 				len(skin.Joints), len(skin.InverseBindMatrices), len(skin.JointEntities))
// 		}
// 	}

// 	// Normal per-frame joint matrix assembly
// 	for _, e := range ents {
// 		s := e.GetComponent((*Skin)(nil))
// 		if s == nil {
// 			continue
// 		}
// 		skin := s.(*Skin)

// 		if len(skin.JointMatrices) != len(skin.Joints) {
// 			skin.JointMatrices = make([][16]float32, len(skin.Joints))
// 		}

// 		for i := range skin.Joints {
// 			jointEnt := skin.JointEntities[i]
// 			if jointEnt == nil {
// 				continue
// 			}

// 			tr := jointEnt.GetComponent((*Transform)(nil))
// 			if tr == nil {
// 				continue
// 			}

// 			jointWorld := tr.(*Transform).WorldMatrix

// 			// jointMatrix = M_jointWorld * B^-1  (column-major, same as engine.MulMat4)
// 			skin.JointMatrices[i] = engine.MulMat4(jointWorld, skin.InverseBindMatrices[i])
// 		}
// 	}
// }

func identityError(m [16]float32) float32 {
	var err float32
	for r := 0; r < 4; r++ {
		for c := 0; c < 4; c++ {
			v := m[c*4+r]
			if r == c {
				err += float32(math.Abs(float64(v - 1.0)))
			} else {
				err += float32(math.Abs(float64(v)))
			}
		}
	}
	return err
}

func (sys *SkinningSystem) Update(dt float32, ents []*Entity) {
	for _, e := range ents {
		// SAFETY CHECK: Don't panic if entity doesn't have a skin
		skinComp := e.GetComponent((*Skin)(nil))
		if skinComp == nil {
			continue
		}
		skin := skinComp.(*Skin)

		meshTr := e.GetTransform()
		if meshTr == nil {
			continue
		}

		// Calculate Inverse World Matrix of the mesh holder
		invMeshWorld := engine.InverseMat4(meshTr.WorldMatrix)

		// Make sure JointMatrices slice is ready
		if len(skin.JointMatrices) != len(skin.JointEntities) {
			skin.JointMatrices = make([][16]float32, len(skin.JointEntities))
		}

		for i := range skin.Joints {
			jointEnt := skin.JointEntities[i]
			if jointEnt == nil {
				continue
			}

			tr := jointEnt.GetTransform()
			if tr == nil {
				continue
			}

			jointWorld := tr.WorldMatrix

			// Correct glTF Math:
			// JointMatrix = InverseMeshWorld * JointWorld * InverseBindMatrix
			modelSpaceJoint := engine.MulMat4(invMeshWorld, jointWorld)
			skin.JointMatrices[i] = engine.MulMat4(modelSpaceJoint, skin.InverseBindMatrices[i])
		}
	}
}
