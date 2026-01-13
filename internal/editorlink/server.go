package editorlink

import (
	"encoding/json"
	"log"
	"net"

	"go-engine/Go-Cordance/internal/ecs"
	"go-engine/Go-Cordance/internal/ecs/gizmo"
	state "go-engine/Go-Cordance/internal/editor/state"

	"go-engine/Go-Cordance/internal/scene"
)

var EditorConn net.Conn
var lastLightVersion = map[uint64]uint64{} // entityID -> version

// StartServer exposes the given Scene to a single editor client.
func StartServer(addr string, sc *scene.Scene) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("editorlink: listen %s: %v", addr, err)
		return
	}
	log.Printf("editorlink: listening on %s", addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("editorlink: accept: %v", err)
			continue
		}
		EditorConn = conn
		log.Printf("editorlink: editor connected from %s", conn.RemoteAddr())
		go handleConn(conn, sc)
	}
}

func handleConn(conn net.Conn, sc *scene.Scene) {
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

		default:
			log.Printf("editorlink: unknown msg type %q", msg.Type)
		}
	}
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
