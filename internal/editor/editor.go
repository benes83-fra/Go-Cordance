package editor

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"go-engine/Go-Cordance/internal/assets"
	"go-engine/Go-Cordance/internal/ecs"
	"go-engine/Go-Cordance/internal/ecs/gizmo"
	"go-engine/Go-Cordance/internal/editor/bridge"
	state "go-engine/Go-Cordance/internal/editor/state"
	"go-engine/Go-Cordance/internal/editor/ui"
	"math"
	"sync"
	"time"

	"go-engine/Go-Cordance/internal/editorlink"
	"net"
	"os"
	"path/filepath"
	"strings"

	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// editor: pending transform applier
var (
	pendingTransformsMu  sync.Mutex
	pendingTransforms    = map[int64]editorlink.MsgSetTransform{} // store last msg per entity
	transformApplierOnce sync.Once
	editorDebug          bool = false // set true for verbose logs
)

// Run starts the editor UI for the provided world.
func Run(world *ecs.World) {
	a := app.NewWithID("Go-Cordance Editor")
	a.Settings().SetTheme(theme.DarkTheme())
	win := a.NewWindow("Go-Cordance Editor")
	win.Resize(fyne.NewSize(1000, 600))

	// state
	st := state.Global
	st.Foldout = map[string]bool{"Position": true, "Rotation": true, "Scale": true}
	st.ShowLightGizmos = true
	st.UpdateLocalMaterial = func(entityID int64, fields map[string]any) {
		updateLocalMaterial(world, entityID, fields)
	}

	var hierarchyWidget fyne.CanvasObject
	var hierarchyList *widget.List

	// Create inspector first so we have the rebuild function available.
	inspectorContainer, inspectorRebuild := ui.NewInspectorPanel()
	assetBrowser, assetList := ui.NewAssetBrowserPanel(st, win)

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
	hierarchyWidget, hierarchyList = ui.NewHierarchyPanel(st, win, func(id int) {
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
	// preserve selected entity ID so selection survives reimport/reorder
	var prevSelectedID int64 = 0
	if state.Global.SelectedIndex >= 0 && state.Global.SelectedIndex < len(state.Global.Entities) {
		prevSelectedID = state.Global.Entities[state.Global.SelectedIndex].ID
	}

	// replace snapshot
	state.Global.Entities = ents

	// restore SelectedIndex to point to the same entity ID (if still present)
	if prevSelectedID != 0 {
		state.Global.SelectedIndex = -1
		for i, e := range state.Global.Entities {
			if e.ID == prevSelectedID {
				state.Global.SelectedIndex = i
				break
			}
		}
		// also keep Selection.ActiveID consistent
		if state.Global.SelectedIndex >= 0 {
			state.Global.Selection.ActiveID = state.Global.Entities[state.Global.SelectedIndex].ID
		} else {
			state.Global.Selection.ActiveID = 0
		}
	}

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
	// Apply snapshot transforms into the ECS world so inspector component UIs read fresh values.
	// This must run even when there is no structural change (reorder/reimport only).
	if world != nil {
		for _, e := range ents {
			ecsEnt := world.FindByID(e.ID)
			if ecsEnt == nil {
				continue
			}
			tr := ecsEnt.GetTransform()
			if tr == nil {
				// create and attach a Transform if missing
				tr = &ecs.Transform{}
				ecsEnt.AddComponent(tr)
			}
			tr.Position = e.Position
			tr.Rotation = e.Rotation
			tr.Scale = e.Scale

			// Clear Euler cache so inspector recomputes Euler angles from the new quaternion
			delete(state.Global.EulerCache, uint64(e.ID))
		}
	}

	if structuralChange {
		SyncEditorWorld(world, ents)
	}
	// Build parent → children map
	// Build parent → children map from snapshot
	parentMap := map[int64]int64{}
	childrenMap := map[int64][]int64{}

	// 1) Use Children arrays as primary source
	for _, e := range ents {
		for _, cid := range e.Children {
			cid64 := int64(cid)
			childrenMap[e.ID] = append(childrenMap[e.ID], cid64)
			// only set parent if not already set
			if _, exists := parentMap[cid64]; !exists {
				parentMap[cid64] = e.ID
			}
		}
	}

	// 2) Use Parent field as fallback (without duplicating children)
	for _, e := range ents {
		if e.Parent == 0 {
			continue
		}
		id := e.ID
		pid := int64(e.Parent)

		if _, exists := parentMap[id]; !exists {
			parentMap[id] = pid
			childrenMap[pid] = append(childrenMap[pid], id)
		}
	}

	state.Global.ParentMap = parentMap
	state.Global.ChildrenMap = childrenMap

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

// editor.go

func UpdateEntityTransform(world *ecs.World, id int64, pos bridge.Vec3, rot bridge.Vec4, scale bridge.Vec3) {
	// Update editor snapshot (so inspector transform fields update)
	for i := range state.Global.Entities {
		if state.Global.Entities[i].ID == id {
			state.Global.Entities[i].Position = pos
			state.Global.Entities[i].Rotation = rot
			state.Global.Entities[i].Scale = scale
			delete(state.Global.EulerCache, uint64(id))
			break
		}
	}

	// Also update the ECS world entity's Transform component so component UIs read the new values
	if world == nil {
		return
	}
	ecsEnt := world.FindByID(id)
	if ecsEnt == nil {
		return
	}
	tr := ecsEnt.GetTransform()
	if tr == nil {
		// create and attach a Transform if missing
		tr = &ecs.Transform{}
		ecsEnt.AddComponent(tr)
	}
	tr.Position = pos
	tr.Rotation = rot
	tr.Scale = scale
}

func startEditorLinkClient(world *ecs.World) {
	conn, err := net.Dial("tcp", "localhost:7777")
	if err != nil {
		log.Fatalf("editor: cannot connect to game: %v", err)
	}

	editorlink.EditorConn = conn
	startTransformApplier(world)

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
			pendingTransformsMu.Lock()
			pendingTransforms[int64(m.ID)] = m
			pendingTransformsMu.Unlock()
			// UpdateEntityTransform(world, int64(m.ID), m.Position, m.Rotation, m.Scale)

		case "SetTransformGizmoFinal":
			var m editorlink.MsgSetTransform
			if err := json.Unmarshal(msg.Data, &m); err != nil {
				log.Printf("editor: bad SetTransformGizmoFinal: %v", err)
				continue
			}
			pendingTransformsMu.Lock()
			pendingTransforms[int64(m.ID)] = m
			pendingTransformsMu.Unlock()
			// fyne.DoAndWait(func() {
			// 	UpdateEntityTransform(world, int64(m.ID), m.Position, m.Rotation, m.Scale)
			// 	if state.Global.RefreshUI != nil {
			// 		state.Global.RefreshUI() // rebuild inspector ONCE
			// 	}
			// })

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
					Parent:     e.Parent,
					Children:   e.Children,
				}
			}
			log.Printf("editor: incoming SceneSnapshot with %v entitites", ents)

			fyne.DoAndWait(func() {
				UpdateEntities(world, ents)
			})
		case "AssetMeshThumbnail":
			var t editorlink.MsgAssetMeshThumbnail
			if err := json.Unmarshal(msg.Data, &t); err != nil {
				log.Printf("editor: bad AssetMeshThumbnail: %v", err)
				continue
			}
			data, err := base64.StdEncoding.DecodeString(t.DataB64)
			if err != nil {
				log.Printf("editor: AssetMeshThumbnail base64 decode error: %v", err)
				continue
			}

			fyne.DoAndWait(func() {
				handleMeshSubThumbnail(t.AssetID, t.MeshID, t.Format, data, t.Hash)
			})

		case "AssetThumbnail":
			var t editorlink.MsgAssetThumbnail
			if err := json.Unmarshal(msg.Data, &t); err != nil {
				log.Printf("editor: bad AssetThumbnail: %v", err)
				continue
			}
			// decode base64
			data, err := base64.StdEncoding.DecodeString(t.DataB64)
			if err != nil {
				log.Printf("editor: AssetThumbnail base64 decode error: %v", err)
				continue
			}

			// call UI handler on main thread
			fyne.DoAndWait(func() {
				handleAssetThumbnail(t.AssetID, t.MeshID, t.Format, data, t.Hash)
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
						ID:        v.ID,
						Path:      v.Path,
						Type:      v.Type,
						MeshIDs:   v.MeshIDs,
						MeshThumb: make(map[string]string),
					}
				}

				st.Assets.Materials = make([]state.AssetView, len(m.Materials))
				for i, v := range m.Materials {
					st.Assets.Materials[i] = state.AssetView{
						ID:           v.ID,
						Path:         v.Path,
						Type:         v.Type,
						MaterialData: v.MaterialData,
					}
				}
				st.Assets.Shaders = make([]state.AssetView, len(m.Shaders))
				for i, v := range m.Shaders {
					st.Assets.Shaders[i] = state.AssetView{
						ID:         v.ID,
						Path:       v.Path,
						Type:       v.Type,
						ShaderData: v.ShaderData,
					}
				}

				if st.RefreshUI != nil {
					st.RefreshUI()
				}
			})

		}
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

