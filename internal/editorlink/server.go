package editorlink

import (
	"encoding/json"
	"log"
	"net"

	"go-engine/Go-Cordance/internal/assets"
	"go-engine/Go-Cordance/internal/ecs"
	"go-engine/Go-Cordance/internal/ecs/gizmo"
	"go-engine/Go-Cordance/internal/editor/bridge"
	state "go-engine/Go-Cordance/internal/editor/state"
	"go-engine/Go-Cordance/internal/editor/undo"
	"go-engine/Go-Cordance/internal/thumbnails"

	"go-engine/Go-Cordance/internal/scene"
)

var EditorConn net.Conn
var lastLightVersion = map[uint64]uint64{} // entityID -> version
var Mgr *thumbnails.Manager

// StartServer exposes the given Scene to a single editor client.
func StartServer(addr string, sc *scene.Scene, camSys *ecs.CameraSystem) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("editorlink: listen %s: %v", addr, err)
		return
	}
	log.Printf("editorlink: listening on %s", addr)
	// create manager with send callback that uses editorlink.WriteAssetThumbnail
	mgr := thumbnails.NewManager(2, 256, func(conn net.Conn, assetID uint64, format string, data []byte, hash string) error {
		// this is inside editorlink package, so we can call WriteAssetThumbnail directly
		return WriteAssetThumbnail(conn, assetID, format, data, hash)
	})
	Mgr = mgr
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("editorlink: accept: %v", err)
			continue
		}
		EditorConn = conn
		log.Printf("editorlink: editor connected from %s", conn.RemoteAddr())
		go handleConn(conn, sc, camSys)
		// After EditorConn = conn
		//go WriteTextureList(conn, ecs.TextureNames, ecs.TextureIDs) Legacy Texture communication
		//log.Printf("game: sent texture list to editor: %v", ecs.TextureNames)

	}
}

