package ecs

import "fmt"

type Skin struct {
	// glTF joint node indices
	Joints []int

	// One inverse bind matrix per joint, column-major 4x4
	InverseBindMatrices [][16]float32
	JointMatrices       [][16]float32
	JointEntities       []*Entity
}

func NewSkin(joints []int, ibm [][16]float32) *Skin {
	fmt.Printf("Skin created: joints=%v ibmCount=%d\n", joints, len(ibm))

	return &Skin{
		Joints:              append([]int(nil), joints...),
		InverseBindMatrices: append([][16]float32(nil), ibm...),
		JointMatrices:       make([][16]float32, len(joints)),
		JointEntities:       make([]*Entity, len(joints)),
	}
}

func (s *Skin) Update(dt float32) { _ = dt }

func (s *Skin) EditorName() string {
	return "Skin"
}

func (s *Skin) EditorFields() map[string]any {
	return map[string]any{
		"JointCount": len(s.Joints),
	}
}
