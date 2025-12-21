package ecs

type MultiMaterial struct {
	Materials map[string]*Material // key = meshID (e.g. "Teapot/0")
}

func NewMultiMaterial() *MultiMaterial {
	return &MultiMaterial{
		Materials: make(map[string]*Material),
	}
}

func (multmat *MultiMaterial) Update(dt float32) { _ = dt }