// SyncEditorWorld rebuilds the editor's ECS world to match the snapshot,
// but reuses existing component instances when possible to avoid losing
// transient editor-side state.
// --- Replace SyncEditorWorld body with this approach ---

func SyncEditorWorld(world *ecs.World, ents []bridge.EntityInfo) {
	// Map existing entities by ID
	oldByID := make(map[int64]*ecs.Entity)
	for _, e := range world.Entities {
		oldByID[e.ID] = e
	}

	// New ordered slice we will assign back to world.Entities
	newEntities := make([]*ecs.Entity, 0, len(ents))

	for _, einfo := range ents {
		// Try to reuse the existing entity pointer
		var ent *ecs.Entity
		if oe, ok := oldByID[einfo.ID]; ok {
			ent = oe
			// Clear components that are no longer present in snapshot
			// Build a set of desired component names
			desired := map[string]bool{}
			for _, cname := range einfo.Components {
				desired[cname] = true
			}
			// Remove components that are not desired
			for _, comp := range ent.Components {
				// assume compTypeName(comp) returns the registry name for the component
				if !desired[ecs.ComponentTypeName(comp)] {
					ent.RemoveComponent(comp)
				}
			}
		} else {
			// create a new entity pointer with same ID (only when not present)
			ent = ecs.NewEntity(einfo.ID)
		}

		// Ensure each component in snapshot exists on ent, reusing instances where possible
		for _, cname := range einfo.Components {
			constructor, ok := ecs.ComponentRegistry[cname]
			if !ok {
				continue
			}
			// If entity already has that component, reuse it
			comp := ent.GetComponent(constructor())
			if comp == nil {
				// create and add
				comp = constructor()
				ent.AddComponent(comp)
			}
			// If it's a Transform, update values in-place
			if cname == "Transform" {
				if tr, ok := comp.(*ecs.Transform); ok {
					tr.Position = einfo.Position
					tr.Rotation = einfo.Rotation
					tr.Scale = einfo.Scale
				}
			}
			// For other components you may want to update cached fields if snapshot carries them
		}

		newEntities = append(newEntities, ent)
	}

	// Replace world.Entities with the new ordered slice (same pointers reused)
	world.Entities = newEntities
}

