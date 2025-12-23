package editor

type EditorState struct {
	Entities      []string // later replace with your ECS entity type
	SelectedIndex int
}

func NewEditorState() *EditorState {
	return &EditorState{
		Entities:      []string{},
		SelectedIndex: -1,
	}
}
func New() *EditorState { return &EditorState{Entities: []string{}, SelectedIndex: -1} }