func handleConn(conn net.Conn, sc *scene.Scene, camSys *ecs.CameraSystem) {
	defer conn.Close()
	// inside the game-side editorlink message handler

	for {
		msg, err := readMsg(conn)
		if err != nil {
			log.Printf("editorlink: read: %v", err)
			return
		}
		log.Printf("editorlink: received msg type=%s data=%s", msg.Type, string(msg.Data))

		switch msg.Type {
		case "RequestSceneSnapshot":
			snap := buildSceneSnapshot(sc)
			resp := MsgSceneSnapshot{Snapshot: snap}
			if err := writeMsg(conn, "SceneSnapshot", resp); err != nil {
				log.Printf("editorlink: write SceneSnapshot: %v", err)
				return
			}
		case "SetTransform":
			var msgST MsgSetTransform
			if err := json.Unmarshal(msg.Data, &msgST); err != nil {
				log.Printf("editorlink: bad SetTransform: %v", err)
				continue
			}

			// Find entity
			ent := sc.World().FindByID(int64(msgST.ID))
			if ent == nil {
				log.Printf("editorlink: SetTransform: entity %d not found", msgST.ID)
				continue
			}

			// Update transform component
			if tr, ok := ent.GetComponent((*ecs.Transform)(nil)).(*ecs.Transform); ok {
				tr.Position = msgST.Position
				tr.Rotation = msgST.Rotation
				tr.Scale = msgST.Scale
			}

			log.Printf("editorlink: updated transform for %d", msgST.ID)

		case "SelectEntity":
			var sel MsgSelectEntity
			if err := json.Unmarshal(msg.Data, &sel); err != nil {
				log.Printf("editorlink: bad SelectEntity: %v", err)
				continue
			}
			ent := sc.World().FindByID(int64(sel.ID))
			if ent != nil {
				sc.Selected = ent
				sc.SelectedEntity = sel.ID
			}
			log.Printf("editorlink: SelectEntity %d ", sel.ID)
		case "SelectEntities":
			var sels MsgSelectEntities
			if err := json.Unmarshal(msg.Data, &sels); err != nil {
				log.Printf("editorlink: bad SelectEntities: %v", err)
				continue
			}
			log.Printf("editorlink: SelectEntities %v", sels.IDs)
			// set selection IDs in gizmo system
			var ids []int64
			for _, id := range sels.IDs {
				ids = append(ids, int64(id))
			}
			gizmo.SetGlobalSelectionIDs(ids)
		case "SetPivotMode":
			var pm MsgSetPivotMode
			if err := json.Unmarshal(msg.Data, &pm); err != nil {
				log.Printf("editorlink: bad SetPivotMode: %v", err)
				continue
			}
			log.Printf("editorlink: SetPivotMode %s", pm.Mode)
			if pm.Mode == "pivot" {
				gizmo.SetGlobalPivotMode(state.PivotModePivot)
			} else if pm.Mode == "center" {
				gizmo.SetGlobalPivotMode(state.PivotModeCenter)
			}
		case "SetComponent":
			var m MsgSetComponent
			if err := json.Unmarshal(msg.Data, &m); err != nil {
				log.Printf("game: bad SetComponent: %v", err)
				continue
			}
			applySetComponent(sc, m)

		case "RemoveComponent":
			var m MsgRemoveComponent
			if err := json.Unmarshal(msg.Data, &m); err != nil {
				log.Printf("game: bad RemoveComponent: %v", err)
				continue
			}
			applyRemoveComponent(sc, m)
		case "SetEditorFlag":
			var m MsgSetEditorFlag
			json.Unmarshal(msg.Data, &m)
			gizmo.SetGlobalShowLightGizmos(m.ShowLightGizmos)
		case "FocusEntity":
			var m MsgFocusEntity
			if err := json.Unmarshal(msg.Data, &m); err != nil {
				log.Printf("editorlink: bad FocusEntity: %v", err)
				continue
			}

			ent := sc.World().FindByID(int64(m.ID))
			if ent == nil {
				log.Printf("editorlink: FocusEntity: entity %d not found", m.ID)
				continue
			}

			// Tell the camera system to focus
			camSys.FocusOn(ent)
		case "DuplicateEntity":
			var m MsgDuplicateEntity
			json.Unmarshal(msg.Data, &m)

			src := sc.World().FindByID(int64(m.ID))
			if src == nil {
				log.Printf("DuplicateEntity: entity %d not found", m.ID)
				continue
			}

			dup := sc.DuplicateEntity(src)

			// Build full EntityInfo from ECS
			//	info := getEntityInfo(dup)
			log.Printf("editorlink: DuplicateEntity created entity %d (from %d)", dup.ID, src.ID)

			// Push undo command

			undo.Global.PushStructural(undo.CreateEntityCommand{Entity: dup})
			log.Printf("editorlink: UndoStack after DuplicateEntity contains %v.", undo.Global)
			sc.Selected = dup
			sc.SelectedEntity = uint64(dup.ID)

			if EditorConn != nil {
				snap := buildSceneSnapshot(sc)
				writeMsg(EditorConn, "SceneSnapshot", MsgSceneSnapshot{Snapshot: snap})
			}

		case "DeleteEntity":
			var m MsgDeleteEntity
			if err := json.Unmarshal(msg.Data, &m); err != nil {
				log.Printf("bad DeleteEntity: %v", err)
				continue
			}

			ent := sc.World().FindByID(int64(m.ID))

			if ent != nil {
				//info  := getEntityInfo(ent)

				undo.Global.PushStructural(undo.DeleteEntityCommand{Entity: ent})
				log.Printf("editorlink: UndoStack after DeleteEntity contains %v.", undo.Global)
			}

			sc.DeleteEntityByID(m.ID)

			if EditorConn != nil {
				snap := buildSceneSnapshot(sc)
				writeMsg(EditorConn, "SceneSnapshot", MsgSceneSnapshot{Snapshot: snap})
			}
		case "RequestAssetList":
			resp := buildAssetList()
			if err := writeMsg(conn, "AssetList", resp); err != nil {
				log.Printf("editorlink: failed to send AssetList: %v", err)
			}
		case "RequestThumbnail":
			thumbnails.HandleRequestThumbnail(msg.Data, EditorConn, Mgr)
		case "RequestMeshThumbnail":
			var m MsgRequestMeshThumbnail
			if err := json.Unmarshal(msg.Data, &m); err != nil {
				log.Printf("game: bad RequestMeshThumbnail: %v", err)
				continue
			}

			data, hash, err := thumbnails.GenerateMeshSubThumbnailBytes(m.AssetID, m.MeshID, m.Size)
			if err != nil {
				log.Printf("game: GenerateMeshSubThumbnailBytes failed: %v", err)
				continue
			}

			if err := SendAssetMeshThumbnail(conn, m.AssetID, m.MeshID, "png", data, hash); err != nil {
				log.Printf("game: SendAssetMeshThumbnail failed: %v", err)
			}
		case "RequestThumbnailMesh":
			thumbnails.HandleRequestThumbnailMesh(msg.Data, conn, Mgr)

		default:
			log.Printf("editorlink: unknown msg type %q", msg.Type)
		}
	}
}