// func SyncEditorWorld(world *ecs.World, ents []bridge.EntityInfo) {

// 	// Build map of old entities by ID for reuse
// 	oldByID := make(map[int64]*ecs.Entity)
// 	for _, e := range world.Entities {
// 		oldByID[e.ID] = e
// 	}

// 	// New list
// 	newEntities := make([]*ecs.Entity, 0, len(ents))

// 	for _, e := range ents {
// 		// Create a fresh entity object (we will reuse components where possible)
// 		ent := ecs.NewEntity(e.ID)

// 		// If we have an old entity, try to reuse its components
// 		var oldEnt *ecs.Entity
// 		if oe, ok := oldByID[e.ID]; ok {
// 			oldEnt = oe
// 		}

// 		for _, cname := range e.Components {
// 			constructor, ok := ecs.ComponentRegistry[cname]
// 			if !ok {
// 				if cname != "Transform" {
// 					log.Printf("editor: no constructor for component %q in registry", cname)
// 				}
// 				continue
// 			}

// 			// Try to reuse existing component instance from old entity
// 			var comp ecs.Component // replace with your component interface type
// 			if oldEnt != nil {
// 				// ask old entity for a component instance of the same type
// 				oldComp := oldEnt.GetComponent(constructor())
// 				if oldComp != nil {
// 					comp = oldComp
// 				}
// 			}

// 			// If not found, construct a new one
// 			if comp == nil {
// 				comp = constructor()
// 			}
// 			if cname == "Transform" {
// 				if tr, ok := comp.(*ecs.Transform); ok {
// 					// e.Position/Rotation/Scale are bridge.Vec3/Vec4 types; convert as needed
// 					tr.Position = e.Position
// 					tr.Rotation = e.Rotation
// 					tr.Scale = e.Scale
// 				} else {
// 					// defensive: if constructor returned a different concrete type, try to set via interface
// 					log.Printf("editor: Transform component type mismatch for entity %d", e.ID)
// 				}
// 			}
// 			ent.AddComponent(comp)
// 		}

