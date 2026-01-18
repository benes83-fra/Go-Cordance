package scene

import "go-engine/Go-Cordance/internal/ecs"

type BillboardSystem struct {
	CameraEntity *ecs.Entity
}

func (bs *BillboardSystem) Update(sc *Scene, dt float32) {
	if bs.CameraEntity == nil {
		return
	}

	camEntity := bs.CameraEntity
	camTransform := camEntity.GetTransform()
	if camTransform == nil {
		return
	}
	camPos := camTransform.Position
	for _, e := range sc.World().Entities {
		if e.GetComponent((*ecs.Billboard)(nil)) == nil {
			continue
		}

		tr := e.GetTransform()
		if tr == nil {
			continue
		}

		// Make billboard face the camera
		tr.LookAt(camPos, [3]float32{0, 1, 0})
	}
}
