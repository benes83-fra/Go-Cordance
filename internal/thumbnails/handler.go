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