// 		newEntities = append(newEntities, ent)
// 	}

// 	// Replace world.Entities with the rebuilt list
// 	world.Entities = newEntities
// }

// userCacheDir returns a writable cache directory for the current user.
// Falls back to the current working directory if os.UserCacheDir fails.
func userCacheDir() string {
	if dir, err := os.UserCacheDir(); err == nil && dir != "" {
		return filepath.Join(dir, "go-cordance-editor")
	}
	// fallback: use a local ".cache" directory in the working dir
	cwd, err := os.Getwd()
	if err != nil {
		return ".cache"
	}
	return filepath.Join(cwd, ".cache", "go-cordance-editor")
}

// handleAssetThumbnail receives decoded thumbnail bytes and stores them in disk cache,
// updates the editor state, and triggers a UI refresh.
func handleAssetThumbnail(assetID uint64, meshID, format string, data []byte, hash string) {

	cacheDir := filepath.Join(userCacheDir(), "thumbs")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		log.Printf("editor: failed to create thumbnail cache dir: %v", err)
	}

	ext := "png"
	if format != "" {
		ext = format
	}

	safeMesh := safeName(meshID)
	fname := filepath.Join(cacheDir, fmt.Sprintf("%d-%s-%s.%s", assetID, hash, safeMesh, ext))

	if _, err := os.Stat(fname); os.IsNotExist(err) {
		if err := os.WriteFile(fname, data, 0644); err != nil {
			log.Printf("editor: failed to write thumbnail file %s: %v", fname, err)
		}
	}

	// Textures: still only asset-level thumbs
	for i := range state.Global.Assets.Textures {
		if state.Global.Assets.Textures[i].ID == assetID {
			state.Global.Assets.Textures[i].Thumbnail = fname
			if state.Global.RefreshUI != nil {
				state.Global.RefreshUI()
			}
			return
		}
	}

	// Meshes: either asset-level or per-meshID
	for i := range state.Global.Assets.Meshes {
		av := &state.Global.Assets.Meshes[i]
		if av.ID != assetID {
			continue
		}

		if meshID == "" {
			// whole-asset thumbnail

			av.Thumbnail = fname
		} else {

			if av.MeshThumb == nil {
				av.MeshThumb = make(map[string]string)
			}
			av.MeshThumb[meshID] = fname
		}
		break
	}
	for i := range state.Global.Assets.Materials {
		if state.Global.Assets.Materials[i].ID == assetID {
			state.Global.Assets.Materials[i].Thumbnail = fname
			state.Global.RefreshUI()
			return
		}
	}

	if state.Global.RefreshUI != nil {
		state.Global.RefreshUI()
	}
}
func handleMeshSubThumbnail(assetID uint64, meshID, format string, data []byte, hash string) {
	cacheDir := filepath.Join(userCacheDir(), "thumbs")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		log.Printf("editor: failed to create thumbnail cache dir: %v", err)
	}

	ext := "png"
	if format != "" {
		ext = format
	}

	safeMesh := safeName(meshID)
	fname := filepath.Join(cacheDir, fmt.Sprintf("%d-%s-%s.%s", assetID, hash, safeMesh, ext))

	if _, err := os.Stat(fname); os.IsNotExist(err) {
		if err := os.WriteFile(fname, data, 0644); err != nil {
			log.Printf("editor: failed to write mesh thumbnail file %s: %v", fname, err)
		}
	}

	// update state.Global.Assets.Meshes[*].MeshThumb[meshID]
	for i := range state.Global.Assets.Meshes {
		av := &state.Global.Assets.Meshes[i]
		if av.ID == assetID {
			if av.MeshThumb == nil {
				av.MeshThumb = make(map[string]string)
			}
			av.MeshThumb[meshID] = fname
			break
		}
	}

	if state.Global.RefreshUI != nil {
		state.Global.RefreshUI()
	}
}

