package ecs

// Component is a minimal interface for components that need per-frame updates.
type Component interface {
	Update(dt float32)
}

// Entity is a simple container of components and an ID.
type Entity struct {
	ID         int64
	Components []Component
}

// NewEntity creates an empty entity with the given id.
func NewEntity(id int64) *Entity {
	return &Entity{
		ID:         id,
		Components: make([]Component, 0, 4),
	}
}

// AddComponent appends a component to the entity.
func (e *Entity) AddComponent(c Component) {
	e.Components = append(e.Components, c)
}

// Update calls Update on all components.
func (e *Entity) Update(dt float32) {
	for _, c := range e.Components {
		c.Update(dt)
	}
}
