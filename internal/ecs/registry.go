package ecs

import "reflect"

var ComponentRegistry = map[string]func() Component{

	"Mesh":           func() Component { return NewMesh("") },
	"Material":       func() Component { return NewMaterial([4]float32{1, 1, 1, 1}) },
	"RigidBody":      func() Component { return NewRigidBody(1) },
	"ColliderSphere": func() Component { return NewColliderSphere(1) },
	"ColliderPlane":  func() Component { return NewColliderPlane(0) },
	"ColliderAABB":   func() Component { return NewColliderAABB([3]float32{0.5, 0.5, 0.5}) },
	"Light":          func() Component { return NewLightComponent() },
	"Billboard":      func() Component { return NewBillboard() },
	"MultiMesh":      func() Component { return NewMultiMesh(nil) },
	"Parent":         func() Component { return &Parent{} },
	"Children":       func() Component { return NewChildren() },
	"Name":           func() Component { return NewName("") },
	"Camera":         func() Component { return NewCamera() },
}

// ComponentNameRegistry maps concrete component types to their registry name.
var ComponentNameRegistry map[reflect.Type]string

func init() {
	ComponentNameRegistry = make(map[reflect.Type]string)
	for name, ctor := range ComponentRegistry {
		// ctor() returns a Component instance; use its concrete type
		ComponentNameRegistry[reflect.TypeOf(ctor())] = name
	}
}

// ComponentTypeName returns the registry name for a concrete component instance.
// Returns empty string if unknown.
func ComponentTypeName(c Component) string {
	if c == nil {
		return ""
	}
	if name, ok := ComponentNameRegistry[reflect.TypeOf(c)]; ok {
		return name
	}
	return ""
}
