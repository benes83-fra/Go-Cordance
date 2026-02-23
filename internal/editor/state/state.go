package state

import (
	"go-engine/Go-Cordance/internal/editor/bridge"
)

type AssetView struct {
	ID           uint64            `json:"id"`
	Path         string            `json:"path"`
	Type         string            `json:"type"`
	Thumbnail    string            `json:"thumbnail,omitempty"` // whole-asset thumb
	ThumbHash    string            `json:"thumb_hash,omitempty"`
	MeshIDs      []string          `json:"mesh_ids,omitempty"`
	MeshThumb    map[string]string `json:"mesh_thumb,omitempty"` // meshID -> file path
	MaterialData map[string]any    `json:"material_data,omitempty"`
	ShaderData   map[string]any    `json:"shader_data,omitempty"`
}

type EditorState struct {
	Entities            []bridge.EntityInfo
	SelectedID          int64
	SelectedIndex       int
	Foldout             map[string]bool
	RefreshUI           func() // <-- add this
	UpdateLocalMaterial func(entityID int64, fields map[string]any)
	Selection           Selection
	SplitOffset         float64
	ShowLightGizmos     bool
	IsRebuilding        bool
	LastComponents      map[int64][]string
	Shaders             []AssetView
	UpdateLocalMesh     func(entityID int64, meshID string)
	// in EditorState

	Assets struct {
		Textures  []AssetView
		Meshes    []AssetView
		Materials []AssetView
		Shaders   []AssetView
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
