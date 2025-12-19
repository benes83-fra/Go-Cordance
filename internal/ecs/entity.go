package ecs

import "reflect"

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
func (e *Entity) GetComponent(target Component) Component {
	for _, c := range e.Components {
		if reflect.TypeOf(c) == reflect.TypeOf(target) {
			return c
		}
	}
	return nil
}
func (e *Entity) GetTransform() *Transform {
	for _, c := range e.Components {
		if t, ok := c.(*Transform); ok {
			return t
		}
	}
	return nil
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
