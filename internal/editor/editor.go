package editor

import (
	"go-engine/Go-Cordance/internal/ecs"
	state "go-engine/Go-Cordance/internal/editor/state"
	"go-engine/Go-Cordance/internal/editor/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func Run(world *ecs.World) {
	a := app.New()
	win := a.NewWindow("Go-Cordance Editor")
	win.Resize(fyne.NewSize(1000, 600))

	st := state.New()
	st.Entities = world.ListEntityInfo()

	inspector := widget.NewLabel("Select an entity")
	viewport := widget.NewLabel("Viewport Placeholder")

	hierarchy := ui.NewHierarchyPanel(st, func(id int) {
		inspector.SetText("Selected: " + st.Entities[id].Name)
	})

	left := container.NewMax(hierarchy)
	center := container.NewVBox(viewport)
	right := container.NewVBox(inspector)

	split := container.NewHSplit(
		container.NewVSplit(left, center),
		right,
	)
	split.Offset = 0.25

	win.SetContent(split)
	win.ShowAndRun()
}
