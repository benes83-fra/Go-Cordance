package ui

import (
	state "go-engine/Go-Cordance/internal/editor/state"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

func NewHierarchyPanel(st *state.EditorState, onSelect func(int)) *widget.List {

	list := widget.NewList(
		func() int {
			return len(st.Entities)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("entity")
		},
		func(i int, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(st.Entities[i])
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		st.SelectedIndex = id
		onSelect(id)
	}

	return list
}
