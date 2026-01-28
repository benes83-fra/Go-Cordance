package state

import (
	"go-engine/Go-Cordance/internal/editor/bridge"
)

type AssetView struct {
	ID   uint64
	Path string
	Type string
}

type EditorState struct {
	Entities        []bridge.EntityInfo
	SelectedID      int64
	SelectedIndex   int
	Foldout         map[string]bool
	RefreshUI       func() // <-- add this
	Selection       Selection
	SplitOffset     float64
	ShowLightGizmos bool
	IsRebuilding    bool
	LastComponents  map[int64][]string

	Assets struct {
		Textures  []AssetView
		Meshes    []AssetView
		Materials []AssetView
	}
}

func NewEditorState() *EditorState {
	return &EditorState{
		Entities:      []bridge.EntityInfo{},
		SelectedIndex: -1,

		LastComponents: make(map[int64][]string),
	}
}

func New() *EditorState {
	return &EditorState{
		Entities:       []bridge.EntityInfo{},
		SelectedIndex:  -1,
		SplitOffset:    0.35,
		LastComponents: make(map[int64][]string),
	}

}

type Selection struct {
	IDs      []int64
	ActiveID int64
	Mode     PivotMode
}

type PivotMode int

const (
	PivotModePivot PivotMode = iota
	PivotModeCenter
)

var Global = New()
