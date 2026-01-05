package bridge

type Vec3 [3]float32
type Vec4 [4]float32

type EntityInfo struct {
	ID         int64    `json:"id"`
	Name       string   `json:"name"`
	Position   Vec3     `json:"position"`
	Rotation   Vec4     `json:"rotation"`
	Scale      Vec3     `json:"scale"`
	Components []string `json:"components"`
}

type ECSProvider interface {
	ListEntities() []EntityInfo
}
