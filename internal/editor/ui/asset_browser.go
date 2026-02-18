package ui

import (
	"go-engine/Go-Cordance/internal/assets"
	"strings"

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

type MeshListEntry struct {
	IsHeader bool
	AssetID  uint64
	MeshID   string // empty for headers
	Label    string
}

type materialItem struct {
	widget.BaseWidget
	img     *canvas.Image
	lbl     *widget.Label
	assetID uint64
}

func findAssetView(st *state.EditorState, id uint64) *state.AssetView {
	for i := range st.Assets.Meshes {
		if st.Assets.Meshes[i].ID == id {
			return &st.Assets.Meshes[i]
		}
	}
	return nil
}
func newMaterialItem() *materialItem {
	img := canvas.NewImageFromResource(theme.FileImageIcon())
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(96, 96))

	lbl := widget.NewLabel("")

	item := &materialItem{
		img: img,
		lbl: lbl,
	}
	item.ExtendBaseWidget(item)
	return item
}

func (m *materialItem) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(
		container.NewHBox(m.img, m.lbl),
	)
}

func buildMeshListEntries(st *state.EditorState) []MeshListEntry {
	out := []MeshListEntry{}

	for _, a := range st.Assets.Meshes {
		// Header
		out = append(out, MeshListEntry{
			IsHeader: true,
			AssetID:  a.ID,
			Label:    filepath.Base(a.Path),
		})

		// Submeshes
		if len(a.MeshIDs) == 0 {
			out = append(out, MeshListEntry{
				IsHeader: false,
				AssetID:  a.ID,
				MeshID:   filepath.Base(a.Path),
				Label:    filepath.Base(a.Path),
			})
		} else {
			for _, mid := range a.MeshIDs {
				out = append(out, MeshListEntry{
					IsHeader: false,
					AssetID:  a.ID,
					MeshID:   mid,
					Label:    mid,
				})
			}
		}
	}

	return out
}

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
									_ = editorlink.WriteRequestThumbnail(editorlink.EditorConn, assetID, 128)
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
								_ = editorlink.WriteRequestMeshThumbnail(editorlink.EditorConn, assetID, meshID, 128)
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
			return newMaterialItem() // ✔ FIXED
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			item := o.(*materialItem) // ✔ now valid
			av := st.Assets.Materials[i]

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
					if editorlink.EditorConn != nil {
						_ = editorlink.WriteRequestThumbnail(editorlink.EditorConn, assetID, 128)
					}
				}(av.ID)
			}
		},
	)

	// (Materials not assignable yet — future feature)
	matList.OnSelected = func(id widget.ListItemID) {
		if st.SelectedIndex < 0 || st.SelectedIndex >= len(st.Entities) {
			return
		}

		ent := st.Entities[st.SelectedIndex]
		av := st.Assets.Materials[id]
		shaderName := ""
		if s, ok := av.MaterialData["shader"].(string); ok {
			shaderName = s
		}

		ecsMat := ConvertMaterialAssetToECS(av)
		if ecsMat == nil {
			log.Printf("Material asset %d has no valid data", av.ID)
			return
		}
		// Derive UseTexture from IDs to avoid bad flags
		useTex := (ecsMat.TextureAsset != 0 || ecsMat.TextureID != 0)

		msg := editorlink.MsgSetComponent{
			EntityID: uint64(ent.ID),
			Name:     "Material",
			Fields: map[string]any{
				"BaseColor":    ecsMat.BaseColor,
				"UseTexture":   useTex,
				"TextureAsset": ecsMat.TextureAsset,
				"TextureID":    int(ecsMat.TextureID),
				"ShaderName":   shaderName,
			},
		}

		if editorlink.EditorConn != nil {
			go editorlink.WriteSetComponent(editorlink.EditorConn, msg)
		}

		if st.UpdateLocalMaterial != nil {
			st.UpdateLocalMaterial(ent.ID, msg.Fields)
		}

		if state.Global.RefreshUI != nil {
			state.Global.RefreshUI()
		}
	}

	shaderList := widget.NewList(
		func() int {
			return len(st.Assets.Shaders)
		},
		func() fyne.CanvasObject {
			return newShaderItem()
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			item := o.(*shaderItem)
			av := st.Assets.Shaders[i]

			item.assetID = av.ID
			item.lbl.SetText(filepath.Base(av.Path))

			// shaders don’t have thumbnails yet → use icon
			name, _ := av.ShaderData["name"].(string)

			item.SetIconForShader(name)
			item.img.Refresh()
		},
	)
	shaderList.OnSelected = func(id widget.ListItemID) {
		av := st.Assets.Shaders[id]

		name, ok := av.ShaderData["name"].(string)
		if !ok {
			log.Printf("Shader asset %d missing name", av.ID)
			return
		}

		if editorlink.EditorConn != nil {
			go editorlink.SendSetGlobalShader(editorlink.EditorConn, name)
		}

		log.Printf("Requested global shader switch to %s", name)
	}

	// --- Tabs ---
	tabs := container.NewAppTabs(
		container.NewTabItem("Textures", texList),
		container.NewTabItem("Meshes", meshList),
		container.NewTabItem("Materials", matList),
		container.NewTabItem("Shaders", shaderList),
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

type shaderItem struct {
	widget.BaseWidget
	img     *canvas.Image
	lbl     *widget.Label
	assetID uint64
}

func newShaderItem() *shaderItem {
	img := canvas.NewImageFromResource(theme.FileImageIcon())
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(96, 96))

	lbl := widget.NewLabel("")

	item := &shaderItem{
		img: img,
		lbl: lbl,
	}
	item.ExtendBaseWidget(item)
	return item
}

func (s *shaderItem) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(
		container.NewHBox(s.img, s.lbl),
	)
}

func (s *shaderItem) SetIconForShader(name string) {
	lower := strings.ToLower(name)

	switch {
	case strings.Contains(lower, "debug"):
		s.img.Resource = theme.WarningIcon()
	case strings.Contains(lower, "depth"):
		s.img.Resource = theme.VisibilityOffIcon()
	case strings.Contains(lower, "shadow"):
		s.img.Resource = theme.VisibilityIcon()
	case strings.Contains(lower, "flat"):
		s.img.Resource = theme.ColorPaletteIcon()
	case strings.Contains(lower, "tbn"):
		s.img.Resource = theme.GridIcon()
	case strings.Contains(lower, "visual"):
		s.img.Resource = theme.InfoIcon()
	case strings.Contains(lower, "light"):
		s.img.Resource = theme.ColorChromaticIcon()
	case strings.Contains(lower, "post"):
		s.img.Resource = theme.MediaReplayIcon()
	default:
		s.img.Resource = theme.ComputerIcon()
	}

	s.img.Refresh()
}
