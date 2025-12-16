package ecs

// Renderable is a component that marks an entity as drawable.
// It holds a mesh/material handle (IDs or references into your renderer).
type Renderable struct {
	MeshID     string
	MaterialID string
}

// Update is a no-op; rendering is handled by the RenderSystem.
func (r *Renderable) Update(dt float32) {
	_ = dt
}