func getEntityInfo(dup *ecs.Entity) bridge.EntityInfo {
	name := dup.GetComponent((*ecs.Name)(nil)).(*ecs.Name).Value

	var pos [3]float32
	var rot [4]float32
	var scale [3]float32
	var comps []string

	if c := dup.GetComponent((*ecs.Transform)(nil)); c != nil {
		tr := c.(*ecs.Transform)
		pos = tr.Position
		rot = tr.Rotation
		scale = tr.Scale
		comps = append(comps, "Transform")
	}
	if dup.GetComponent((*ecs.Material)(nil)) != nil {
		comps = append(comps, "Material")
	}
	if dup.GetComponent((*ecs.RigidBody)(nil)) != nil {
		comps = append(comps, "RigidBody")
	}
	if dup.GetComponent((*ecs.ColliderSphere)(nil)) != nil {
		comps = append(comps, "ColliderSphere")
	}
	if dup.GetComponent((*ecs.ColliderAABB)(nil)) != nil {
		comps = append(comps, "ColliderAABB")
	}
	if dup.GetComponent((*ecs.ColliderPlane)(nil)) != nil {
		comps = append(comps, "ColliderPlane")
	}
	if dup.GetComponent((*ecs.LightComponent)(nil)) != nil {
		comps = append(comps, "Light")
	}

	info := bridge.EntityInfo{
		ID:         int64(dup.ID),
		Name:       name,
		Position:   bridge.Vec3(pos),
		Rotation:   bridge.Vec4(rot),
		Scale:      bridge.Vec3(scale),
		Components: comps,
	}
	log.Printf("editorlink: getEntityInfo: build info for entity %d: %+v", dup.ID, info)
	return info
}

func buildSceneSnapshot(sc *scene.Scene) SceneSnapshot {
	snap := SceneSnapshot{
		Entities: make([]EntityView, 0),
		Selected: 0,
	}

	for _, ent := range sc.World().Entities {
		view := EntityView{
			ID: uint64(ent.ID),
		}

		if c := ent.GetComponent((*ecs.Name)(nil)); c != nil {
			view.Name = c.(*ecs.Name).Value
		}
		if c := ent.GetComponent((*ecs.Transform)(nil)); c != nil {
			tr := c.(*ecs.Transform)
			view.Position = Vec3(tr.Position)
			view.Components = append(view.Components, "Transform")
		}
		if c := ent.GetComponent((*ecs.Material)(nil)); c != nil {
			mat := c.(*ecs.Material)
			view.BaseColor = Vec4(mat.BaseColor)
			view.Components = append(view.Components, "Material")
		}
		if c := ent.GetComponent((*ecs.RigidBody)(nil)); c != nil {
			view.Components = append(view.Components, "RigidBody")
		}
		if ent.GetComponent((*ecs.ColliderSphere)(nil)) != nil {
			view.Components = append(view.Components, "ColliderSphere")
		}
		if ent.GetComponent((*ecs.ColliderAABB)(nil)) != nil {
			view.Components = append(view.Components, "ColliderAABB")
		}
		if ent.GetComponent((*ecs.ColliderPlane)(nil)) != nil {
			view.Components = append(view.Components, "ColliderPlane")
		}
		if ent.GetComponent((*ecs.LightComponent)(nil)) != nil {
			view.Components = append(view.Components, "Light")
		}

		snap.Entities = append(snap.Entities, view)
	}

	return snap
}

