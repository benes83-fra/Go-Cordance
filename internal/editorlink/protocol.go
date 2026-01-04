package editorlink

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
)

type Vec3 [3]float32
type Vec4 [4]float32

// This is the “editor view” of an entity: what Fyne’s inspector needs.
type EntityView struct {
	ID        uint64 `json:"id"`
	Name      string `json:"name"`
	Position  Vec3   `json:"position"`
	Rotation  Vec4   `json:"rotation"` // fill later if you want
	Scale     Vec3   `json:"scale"`    // fill later if you want
	BaseColor Vec4   `json:"baseColor"`
}

type SceneSnapshot struct {
	Entities []EntityView `json:"entities"`
	Selected uint64       `json:"selected"`
}

// Generic envelope.
type Msg struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// Requests/responses

// Editor -> Renderer
type MsgRequestSceneSnapshot struct{}

type MsgSelectEntity struct {
	ID uint64 `json:"id"`
}

type MsgSelectEntities struct {
	IDs []uint64 `json:"ids"`
}

// Renderer -> Editor
type MsgSceneSnapshot struct {
	Snapshot SceneSnapshot `json:"snapshot"`
}

type MsgSetTransform struct {
	ID       uint64     `json:"id"`
	Position [3]float32 `json:"position"`
	Rotation [4]float32 `json:"rotation"`
	Scale    [3]float32 `json:"scale"`
}
type MsgSetPivotMode struct {
	Mode string `json:"mode"` // "pivot" or "center"
}
type MsgSetComponent struct {
	EntityID uint64
	Name     string
	Fields   map[string]any
}
type MsgRemoveComponent struct {
	EntityID uint64
	Name     string
}

func readMsg(conn net.Conn) (Msg, error) {
	var m Msg
	r := bufio.NewReader(conn)
	line, err := r.ReadBytes('\n')
	if err != nil {
		return m, err
	}
	if err := json.Unmarshal(line, &m); err != nil {
		return m, fmt.Errorf("unmarshal msg: %w", err)
	}
	return m, nil
}

func writeMsg(conn net.Conn, msgType string, payload any) error {
	var data json.RawMessage
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("marshal payload: %w", err)
		}
		data = b
	}
	m := Msg{Type: msgType, Data: data}
	b, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("marshal msg: %w", err)
	}
	b = append(b, '\n')
	_, err = conn.Write(b)
	return err
}

func WriteSelectEntity(conn net.Conn, id int64) error {
	sel := MsgSelectEntity{ID: uint64(id)}
	fmt.Printf("Writinge selection %s", sel)
	return writeMsg(conn, "SelectEntity", sel)
}

func WriteSelectEntities(conn net.Conn, ids []int64) error {
	uids := make([]uint64, len(ids))
	for i, id := range ids {
		uids[i] = uint64(id)
	}
	sel := MsgSelectEntities{IDs: uids}
	fmt.Printf("Writinge selection %s", sel)
	return writeMsg(conn, "SelectEntities", sel)
}

func WriteSetPivotMode(conn net.Conn, mode string) error {
	msg := MsgSetPivotMode{Mode: mode}
	return writeMsg(conn, "SetPivotMode", msg)
}

// Public client helpers:

func ReadMsg(conn net.Conn) (Msg, error) { return readMsg(conn) }

func WriteRequestSceneSnapshot(conn net.Conn) error {
	return writeMsg(conn, "RequestSceneSnapshot", MsgRequestSceneSnapshot{})
}

func WriteSetTransform(conn net.Conn, msg MsgSetTransform) error {
	return writeMsg(conn, "SetTransform", msg)
}

func WriteTransformFromGame(conn net.Conn, id int64, position [3]float32, rotation [4]float32, scale [3]float32) error {
	msg := MsgSetTransform{
		ID:       uint64(id),
		Position: position,
		Rotation: rotation,
		Scale:    scale,
	}
	return writeMsg(conn, "SetTransformGizmo", msg)
}
func WriteSetTransformFinal(conn net.Conn, m MsgSetTransform) error {
	return writeMsg(conn, "SetTransformGizmoFinal", m)
}
func WriteSetComponent(conn net.Conn, msg MsgSetComponent) error {
	return writeMsg(conn, "SetComponent", msg)
}

func WriteRemoveComponent(conn net.Conn, msg MsgRemoveComponent) error {
	return writeMsg(conn, "RemoveComponent", msg)
}
