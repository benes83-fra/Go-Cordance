package registry

import (
	"go-engine/Go-Cordance/internal/ecs"
)

type ComponentFactory func() ecs.Component

var Components = map[string]ComponentFactory{
	"Name":         func() ecs.Component { return ecs.NewName("New Entity") },
	"Transform":    func() ecs.Component { return ecs.NewTransform([3]float32{0, 0, 0}) },
	"Mesh":         func() ecs.Component { return ecs.NewMesh("cube") },
	"Material":     func() ecs.Component { return ecs.NewMaterial([4]float32{1, 1, 1, 1}) },
	"RigidBody":    func() ecs.Component { return ecs.NewRigidBody(1.0) },
	"ColliderAABB": func() ecs.Component { return ecs.NewColliderAABB([3]float32{0.5, 0.5, 0.5}) },
	// Add more as needed
}
