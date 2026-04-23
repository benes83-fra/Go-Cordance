package gltf

import (
	"go-engine/Go-Cordance/internal/assets"
	"go-engine/Go-Cordance/internal/ecs"
	"go-engine/Go-Cordance/internal/engine"
	"go-engine/Go-Cordance/internal/scene"
	"log"
	"path/filepath"
	"strings"
)

func LoadGLTFMulti(sc *scene.Scene, path string) (*ecs.Entity, error) {
	_, meshIDs, err := assets.ImportGLTFMulti(path, engine.GlobalMeshManager)
	if err != nil {
		return nil, err
	}

	trs, err := engine.ExtractGLTFMeshTRS(path)
	if err != nil {
		return nil, err
	}

	mats, err := engine.LoadGLTFMaterialsMulti(path)
	if err != nil {
		return nil, err
	}
	skins, err := ExtractGLTFSkins(path)
	if err != nil {
		// To keep this step "safe", you can log and continue instead of failing:
		// log.Printf("ExtractGLTFSkins(%s) failed: %v", path, err)
		// skins = nil
		return nil, err
	}
	matByMesh := map[string]engine.LoadedMeshMaterial{}
	for _, m := range mats {
		matByMesh[m.MeshID] = m
	}

	root := SpawnMultiMesh(sc, meshIDs, nil, trs)
	name := filepath.Base(path)                         // "sofa.gltf"
	name = strings.TrimSuffix(name, filepath.Ext(name)) // "sofa"
	root.AddComponent(ecs.NewName(name))
	children := root.GetComponent((*ecs.Children)(nil)).(*ecs.Children)

	for _, child := range children.Entities {
		mesh := child.GetComponent((*ecs.Mesh)(nil)).(*ecs.Mesh)
		// Step F: load JOINTS_0 + WEIGHTS_0 if present
		if js, ok := engine.GlobalMeshManager.JointData[mesh.ID]; ok {
			mesh.Joints = js
		}
		if ws, ok := engine.GlobalMeshManager.WeightData[mesh.ID]; ok {
			mesh.Weights = ws
		}

		info, ok := matByMesh[mesh.ID]
		if !ok {
			continue
		}

		// Ensure entity has a Material component
		matComp := child.GetComponent((*ecs.Material)(nil))
		if matComp == nil {
			m := ecs.NewMaterial(info.BaseColor)
			child.AddComponent(m)
			matComp = m
		}
		if skins != nil {
			if skinComp, ok := skins[mesh.ID]; ok {
				child.AddComponent(skinComp)
			}
		}
		m := matComp.(*ecs.Material)

		// Base color
		m.BaseColor = info.BaseColor

		// Store paths for debugging / editor
		m.DiffuseTexturePath = info.DiffuseTexturePath
		m.NormalTexturePath = info.NormalTexturePath
		m.OcclusionTexturePath = info.OcclusionTexturePath
		m.MetallicRoughnessTexturePath = info.MetallicRoughnessTexturePath

		// Texcoord mapping and UV transforms
		if info.TexCoordMap != nil {
			if m.TexCoordMap == nil {
				m.TexCoordMap = map[string]int{}
			}
			for k, v := range info.TexCoordMap {
				m.TexCoordMap[k] = v
			}
		}
		if info.UVScale != nil {
			if m.UVScale == nil {
				m.UVScale = map[string][2]float32{}
			}
			for k, v := range info.UVScale {
				m.UVScale[k] = v
			}
		}
		if info.UVOffset != nil {
			if m.UVOffset == nil {
				m.UVOffset = map[string][2]float32{}
			}
			for k, v := range info.UVOffset {
				m.UVOffset[k] = v
			}
		}

		// Normal scale
		if info.NormalScale != 0 {
			m.NormalScale = info.NormalScale
		}

		// Sheen / specular
		if info.SheenRoughness != 0 {
			m.SheenRoughness = info.SheenRoughness
		}
		if info.SheenColor != [3]float32{} {
			m.SheenColor = info.SheenColor
		}
		if info.SpecularFactor != 0 {
			m.SpecularFactor = info.SpecularFactor
		}

		// Import textures with correct color space
		// Base color / albedo -> sRGB
		// Base color (sRGB)
		if info.DiffuseTexturePath != "" {
			assetID, glID, err := assets.ImportTextureWithSRGB(info.DiffuseTexturePath, true)
			if err == nil {
				m.UseTexture = true
				m.TextureID = glID
				m.TextureAsset = assetID
			}
		}

		// Normal map (linear)
		if info.NormalTexturePath != "" {
			assetID, glID, err := assets.ImportTextureWithSRGB(info.NormalTexturePath, false)
			if err == nil {
				m.UseNormal = true
				m.NormalID = glID
				m.NormalAsset = assetID
			}
		}

		// Occlusion (linear)
		if info.OcclusionTexturePath != "" {
			assetID, glID, err := assets.ImportTextureWithSRGB(info.OcclusionTexturePath, false)
			if err == nil {
				m.OcclusionTexturePath = info.OcclusionTexturePath
				m.OcclusionAsset = assetID
				m.OcclusionID = glID
			}
		}

		// MetallicRoughness (linear)
		if info.MetallicRoughnessTexturePath != "" {
			assetID, glID, err := assets.ImportTextureWithSRGB(info.MetallicRoughnessTexturePath, false)
			if err == nil {
				m.MetallicRoughnessTexturePath = info.MetallicRoughnessTexturePath
				m.MetallicRoughnessAsset = assetID
				m.MetallicRoughnessID = glID
			}
		}

		// Mark material dirty so renderer/editor can pick up changes
		m.Dirty = true

	}

	return root, nil
}

