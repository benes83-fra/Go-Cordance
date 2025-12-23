package ecs

// Name is a simple component that gives an entity a human readable label.
type Name struct {
	Value string
}

func NewName(v string) *Name { return &Name{Value: v} }

func (n *Name) Update(dt float32) {
	_ = dt
}
