package ecs

import (
	"go-engine/Go-Cordance/internal/engine"

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

		model := mgl32.Translate3D(t.Position[0], t.Position[1], t.Position[2])
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
		gl.DrawElements(gl.TRIANGLES, count, gl.UNSIGNED_INT, gl.PtrOffset(0))
		gl.BindVertexArray(0)
	}
}
