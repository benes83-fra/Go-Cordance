package ecs

type TransformSystem struct{}

func NewTransformSystem() *TransformSystem {
	return &TransformSystem{}
}

/*
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
*/
func (ts *TransformSystem) Update(dt float32, entities []*Entity) {
	_ = dt
	identity := IdentityMatrix()

	// Update roots (and their subtrees)
	for _, e := range entities {
		if e.GetComponent((*Parent)(nil)) != nil {
			continue
		}
		ts.updateRecursive(e, &identity)
	}
}

func (ts *TransformSystem) updateRecursive(e *Entity, parentWorld *[16]float32) {
	tr := e.GetTransform()
	if tr != nil {
		// Always rebuild local if dirty
		if tr.Dirty {
			tr.RecalculateLocal()
		}
		tr.WorldMatrix = MulMat4(*parentWorld, tr.LocalMatrix)
	}

	children, ok := e.GetComponent((*Children)(nil)).(*Children)
	if !ok || children == nil {
		return
	}

	nextParent := parentWorld
	if tr != nil {
		nextParent = &tr.WorldMatrix
	}

	for _, c := range children.Entities {
		ts.updateRecursive(c, nextParent)
	}
}
