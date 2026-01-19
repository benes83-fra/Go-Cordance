package ui

import (
	"go-engine/Go-Cordance/internal/ecs/gizmo"
	state "go-engine/Go-Cordance/internal/editor/state"
	"go-engine/Go-Cordance/internal/editorlink"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var lastClickTime time.Time
var lastClickIndex = -1

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
				now := time.Now()

				// Detect double-click on the same row
				if lastClickIndex == i && now.Sub(lastClickTime) < 300*time.Millisecond {
					// Double-click detected → send FocusEntity
					if editorlink.EditorConn != nil {
						go editorlink.WriteFocusEntity(editorlink.EditorConn, ent.ID)
					}
					return
				}

				// Update click tracking
				lastClickTime = now
				lastClickIndex = i

				// --- Normal single-click behavior ---
				st.SelectedIndex = i
				st.Selection.ActiveID = ent.ID

				// Reset multi-selection to only this entity
				st.Selection.IDs = []int64{ent.ID}

				// Notify the game
				if editorlink.EditorConn != nil {
					go editorlink.WriteSelectEntity(editorlink.EditorConn, ent.ID)
					go editorlink.WriteSelectEntities(editorlink.EditorConn, st.Selection.IDs)
				}

				// Refresh inspector
				onSelect(i)
			}

		},
	)

	return list
}
