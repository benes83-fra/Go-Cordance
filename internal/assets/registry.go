package assets

import "sync/atomic"

var (
	registry    = map[AssetID]*Asset{}
	nextAssetID uint64
)

func Register(t AssetType, path string, data any) AssetID {
	id := AssetID(atomic.AddUint64(&nextAssetID, 1))
	registry[id] = &Asset{
		ID:   id,
		Type: t,
		Path: path,
		Data: data,
	}
	return id
}

func Get(id AssetID) *Asset {
	return registry[id]
}

func All() []*Asset {
	out := make([]*Asset, 0, len(registry))
	for _, a := range registry {
		out = append(out, a)
	}
	return out
}