func updateLocalMaterial(world *ecs.World, entityID int64, fields map[string]any) {
	var ent *ecs.Entity
	for _, e := range world.Entities {
		if e.ID == entityID {
			ent = e
			break
		}
	}
	if ent == nil {
		return
	}

	comp := ent.GetComponent(&ecs.Material{})
	var mat *ecs.Material
	if comp == nil {
		mat = &ecs.Material{}
		ent.AddComponent(mat)
	} else {
		var ok bool
		mat, ok = comp.(*ecs.Material)
		if !ok {
			return
		}
	}

	// --- FULL SYNC OF ALL MATERIAL FIELDS ---

	if v, ok := fields["BaseColor"].([4]float32); ok {
		mat.BaseColor = v
	}
	if v, ok := fields["Ambient"].(float32); ok {
		mat.Ambient = v
	}
	if v, ok := fields["Diffuse"].(float32); ok {
		mat.Diffuse = v
	}
	if v, ok := fields["Specular"].(float32); ok {
		mat.Specular = v
	}
	if v, ok := fields["Shininess"].(float32); ok {
		mat.Shininess = v
	}

	// Texture flags + IDs/assets
	useTextureSet := false
	if v, ok := fields["UseTexture"].(bool); ok {
		mat.UseTexture = v
		useTextureSet = true
	}
	if v, ok := fields["TextureAsset"].(int); ok {
		mat.TextureAsset = assets.AssetID(v)
	}
	if v, ok := fields["TextureID"].(int); ok {
		mat.TextureID = uint32(v)
	}
	// If caller didn’t explicitly set UseTexture, derive it from IDs
	if !useTextureSet {
		mat.UseTexture = (mat.TextureID != 0 || mat.TextureAsset != 0)
	}
	if v, ok := fields["ShaderName"].(string); ok {
		mat.ShaderName = v
	}
	// Normal flags + IDs/assets
	useNormalSet := false
	if v, ok := fields["UseNormal"].(bool); ok {
		mat.UseNormal = v
		useNormalSet = true
	}
	if v, ok := fields["NormalID"].(int); ok {
		mat.NormalID = uint32(v)
	}
	if v, ok := fields["NormalAsset"].(int); ok {
		mat.NormalAsset = assets.AssetID(v)
	}
	if !useNormalSet {
		mat.UseNormal = (mat.NormalID != 0 || mat.NormalAsset != 0)
	}

	mat.Dirty = true

}

func safeName(s string) string {
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")
	return s
}

func transformDifferent(pos bridge.Vec3, newPos [3]float32, rot bridge.Vec4, newRot [4]float32, scale bridge.Vec3, newScale [3]float32) bool {
	const eps = 1e-4
	// positions
	for i := 0; i < 3; i++ {
		if math.Abs(float64(pos[i]-newPos[i])) > eps {
			return true
		}
	}
	// rotation quaternion
	for i := 0; i < 4; i++ {
		if math.Abs(float64(rot[i]-newRot[i])) > eps {
			return true
		}
	}
	// scale
	for i := 0; i < 3; i++ {
		if math.Abs(float64(scale[i]-newScale[i])) > eps {
			return true
		}
	}
	return false
}

func startTransformApplier(world *ecs.World) {
	transformApplierOnce.Do(func() {
		go func() {
			ticker := time.NewTicker(33 * time.Millisecond) // ~30Hz
			defer ticker.Stop()
			for range ticker.C {
				// drain pending map
				pendingTransformsMu.Lock()
				if len(pendingTransforms) == 0 {
					pendingTransformsMu.Unlock()
					continue
				}
				// copy and clear
				copyMap := make(map[int64]editorlink.MsgSetTransform, len(pendingTransforms))
				for id, m := range pendingTransforms {
					copyMap[id] = m
				}
				pendingTransforms = map[int64]editorlink.MsgSetTransform{}
				pendingTransformsMu.Unlock()

				// Determine which entities actually changed (epsilon compare)
				changedIDs := make([]int64, 0, len(copyMap))
				for id, m := range copyMap {
					// find snapshot entity and compare
					for _, ent := range state.Global.Entities {
						if ent.ID == id {
							// cheap epsilon compare for position/rotation/scale
							if transformDifferent(ent.Position, m.Position, ent.Rotation, m.Rotation, ent.Scale, m.Scale) {
								changedIDs = append(changedIDs, id)
							}
							break
						}
					}
				}
				if len(changedIDs) == 0 {
					continue
				}

				// Apply all changes on UI thread in one shot
				fyne.DoAndWait(func() {
					// apply to editor snapshot and ECS world
					for _, id := range changedIDs {
						m := copyMap[id]
						UpdateEntityTransform(world, int64(m.ID), m.Position, m.Rotation, m.Scale)
					}
					// clear Euler cache for changed IDs (UpdateEntityTransform already deletes per-id)
					// single RefreshUI call
					if state.Global.RefreshUI != nil {
						state.Global.RefreshUI()
					}
				})
			}
		}()
	})
}
