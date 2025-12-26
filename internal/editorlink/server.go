package editorlink

import (
	"encoding/json"
	"log"
	"net"

	"go-engine/Go-Cordance/internal/ecs"
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
		log.Printf("editorlink: editor connected from %s", conn.RemoteAddr())
		go handleConn(conn, sc)
	}
}

func handleConn(conn net.Conn, sc *scene.Scene) {
	defer conn.Close()

	for {
		msg, err := readMsg(conn)
		if err != nil {
			log.Printf("editorlink: read: %v", err)
			return
		}

		switch msg.Type {
		case "RequestSceneSnapshot":
			snap := buildSceneSnapshot(sc)
			resp := MsgSceneSnapshot{Snapshot: snap}
			if err := writeMsg(conn, "SceneSnapshot", resp); err != nil {
				log.Printf("editorlink: write SceneSnapshot: %v", err)
				return
			}

		case "SelectEntity":
			var sel MsgSelectEntity
			if err := json.Unmarshal(msg.Data, &sel); err != nil {
				log.Printf("editorlink: bad SelectEntity: %v", err)
				continue
			}
			for _, sys := range sc.Systems().Systems() {
				if rs, ok := sys.(*ecs.RenderSystem); ok {
					rs.SelectedEntity = sel.ID
				}
			}
			log.Printf("editorlink: SelectEntity %d ", sel.ID)

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
