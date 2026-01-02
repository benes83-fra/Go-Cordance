package ui

import (
	"go-engine/Go-Cordance/internal/ecs/gizmo"
	state "go-engine/Go-Cordance/internal/editor/state"
	"go-engine/Go-Cordance/internal/editorlink"
	"log"

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
					// add if not present
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
					if editorlink.EditorConn != nil {
						go editorlink.WriteSelectEntities(editorlink.EditorConn, st.Selection.IDs)
					}
				} else {
					// remove if present
					newIDs := make([]int64, 0, len(st.Selection.IDs))
					for _, id := range st.Selection.IDs {
						if id != ent.ID {
							newIDs = append(newIDs, id)
						}
					}
					st.Selection.IDs = newIDs

				}

				// keep ActiveID consistent: if nothing selected, clear; otherwise keep existing active or set to this id
				if len(st.Selection.IDs) == 0 {
					st.Selection.ActiveID = 0
				} else {
					// if active not in selection, set active to first selected
					foundActive := false
					for _, id := range st.Selection.IDs {
						if id == st.Selection.ActiveID {
							foundActive = true
							break
						}
					}
					if !foundActive {
						st.Selection.ActiveID = st.Selection.IDs[0]
					}
				}

				// forward selection to gizmo bridge
				gizmo.SetGlobalSelectionIDs(st.Selection.IDs)
			}

			// clicking the button sets the active index (inspector) and notifies remote
			btn.OnTapped = func() {
				st.SelectedIndex = i
				log.Printf("hierarchy: button tapped, setting SelectedIndex to %d", i)
				onSelect(i)
				if editorlink.EditorConn != nil {
					go editorlink.WriteSelectEntity(editorlink.EditorConn, ent.ID)
				}
				check.SetChecked(true)
				// also make this the active selection in the multi-select structure
				st.Selection.ActiveID = ent.ID
				// ensure active is included in selection IDs (optional: keep single-click as "make active only")
				// If you prefer single-click to select only this entity, uncomment:
				//st.Selection.IDs = []int64{ent.ID}
				log.Printf("editor.hierarchy: checkbox toggled, selection now %v", st.Selection.IDs)
				if editorlink.EditorConn != nil {
					go editorlink.WriteSelectEntities(editorlink.EditorConn, st.Selection.IDs)
				}
				log.Printf("editor.hierarchy: checkbox toggled, selection now %v", st.Selection.IDs)
			}
		},
	)

	return list
}
