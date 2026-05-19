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
		skinComp := e.GetComponent((*Skin)(nil))
		if skinComp == nil {
			continue
		}
		skin := skinComp.(*Skin)

		meshTr := e.GetTransform()
		if meshTr == nil {
			continue
		}

		invMeshWorld := engine.InverseMat4(meshTr.WorldMatrix)

		if len(skin.JointMatrices) != len(skin.JointEntities) {
			skin.JointMatrices = make([][16]float32, len(skin.JointEntities))
		}

		for i := range skin.Joints {
			jointEnt := skin.JointEntities[i]

			if jointEnt == nil {
				skin.JointMatrices[i] = skin.InverseBindMatrices[i]
				continue
			}

			tr := jointEnt.GetTransform()
			if tr == nil {
				skin.JointMatrices[i] = skin.InverseBindMatrices[i]
				continue
			}

			// ← ADD THIS: Check for zero or near-zero scale
			if tr.Scale[0] < 0.001 && tr.Scale[1] < 0.001 && tr.Scale[2] < 0.001 {
				skin.JointMatrices[i] = skin.InverseBindMatrices[i]
				continue
			}

			jointWorld := tr.WorldMatrix
			modelSpaceJoint := engine.MulMat4(invMeshWorld, jointWorld)
			skin.JointMatrices[i] = engine.MulMat4(modelSpaceJoint, skin.InverseBindMatrices[i])
		}
	}
}
