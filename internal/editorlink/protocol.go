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

// Renderer -> Editor
type MsgSceneSnapshot struct {
	Snapshot SceneSnapshot `json:"snapshot"`
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

// Public client helpers:

func ReadMsg(conn net.Conn) (Msg, error) { return readMsg(conn) }

func WriteRequestSceneSnapshot(conn net.Conn) error {
	return writeMsg(conn, "RequestSceneSnapshot", MsgRequestSceneSnapshot{})
}
