package scene

import (
	"go-engine/Go-Cordance/internal/ecs"
	"sync/atomic"
)

// Camera is a minimal camera placeholder for the prototype.
type Camera struct {
	Position [3]float32
	Target   [3]float32
	Up       [3]float32
	Fov      float32
	Near     float32
	Far      float32
}

// Scene holds entities and a camera.
type Scene struct {
	entities       []*ecs.Entity
	world          *ecs.World
	camera         Camera
	nextID         int64
	sysMgr         *ecs.SystemManager
	Selected       *ecs.Entity
	SelectedEntity uint64
}

// New returns a basic scene with a default camera.
func New() *Scene {
	s := &Scene{
		entities: make([]*ecs.Entity, 0, 16),
		world:    ecs.NewWorld(),
		camera: Camera{
			Position: [3]float32{0, 0, 3},
			Target:   [3]float32{0, 0, 0},
			Up:       [3]float32{0, 1, 0},
			Fov:      60,
			Near:     0.1,
			Far:      100,
		},
		nextID: 1,
		sysMgr: ecs.NewSystemManager(),
	}
	s.sysMgr.AddSystem(ecs.NewTransformSystem())
	return s
}

func (s *Scene) World() *ecs.World {
	return s.world
}
func (s *Scene) Systems() *ecs.SystemManager {
	return s.sysMgr
}

// AddEntity creates a new entity, appends it to the scene, and returns it.
func (s *Scene) AddEntity() *ecs.Entity {
	id := atomic.AddInt64(&s.nextID, 1)
	e := ecs.NewEntity(id)
	s.entities = append(s.entities, e)
	if s.world != nil {
		s.world.AddEntity(e)
	}
	return e
}

// AddExisting adds an already-created entity to the scene.
func (s *Scene) AddExisting(e *ecs.Entity) {
	s.entities = append(s.entities, e)
	if s.world != nil {
		s.world.AddEntity(e)
	}
}

// Entities returns a snapshot slice of entities.
func (s *Scene) Entities() []*ecs.Entity {
	return s.entities
}

// Camera returns a pointer to the scene camera for configuration.
func (s *Scene) Camera() *Camera {
	return &s.camera
}

// Update runs per-frame updates on entities. dt is seconds since last frame.
func (s *Scene) Update(dt float32) {
	// Update entity-local components
	for _, e := range s.entities {
		e.Update(dt)
	}
	// Run global systems
	s.sysMgr.Update(dt, s.entities)
}
func (s *Scene) contains(e *ecs.Entity) bool {
	for _, ex := range s.entities {
		if ex == e {
			return true
		}
	}
	return false
}
func (s *Scene) DuplicateEntity(src *ecs.Entity) *ecs.Entity {
	// 1. Create a new entity with a new ID
	dup := s.AddEntity()

	// 2. Clone components (shallow copy is fine for your architecture)
	for _, comp := range src.Components {
		switch c := comp.(type) {

		case *ecs.Transform:
			nc := *c
			nc.Position[0] += 0.5 // small offset so it's visible
			dup.AddComponent(&nc)

		case *ecs.Name:
			nc := *c
			nc.Value = c.Value + " Copy"
			dup.AddComponent(&nc)

		case *ecs.Material:
			nc := *c
			dup.AddComponent(&nc)

		case *ecs.LightComponent:
			nc := *c
			dup.AddComponent(&nc)

		case *ecs.ColliderSphere:
			nc := *c
			dup.AddComponent(&nc)

		case *ecs.ColliderAABB:
			nc := *c
			dup.AddComponent(&nc)

		case *ecs.ColliderPlane:
			nc := *c
			dup.AddComponent(&nc)
		case *ecs.Mesh:
			nc := *c
			dup.AddComponent(&nc)

		// Add more as needed

		default:
			// Unknown component type — skip
		}
	}

	return dup
}

func (s *Scene) DeleteEntityByID(id int64) {
	// Remove from scene.entities
	newList := make([]*ecs.Entity, 0, len(s.entities))
	for _, e := range s.entities {
		if e.ID != id {
			newList = append(newList, e)
		}
	}
	s.entities = newList

	// Remove from ECS world
	if s.world != nil {
		s.world.RemoveEntityByID(id)
	}

	// Clear selection if needed
	if s.Selected != nil && s.Selected.ID == id {
		s.Selected = nil
	}
}
