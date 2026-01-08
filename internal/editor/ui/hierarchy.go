package ui

import (
	"go-engine/Go-Cordance/internal/ecs/gizmo"
	state "go-engine/Go-Cordance/internal/editor/state"
	"go-engine/Go-Cordance/internal/editorlink"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// NewHierarchyPanel returns a *widget.List that displays st.Entities and calls
// onSelect(index) when the user selects an item. This version adds a checkbox
// per row to support multi-selection without modifier keys.
func NewHierarchyPanel(st *state.EditorState, onSelect func(int)) *widget.List {
	// item template: checkbox + label
	// item template: checkbox + button (acts like a clickable label)
	makeItem := func() fyne.CanvasObject {
		check := widget.NewCheck("", nil)
		btn := widget.NewButton("", nil)
		btn.Importance = widget.LowImportance
		row := container.NewHBox(check, btn)
		return row
	}

	list := widget.NewList(
		func() int { return len(st.Entities) },
		makeItem,
		func(i int, o fyne.CanvasObject) {
			// defensive: ensure index in range
			if i < 0 || i >= len(st.Entities) {
				// clear row
				row := o.(*fyne.Container)
				row.Objects[0].(*widget.Check).SetChecked(false)
				row.Objects[1].(*widget.Label).SetText("")
				return
			}

			ent := st.Entities[i]
			row := o.(*fyne.Container)
			check := row.Objects[0].(*widget.Check)
			btn := row.Objects[1].(*widget.Button)

			// set button text
			btn.SetText(ent.Name)

			// set checkbox state based on st.Selection.IDs
			checked := false
			for _, id := range st.Selection.IDs {
				if id == ent.ID {
					checked = true
					break
				}
			}
			// avoid triggering OnChanged when programmatically setting state
			check.OnChanged = nil
			check.SetChecked(checked)

			// checkbox toggles membership in selection
			check.OnChanged = func(checked bool) {
				if checked {
					// Add to multi-selection
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
					// Remove from multi-selection
					newIDs := make([]int64, 0, len(st.Selection.IDs))
					for _, id := range st.Selection.IDs {
						if id != ent.ID {
							newIDs = append(newIDs, id)
						}
					}
					st.Selection.IDs = newIDs
				}

				// Do NOT change ActiveID here
				// Do NOT change SelectedIndex here

				if editorlink.EditorConn != nil {
					go editorlink.WriteSelectEntities(editorlink.EditorConn, st.Selection.IDs)
				}

				gizmo.SetGlobalSelectionIDs(st.Selection.IDs)
			}

			// clicking the button sets the active index (inspector) and notifies remote
			btn.OnTapped = func() {
				// 1. Set primary selection
				st.SelectedIndex = i
				st.Selection.ActiveID = ent.ID

				// 2. Reset multi-selection to only this entity
				st.Selection.IDs = []int64{ent.ID}

				// 3. Uncheck all rows except this one
				for row := 0; row < len(st.Entities); row++ {
					if row == i {
						// This row should be checked
						// We must refresh the list AFTER updating state
						continue
					}
					// Uncheck others by removing them from selection
					// (the UI will reflect this on next Refresh)
				}

				// 4. Notify the game
				if editorlink.EditorConn != nil {
					go editorlink.WriteSelectEntity(editorlink.EditorConn, ent.ID)
					go editorlink.WriteSelectEntities(editorlink.EditorConn, st.Selection.IDs)
				}

				// 5. Refresh UI
				onSelect(i)
			}

		},
	)

	return list
}
