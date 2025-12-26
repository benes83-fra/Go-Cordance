package ui

import (
	"fmt"
	"sort"
	"strconv"

	"go-engine/Go-Cordance/internal/ecs"
	"go-engine/Go-Cordance/internal/editor/registry"
	state "go-engine/Go-Cordance/internal/editor/state"
	"go-engine/Go-Cordance/internal/editorlink"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// NewInspectorPanel returns an inspector container and a rebuild function.
// The rebuild function signature is:
//
//	func(world *ecs.World, st *state.EditorState, hierarchy *widget.List)
func NewInspectorPanel() (*fyne.Container, func(world *ecs.World, st *state.EditorState, hierarchy *widget.List)) {
	cont := container.NewVBox()
	cont.Add(widget.NewLabel("Select an entity"))

	var rebuild func(world *ecs.World, st *state.EditorState, hierarchy *widget.List)

	rebuild = func(world *ecs.World, st *state.EditorState, hierarchy *widget.List) {
		cont.Objects = nil
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

		// -------------------------
		// ADD COMPONENT SECTION
		// -------------------------
		cont.Add(widget.NewSeparator())
		cont.Add(widget.NewLabel("Add Component"))

		componentNames := make([]string, 0, len(registry.Components))
		for name := range registry.Components {
			componentNames = append(componentNames, name)
		}
		sort.Strings(componentNames)

		addSelect := widget.NewSelect(componentNames, nil)
		addButton := widget.NewButton("Add", func() {
			if addSelect.Selected == "" {
				return
			}
			factory := registry.Components[addSelect.Selected]
			if factory != nil {
				e.AddComponent(factory())
				st.Entities = state.Global.Entities
				if hierarchy != nil {
					hierarchy.Refresh()
				}
				rebuild(world, st, hierarchy)
			}
		})

		cont.Add(container.NewHBox(addSelect, addButton))

		// -------------------------
		// NAME COMPONENT
		// -------------------------
		if nComp := e.GetComponent(&ecs.Name{}); nComp != nil {
			nameComp := nComp.(*ecs.Name)

			header := container.NewHBox(
				widget.NewLabel("Name"),
				widget.NewButton("Remove", func() {
					e.RemoveComponent(&ecs.Name{})
					st.Entities = world.ListEntityInfo()
					if hierarchy != nil {
						hierarchy.Refresh()
					}
					rebuild(world, st, hierarchy)
				}),
			)
			cont.Add(header)

			nameEntry := widget.NewEntry()
			nameEntry.SetText(nameComp.Value)
			nameEntry.OnChanged = func(val string) {
				nameComp.Value = val
				st.Entities = world.ListEntityInfo()
				if hierarchy != nil {
					hierarchy.Refresh()
				}
			}
			cont.Add(nameEntry)
		}

		// -------------------------
		// TRANSFORM COMPONENT
		// -------------------------
		if t := e.GetTransform(); t != nil {
			header := container.NewHBox(
				widget.NewLabel("Transform"),
				widget.NewButton("Remove", func() {
					e.RemoveComponent(&ecs.Transform{})
					rebuild(world, st, hierarchy)
				}),
			)
			cont.Add(header)

			x := widget.NewEntry()
			y := widget.NewEntry()
			z := widget.NewEntry()
			x.SetText(fmt.Sprintf("%g", t.Position[0]))
			y.SetText(fmt.Sprintf("%g", t.Position[1]))
			z.SetText(fmt.Sprintf("%g", t.Position[2]))
			x.OnChanged = func(val string) {
				f, _ := strconv.ParseFloat(val, 32)
				st.Entities[st.SelectedIndex].Position[0] = float32(f)

				if editorlink.EditorConn != nil {
					go editorlink.WriteSetTransform(editorlink.EditorConn, editorlink.MsgSetTransform{
						ID:       uint64(st.Entities[st.SelectedIndex].ID), // cast to uint64
						Position: [3]float32(st.Entities[st.SelectedIndex].Position),
						Rotation: [4]float32(st.Entities[st.SelectedIndex].Rotation),
						Scale:    [3]float32(st.Entities[st.SelectedIndex].Scale),
					})
				}
			}

			y.OnChanged = func(val string) {
				f, _ := strconv.ParseFloat(val, 32)
				st.Entities[st.SelectedIndex].Position[1] = float32(f)

				if editorlink.EditorConn != nil {
					go editorlink.WriteSetTransform(editorlink.EditorConn, editorlink.MsgSetTransform{
						ID:       uint64(st.Entities[st.SelectedIndex].ID), // cast to uint64
						Position: [3]float32(st.Entities[st.SelectedIndex].Position),
						Rotation: [4]float32(st.Entities[st.SelectedIndex].Rotation),
						Scale:    [3]float32(st.Entities[st.SelectedIndex].Scale),
					})
				}
			}

			z.OnChanged = func(val string) {
				f, _ := strconv.ParseFloat(val, 32)
				st.Entities[st.SelectedIndex].Position[2] = float32(f)

				if editorlink.EditorConn != nil {
					go editorlink.WriteSetTransform(editorlink.EditorConn, editorlink.MsgSetTransform{
						ID:       uint64(st.Entities[st.SelectedIndex].ID), // cast to uint64
						Position: [3]float32(st.Entities[st.SelectedIndex].Position),
						Rotation: [4]float32(st.Entities[st.SelectedIndex].Rotation),
						Scale:    [3]float32(st.Entities[st.SelectedIndex].Scale),
					})
				}
			}

			apply := widget.NewButton("Apply", func() {
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

			cont.Add(container.NewVBox(
				container.NewHBox(widget.NewLabel("X:"), x),
				container.NewHBox(widget.NewLabel("Y:"), y),
				container.NewHBox(widget.NewLabel("Z:"), z),
				apply,
			))
		}

		// -------------------------
		// TODO: Add more components here
		// -------------------------

		cont.Refresh()
	}

	return cont, rebuild
}
