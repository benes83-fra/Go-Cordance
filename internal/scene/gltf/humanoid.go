package gltf

import (
	"go-engine/Go-Cordance/internal/ecs"
	"go-engine/Go-Cordance/internal/engine"
)

type HumanoidBone int

const (
	HumanoidHips HumanoidBone = iota
	HumanoidSpine
	HumanoidChest
	HumanoidNeck
	HumanoidHead
	HumanoidLeftShoulder
	HumanoidLeftUpperArm
	HumanoidLeftLowerArm
	HumanoidLeftHand
	HumanoidRightShoulder
	HumanoidRightUpperArm
	HumanoidRightLowerArm
	HumanoidRightHand
	HumanoidLeftUpperLeg
	HumanoidLeftLowerLeg
	HumanoidLeftFoot
	HumanoidRightUpperLeg
	HumanoidRightLowerLeg
	HumanoidRightFoot
)

type HumanoidRig struct {
	Nodes      []*ecs.Entity
	BoneToNode map[HumanoidBone]int
	NodeToBone map[int]HumanoidBone
}

func BuildHumanoidRigFromGLTF(g *engine.GltfRoot, nodes []*ecs.Entity) *HumanoidRig {
	rig := &HumanoidRig{
		Nodes:      nodes,
		BoneToNode: make(map[HumanoidBone]int),
		NodeToBone: make(map[int]HumanoidBone),
	}

	find := func(pred func(name string) bool) int {
		for i, n := range g.Nodes {
			if pred(n.Name) {
				return i
			}
		}
		return -1
	}

	rig.BoneToNode[HumanoidHips] = find(func(n string) bool { return n == "pelvis" })
	rig.BoneToNode[HumanoidSpine] = find(func(n string) bool { return n == "spine_01" })
	rig.BoneToNode[HumanoidChest] = find(func(n string) bool { return n == "spine_03" })
	rig.BoneToNode[HumanoidNeck] = find(func(n string) bool { return n == "neck_01" })
	rig.BoneToNode[HumanoidHead] = find(func(n string) bool { return n == "head" })

	rig.BoneToNode[HumanoidLeftShoulder] = find(func(n string) bool { return n == "clavicle_l" })
	rig.BoneToNode[HumanoidLeftUpperArm] = find(func(n string) bool { return n == "upperarm_l" })
	rig.BoneToNode[HumanoidLeftLowerArm] = find(func(n string) bool { return n == "lowerarm_l" })
	rig.BoneToNode[HumanoidLeftHand] = find(func(n string) bool { return n == "hand_l" })

	rig.BoneToNode[HumanoidRightShoulder] = find(func(n string) bool { return n == "clavicle_r" })
	rig.BoneToNode[HumanoidRightUpperArm] = find(func(n string) bool { return n == "upperarm_r" })
	rig.BoneToNode[HumanoidRightLowerArm] = find(func(n string) bool { return n == "lowerarm_r" })
	rig.BoneToNode[HumanoidRightHand] = find(func(n string) bool { return n == "hand_r" })

	rig.BoneToNode[HumanoidLeftUpperLeg] = find(func(n string) bool { return n == "thigh_l" })
	rig.BoneToNode[HumanoidLeftLowerLeg] = find(func(n string) bool { return n == "calf_l" })
	rig.BoneToNode[HumanoidLeftFoot] = find(func(n string) bool { return n == "foot_l" })

	rig.BoneToNode[HumanoidRightUpperLeg] = find(func(n string) bool { return n == "thigh_r" })
	rig.BoneToNode[HumanoidRightLowerLeg] = find(func(n string) bool { return n == "calf_r" })
	rig.BoneToNode[HumanoidRightFoot] = find(func(n string) bool { return n == "foot_r" })

	for b, idx := range rig.BoneToNode {
		if idx >= 0 {
			rig.NodeToBone[idx] = b
		}
	}
	return rig
}

// Simple naive retargeting - copies keyframes directly
func RetargetClip(srcRig, dstRig *HumanoidRig, srcClip *ecs.AnimationClip) *ecs.AnimationClip {
	boneMap := BuildBoneMap(srcRig, dstRig)

	dstClip := &ecs.AnimationClip{
		Name:     srcClip.Name + "_retargeted",
		Duration: srcClip.Duration,
		Tracks:   []ecs.AnimationTrack{},
	}

	for _, track := range srcClip.Tracks {
		dstNode, ok := boneMap[track.NodeIndex]
		if !ok {
			continue
		}
		// Naive copy - may cause deformation but compiles and runs
		retargeted := track
		retargeted.NodeIndex = dstNode
		dstClip.Tracks = append(dstClip.Tracks, retargeted)
	}
	return dstClip
}

func BuildBoneMap(src, dst *HumanoidRig) map[int]int {
	out := map[int]int{}
	for bone, srcNode := range src.BoneToNode {
		dstNode := dst.BoneToNode[bone]
		if srcNode >= 0 && dstNode >= 0 {
			out[srcNode] = dstNode
		}
	}
	return out
}