func LoadGLTFMultiSkinned(
	sc *scene.Scene,
	path string,
) (*ecs.Entity, []*ecs.Entity, []*ecs.Entity, error) {

	root, err := LoadGLTFMulti(sc, path)
	if err != nil {
		return nil, nil, nil, err
	}

	g, _, err := engine.LoadGLTFOrGLB(path)
	if err != nil {
		return nil, nil, nil, err
	}

	nodeEntities := BuildNodeEntities(sc, g)

	// attach skeleton to root (CesiumMan entity)
	root.AddComponent(&ecs.Skeleton{
		Nodes: nodeEntities,
	})

	children := root.GetComponent((*ecs.Children)(nil)).(*ecs.Children)
	var skinEntities []*ecs.Entity
	for _, child := range children.Entities {
		if child.GetComponent((*ecs.Skin)(nil)) != nil {
			skinEntities = append(skinEntities, child)
		}
	}

	for _, skinEnt := range skinEntities {
		skin := skinEnt.GetComponent((*ecs.Skin)(nil)).(*ecs.Skin)
		if len(skin.JointEntities) != len(skin.Joints) {
			skin.JointEntities = make([]*ecs.Entity, len(skin.Joints))
		}

		for i, nodeIndex := range skin.Joints {
			if nodeIndex < 0 || nodeIndex >= len(nodeEntities) {
				log.Printf("Skin joint %d -> invalid node index %d", i, nodeIndex)
				continue
			}
			ent := nodeEntities[nodeIndex]
			if ent == nil {
				log.Printf("Skin joint %d -> node entity nil for node %d", i, nodeIndex)
			} else {
				log.Printf("Skin joint %d -> entity ID %d (node %d)", i, ent.ID, nodeIndex)
			}
			skin.JointEntities[i] = ent
		}

	}

	return root, nodeEntities, skinEntities, nil
}

func BuildNodeEntities(sc *scene.Scene, g *engine.GltfRoot) []*ecs.Entity {
	nodeEntities := make([]*ecs.Entity, len(g.Nodes))

	// 1. Create one ECS entity per glTF node
	for i, n := range g.Nodes {
		ent := sc.AddEntity()

		// Build the correct local matrix from glTF TRS or matrix
		localMat := engine.ComposeNodeTransform(n)

		// Decompose into TRS for your Transform component
		pos, rot, scale := engine.DecomposeTRS(localMat)

		tr := &ecs.Transform{
			Position: pos,
			Rotation: rot,
			Scale:    scale,
		}
		tr.RecalculateLocal()
		ent.AddComponent(tr)

		ent.AddComponent(ecs.NewChildren())
		nodeEntities[i] = ent
	}

	// 2. Build parent-child relationships
	for i, n := range g.Nodes {
		for _, childIdx := range n.Children {
			parent := nodeEntities[i]
			child := nodeEntities[childIdx]

			child.AddComponent(ecs.NewParent(parent))
			pc := parent.GetComponent((*ecs.Children)(nil)).(*ecs.Children)
			pc.AddChild(child)
		}
	}

	return nodeEntities
}
