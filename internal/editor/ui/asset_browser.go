package ui

import (
	"go-engine/Go-Cordance/internal/assets"
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

		// Resolve GL texture id from asset registry (assets.ResolveTextureGLID must exist)
		glID := int(assets.ResolveTextureGLID(assets.AssetID(asset.ID)))

		msg := editorlink.MsgSetComponent{
			EntityID: uint64(ent.ID),
			Name:     "Material",
			Fields: map[string]any{
				"UseTexture":   true,
				"TextureAsset": int(asset.ID), // editor-side AssetID
				"TextureID":    glID,          // compatibility: raw GL ID for renderer/inspector
			},
		}

		if editorlink.EditorConn != nil {
			go func(m editorlink.MsgSetComponent) {
				_ = editorlink.WriteSetComponent(editorlink.EditorConn, m)
			}(msg)
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
