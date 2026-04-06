package ui

import (
	"go-engine/Go-Cordance/internal/ecs/gizmo"
	"go-engine/Go-Cordance/internal/editor/bridge"
	state "go-engine/Go-Cordance/internal/editor/state"
	"go-engine/Go-Cordance/internal/editorlink"
	"log"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

var lastClickTime time.Time
var lastClickIndex = -1

type HierarchyRow struct {
	ID        int64
	Name      string
	IsVirtual bool
	ParentID  int64
	Depth     int
}

// NewHierarchyPanel returns BOTH:
// 1) The UI container (button + list)
// 2) The underlying *widget.List for inspector rebuild
func NewHierarchyPanel(st *state.EditorState, win fyne.Window, onSelect func(int)) (fyne.CanvasObject, *widget.List) {
	state.Global.RenameIndex = -1

	dupBtn := widget.NewButton("Duplicate", func() {
		if st.Selection.ActiveID != 0 && editorlink.EditorConn != nil {
			go editorlink.WriteDuplicateEntity(editorlink.EditorConn, st.Selection.ActiveID)
		}
	})

	delBtn := widget.NewButton("Delete", func() {
		if st.Selection.ActiveID != 0 && editorlink.EditorConn != nil {
			var deleted bridge.EntityInfo
			for _, e := range st.Entities {
				if e.ID == st.Selection.ActiveID {
					deleted = e
					break
				}
			}
			go editorlink.WriteDeleteEntity(editorlink.EditorConn, st.Selection.ActiveID, deleted.Name)
		}
	})
	saveBtn := widget.NewButton("Save Scene", func() {
		dialog.ShowFileSave(func(uc fyne.URIWriteCloser, err error) {
			if uc == nil {
				return
			}
			path := uc.URI().Path()
			editorlink.WriteSaveScene(editorlink.EditorConn, path)
		}, win)
	})

	loadBtn := widget.NewButton("Load Scene", func() {
		dialog.ShowFileOpen(func(ur fyne.URIReadCloser, err error) {
			if ur == nil {
				return
			}
			path := ur.URI().Path()
			editorlink.WriteLoadScene(editorlink.EditorConn, path)
		}, win)
	})

	createBtn := widget.NewButton("Create Empty", func() {
		if editorlink.EditorConn != nil {
			go editorlink.WriteCreateEntity(editorlink.EditorConn, "Empty")
		}
	})

	makeItem := func() fyne.CanvasObject {
		return newHierarchyDropItem()
	}

	var rows []HierarchyRow
	var list *widget.List

	list = widget.NewList(
		func() int {
			rows = BuildHierarchyRows(st)
			return len(rows)
		},
		makeItem,
		func(i int, o fyne.CanvasObject) {
			item := o.(*hierarchyDropItem)
			row := rows[i]
			item.entityID = row.ID

			if i < 0 || i >= len(rows) {
				item.check.SetChecked(false)
				item.btn.SetText("")
				item.entry.Hide()
				item.btn.Show()
				return
			}

			check := item.check
			btn := item.btn
			entry := item.entry
			item.entityID = row.ID
			ent := st.Entities[i]
			indent := strings.Repeat("  ", row.Depth)
			btn.SetText(indent + row.Name)

			// Inline rename mode
			if state.Global.RenameIndex == i {
				btn.Hide()
				entry.Show()
				entry.SetText(row.Name)

				entry.OnSubmitted = func(newName string) {
					state.Global.RenameIndex = -1
					entry.Hide()
					btn.Show()

					msg := editorlink.MsgSetComponent{
						EntityID: uint64(row.ID),
						Name:     "Name",
						Fields: map[string]any{
							"Value": newName,
						},
					}
					go editorlink.WriteSetComponent(editorlink.EditorConn, msg)
				}
			} else {
				entry.Hide()
				btn.Show()
			}

			// checkbox state
			checked := false
			for _, id := range st.Selection.IDs {
				if id == row.ID {
					checked = true
					break
				}
			}

			check.OnChanged = nil
			check.SetChecked(checked)

			check.OnChanged = func(checked bool) {
				if checked {
					found := false
					for _, id := range st.Selection.IDs {
						if id == row.ID {
							found = true
							break
						}
					}
					if !found {
						st.Selection.IDs = append(st.Selection.IDs, row.ID)
					}
				} else {
					newIDs := make([]int64, 0, len(st.Selection.IDs))
					for _, id := range st.Selection.IDs {
						if id != row.ID {
							newIDs = append(newIDs, id)
						}
					}
					st.Selection.IDs = newIDs
				}

				if editorlink.EditorConn != nil {
					go editorlink.WriteSelectEntities(editorlink.EditorConn, st.Selection.IDs)
				}

				gizmo.SetGlobalSelectionIDs(st.Selection.IDs)

			}

			btn.OnTapped = func() {
				now := time.Now()

				// Double-click → enter rename mode
				if lastClickIndex == i && now.Sub(lastClickTime) < 300*time.Millisecond {
					state.Global.RenameIndex = i
					list.Refresh()
					if editorlink.EditorConn != nil {
						go editorlink.WriteFocusEntity(editorlink.EditorConn, ent.ID)
					}

					return
				}

				lastClickTime = now
				lastClickIndex = i

				// Single-click → select
				entIndex := -1
				for idx, e := range st.Entities {
					if e.ID == row.ID {
						entIndex = idx
						break
					}
				}

				st.Selection.ActiveID = row.ID
				st.Selection.IDs = []int64{row.ID}
				st.SelectedIndex = entIndex

				if editorlink.EditorConn != nil {
					go editorlink.WriteSelectEntity(editorlink.EditorConn, row.ID)
					go editorlink.WriteSelectEntities(editorlink.EditorConn, st.Selection.IDs)
				}

				if entIndex >= 0 {
					onSelect(entIndex)
				}
			}
			// Right-click → context menu
			item.OnTappedSecondary = func(ev *fyne.PointEvent) {
				showHierarchyContextMenu(item, row, st, list, onSelect)
			}

		},
	)

	topBar := container.NewHBox(createBtn, dupBtn, delBtn, saveBtn, loadBtn)

	panel := container.NewBorder(topBar, nil, nil, nil, list)

	return panel, list
}

func BuildHierarchyRows(st *state.EditorState) []HierarchyRow {
	ents := st.Entities
	parentMap := st.ParentMap
	childrenMap := st.ChildrenMap

	byID := map[int64]bridge.EntityInfo{}
	for _, e := range ents {
		byID[e.ID] = e
	}

	var rows []HierarchyRow
	seen := map[int64]bool{}

	var walk func(id int64, depth int)
	walk = func(id int64, depth int) {
		if seen[id] {
			return
		}
		seen[id] = true

		e, ok := byID[id]
		if !ok {
			return
		}

		rows = append(rows, HierarchyRow{
			ID:        id,
			Name:      e.Name,
			IsVirtual: false,
			ParentID:  parentMap[id],
			Depth:     depth,
		})

		for _, childID := range childrenMap[id] {
			walk(childID, depth+1)
		}
	}

	for _, e := range ents {
		if e.Parent == 0 {
			walk(e.ID, 0)
		}
	}

	for _, e := range ents {
		if !seen[e.ID] {
			walk(e.ID, 0)
		}
	}

	return rows
}

func buildHierarchyOrder(ents []bridge.EntityInfo, childrenMap map[int64][]int64) []int64 {
	var out []int64

	roots := []int64{}
	for _, e := range ents {
		if e.Parent == 0 {
			roots = append(roots, e.ID)
		}
	}

	var walk func(id int64)
	walk = func(id int64) {
		out = append(out, id)
		for _, child := range childrenMap[id] {
			walk(child)
		}
	}

	for _, r := range roots {
		walk(r)
	}

	return out
}

type hierarchyDropItem struct {
	widget.BaseWidget
	check    *widget.Check
	btn      *widget.Button
	entry    *widget.Entry
	entityID int64

	OnTappedSecondary func(*fyne.PointEvent)
}

func (h *hierarchyDropItem) TappedSecondary(ev *fyne.PointEvent) {
	if h.OnTappedSecondary != nil {
		h.OnTappedSecondary(ev)
	}
}

func newHierarchyDropItem() *hierarchyDropItem {
	check := widget.NewCheck("", nil)
	btn := widget.NewButton("", nil)
	btn.Importance = widget.LowImportance
	entry := widget.NewEntry()
	entry.Hide()

	item := &hierarchyDropItem{
		check: check,
		btn:   btn,
		entry: entry,
	}
	item.ExtendBaseWidget(item)
	return item
}

func (h *hierarchyDropItem) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(
		container.NewHBox(
			h.check,
			container.NewMax(h.btn, h.entry),
		),
	)
}

