package editor

import (
	"encoding/json"
	"go-engine/Go-Cordance/internal/ecs"
	"go-engine/Go-Cordance/internal/ecs/gizmo"
	"go-engine/Go-Cordance/internal/editor/bridge"
	state "go-engine/Go-Cordance/internal/editor/state"
	"go-engine/Go-Cordance/internal/editor/ui"
	"go-engine/Go-Cordance/internal/editorlink"
	"net"
	"time"

	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var lastTransformUIRedraw time.Time
var incomingTransforms = make(chan editorlink.MsgSetTransform, 128)

// Run starts the editor UI for the provided world.
func Run(world *ecs.World) {
	a := app.New()
	a.Settings().SetTheme(theme.DarkTheme())
	win := a.NewWindow("Go-Cordance Editor")
	win.Resize(fyne.NewSize(1000, 600))
	startEditorLinkClient()

	// state
	st := state.Global
	st.Foldout = map[string]bool{"Position": true, "Rotation": true, "Scale": true}
	var hierarchyWidget *widget.List
	// Create inspector first so we have the rebuild function available.
	inspectorContainer, inspectorRebuild := ui.NewInspectorPanel()

	state.Global.RefreshUI = func() {
		hierarchyWidget.Refresh()
		inspectorRebuild(world, st, hierarchyWidget)

	}

	// Now create the hierarchy and pass a callback that calls inspectorRebuild.
	hierarchyWidget = ui.NewHierarchyPanel(st, func(id int) {
		// This callback runs on the UI goroutine (Fyne), so it's safe to call rebuild directly.
		st.SelectedIndex = id
		inspectorRebuild(world, st, hierarchyWidget)
	})

	// viewport placeholder
	viewport := widget.NewLabel("Viewport Placeholder")

	left := container.NewMax(hierarchyWidget)
	center := container.NewVBox(viewport)
	right := container.NewVBox(inspectorContainer)

	split := container.NewHSplit(container.NewVSplit(left, center), right)
	split.Offset = 0.25

	win.SetContent(split)
	win.Show()

	// initial build (no selection)
	inspectorRebuild(world, st, hierarchyWidget)
	win.SetCloseIntercept(func() {
		win.Close()
	})

	win.ShowAndRun()
}

func UpdateEntities(ents []bridge.EntityInfo) {
	log.Printf("editor: UpdateEntities called with %d entities", len(ents))
	for i, e := range ents {
		log.Printf("  Entity %d: ID=%d, Name=%s", i, e.ID, e.Name)
	}
	state.Global.Entities = ents
	// prune selection IDs that no longer exist
	valid := map[int64]bool{}
	for _, e := range ents {
		valid[e.ID] = true
	}
	newIDs := make([]int64, 0, len(state.Global.Selection.IDs))
	for _, id := range state.Global.Selection.IDs {
		if valid[id] {
			newIDs = append(newIDs, id)
		}
	}
	state.Global.Selection.IDs = newIDs

	// If there is a selected index, forward it to the gizmo; otherwise clear selection.
	// Forward current multi-selection to gizmo if present; otherwise fall back to single SelectedIndex.
	if len(state.Global.Selection.IDs) > 0 {
		gizmo.SetGlobalSelectionIDs(state.Global.Selection.IDs)
	} else if state.Global.SelectedIndex >= 0 && state.Global.SelectedIndex < len(state.Global.Entities) {
		id := state.Global.Entities[state.Global.SelectedIndex].ID
		gizmo.SetGlobalSelectionIDs([]int64{id})
	} else {
		gizmo.SetGlobalSelectionIDs(nil)
	}

	if state.Global.RefreshUI != nil {
		log.Printf("editor: RefreshUI triggered")
		state.Global.RefreshUI()
	}
}

func UpdateEntityTransform(id int64, pos bridge.Vec3, rot bridge.Vec4, scale bridge.Vec3) {
	for i := range state.Global.Entities {
		if state.Global.Entities[i].ID == id {
			state.Global.Entities[i].Position = pos
			state.Global.Entities[i].Rotation = rot
			state.Global.Entities[i].Scale = scale
			return
		}
	}
}
func startEditorLinkClient() {
	conn, err := net.Dial("tcp", "localhost:7777")
	if err != nil {
		log.Fatalf("editor: cannot connect to game: %v", err)
	}

	go editorReadLoop(conn)
}
func editorReadLoop(conn net.Conn) {
	for {
		msg, err := editorlink.ReadMsg(conn)
		if err != nil {
			log.Printf("editor: read error: %v", err)
			return
		}

		switch msg.Type {
		case "SetTransformGizmo":
			var m editorlink.MsgSetTransform
			if err := json.Unmarshal(msg.Data, &m); err != nil {
				log.Printf("editor: bad SetTransformGizmo: %v", err)
				continue
			}

			fyne.DoAndWait(func() {
				UpdateEntityTransform(int64(m.ID), m.Position, m.Rotation, m.Scale)
				if state.Global.RefreshUI != nil {
					state.Global.RefreshUI()
				}
			})

		}
	}
}
