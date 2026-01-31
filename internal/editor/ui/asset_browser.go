package ui

import (
	"go-engine/Go-Cordance/internal/assets"
	"go-engine/Go-Cordance/internal/editor/state"
	"go-engine/Go-Cordance/internal/editorlink"
	"log"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func NewAssetBrowserPanel(st *state.EditorState) (fyne.CanvasObject, *widget.List) {

	// --- TEXTURES LIST (image + label cells, lazy thumbnail requests) ---

	texList := widget.NewList(
		func() int {
			return len(st.Assets.Textures)
		},
		func() fyne.CanvasObject {
			return makeTextureItem()
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			row := o.(*fyne.Container)            // HBox
			img := row.Objects[0].(*canvas.Image) // first child
			lbl := row.Objects[1].(*widget.Label) // second child

			av := st.Assets.Textures[i]
			log.Printf("texture[%d] label = %q", i, filepath.Base(av.Path))

			lbl.SetText(filepath.Base(av.Path))

			if av.Thumbnail != "" {
				img.File = av.Thumbnail
				img.Resource = nil
				img.Refresh()
			} else {
				img.Resource = theme.FileImageIcon()
				img.File = ""
				img.Refresh()

				go func(assetID uint64) {
					if editorlink.EditorConn == nil {
						return
					}
					if err := editorlink.WriteRequestThumbnail(editorlink.EditorConn, assetID, 128); err != nil {
						log.Printf("failed to request thumbnail for asset %d: %v", assetID, err)
					}
				}(av.ID)
			}
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
	// Hook into global RefreshUI

	return tabs, texList
}

func makeTextureItem() fyne.CanvasObject {
	img := canvas.NewImageFromResource(theme.FileImageIcon())
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(96, 96)) // was 64, now larger

	lbl := widget.NewLabel("")
	//lbl.Wrapping = fyne.TextTruncate

	// EXACTLY two children: [img, lbl]
	return container.NewHBox(img, lbl)
}

func makeTextureGridItem(av state.AssetView) fyne.CanvasObject {
	img := canvas.NewImageFromFile(av.Thumbnail)
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(128, 128))

	lbl := widget.NewLabel(filepath.Base(av.Path))
	lbl.Wrapping = fyne.TextTruncate

	return container.NewVBox(img, lbl)
}

func rebuildTextureGrid(st *state.EditorState, grid *fyne.Container) {
	grid.Objects = nil

	for _, av := range st.Assets.Textures {
		// request thumbnail if missing
		if av.Thumbnail == "" && editorlink.EditorConn != nil {
			go editorlink.WriteRequestThumbnail(editorlink.EditorConn, av.ID, 128)
		}

		grid.Add(makeTextureGridItem(av))
	}

	grid.Refresh()
}
