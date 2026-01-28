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

	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Run starts the editor UI for the provided world.
func Run(world *ecs.World) {
	a := app.New()
	a.Settings().SetTheme(theme.DarkTheme())
	win := a.NewWindow("Go-Cordance Editor")
	win.Resize(fyne.NewSize(1000, 600))

	// state
	st := state.Global
	st.Foldout = map[string]bool{"Position": true, "Rotation": true, "Scale": true}
	st.ShowLightGizmos = true

	var hierarchyWidget fyne.CanvasObject
	var hierarchyList *widget.List

	// Create inspector first so we have the rebuild function available.
	inspectorContainer, inspectorRebuild := ui.NewInspectorPanel()
	assetBrowser, assetList := ui.NewAssetBrowserPanel(st)

	state.Global.RefreshUI = func() {
		if hierarchyList != nil {
			hierarchyList.Refresh()
		}
		if assetList != nil {
			assetList.Refresh()
		}
		inspectorRebuild(world, st, hierarchyList)
	}

	// Now create the hierarchy and pass a callback that calls inspectorRebuild.
	hierarchyWidget, hierarchyList = ui.NewHierarchyPanel(st, func(id int) {
		st.SelectedIndex = id
		inspectorRebuild(world, st, hierarchyList)
	})

	// viewport placeholder
	viewport := widget.NewLabel("Viewport Placeholder")

	// toolbar / settings row
	showGizmosCheck := widget.NewCheck("Show Light Gizmos", func(v bool) {
		st.ShowLightGizmos = v
		log.Printf("ShowLightGizmos set to %v", v)

		if editorlink.EditorConn != nil {
			editorlink.WriteSetEditorFlag(
				editorlink.EditorConn,
				editorlink.MsgSetEditorFlag{ShowLightGizmos: v},
			)
		}
	})

	showGizmosCheck.SetChecked(st.ShowLightGizmos)

	viewportColumn := container.NewVBox(
		container.NewHBox(showGizmosCheck, layout.NewSpacer()),
		viewport,
	)

	left := container.NewMax(hierarchyWidget)
	center := container.NewVBox(viewportColumn)
	right := container.NewAppTabs(
		container.NewTabItem("Inspector", inspectorContainer),
		container.NewTabItem("Assets", assetBrowser),
	)
	right.SetTabLocation(container.TabLocationTop)

	split := container.NewHSplit(container.NewVSplit(left, center), right)
	split.Offset = 0.25

	win.SetContent(split)
	win.Show()
	go startEditorLinkClient(world)
	// initial build (no selection)
	inspectorRebuild(world, st, hierarchyList)

	win.SetCloseIntercept(func() {
		st.SplitOffset = split.Offset
		win.Close()
	})

	win.ShowAndRun()
}

func UpdateEntities(world *ecs.World, ents []bridge.EntityInfo) {
	log.Printf("editor: UpdateEntities called with %d entities", len(ents))
	for i, e := range ents {
		log.Printf(" Entity %d: ID=%d, Name=%s, Components=%v", i, e.ID, e.Name, e.Components)
	}
	state.Global.Entities = ents

	structuralChange := false
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
	// ❗ NEW: fix SelectedIndex if it points to a deleted entity
	if state.Global.SelectedIndex >= len(ents) {
		state.Global.SelectedIndex = -1
		state.Global.Selection.ActiveID = 0
		state.Global.Selection.IDs = nil
	}

	for _, e := range ents {
		last, ok := state.Global.LastComponents[e.ID]
		if !ok || !equalStringSlices(last, e.Components) {
			structuralChange = true
			state.Global.LastComponents[e.ID] = append([]string{}, e.Components...)
		}
	}
	if structuralChange {
		SyncEditorWorld(world, ents)
	}
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

func startEditorLinkClient(world *ecs.World) {
	conn, err := net.Dial("tcp", "localhost:7777")
	if err != nil {
		log.Fatalf("editor: cannot connect to game: %v", err)
	}

	editorlink.EditorConn = conn

	// Request initial snapshot
	go editorlink.WriteRequestSceneSnapshot(conn)
	go editorlink.WriteRequestAssetList(conn)

	go editorReadLoop(conn, world)
}

func editorReadLoop(conn net.Conn, world *ecs.World) {
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
			UpdateEntityTransform(int64(m.ID), m.Position, m.Rotation, m.Scale)

		case "SetTransformGizmoFinal":
			var m editorlink.MsgSetTransform
			if err := json.Unmarshal(msg.Data, &m); err != nil {
				log.Printf("editor: bad SetTransformGizmoFinal: %v", err)
				continue
			}

			fyne.DoAndWait(func() {
				UpdateEntityTransform(int64(m.ID), m.Position, m.Rotation, m.Scale)
				if state.Global.RefreshUI != nil {
					state.Global.RefreshUI() // rebuild inspector ONCE
				}
			})

		case "SceneSnapshot":
			var snap editorlink.MsgSceneSnapshot
			if err := json.Unmarshal(msg.Data, &snap); err != nil {
				log.Printf("editor: bad SceneSnapshot: %v", err)
				continue
			}

			// Convert snapshot to bridge.EntityInfo
			ents := make([]bridge.EntityInfo, len(snap.Snapshot.Entities))
			for i, e := range snap.Snapshot.Entities {
				ents[i] = bridge.EntityInfo{
					ID:         int64(e.ID),
					Name:       e.Name,
					Position:   bridge.Vec3(e.Position),
					Rotation:   bridge.Vec4(e.Rotation),
					Scale:      bridge.Vec3(e.Scale),
					Components: e.Components,
				}
			}

			fyne.DoAndWait(func() {
				UpdateEntities(world, ents)
			})
		case "AssetList":
			var m editorlink.MsgAssetList
			json.Unmarshal(msg.Data, &m)

			fyne.DoAndWait(func() {
				st := state.Global

				// Convert message → editor state
				st.Assets.Textures = make([]state.AssetView, len(m.Textures))
				for i, v := range m.Textures {
					st.Assets.Textures[i] = state.AssetView{
						ID:   v.ID,
						Path: v.Path,
						Type: v.Type,
					}
					log.Printf("Loaded texture asset: ID=%d, Path=%s, Type=%s", v.ID, v.Path, v.Type)
				}

				st.Assets.Meshes = make([]state.AssetView, len(m.Meshes))
				for i, v := range m.Meshes {
					st.Assets.Meshes[i] = state.AssetView{
						ID:   v.ID,
						Path: v.Path,
						Type: v.Type,
					}
				}

				st.Assets.Materials = make([]state.AssetView, len(m.Materials))
				for i, v := range m.Materials {
					st.Assets.Materials[i] = state.AssetView{
						ID:   v.ID,
						Path: v.Path,
						Type: v.Type,
					}
				}
				if st.RefreshUI != nil {
					st.RefreshUI()
				}
			})

		}
	}
}

// SyncEditorWorld rebuilds the editor's ECS world to match the snapshot.
func SyncEditorWorld(world *ecs.World, ents []bridge.EntityInfo) {
	// Clear the editor ECS
	world.Entities = nil

	// Rebuild entities
	for _, e := range ents {
		ent := ecs.NewEntity(e.ID)

		// Add components based on snapshot
		for _, cname := range e.Components {
			constructor, ok := ecs.ComponentRegistry[cname]
			if !ok {
				log.Printf("editor: no constructor for component %q in registry", cname)
				continue
			}
			comp := constructor()
			log.Printf("editor: Snapshot Components %v", comp)
			ent.AddComponent(comp)
		}

		world.Entities = append(world.Entities, ent)
	}
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
