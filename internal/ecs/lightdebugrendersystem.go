package ecs

import (
	"go-engine/Go-Cordance/internal/engine"
	"math"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type LightDebugRenderSystem struct {
	Renderer     *engine.DebugRenderer
	MeshManager  *engine.MeshManager
	CameraSystem *CameraSystem
	Enabled      bool
	tracked      []*Entity
	Colors       map[*Entity][4]float32
}

func NewLightDebugRenderSystem(r *engine.DebugRenderer, mm *engine.MeshManager, cs *CameraSystem) *LightDebugRenderSystem {
	return &LightDebugRenderSystem{
		Renderer:     r,
		MeshManager:  mm,
		CameraSystem: cs,
		Enabled:      true,
		tracked:      []*Entity{},
		Colors:       make(map[*Entity][4]float32),
	}
}

// Register an entity for gizmo rendering
func (lds *LightDebugRenderSystem) Track(e *Entity) {
	lds.tracked = append(lds.tracked, e)
}

// Optional per-entity color
func (lds *LightDebugRenderSystem) SetColor(e *Entity, col [4]float32) {
	lds.Colors[e] = col
}

func (lds *LightDebugRenderSystem) Update(_ float32, _ []*Entity) {
	if !lds.Enabled {
		return
	}

	gl.UseProgram(lds.Renderer.Program)
	view := lds.CameraSystem.View
	proj := lds.CameraSystem.Projection

	for _, e := range lds.tracked {
		var t *Transform
		var mesh *Mesh
		for _, c := range e.Components {
			switch v := c.(type) {
			case *Transform:
				t = v
			case *Mesh:
				mesh = v
			}
		}
		if t == nil || mesh == nil {
			continue
		}

		// Base model from position
		model := mgl32.Translate3D(t.Position[0], t.Position[1], t.Position[2])

		// If it's the light arrow, align unit Z to the desired direction
		if mesh.ID == "line" {
			// Expect t.Scale to carry the endpoint vector (dir * length)
			dir := mgl32.Vec3{t.Scale[0], t.Scale[1], t.Scale[2]}
			length := dir.Len()
			if length > 0 {
				n := dir.Normalize()
				z := mgl32.Vec3{0, 0, 1}
				// Compute rotation from Z to n
				c := z.Cross(n)
				d := z.Dot(n)
				if d < 1.0-1e-6 {
					// Angle-axis rotation: axis=c normalized, angle=acos(d)
					axis := c.Normalize()
					angle := float32(math.Acos(float64(d)))
					rot := mgl32.HomogRotate3D(angle, axis)
					model = model.Mul4(rot)
				}
				// Scale only along Z to reach the length (line goes from 0→1 in Z)
				scale := mgl32.Scale3D(1, 1, length)
				model = model.Mul4(scale)
			}

			// Make sure we render as lines and set a visible width
			gl.LineWidth(2)
		} else {
			// Optional: apply regular scale if you use t.Scale elsewhere
			if t.Scale != ([3]float32{0, 0, 0}) {
				s := mgl32.Scale3D(t.Scale[0], t.Scale[1], t.Scale[2])
				model = model.Mul4(s)
			}
		}

		gl.UniformMatrix4fv(lds.Renderer.LocModel, 1, false, &model[0])
		gl.UniformMatrix4fv(lds.Renderer.LocView, 1, false, &view[0])
		gl.UniformMatrix4fv(lds.Renderer.LocProj, 1, false, &proj[0])

		col := [4]float32{1, 1, 1, 1}
		if c, ok := lds.Colors[e]; ok {
			col = c
		}
		gl.Uniform4fv(lds.Renderer.LocColor, 1, &col[0])

		vao := lds.MeshManager.GetVAO(mesh.ID)
		gl.BindVertexArray(vao)
		count := lds.MeshManager.GetCount(mesh.ID)

		if mesh.ID == "line" {
			gl.DrawElements(gl.LINES, count, gl.UNSIGNED_INT, gl.PtrOffset(0))
		} else {
			gl.DrawElements(gl.TRIANGLES, count, gl.UNSIGNED_INT, gl.PtrOffset(0))
		}
		gl.BindVertexArray(0)
	}
}
