package ecs

import (
	"fmt"
	"go-engine/Go-Cordance/internal/editor/bridge"
)


type World struct {
	Entities []*Entity
}

func NewWorld() *World {
	return &World{
		Entities: make([]*Entity, 0, 128),
	}
}

func (w *World) AddEntity(e *Entity) {
	w.Entities = append(w.Entities, e)
}

func (w *World) ListEntityInfo() []bridge.EntityInfo {
	out := make([]bridge.EntityInfo, 0, len(w.Entities))
	for _, e := range w.Entities {
		name := "Entity " + fmt.Sprint(e.ID)
		if n := e.GetComponent(&Name{}); n != nil {
			if nn, ok := n.(*Name); ok && nn.Value != "" {
				name = nn.Value
			}
		}
		out = append(out, bridge.EntityInfo{ID: e.ID, Name: name})
	}
	return out
}

func (w *World) FindByID(id int64) *Entity {
	for _, e := range w.Entities {
		if e.ID == id {
			return e
		}
	}
	return nil
}
