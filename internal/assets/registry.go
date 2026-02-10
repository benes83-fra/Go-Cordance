package assets

import (
	"path/filepath"
	"sync/atomic"
)

var (
	registry    = map[AssetID]*Asset{}
	nextAssetID uint64
)

func normalize(p string) string {
	return filepath.ToSlash(p)
}

func Register(t AssetType, path string, data any) AssetID {
	path = normalize(path)
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

func FindAssetByPath(path string) *Asset {
	path = normalize(path)
	for _, a := range registry {
		if a.Path == path {
			return a
		}
	}
	return nil
}
