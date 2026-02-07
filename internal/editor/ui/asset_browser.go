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
			item := o.(*textureDragItem)
			av := st.Assets.Textures[i]
			item.assetID = av.ID
			item.lbl.SetText(filepath.Base(av.Path))

			if av.Thumbnail != "" {
				item.img.File = av.Thumbnail
				item.img.Resource = nil
				item.img.Refresh()
			} else {
				item.img.Resource = theme.FileImageIcon()
				item.img.File = ""
				item.img.Refresh()

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
		if st.UpdateLocalMaterial != nil {
			st.UpdateLocalMaterial(ent.ID, msg.Fields)
		}
		if state.Global.RefreshUI != nil {
			state.Global.RefreshUI()
		}
	}

	// --- MESH LIST ---
	// --- MESH LIST ---
	// --- MESH LIST ---
	meshList := widget.NewList(
		func() int {
			// total number of submeshes across all mesh assets
			count := 0
			for _, a := range st.Assets.Meshes {
				if len(a.MeshIDs) == 0 {
					// single-mesh asset with no MeshIDs? treat as 1 entry using Path basename
					count++
				} else {
					count += len(a.MeshIDs)
				}
			}
			return count
		},
		func() fyne.CanvasObject {
			return newMeshDragItem()
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			item := o.(*meshDragItem)
			idx := int(i)

			for _, a := range st.Assets.Meshes {
				if len(a.MeshIDs) == 0 {
					// single-mesh asset
					if idx == 0 {
						item.assetID = a.ID
						item.meshID = filepath.Base(a.Path)
						item.lbl.SetText(item.meshID)

						thumb := a.Thumbnail
						if thumb != "" {
							item.img.File = thumb
							item.img.Resource = nil
							item.img.Refresh()
						} else {
							item.img.Resource = theme.FileImageIcon()
							item.img.File = ""
							item.img.Refresh()

							go func(assetID uint64) {
								if editorlink.EditorConn != nil {
									_ = editorlink.WriteRequestThumbnail(editorlink.EditorConn, assetID, 128 /* no meshID */)
								}
							}(a.ID)
						}
						return
					}
					idx--
					continue
				}

				if idx < len(a.MeshIDs) {
					meshID := a.MeshIDs[idx]
					item.assetID = a.ID
					item.meshID = meshID
					item.lbl.SetText(meshID)

					// prefer per-mesh thumbnail
					thumb := ""
					if a.MeshThumb != nil {
						thumb = a.MeshThumb[meshID]
					}
					if thumb == "" {
						// optional fallback to asset-level thumb
						thumb = a.Thumbnail
					}

					if thumb != "" {
						item.img.File = thumb
						item.img.Resource = nil
						item.img.Refresh()
					} else {
						item.img.Resource = theme.FileImageIcon()
						item.img.File = ""
						item.img.Refresh()

						go func(assetID uint64, meshID string) {
							if editorlink.EditorConn != nil {
								_ = editorlink.WriteRequestThumbnailWithMesh(editorlink.EditorConn, assetID, meshID, 128)
							}
						}(a.ID, meshID)
					}
					return
				}
				idx -= len(a.MeshIDs)
			}

			// fallback
			item.assetID = 0
			item.meshID = ""
			item.lbl.SetText("<invalid>")
			item.img.Resource = theme.FileImageIcon()
			item.img.File = ""
			item.img.Refresh()
		},
	)

	meshList.OnSelected = func(id widget.ListItemID) {
		if st.SelectedIndex < 0 || st.SelectedIndex >= len(st.Entities) {
			return
		}

		// map flat index id -> (asset, meshID) again
		idx := int(id)

		var meshID string

		for _, a := range st.Assets.Meshes {
			if len(a.MeshIDs) == 0 {
				if idx == 0 {

					meshID = filepath.Base(a.Path)
					break
				}
				idx--
				continue
			}

			if idx < len(a.MeshIDs) {

				meshID = a.MeshIDs[idx]
				break
			}
			idx -= len(a.MeshIDs)
		}

		if meshID == "" {
			log.Printf("MeshList.OnSelected: no meshID resolved for index %d", id)
			return
		}

		ent := st.Entities[st.SelectedIndex]

		msg := editorlink.MsgSetComponent{
			EntityID: uint64(ent.ID),
			Name:     "Mesh",
			Fields: map[string]any{
				"MeshID": meshID,
			},
		}

		if editorlink.EditorConn != nil {
			go editorlink.WriteSetComponent(editorlink.EditorConn, msg)
		}

		if state.Global.RefreshUI != nil {
			state.Global.RefreshUI()
		}
	}

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
	return newTextureDragItem()
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

type textureDragItem struct {
	widget.BaseWidget
	img     *canvas.Image
	lbl     *widget.Label
	assetID uint64
}

func newTextureDragItem() *textureDragItem {
	img := canvas.NewImageFromResource(theme.FileImageIcon())
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(96, 96))

	lbl := widget.NewLabel("")

	item := &textureDragItem{
		img: img,
		lbl: lbl,
	}
	item.ExtendBaseWidget(item)
	return item
}

func (t *textureDragItem) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(
		container.NewHBox(t.img, t.lbl),
	)
}

func (t *textureDragItem) Dragged(ev *fyne.DragEvent) {}
func (t *textureDragItem) DragEnd()                   {}

func (t *textureDragItem) DragData() interface{} {
	return DragAsset{
		ID:   t.assetID,
		Type: "texture",
	}
}

type meshDragItem struct {
	widget.BaseWidget
	img     *canvas.Image
	lbl     *widget.Label
	assetID uint64
	meshID  string
}

func newMeshDragItem() *meshDragItem {
	img := canvas.NewImageFromResource(theme.FileImageIcon())
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(96, 96))

	lbl := widget.NewLabel("")

	item := &meshDragItem{
		img: img,
		lbl: lbl,
	}
	item.ExtendBaseWidget(item)
	return item
}

func (m *meshDragItem) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(
		container.NewHBox(m.img, m.lbl),
	)
}

func (m *meshDragItem) Dragged(ev *fyne.DragEvent) {}
func (m *meshDragItem) DragEnd()                   {}

func (m *meshDragItem) DragData() interface{} {
	return DragAsset{
		ID:   m.assetID,
		Type: "mesh",
	}
}

type DragAsset struct {
	ID   uint64
	Type string // "texture", "mesh", "material"
}
