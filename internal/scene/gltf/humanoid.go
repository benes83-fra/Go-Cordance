package gltf

import (
	"go-engine/Go-Cordance/internal/ecs"
	"go-engine/Go-Cordance/internal/engine"
)

// package gltf or ecs/humanoid.go

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
	// extend later (toes, fingers, etc.)
)

type HumanoidRig struct {
	Nodes      []*ecs.Entity        // same as Skeleton.Nodes
	BoneToNode map[HumanoidBone]int // bone -> glTF node index
	NodeToBone map[int]HumanoidBone // optional reverse map
}

func BuildHumanoidRigFromGLTF(g *engine.GltfRoot, nodes []*ecs.Entity) *HumanoidRig

func RetargetClip(srcRig, dstRig *HumanoidRig, srcClip *ecs.AnimationClip) *ecs.AnimationClip
