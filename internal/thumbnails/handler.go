package thumbnails

import (
	"encoding/json"

	"log"
	"net"
)

func HandleRequestThumbnail(msgData []byte, conn net.Conn, mgr *Manager) {
	var req struct {
		AssetID uint64 `json:"asset_id"`
		Size    int    `json:"size,omitempty"`
	}
	if err := json.Unmarshal(msgData, &req); err != nil {
		log.Printf("bad RequestThumbnail: %v", err)
		return
	}
	size := req.Size
	if size <= 0 {
		size = 128
	}
	mgr.Enqueue(req.AssetID, size, conn)
}

func HandleRequestThumbnailMesh(data []byte, conn net.Conn, mgr *Manager) {
	var req struct {
		AssetID uint64 `json:"asset_id"`
		MeshID  string `json:"mesh_id"`
		Size    int    `json:"size"`
	}
	if err := json.Unmarshal(data, &req); err != nil {
		log.Printf("bad RequestThumbnailMesh: %v", err)
		return
	}

	if req.Size <= 0 {
		req.Size = 128
	}

	// Just render and return the bytes
	bytes, hash, err := GenerateMeshSubThumbnailBytes(req.AssetID, req.MeshID, req.Size)
	if err != nil {
		log.Printf("mesh thumbnail failed: %v", err)
		return
	}

	// DO NOT send it here.
	// DO NOT import editorlink.
	// Return the data to the caller (editorlink).
	mgr.send(conn, req.AssetID, "png", bytes, hash)
}
