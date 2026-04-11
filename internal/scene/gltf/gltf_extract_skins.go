package gltf

import (
	"fmt"
	"go-engine/Go-Cordance/internal/ecs"
	"go-engine/Go-Cordance/internal/engine"
)

// ExtractGLTFSkins builds Skin components keyed by meshID ("MeshName/primitiveIndex").
func ExtractGLTFSkins(path string) (map[string]*ecs.Skin, error) {
	g, buffers, err := engine.LoadGLTFOrGLB(path)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*ecs.Skin)

	// Pre-decode all skins' inverse bind matrices
	type skinData struct {
		joints []int
		ibm    [][16]float32
	}
	skins := make([]skinData, len(g.Skins))

	for si, s := range g.Skins {
		data := skinData{
			joints: append([]int(nil), s.Joints...),
		}

		if s.InverseBindMatrices >= 0 {
			acc, err := engine.GetAccessor(g, buffers, s.InverseBindMatrices)
			if err != nil {
				return nil, fmt.Errorf("skin %d inverseBindMatrices: %w", si, err)
			}
			if acc.Acc.Type != "MAT4" || acc.Acc.ComponentType != 5126 {
				return nil, fmt.Errorf("skin %d inverseBindMatrices: expected FLOAT MAT4", si)
			}

			for i := 0; i < acc.Acc.Count; i++ {
				off := acc.Base + i*acc.Stride
				var m [16]float32
				for c := 0; c < 16; c++ {
					m[c] = engine.BytesToFloat32(acc.Buf[off+4*c:])
				}
				data.ibm = append(data.ibm, m)
			}
		}

		skins[si] = data
	}

	// Map skins to meshIDs via nodes that reference them
	for _, n := range g.Nodes {
		if n.Skin < 0 || n.Skin >= len(g.Skins) {
			continue
		}
		if n.Mesh < 0 || n.Mesh >= len(g.Meshes) {
			continue
		}

		mesh := g.Meshes[n.Mesh]
		meshName := mesh.Name
		if meshName == "" {
			meshName = fmt.Sprintf("mesh_%d", n.Mesh)
		}

		sd := skins[n.Skin]
		skinComp := ecs.NewSkin(sd.joints, sd.ibm)

		// One Skin per primitive, keyed like RegisterGLTFMulti / loadGLTFInternal
		for pi := range mesh.Primitives {
			meshID := fmt.Sprintf("%s/%d", meshName, pi)
			result[meshID] = skinComp
		}
	}

	return result, nil
}
