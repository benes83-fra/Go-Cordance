package ecs

import (
	"go-engine/Go-Cordance/internal/engine"
	"log"
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
		if len(skin.Joints) != len(skin.InverseBindMatrices) {
			log.Printf("❌ Skin mismatch: joints=%d ibm=%d entity=%d",
				len(skin.Joints), len(skin.InverseBindMatrices), e.ID)
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

			// If scale is zero → neutralize scale in the world matrix
			if tr.Scale[0] < 0.001 || tr.Scale[1] < 0.001 || tr.Scale[2] < 0.001 {
				// Copy world matrix
				jointWorld := tr.WorldMatrix

				// Remove scale from the 3×3 rotation/scale block
				// Column-major: each column is a basis vector
				for c := 0; c < 3; c++ {
					// length of column vector
					x := jointWorld[c*4+0]
					y := jointWorld[c*4+1]
					z := jointWorld[c*4+2]
					l := float32(math.Sqrt(float64(x*x + y*y + z*z)))
					if l > 0.00001 {
						jointWorld[c*4+0] = x / l
						jointWorld[c*4+1] = y / l
						jointWorld[c*4+2] = z / l
					}
				}

				modelSpaceJoint := engine.MulMat4(invMeshWorld, jointWorld)
				skin.JointMatrices[i] = engine.MulMat4(modelSpaceJoint, skin.InverseBindMatrices[i])
				continue
			}

			jointWorld := tr.WorldMatrix
			modelSpaceJoint := engine.MulMat4(invMeshWorld, jointWorld)
			skin.JointMatrices[i] = engine.MulMat4(modelSpaceJoint, skin.InverseBindMatrices[i])
		}
	}
}
