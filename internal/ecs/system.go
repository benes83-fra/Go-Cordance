package ecs

// System is an interface for global systems that operate on entities/components.
type System interface {
	Update(dt float32, entities []*Entity)
}

// SystemManager holds registered systems and runs them each frame.
type SystemManager struct {
	systems []System
}

// NewSystemManager creates an empty manager.
func NewSystemManager() *SystemManager {
	return &SystemManager{systems: make([]System, 0, 8)}
}

// AddSystem registers a new system.
func (sm *SystemManager) AddSystem(s System) {
	sm.systems = append(sm.systems, s)
}

// Update runs all systems on the given entities.
func (sm *SystemManager) Update(dt float32, entities []*Entity) {
	for _, sys := range sm.systems {
		sys.Update(dt, entities)
	}
}
