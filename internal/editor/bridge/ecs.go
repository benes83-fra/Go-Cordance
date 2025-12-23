package bridge

type EntityInfo struct {
	ID   int
	Name string
}

type ECSProvider interface {
	ListEntities() []EntityInfo
}
