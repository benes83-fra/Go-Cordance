package ecs

type Texture struct {
	ID uint32
}

func NewTexture(id uint32) *Texture {
	return &Texture{ID: id}
}

func (av *Texture) Update(dt float32) {
	_ = dt
}
