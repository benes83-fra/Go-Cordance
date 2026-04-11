package ecs

import "go-engine/Go-Cordance/internal/engine"

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
func (sys *SkinningSystem) Update(dt float32, ents []*Entity) {
	for _, e := range ents {
		s := e.GetComponent((*Skin)(nil))
		if s == nil {
			continue
		}
		skin := s.(*Skin)

		// Ensure JointMatrices slice exists
		if len(skin.JointMatrices) != len(skin.Joints) {
			skin.JointMatrices = make([][16]float32, len(skin.Joints))
		}

		for i, jointNodeIndex := range skin.Joints {
			jointEnt := sys.world.FindByID(int64(jointNodeIndex))
			if jointEnt == nil {
				continue
			}

			tr := jointEnt.GetComponent((*Transform)(nil))
			if tr == nil {
				continue
			}

			world := tr.(*Transform).WorldMatrix

			skin.JointMatrices[i] = engine.MulMat4(world, skin.InverseBindMatrices[i])
		}
	}
}
