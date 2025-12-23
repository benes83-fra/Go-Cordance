package ecs

func BootstrapGameWorld() *World {
	world := NewWorld()

	// All your game entities:
	world.AddEntity(NewEntity(1))
	world.AddEntity(NewEntity(2))
	world.AddEntity(NewEntity(3))

	return world
}
