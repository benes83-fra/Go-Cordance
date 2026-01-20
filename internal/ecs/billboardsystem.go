package ecs

import "github.com/go-gl/mathgl/mgl32"

type BillboardSystem struct {
	camSys *CameraSystem
}

func NewBillboardSystem(camSys *CameraSystem) *BillboardSystem {
	return &BillboardSystem{camSys: camSys}
}

func (bs *BillboardSystem) Update(dt float32, entities []*Entity) {
	if bs.camSys == nil {
		return
	}

	camPos := mgl32.Vec3{
		bs.camSys.Position[0],
		bs.camSys.Position[1],
		bs.camSys.Position[2],
	}

	for _, e := range entities {
		if e.GetComponent((*Billboard)(nil)) == nil {
			continue
		}

		tr := e.GetTransform()
		if tr == nil {
			continue
		}

		tr.LookAt(
			[3]float32{camPos.X(), camPos.Y(), camPos.Z()},
			[3]float32{0, 1, 0},
		)
	}
}
