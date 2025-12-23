package editor

import (
	state "go-engine/Go-Cordance/internal/editor/state"
	"go-engine/Go-Cordance/internal/editor/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func Run() {
	a := app.New()
	w := a.NewWindow("Go-Cordance Editor")
	w.Resize(fyne.NewSize(1000, 600))

	st := state.New()
	st.Entities = []string{"Camera", "Player", "Light", "Cube", "Enemy"}

	inspector := widget.NewLabel("Select an entity")
	viewport := widget.NewLabel("Viewport Placeholder")

	hierarchy := ui.NewHierarchyPanel(st, func(id int) {
		inspector.SetText("Selected: " + st.Entities[id])
	})

	left := container.NewMax(hierarchy)
	center := container.NewVBox(viewport)
	right := container.NewVBox(inspector)

	split := container.NewHSplit(
		container.NewVSplit(left, center),
		right,
	)
	split.Offset = 0.25

	w.SetContent(split)
	w.ShowAndRun()
}
