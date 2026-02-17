package editorlink

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
)

type Vec3 [3]float32
type Vec4 [4]float32

// This is the “editor view” of an entity: what Fyne’s inspector needs.
type EntityView struct {
	ID       uint64 `json:"id"`
	Name     string `json:"name"`
	Position Vec3   `json:"position"`
	Rotation Vec4   `json:"rotation"`
	Scale    Vec3   `json:"scale"`

	// Material fields
	BaseColor Vec4 `json:"baseColor"`

	Components []string `json:"components"`
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

type MsgSetEditorFlag struct {
	ShowLightGizmos bool
}
type MsgTextureList struct {
	Names []string `json:"names"`
	IDs   []uint32 `json:"ids"`
}
type MsgFocusEntity struct {
	ID uint64 `json:"id"`
}

type MsgAssetList struct {
	Textures  []AssetView `json:"textures"`
	Meshes    []AssetView `json:"meshes"`
	Materials []AssetView `json:"materials"`
	Shaders   []AssetView `json:"shaders"`
}

type AssetView struct {
	ID           uint64         `json:"id"`
	Path         string         `json:"path"`
	Type         string         `json:"type"`
	MeshIDs      []string       `json:"mesh_ids,omitempty"`
	MaterialData map[string]any `json:"material_data,omitempty"`
	ShaderData   map[string]any `json:"shader_data,omitempty"`
}
type MsgRequestAssetList struct{}

// MsgRequestThumbnail asks the game to generate/send a thumbnail for AssetID.
type MsgRequestThumbnail struct {
	AssetID uint64 `json:"asset_id"`
	Size    int    `json:"size,omitempty"`
	MeshID  string `json:"mesh_id,omitempty"`
}

type MsgRequestMeshThumbnail struct {
	AssetID uint64 `json:"asset_id"`
	MeshID  string `json:"mesh_id"`
	Size    int    `json:"size"`
}
type MsgAssetMeshThumbnail struct {
	AssetID uint64 `json:"asset_id"`
	MeshID  string `json:"mesh_id"`
	Format  string `json:"format"` // "png"
	DataB64 string `json:"data_b64"`
	Hash    string `json:"hash"`
}

// MsgAssetThumbnail carries a generated thumbnail from game -> editor.
// Data is base64-encoded PNG/JPEG bytes to keep the message JSON-friendly.
type MsgAssetThumbnail struct {
	AssetID uint64 `json:"asset_id"`
	MeshID  string `json:"mesh_id,omitempty"`
	Format  string `json:"format"`         // e.g. "png" or "jpeg"
	DataB64 string `json:"data_b64"`       // base64-encoded image bytes
	Hash    string `json:"hash,omitempty"` // optional checksum for cache validation
}
type MsgMeshList struct {
	Meshes []struct {
		ID   string `json:"id"`
		Path string `json:"path"`
	} `json:"meshes"`
}
type MsgDeleteEntity struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type MsgDuplicateEntity struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

type MsgSetGlobalShader struct {
	Name string `json:"name"`
}

func readMsg(conn net.Conn) (Msg, error) {
	var m Msg
	header := make([]byte, 4)
	if _, err := io.ReadFull(conn, header); err != nil {
		return m, err
	}

	length := uint32(header[0])<<24 |
		uint32(header[1])<<16 |
		uint32(header[2])<<8 |
		uint32(header[3])

	buf := make([]byte, length)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return m, err
	}

	if err := json.Unmarshal(buf, &m); err != nil {
		return m, err
	}

	return m, nil
}

func writeMsg(conn net.Conn, msgType string, payload any) error {
	var data json.RawMessage
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		data = b
	}

	m := Msg{Type: msgType, Data: data}
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}

	// prefix with length
	length := uint32(len(b))
	header := []byte{
		byte(length >> 24),
		byte(length >> 16),
		byte(length >> 8),
		byte(length),
	}

	_, err = conn.Write(append(header, b...))
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

func WriteDuplicateEntity(conn net.Conn, id int64) error {
	msg := MsgDuplicateEntity{ID: uint64(id)}
	return writeMsg(conn, "DuplicateEntity", msg)
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

func WriteSetEditorFlag(conn net.Conn, msg MsgSetEditorFlag) error {
	return writeMsg(conn, "SetEditorFlag", msg)
}
func WriteTextureList(conn net.Conn, names []string, ids []uint32) error {

	msg := MsgTextureList{
		Names: names,
		IDs:   ids,
	}
	return writeMsg(conn, "TextureList", msg)
}
func WriteFocusEntity(conn net.Conn, id int64) error {
	msg := MsgFocusEntity{ID: uint64(id)}
	return writeMsg(conn, "FocusEntity", msg)
}

func WriteDeleteEntity(conn net.Conn, id int64, name string) error {
	msg := MsgDeleteEntity{
		ID:   id,
		Name: name,
	}
	return writeMsg(conn, "DeleteEntity", msg)
}
func WriteRequestAssetList(conn net.Conn) error {
	return writeMsg(conn, "RequestAssetList", MsgRequestAssetList{})
}

// WriteRequestThumbnail sends a thumbnail request to the game.
// conn must be a live net.Conn (editorlink.EditorConn).
func WriteRequestThumbnail(conn net.Conn, assetID uint64, size int) error {
	if conn == nil {
		return fmt.Errorf("editorlink: nil connection")
	}
	msg := MsgRequestThumbnail{
		AssetID: assetID,
		Size:    size,
	}
	return writeMsg(conn, "RequestThumbnail", msg)
}

// WriteAssetThumbnail sends a generated thumbnail to the editor.
func WriteAssetThumbnail(conn net.Conn, assetID uint64, format string, data []byte, hash string) error {
	if conn == nil {
		return fmt.Errorf("editorlink: nil connection")
	}
	b64 := base64.StdEncoding.EncodeToString(data)
	msg := MsgAssetThumbnail{
		AssetID: assetID,
		Format:  format,
		DataB64: b64,
		Hash:    hash,
	}
	return writeMsg(conn, "AssetThumbnail", msg)
}

func WriteMeshList(conn net.Conn, meshes []struct {
	ID   string `json:"id"`
	Path string `json:"path"`
}) error {
	if conn == nil {
		return fmt.Errorf("editorlink: nil connection")
	}
	msg := MsgMeshList{
		Meshes: meshes,
	}
	return writeMsg(conn, "MeshList", msg)
}
func WriteRequestMeshThumbnail(conn net.Conn, assetID uint64, meshID string, size int) error {
	m := MsgRequestMeshThumbnail{
		AssetID: assetID,
		MeshID:  meshID,
		Size:    size,
	}
	return writeMsg(conn,
		"RequestMeshThumbnail",
		m)

}

func SendAssetMeshThumbnail(conn net.Conn, assetID uint64, meshID string, format string, data []byte, hash string) error {
	m := MsgAssetMeshThumbnail{
		AssetID: assetID,
		MeshID:  meshID,
		Format:  format,
		DataB64: base64.StdEncoding.EncodeToString(data),
		Hash:    hash,
	}
	return writeMsg(conn,
		"AssetMeshThumbnail",
		m)
}

func SendSetGlobalShader(conn net.Conn, name string) error {
	msg := MsgSetGlobalShader{Name: name}
	return writeMsg(conn, "SetGlobalShader", msg)
}