func (h *hierarchyDropItem) Dragged(ev *fyne.DragEvent) {}
func (h *hierarchyDropItem) DragEnd()                   {}

func (h *hierarchyDropItem) DragAccept(data interface{}) bool {
	log.Printf("DragAccept on entity %d with data %#v", h.entityID, data)
	_, ok := data.(DragAsset)
	log.Printf("DragAccept type assertion ok=%v", ok)
	return ok
}

func (h *hierarchyDropItem) Drop(data interface{}) {
	da, ok := data.(DragAsset)
	if !ok {
		log.Printf("Drop: wrong type: %#v", data)
		return
	}

	log.Printf("Drop on entity %d: %+v", h.entityID, da)

	switch da.Type {
	case "texture":
		msg := editorlink.MsgSetComponent{
			EntityID: uint64(h.entityID),
			Name:     "Material",
			Fields: map[string]any{
				"UseTexture":   true,
				"TextureAsset": int(da.ID),
			},
		}
		go editorlink.WriteSetComponent(editorlink.EditorConn, msg)

	case "mesh":
		av := findMeshAssetByID(state.Global.Assets.Meshes, da.ID)
		meshID := filepath.Base(av.Path)

		msg := editorlink.MsgSetComponent{
			EntityID: uint64(h.entityID),
			Name:     "Mesh",
			Fields: map[string]any{
				"MeshID": meshID,
			},
		}
		go editorlink.WriteSetComponent(editorlink.EditorConn, msg)
	}
}

