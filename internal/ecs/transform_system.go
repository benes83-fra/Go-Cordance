package ecs

type TransformSystem struct{}

func NewTransformSystem() *TransformSystem {
	return &TransformSystem{}
}

func (ts *TransformSystem) Update(dt float32, entities []*Entity) {
	_ = dt

	// Find root entities (no Parent component)
	roots := make([]*Entity, 0)
	for _, e := range entities {
		if e.GetComponent((*Parent)(nil)) == nil {
			roots = append(roots, e)
		}
	}

	identity := IdentityMatrix()

	for _, r := range roots {
		ts.updateRecursive(r, &identity)
	}
}

func (ts *TransformSystem) updateRecursive(e *Entity, parentWorld *[16]float32) {
	tr := e.GetTransform()
	if tr != nil {
		if tr.Dirty {
			tr.RecalculateLocal()
		}
		world := MulMat4(*parentWorld, tr.LocalMatrix)
		tr.WorldMatrix = world
	}

	children, ok := e.GetComponent((*Children)(nil)).(*Children)
	if ok && children != nil {
		var nextParent *[16]float32
		if tr != nil {
			nextParent = &tr.WorldMatrix
		} else {
			nextParent = parentWorld
		}

		for _, c := range children.Entities {
			ts.updateRecursive(c, nextParent)
		}
	}
}
