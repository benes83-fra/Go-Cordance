package ui

import (
	"fmt"
	"strconv"

	"go-engine/Go-Cordance/internal/ecs"
	state "go-engine/Go-Cordance/internal/editor/state"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// NewInspectorPanel returns an inspector container and a rebuild function.
// The rebuild function signature is:
//
//	func(world *ecs.World, st *state.EditorState, hierarchy *widget.List)
//
// Call rebuild on the UI goroutine (Fyne ensures callbacks run on the UI thread).
func NewInspectorPanel() (*fyne.Container, func(world *ecs.World, st *state.EditorState, hierarchy *widget.List)) {
	cont := container.NewVBox()
	cont.Add(widget.NewLabel("Select an entity"))

	var rebuild func(world *ecs.World, st *state.EditorState, hierarchy *widget.List)

	rebuild = func(world *ecs.World, st *state.EditorState, hierarchy *widget.List) {
		// Caller must ensure this runs on the UI/main goroutine.
		cont.Objects = nil // clear
		cont.Refresh()

		// No selection
		if st.SelectedIndex < 0 || st.SelectedIndex >= len(st.Entities) {
			cont.Add(widget.NewLabel("Select an entity"))
			cont.Refresh()
			return
		}

		selected := st.Entities[st.SelectedIndex]
		e := world.FindByID(selected.ID)
		if e == nil {
			cont.Add(widget.NewLabel("Entity not found"))
			cont.Refresh()
			return
		}

		// Header
		cont.Add(widget.NewLabel(fmt.Sprintf("Entity %d", e.ID)))

		// Name editor (if Name component exists)
		if nComp := e.GetComponent(&ecs.Name{}); nComp != nil {
			if nameComp, ok := nComp.(*ecs.Name); ok {
				nameEntry := widget.NewEntry()
				nameEntry.SetText(nameComp.Value)
				nameEntry.OnChanged = func(val string) {
					nameComp.Value = val
					// update state list and refresh hierarchy
					st.Entities = world.ListEntityInfo()
					if hierarchy != nil {
						hierarchy.Refresh()
					}
				}
				cont.Add(widget.NewLabel("Name"))
				cont.Add(nameEntry)
			}
		} else {
			// Offer to add a Name component
			addBtn := widget.NewButton("Add Name", func() {
				e.AddComponent(ecs.NewName(fmt.Sprintf("Entity %d", e.ID)))
				st.Entities = world.ListEntityInfo()
				if hierarchy != nil {
					hierarchy.Refresh()
				}
				// Rebuild to show the new name field
				rebuild(world, st, hierarchy)
			})
			cont.Add(addBtn)
		}

		// Transform editor (position only)
		if t := e.GetTransform(); t != nil {
			cont.Add(widget.NewLabel("Transform Position"))

			x := widget.NewEntry()
			y := widget.NewEntry()
			z := widget.NewEntry()
			x.SetText(fmt.Sprintf("%g", t.Position[0]))
			y.SetText(fmt.Sprintf("%g", t.Position[1]))
			z.SetText(fmt.Sprintf("%g", t.Position[2]))

			apply := widget.NewButton("Apply Position", func() {
				if fx, err := strconv.ParseFloat(x.Text, 32); err == nil {
					t.Position[0] = float32(fx)
				}
				if fy, err := strconv.ParseFloat(y.Text, 32); err == nil {
					t.Position[1] = float32(fy)
				}
				if fz, err := strconv.ParseFloat(z.Text, 32); err == nil {
					t.Position[2] = float32(fz)
				}
				t.Dirty = true
			})

			form := container.NewVBox(
				container.NewHBox(widget.NewLabel("X:"), x),
				container.NewHBox(widget.NewLabel("Y:"), y),
				container.NewHBox(widget.NewLabel("Z:"), z),
				apply,
			)
			cont.Add(form)
		} else {
			cont.Add(widget.NewLabel("No Transform component"))
		}

		cont.Refresh()
	}

	return cont, rebuild
}
