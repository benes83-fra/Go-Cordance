package editor

import "go-engine/Go-Cordance/internal/editor/bridge"

type EditorState struct {
	Entities      []bridge.EntityInfo
	SelectedIndex int
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
