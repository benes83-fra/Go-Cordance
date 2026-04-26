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
	Nodes      []*ecs.Entity        // same as Skeleton.Nodes
	BoneToNode map[HumanoidBone]int // bone -> glTF node index
	NodeToBone map[int]HumanoidBone // optional reverse map
}

// BuildHumanoidRigFromGLTF builds a rig using node names.
func BuildHumanoidRigFromGLTF(g *engine.GltfRoot, nodes []*ecs.Entity) *HumanoidRig {
	rig := &HumanoidRig{
		Nodes:      nodes,
		BoneToNode: make(map[HumanoidBone]int),
		NodeToBone: make(map[int]HumanoidBone),
	}

	// helper
	find := func(pred func(name string) bool) int {
		for i, n := range g.Nodes {
			if pred(n.Name) {
				return i
			}
		}
		return -1
	}

	// crawling_man naming (from your dump):
	// pelvis, spine_01, spine_02, spine_03, neck_01, head,
	// clavicle_l, upperarm_l, lowerarm_l, hand_l,
	// clavicle_r, upperarm_r, lowerarm_r, hand_r,
	// thigh_l, calf_l, foot_l,
	// thigh_r, calf_r, foot_r

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

	// build reverse map
	for b, idx := range rig.BoneToNode {
		if idx >= 0 {
			rig.NodeToBone[idx] = b
		}
	}

	return rig
}
func RetargetClip(srcRig, dstRig *HumanoidRig, srcClip *ecs.AnimationClip) *ecs.AnimationClip {
	// TEMP: just return the source clip for now
	// Next step will be: build a new clip with curves remapped bone-by-bone.
	return srcClip
}
