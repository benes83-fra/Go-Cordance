package ui

import (
	"go-engine/Go-Cordance/internal/editor/state"
	"go-engine/Go-Cordance/internal/editorlink"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func NewAssetBrowserPanel(st *state.EditorState) (fyne.CanvasObject, *widget.List) {

	// --- TEXTURES LIST ---
	texList := widget.NewList(
		func() int {
			return len(st.Assets.Textures)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("texture")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(st.Assets.Textures[i].Path)
		},
	)

	texList.OnSelected = func(id widget.ListItemID) {
		if st.SelectedIndex < 0 || st.SelectedIndex >= len(st.Entities) {
			return
		}

		ent := st.Entities[st.SelectedIndex]
		asset := st.Assets.Textures[id]

		// For now: assign raw TextureID via inspector workflow
		// (later: switch to AssetID)
		msg := editorlink.MsgSetComponent{
			EntityID: uint64(ent.ID),
			Name:     "Material",
			Fields: map[string]any{
				"UseTexture": true,
				"TextureID":  int(asset.ID), // asset.ID currently stores GL texture ID
			},
		}

		if editorlink.EditorConn != nil {
			_ = editorlink.WriteSetComponent(editorlink.EditorConn, msg)
		}
	}

	// --- MESH LIST ---
	meshList := widget.NewList(
		func() int {
			return len(st.Assets.Meshes)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("mesh")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(st.Assets.Meshes[i].Path)
		},
	)
	log.Printf("Texture count: %d", len(st.Assets.Textures))

	// (Meshes are not assignable yet — future feature)
	meshList.OnSelected = func(id widget.ListItemID) {}

	// --- MATERIAL LIST ---
	matList := widget.NewList(
		func() int {
			return len(st.Assets.Materials)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("material")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(st.Assets.Materials[i].Path)
		},
	)

	// (Materials not assignable yet — future feature)
	matList.OnSelected = func(id widget.ListItemID) {}

	// --- Tabs ---
	tabs := container.NewAppTabs(
		container.NewTabItem("Textures", texList),
		container.NewTabItem("Meshes", meshList),
		container.NewTabItem("Materials", matList),
	)

	tabs.SetTabLocation(container.TabLocationTop)

	return tabs, texList
}
