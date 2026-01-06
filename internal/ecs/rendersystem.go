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
	SelectedEntity uint64

	DebugShowMode  int32 // 0..6 as in shader
	DebugFlipGreen bool
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

	//debug stuff
	gl.Uniform1i(rs.Renderer.LocShowMode, rs.DebugShowMode)
	if rs.DebugFlipGreen {
		gl.Uniform1i(rs.Renderer.LocFlipNormalG, 1)
	} else {
		gl.Uniform1i(rs.Renderer.LocFlipNormalG, 0)
	}
	//back to normal
	view := rs.CameraSystem.View
	proj := rs.CameraSystem.Projection

	// Only use orbital light if NO LightComponents exist
	hasLight := false
	for _, e := range entities {
		if _, ok := e.GetComponent((*LightComponent)(nil)).(*LightComponent); ok {
			hasLight = true
			break
		}
	}

	if rs.OrbitalEnabled && !hasLight {
		angle := float32(glfw.GetTime())
		rs.LightDir[0] = float32(math.Cos(float64(angle)))
		rs.LightDir[2] = float32(math.Sin(float64(angle)))
		rs.LightDir[1] = -0.7
	}

	rs.Renderer.LightColor = [3]float32{1, 1, 1}
	rs.Renderer.LightIntensity = 1.0

	// Find first LightComponent in the scene
	lights := make([]engine.LightData, 0)

	for _, e := range entities {
		lc, ok := e.GetComponent((*LightComponent)(nil)).(*LightComponent)
		if !ok {
			continue
		}

		tr, _ := e.GetComponent((*Transform)(nil)).(*Transform)

		// Compute direction from transform
		dir := [3]float32{0, 0, -1} // default forward

		if tr != nil {
			q := mgl32.Quat{
				W: tr.Rotation[0],
				V: mgl32.Vec3{tr.Rotation[1], tr.Rotation[2], tr.Rotation[3]},
			}
			fwd := q.Rotate(mgl32.Vec3{0, 0, -1})
			dir = [3]float32{fwd.X(), fwd.Y(), fwd.Z()}
		}

		lights = append(lights, engine.LightData{
			Type:      int32(lc.Type),
			Color:     lc.Color,
			Intensity: lc.Intensity,
			Direction: dir,
		})
	}
	// Upload light count
	gl.Uniform1i(rs.Renderer.LocLightCount, int32(len(lights)))

	// Upload each light
	for i, L := range lights {
		gl.Uniform3f(rs.Renderer.LocLightColor[i], L.Color[0], L.Color[1], L.Color[2])
		gl.Uniform1f(rs.Renderer.LocLightIntensity[i], L.Intensity)
		gl.Uniform3f(rs.Renderer.LocLightDir[i], L.Direction[0], L.Direction[1], L.Direction[2])
	}

	// Upload camera position once
	camPos := rs.CameraSystem.Position
	gl.Uniform3fv(rs.Renderer.LocViewPos, 1, &camPos[0])
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
		// Build model matrix from Position, Rotation (quat) and Scale
		// Translation
		model := mgl32.Translate3D(t.Position[0], t.Position[1], t.Position[2])

		// Rotation (assuming t.Rotation is [4]float32{w, x, y, z})
		// Only apply if it’s non-zero to avoid producing NaNs for an uninitialized quaternion.
		if t.Rotation != [4]float32{0, 0, 0, 0} {
			q := mgl32.Quat{
				W: t.Rotation[0],
				V: mgl32.Vec3{t.Rotation[1], t.Rotation[2], t.Rotation[3]},
			}
			model = model.Mul4(q.Mat4())
		}

		// Scale (default to 1 if zero to avoid collapsing the mesh)
		sx, sy, sz := t.Scale[0], t.Scale[1], t.Scale[2]
		if sx == 0 {
			sx = 1
		}
		if sy == 0 {
			sy = 1
		}
		if sz == 0 {
			sz = 1
		}
		model = model.Mul4(mgl32.Scale3D(sx, sy, sz))

		//lightDir := [3]float32{1.0, -0.7, -0.3}

		gl.UniformMatrix4fv(rs.Renderer.LocModel, 1, false, &model[0])
		gl.UniformMatrix4fv(rs.Renderer.LocView, 1, false, &view[0])
		gl.UniformMatrix4fv(rs.Renderer.LocProj, 1, false, &proj[0])

		// Upload material color
		// Defaults

		if mat != nil {
			// normal material
			gl.Uniform4fv(rs.Renderer.LocBaseCol, 1, &mat.BaseColor[0])

			// highlight override
			if uint64(e.ID) == rs.SelectedEntity {
				highlight := [4]float32{1, 1, 0, 1} // bright yellow
				gl.Uniform4fv(rs.Renderer.LocBaseCol, 1, &highlight[0])
			}

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
		// bind normal map if present ----debug code until Draw
		var normalMapComp *NormalMap
		if nm, ok := e.GetComponent((*NormalMap)(nil)).(*NormalMap); ok && nm != nil && nm.ID != 0 {
			normalMapComp = nm
		}

		if normalMapComp != nil {
			gl.ActiveTexture(gl.TEXTURE1)
			gl.BindTexture(gl.TEXTURE_2D, normalMapComp.ID)
			gl.Uniform1i(rs.Renderer.LocNormalMap, 1) // sampler unit 1
			gl.Uniform1i(rs.Renderer.LocUseNormalMap, 1)
		} else {
			gl.Uniform1i(rs.Renderer.LocUseNormalMap, 0)
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
