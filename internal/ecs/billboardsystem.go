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
		bb, ok := e.GetComponent((*Billboard)(nil)).(*Billboard)
		if !ok {
			continue
		}

		tr := e.GetTransform()
		if tr == nil {
			continue
		}

		switch bb.Mode {

		case BillboardSpherical:
			// Full 3D look-at
			tr.LookAt(
				[3]float32{camPos.X(), camPos.Y(), camPos.Z()},
				[3]float32{0, 1, 0},
			)

		case BillboardCylindrical:
			// Lock Y-axis: ignore camera Y
			target := mgl32.Vec3{
				camPos.X(),
				tr.Position[1], // keep billboard's own Y
				camPos.Z(),
			}
			tr.LookAt(
				[3]float32{target.X(), target.Y(), target.Z()},
				[3]float32{0, 1, 0},
			)

		case BillboardAxial:
			// Rotate only around a custom axis
			axis := mgl32.Vec3{bb.Axis[0], bb.Axis[1], bb.Axis[2]}.Normalize()
			forward := camPos.Sub(mgl32.Vec3{
				tr.Position[0],
				tr.Position[1],
				tr.Position[2],
			}).Normalize()

			// Project forward onto plane perpendicular to axis
			proj := forward.Sub(axis.Mul(forward.Dot(axis))).Normalize()

			// Build a rotation matrix from axis + projected forward
			tr.SetForward([3]float32{proj.X(), proj.Y(), proj.Z()}, bb.Axis)
		}
	}
}
