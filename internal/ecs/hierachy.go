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

func (c *Children) Remove(e *Entity) {
	for i, child := range c.Entities {
		if child == e {
			c.Entities = append(c.Entities[:i], c.Entities[i+1:]...)
			return
		}
	}
}