func findTextureAssetByID(textures []state.AssetView, id uint64) state.AssetView {
	for _, t := range textures {
		if t.ID == id {
			return t
		}
	}
	return state.AssetView{}
}

func findMeshAssetByID(meshes []state.AssetView, id uint64) state.AssetView {
	for _, m := range meshes {
		if m.ID == id {
			return m
		}
	}
	return state.AssetView{}
}
func showHierarchyContextMenu(
	item *hierarchyDropItem,
	row HierarchyRow,
	st *state.EditorState,
	list *widget.List,
	onSelect func(int),
) {
	rename := fyne.NewMenuItem("Rename", func() {
		// Trigger inline rename mode
		for idx, e := range st.Entities {
			if e.ID == row.ID {
				// Set renameIndex and refresh
				// We need to store renameIndex globally or in EditorState
				state.Global.RenameIndex = idx
				list.Refresh()
				return
			}
		}
	})

	duplicate := fyne.NewMenuItem("Duplicate", func() {
		if editorlink.EditorConn != nil {
			go editorlink.WriteDuplicateEntity(editorlink.EditorConn, row.ID)
		}
	})

	delete := fyne.NewMenuItem("Delete", func() {
		if editorlink.EditorConn != nil {
			go editorlink.WriteDeleteEntity(editorlink.EditorConn, row.ID, row.Name)
		}
	})
	savePrefab := fyne.NewMenuItem("Save as Prefab…", func() {
		if editorlink.EditorConn != nil {
			name := row.Name
			path := filepath.Join("prefabs", name+".json")
			go editorlink.WriteSavePrefab(editorlink.EditorConn, row.ID, path)
		}
	})
	menu := fyne.NewMenu("", rename, duplicate, delete, savePrefab)
	widget.ShowPopUpMenuAtPosition(menu, fyne.CurrentApp().Driver().CanvasForObject(item), item.Position())
}
