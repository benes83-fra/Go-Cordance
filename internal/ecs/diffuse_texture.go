package ecs

type DiffuseTexture struct {
	ID uint32
}

func NewDiffuseTexture(id uint32) *DiffuseTexture {
	return &DiffuseTexture{ID: id}
}

func (dtex *DiffuseTexture) Update(dt float32) { _ = dt }
