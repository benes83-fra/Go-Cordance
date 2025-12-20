package ecs

import (
	"go-engine/Go-Cordance/internal/engine"
	"math"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

type RenderSystem struct {
	Renderer       *engine.Renderer
	MeshManager    *engine.MeshManager
	CameraSystem   *CameraSystem
	LightDir       [3]float32
	LightEntity    *Entity
	LightArrow     *Entity
	OrbitalEnabled bool
}

func NewRenderSystem(r *engine.Renderer, mm *engine.MeshManager, cs *CameraSystem) *RenderSystem {
	return &RenderSystem{
		Renderer:       r,
		MeshManager:    mm,
		CameraSystem:   cs,
		LightDir:       [3]float32{1.0, -0.7, -0.3}, // starting direction
		OrbitalEnabled: true,
	}
}

func (rs *RenderSystem) Update(_ float32, entities []*Entity) {
	gl.UseProgram(rs.Renderer.Program)
	view := rs.CameraSystem.View
	proj := rs.CameraSystem.Projection

	if rs.OrbitalEnabled {
		angle := float32(glfw.GetTime()) // seconds since start
		rs.LightDir[0] = float32(math.Cos(float64(angle)))
		rs.LightDir[2] = float32(math.Sin(float64(angle)))
		rs.LightDir[1] = -0.7 // keep some downward tilt
	}
	for _, e := range entities {
		var t *Transform
		var mesh *Mesh
		var mat *Material
		var tex *Texture
		if rs.LightEntity != nil {
			if t, ok := rs.LightEntity.GetComponent((*Transform)(nil)).(*Transform); ok {
				t.Position[0] = rs.LightDir[0] * 5
				t.Position[1] = rs.LightDir[1] * 5
				t.Position[2] = rs.LightDir[2] * 5
			}
		}
		if rs.LightArrow != nil {
			if t, ok := rs.LightArrow.GetComponent((*Transform)(nil)).(*Transform); ok {
				// scale the line to point in LightDir
				t.Scale = [3]float32{rs.LightDir[0] * 5, rs.LightDir[1] * 5, rs.LightDir[2] * 5}
			}
		}
		for _, c := range e.Components {
			switch v := c.(type) {
			case *Transform:
				t = v
			case *Mesh:
				mesh = v
			case *Material:
				mat = v
			case *Texture:
				tex = v
			}
		}
		if t == nil || mesh == nil || mat == nil {
			continue
		}

		// Build MVP
		model := mgl32.Translate3D(t.Position[0], t.Position[1], t.Position[2])
		//lightDir := [3]float32{1.0, -0.7, -0.3}

		gl.Uniform3fv(rs.Renderer.LocLightDir, 1, &rs.LightDir[0])

		camPos := rs.CameraSystem.Position // from your Camera component
		gl.Uniform3fv(rs.Renderer.LocViewPos, 1, &camPos[0])

		gl.UniformMatrix4fv(rs.Renderer.LocModel, 1, false, &model[0])
		gl.UniformMatrix4fv(rs.Renderer.LocView, 1, false, &view[0])
		gl.UniformMatrix4fv(rs.Renderer.LocProj, 1, false, &proj[0])

		// Upload material color

		if mat != nil {
			gl.Uniform4fv(rs.Renderer.LocBaseCol, 1, &mat.BaseColor[0])
			gl.Uniform1f(rs.Renderer.LocAmbient, mat.Ambient)
			gl.Uniform1f(rs.Renderer.LocDiffuse, mat.Diffuse)
			gl.Uniform1f(rs.Renderer.LocSpecular, mat.Specular)
			gl.Uniform1f(rs.Renderer.LocShininess, mat.Shininess)
		}
		if tex != nil {
			gl.ActiveTexture(gl.TEXTURE0)
			gl.BindTexture(gl.TEXTURE_2D, tex.ID)
			gl.Uniform1i(rs.Renderer.LocDiffuseTex, 0)
			gl.Uniform1i(rs.Renderer.LocUseTexture, 1)
		} else {
			gl.Uniform1i(rs.Renderer.LocUseTexture, 0)
		}

		// Draw
		vao := rs.MeshManager.GetVAO(mesh.ID)
		gl.BindVertexArray(vao)
		count := rs.MeshManager.GetCount(mesh.ID)
		//gl.DrawArrays(gl.TRIANGLES, 0, 3)
		if mesh.ID == "line" {
			gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
			gl.DrawElements(gl.LINES, count, gl.UNSIGNED_INT, gl.PtrOffset(0))
			gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
		} else {
			gl.DrawElements(gl.TRIANGLES, count, gl.UNSIGNED_INT, gl.PtrOffset(0))
		}

		gl.BindVertexArray(0)

	}
}
