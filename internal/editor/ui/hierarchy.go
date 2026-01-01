package ui

import (
	"go-engine/Go-Cordance/internal/ecs/gizmo"
	state "go-engine/Go-Cordance/internal/editor/state"
	"go-engine/Go-Cordance/internal/editorlink"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// NewHierarchyPanel returns a *widget.List that displays st.Entities and calls
// onSelect(index) when the user selects an item.
func NewHierarchyPanel(st *state.EditorState, onSelect func(int)) *widget.List {
	list := widget.NewList(
		func() int { return len(st.Entities) },
		func() fyne.CanvasObject { return widget.NewLabel("entity") },
		func(i int, o fyne.CanvasObject) {
			// defensive: ensure index in range
			if i >= 0 && i < len(st.Entities) {
				o.(*widget.Label).SetText(st.Entities[i].Name)
			} else {
				o.(*widget.Label).SetText("")
			}
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		st.SelectedIndex = id
		onSelect(int(id))
		if editorlink.EditorConn != nil {
			go editorlink.WriteSelectEntity(editorlink.EditorConn, st.Entities[id].ID)
		}
		gizmo.SetGlobalSelectionIDs([]int64{st.Entities[id].ID})
	}

	return list
}
