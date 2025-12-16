package ecs

import "go-engine/Go-Cordance/internal/engine"

// RenderSystem draws all entities with a Renderable component.
type RenderSystem struct {
	Renderer *engine.Renderer
}

// NewRenderSystem creates a render system bound to a renderer.
func NewRenderSystem(r *engine.Renderer) *RenderSystem {
	return &RenderSystem{Renderer: r}
}

// Update iterates entities and draws those with Renderable.
func (rs *RenderSystem) Update(dt float32, entities []*Entity) {
	_ = dt
	for _, e := range entities {
		for _, c := range e.Components {
			if _, ok := c.(*Renderable); ok {
				// For now, just draw the prototype triangle.
				// Later, use MeshID/MaterialID to select buffers/shaders.
				rs.Renderer.Draw()
			}
		}
	}
}
