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

var (
	TextureNames []string
	TextureIDs   []uint32
)

func RegisterTexture(name string, id uint32) {
	TextureNames = append(TextureNames, name)
	TextureIDs = append(TextureIDs, id)
}
