package bridge

type EntityInfo struct {
	ID   int64
	Name string
}

type ECSProvider interface {
	ListEntities() []EntityInfo
}