func applySetComponent(sc *scene.Scene, m MsgSetComponent) {
	ent := sc.World().FindByID(int64(m.EntityID))
	if ent == nil {
		log.Printf("game: SetComponent: entity %d not found", m.EntityID)
		return
	}

	// Find the component by name
	constructor, ok := ecs.ComponentRegistry[m.Name]
	if !ok {
		log.Printf("game: SetComponent: unknown component %s", m.Name)
		return
	}

	// Ensure the entity has the component
	comp := ent.GetComponent(constructor())
	if comp == nil {
		// Component doesn't exist yet → create it
		comp = constructor()
		ent.AddComponent(comp)
	}
	// Push updated snapshot back to editor

	// Apply fields
	if insp, ok := comp.(ecs.EditorInspectable); ok {
		for key, val := range m.Fields {
			insp.SetEditorField(key, val)
		}
	} else {
		log.Printf("game: SetComponent: component %s is not EditorInspectable", m.Name)
	}
	if EditorConn != nil {
		snap := buildSceneSnapshot(sc)
		resp := MsgSceneSnapshot{Snapshot: snap}
		if err := writeMsg(EditorConn, "SceneSnapshot", resp); err != nil {
			log.Printf("editorlink: failed to send SceneSnapshot: %v", err)
		}
	}
}
func applyRemoveComponent(sc *scene.Scene, m MsgRemoveComponent) {
	ent := sc.World().FindByID(int64(m.EntityID))
	if ent == nil {
		log.Printf("game: RemoveComponent: entity %d not found", m.EntityID)
		return
	}

	constructor, ok := ecs.ComponentRegistry[m.Name]
	if !ok {
		log.Printf("game: RemoveComponent: unknown component %s", m.Name)
		return
	}

	comp := ent.GetComponent(constructor())
	if comp == nil {
		log.Printf("game: RemoveComponent: component %s not present", m.Name)
		return
	}

	ent.RemoveComponent(comp)

	// Push updated snapshot back to editor

	log.Printf("game: sending SceneSnapshot after RemoveComponent %s on %d", m.Name, m.EntityID)
	if EditorConn != nil {
		snap := buildSceneSnapshot(sc)
		resp := MsgSceneSnapshot{Snapshot: snap}
		if err := writeMsg(EditorConn, "SceneSnapshot", resp); err != nil {
			log.Printf("editorlink: failed to send SceneSnapshot: %v", err)
		}
	}
}

// editorlink/server.go

func SendFullSnapshot(sc *scene.Scene) {
	if EditorConn == nil {
		return
	}
	snap := buildSceneSnapshot(sc)
	resp := MsgSceneSnapshot{Snapshot: snap}
	if err := writeMsg(EditorConn, "SceneSnapshot", resp); err != nil {
		log.Printf("editorlink: failed to send SceneSnapshot: %v", err)
	}
}

func buildAssetList() MsgAssetList {
	out := MsgAssetList{
		Textures:  []AssetView{},
		Meshes:    []AssetView{},
		Materials: []AssetView{},
	}

	for _, a := range assets.All() {
		view := AssetView{
			ID:   uint64(a.ID),
			Path: a.Path,
			Type: assetTypeToString(a.Type),
		}

		switch a.Type {
		case assets.AssetTexture:
			out.Textures = append(out.Textures, view)

		case assets.AssetMesh:
			// Data can be string (single mesh) or []string (multi)
			switch v := a.Data.(type) {
			case string:
				view.MeshIDs = []string{v}
			case []string:
				view.MeshIDs = v
			default:
				// no mesh IDs, leave empty
			}
			out.Meshes = append(out.Meshes, view)

		case assets.AssetMaterial:
			out.Materials = append(out.Materials, view)
		}
	}

	return out
}

func assetTypeToString(t assets.AssetType) string {
	switch t {
	case assets.AssetTexture:
		return "Texture"
	case assets.AssetMesh:
		return "Mesh"
	case assets.AssetMaterial:
		return "Material"
	}
	return "Unknown"
}
