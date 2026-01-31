package ui

import (
	"go-engine/Go-Cordance/internal/ecs/gizmo"
	"go-engine/Go-Cordance/internal/editor/bridge"
	state "go-engine/Go-Cordance/internal/editor/state"
	"go-engine/Go-Cordance/internal/editorlink"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var lastClickTime time.Time
var lastClickIndex = -1

// NewHierarchyPanel returns BOTH:
// 1) The UI container (button + list)
// 2) The underlying *widget.List for inspector rebuild
func NewHierarchyPanel(st *state.EditorState, onSelect func(int)) (fyne.CanvasObject, *widget.List) {

	dupBtn := widget.NewButton("Duplicate", func() {
		if st.Selection.ActiveID != 0 && editorlink.EditorConn != nil {

			// Capture source entity info BEFORE duplication

			go editorlink.WriteDuplicateEntity(editorlink.EditorConn, st.Selection.ActiveID)
		}
	})

	delBtn := widget.NewButton("Delete", func() {
		if st.Selection.ActiveID != 0 && editorlink.EditorConn != nil {

			// Capture entity info for undo
			var deleted bridge.EntityInfo
			for _, e := range st.Entities {
				if e.ID == st.Selection.ActiveID {
					deleted = e
					break
				}
			}

			// Push structural undo

			// Send delete request to game
			go editorlink.WriteDeleteEntity(editorlink.EditorConn, st.Selection.ActiveID, deleted.Name)
		}
	})

	makeItem := func() fyne.CanvasObject {
		return newHierarchyDropItem()
	}

	list := widget.NewList(
		func() int { return len(st.Entities) },
		makeItem,
		func(i int, o fyne.CanvasObject) {

			if i < 0 || i >= len(st.Entities) {
				row := o.(*fyne.Container)
				row.Objects[0].(*widget.Check).SetChecked(false)
				row.Objects[1].(*widget.Button).SetText("")
				return
			}

			ent := st.Entities[i]
			item := o.(*hierarchyDropItem)
			check := item.check
			btn := item.btn
			item.entityID = ent.ID

			btn.SetText(ent.Name)

			// checkbox state
			checked := false
			for _, id := range st.Selection.IDs {
				if id == ent.ID {
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
						if id == ent.ID {
							found = true
							break
						}
					}
					if !found {
						st.Selection.IDs = append(st.Selection.IDs, ent.ID)
					}
				} else {
					newIDs := make([]int64, 0, len(st.Selection.IDs))
					for _, id := range st.Selection.IDs {
						if id != ent.ID {
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

				if lastClickIndex == i && now.Sub(lastClickTime) < 300*time.Millisecond {
					if editorlink.EditorConn != nil {
						go editorlink.WriteFocusEntity(editorlink.EditorConn, ent.ID)
					}
					return
				}

				lastClickTime = now
				lastClickIndex = i

				st.SelectedIndex = i
				st.Selection.ActiveID = ent.ID
				st.Selection.IDs = []int64{ent.ID}

				if editorlink.EditorConn != nil {
					go editorlink.WriteSelectEntity(editorlink.EditorConn, ent.ID)
					go editorlink.WriteSelectEntities(editorlink.EditorConn, st.Selection.IDs)
				}

				onSelect(i)
			}
		},
	)

	topBar := container.NewHBox(dupBtn, delBtn)
	panel := container.NewBorder(topBar, nil, nil, nil, list)

	return panel, list
}

type hierarchyDropItem struct {
	widget.BaseWidget
	check    *widget.Check
	btn      *widget.Button
	entityID int64
}

func newHierarchyDropItem() *hierarchyDropItem {
	check := widget.NewCheck("", nil)
	btn := widget.NewButton("", nil)
	btn.Importance = widget.LowImportance

	item := &hierarchyDropItem{
		check: check,
		btn:   btn,
	}
	item.ExtendBaseWidget(item)
	return item
}

func (h *hierarchyDropItem) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(
		container.NewHBox(h.check, h.btn),
	)
}

func (h *hierarchyDropItem) Dragged(ev *fyne.DragEvent) {}
func (h *hierarchyDropItem) DragEnd()                   {}

func (h *hierarchyDropItem) DragAccept(data interface{}) bool {
	_, ok := data.(uint64)
	return ok
}

func (h *hierarchyDropItem) Drop(data interface{}) {
	assetID := data.(uint64)

	msg := editorlink.MsgSetComponent{
		EntityID: uint64(h.entityID),
		Name:     "Material",
		Fields: map[string]any{
			"UseTexture":   true,
			"TextureAsset": int(assetID),
		},
	}

	if editorlink.EditorConn != nil {
		go editorlink.WriteSetComponent(editorlink.EditorConn, msg)
	}
}
