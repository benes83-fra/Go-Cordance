package ecs

type MultiMesh struct {
	Meshes []string
}

func NewMultiMesh(meshes []string) *MultiMesh {
	return &MultiMesh{Meshes: meshes}
}

func (mm *MultiMesh) Update(dt float32) {
	_ = dt
}
