package thumbnails

import (
	"log"
	"net"
)

type SendFunc func(conn net.Conn, assetID uint64, format string, data []byte, hash string) error

type Manager struct {
	send SendFunc
}

// NewManager creates a thumbnail manager. sendFn is called to deliver generated thumbnails.
// NOTE: this version is synchronous – no background goroutines, no GL from worker threads.
func NewManager(workers int, queueSize int, sendFn SendFunc) *Manager {
	return &Manager{
		send: sendFn,
	}
}

func (m *Manager) Enqueue(assetID uint64, size int, conn net.Conn) {
	data, hash, err := GenerateThumbnailBytes(assetID, size)
	if err != nil {
		log.Printf("thumbnail generation failed for %d: %v", assetID, err)
		return
	}
	if m.send != nil {
		if err := m.send(conn, assetID, "png", data, hash); err != nil {
			log.Printf("failed to send thumbnail %d: %v", assetID, err)
		}
	}
}
