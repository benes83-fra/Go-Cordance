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

// Run starts the editor UI for the provided world.
func Run(world *ecs.World) {
	a := app.New()
	win := a.NewWindow("Go-Cordance Editor")
	win.Resize(fyne.NewSize(1000, 600))

	// state
	st := state.New()
	st.Entities = world.ListEntityInfo()
	var hierarchyWidget *widget.List
	// Create inspector first so we have the rebuild function available.
	inspectorContainer, inspectorRebuild := ui.NewInspectorPanel()

	// Now create the hierarchy and pass a callback that calls inspectorRebuild.
	hierarchyWidget = ui.NewHierarchyPanel(st, func(id int) {
		// This callback runs on the UI goroutine (Fyne), so it's safe to call rebuild directly.
		st.SelectedIndex = id
		inspectorRebuild(world, st, hierarchyWidget)
	})

	// viewport placeholder
	viewport := widget.NewLabel("Viewport Placeholder")

	// layout
	left := container.NewMax(hierarchyWidget)
	center := container.NewVBox(viewport)
	right := container.NewVBox(inspectorContainer)

	split := container.NewHSplit(container.NewVSplit(left, center), right)
	split.Offset = 0.25

	win.SetContent(split)
	win.Show()

	// initial build (no selection)
	inspectorRebuild(world, st, hierarchyWidget)

	win.ShowAndRun()
}
