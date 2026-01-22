package main

import (
	"go-engine/Go-Cordance/internal/editor"
	"go-engine/Go-Cordance/internal/scene"
)

func main() {

	sc, _ := scene.BootstrapScene()
	world := sc.World() // or sc.Entities if you expose them directly
	editor.Run(world)

}

/*
func connectAndRequestSnapshot() error {
	var err error
	editorlink.EditorConn, err = net.Dial("tcp", "localhost:7777")
	if err != nil {
		return err
	}

	if err := editorlink.WriteRequestSceneSnapshot(editorlink.EditorConn); err != nil {
		return err
	}

	msg, err := editorlink.ReadMsg(editorlink.EditorConn)
	if err != nil {
		return err
	}
	if msg.Type != "SceneSnapshot" {
		log.Printf("editor: unexpected msg type %q", msg.Type)
		return nil
	}

	var resp editorlink.MsgSceneSnapshot
	if err := json.Unmarshal(msg.Data, &resp); err != nil {
		return err
	}

	log.Printf("editor: got snapshot with %d entities (selected=%d)",
		len(resp.Snapshot.Entities), resp.Snapshot.Selected)
	// NEXT STEP: hand resp.Snapshot to internal/editor
	// so it can populate the hierarchy.
	ents := snapshotToEntityInfo(resp.Snapshot)
	editor.UpdateEntities(ents)

	return nil
}

func snapshotToEntityInfo(snap editorlink.SceneSnapshot) []bridge.EntityInfo {
	out := make([]bridge.EntityInfo, len(snap.Entities))
	for i, e := range snap.Entities {
		out[i] = bridge.EntityInfo{
			ID:         int64(e.ID),
			Name:       e.Name,
			Position:   bridge.Vec3(e.Position),
			Rotation:   bridge.Vec4(e.Rotation),
			Scale:      bridge.Vec3(e.Scale),
			Components: e.Components, // <-- THIS LINE FIXES EVERYTHING
		}

	}
	return out
}
*/
