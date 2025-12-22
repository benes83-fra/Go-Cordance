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
	entities []*ecs.Entity
	camera   Camera
	nextID   int64
	sysMgr   *ecs.SystemManager
}

// New returns a basic scene with a default camera.
func New() *Scene {
	s := &Scene{
		entities: make([]*ecs.Entity, 0, 16),
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

func (s *Scene) Systems() *ecs.SystemManager {
	return s.sysMgr
}

// AddEntity creates a new entity, appends it to the scene, and returns it.
func (s *Scene) AddEntity() *ecs.Entity {
	id := atomic.AddInt64(&s.nextID, 1)
	e := ecs.NewEntity(id)
	s.entities = append(s.entities, e)
	return e
}

// AddExisting adds an already-created entity to the scene.
func (s *Scene) AddExisting(e *ecs.Entity) {
	s.entities = append(s.entities, e)
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
