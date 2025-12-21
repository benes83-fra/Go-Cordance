package ecs

// Parent marks an entity as having a parent in the scene graph.
type Parent struct {
	Entity *Entity
}

func NewParent(e *Entity) *Parent {
	return &Parent{Entity: e}
}

func (p *Parent) Update(dt float32) { _ = dt }

// Children holds a list of child entities of a parent.
type Children struct {
	Entities []*Entity
}

func NewChildren() *Children {
	return &Children{Entities: make([]*Entity, 0)}
}

func (c *Children) AddChild(e *Entity) {
	c.Entities = append(c.Entities, e)
}

func (c *Children) Update(dt float32) { _ = dt }
