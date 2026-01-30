package thumbnails

import (
	"log"
	"net"
	"sync"
)

type SendFunc func(conn net.Conn, assetID uint64, format string, data []byte, hash string) error

type Manager struct {
	queue    chan job
	inflight map[uint64]bool
	mu       sync.Mutex
	workers  int
	send     SendFunc
}

type job struct {
	AssetID uint64
	Size    int
	Conn    net.Conn
}

// NewManager creates a thumbnail manager. sendFn is called to deliver generated thumbnails.
func NewManager(workers int, queueSize int, sendFn SendFunc) *Manager {
	m := &Manager{
		queue:    make(chan job, queueSize),
		inflight: make(map[uint64]bool),
		workers:  workers,
		send:     sendFn,
	}
	for i := 0; i < workers; i++ {
		go m.worker()
	}
	return m
}

func (m *Manager) Enqueue(assetID uint64, size int, conn net.Conn) {
	m.mu.Lock()
	if m.inflight[assetID] {
		m.mu.Unlock()
		return
	}
	m.inflight[assetID] = true
	m.mu.Unlock()

	select {
	case m.queue <- job{AssetID: assetID, Size: size, Conn: conn}:
	default:
		// queue full: drop and clear inflight
		log.Printf("thumbnail queue full, dropping request %d", assetID)
		m.mu.Lock()
		delete(m.inflight, assetID)
		m.mu.Unlock()
	}
}

func (m *Manager) worker() {
	for j := range m.queue {
		data, hash, err := GenerateThumbnailBytes(j.AssetID, j.Size)
		if err != nil {
			log.Printf("thumbnail generation failed for %d: %v", j.AssetID, err)
		} else {
			// Use the send callback provided by the caller (editorlink)
			if m.send != nil {
				if err := m.send(j.Conn, j.AssetID, "png", data, hash); err != nil {
					log.Printf("failed to send thumbnail %d: %v", j.AssetID, err)
				}
			}
		}
		m.mu.Lock()
		delete(m.inflight, j.AssetID)
		m.mu.Unlock()
	}
}
