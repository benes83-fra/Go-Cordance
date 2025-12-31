package state

import (
	"go-engine/Go-Cordance/internal/editor/bridge"
)

type EditorState struct {
	Entities      []bridge.EntityInfo
	SelectedID    int64
	SelectedIndex int
	Foldout       map[string]bool
	RefreshUI     func() // <-- add this
}

func NewEditorState() *EditorState {
	return &EditorState{
		Entities:      []bridge.EntityInfo{},
		SelectedIndex: -1,
	}
}
func New() *EditorState {
	return &EditorState{
		Entities:      []bridge.EntityInfo{},
		SelectedIndex: -1,
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
