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

		default:
			log.Printf("editorlink: unknown msg type %q", msg.Type)
		}
	}
}

func buildSceneSnapshot(sc *scene.Scene) SceneSnapshot {
	snap := SceneSnapshot{
		Entities: make([]EntityView, 0),
		Selected: 0, // TODO: get from your existing selection tracking
	}

	// This depends on your actual Scene/ECS layout.
	// I’ll sketch it assuming you have something like sc.Entities or sc.World.
	for _, ent := range sc.World().Entities { // adjust this line to your real API
		view := EntityView{
			ID: uint64(ent.ID),
		}

		if c := ent.GetComponent((*ecs.Name)(nil)); c != nil {
			view.Name = c.(*ecs.Name).Value
		}
		if c := ent.GetComponent((*ecs.Transform)(nil)); c != nil {
			tr := c.(*ecs.Transform)
			view.Position = Vec3(tr.Position)
			// TODO Rotation/Scale
		}
		if c := ent.GetComponent((*ecs.Material)(nil)); c != nil {
			mat := c.(*ecs.Material)
			view.BaseColor = Vec4(mat.BaseColor)
		}

		snap.Entities = append(snap.Entities, view)
	}

	return snap
}
