package ecs

var ComponentRegistry = map[string]func() Component{
	"Mesh":           func() Component { return NewMesh("") },
	"Material":       func() Component { return NewMaterial([4]float32{1, 1, 1, 1}) },
	"RigidBody":      func() Component { return NewRigidBody(1) },
	"ColliderSphere": func() Component { return NewColliderSphere(1) },
	"ColliderPlane":  func() Component { return NewColliderPlane(0) },
	"ColliderAABB":   func() Component { return NewColliderAABB([3]float32{0.5, 0.5, 0.5}) },
	"Light":          func() Component { return NewLightComponent() },
}
