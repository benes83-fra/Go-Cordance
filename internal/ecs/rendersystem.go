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
	// --- SHADOW PASS ----------------------------------------------------
	// Only run if shadow resources are initialized
	if rs.Renderer != nil && rs.Renderer.ShadowFBO != 0 && rs.Renderer.ShadowProgram != 0 {
		// Choose scene center for shadow (use camera position as a simple center)
		sceneCenter := mgl32.Vec3{rs.CameraSystem.Position[0], rs.CameraSystem.Position[1], rs.CameraSystem.Position[2]}

		// Compute light direction vector (use first directional light or orbital)
		lightDir := mgl32.Vec3{rs.LightDir[0], rs.LightDir[1], rs.LightDir[2]}

		// extent controls orthographic box size (tweakable)
		extent := float32(20.0)

		lightSpace := engine.ComputeDirectionalLightSpaceMatrix(lightDir, sceneCenter, extent)

		// Bind shadow FBO and render depth
		gl.Viewport(0, 0, int32(rs.Renderer.ShadowWidth), int32(rs.Renderer.ShadowHeight))
		gl.BindFramebuffer(gl.FRAMEBUFFER, rs.Renderer.ShadowFBO)
		gl.Clear(gl.DEPTH_BUFFER_BIT)

		gl.UseProgram(rs.Renderer.ShadowProgram)
		// set lightSpace uniform on shadow program
		locLS := gl.GetUniformLocation(rs.Renderer.ShadowProgram, gl.Str("lightSpaceMatrix\x00"))
		gl.UniformMatrix4fv(locLS, 1, false, &lightSpace[0])

		// Render all meshes to depth map (same iteration as main pass)
		for _, e := range entities {
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

			// Build model matrix (same as main pass)
			model := mgl32.Translate3D(t.Position[0], t.Position[1], t.Position[2])
			if t.Rotation != [4]float32{0, 0, 0, 0} {
				q := mgl32.Quat{
					W: t.Rotation[0],
					V: mgl32.Vec3{t.Rotation[1], t.Rotation[2], t.Rotation[3]},
				}
				model = model.Mul4(q.Mat4())
			}
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

			// upload model matrix to shadow shader
			locModel := gl.GetUniformLocation(rs.Renderer.ShadowProgram, gl.Str("model\x00"))
			gl.UniformMatrix4fv(locModel, 1, false, &model[0])

			// draw mesh
			vao := rs.MeshManager.GetVAO(mesh.ID)
			gl.BindVertexArray(vao)
			count := rs.MeshManager.GetCount(mesh.ID)
			if mesh.ID == "line" {
				gl.DrawElements(gl.LINES, count, gl.UNSIGNED_INT, gl.PtrOffset(0))
			} else {
				gl.DrawElements(gl.TRIANGLES, count, gl.UNSIGNED_INT, gl.PtrOffset(0))
			}
			gl.BindVertexArray(0)
		}

		// Unbind FBO and restore viewport
		gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
		gl.Viewport(0, 0, int32(rs.Renderer.ScreenWidth), int32(rs.Renderer.ScreenHeight))

		// Bind shadow map to a texture unit for the main pass (we'll use unit 2)
		gl.ActiveTexture(gl.TEXTURE2)
		gl.BindTexture(gl.TEXTURE_2D, rs.Renderer.ShadowTex)
		// set uniform in main shader later (after gl.UseProgram for main shader)
		// store lightSpace in a local variable for upload to main shader below
		// (we'll upload it after switching to main program)
		// keep lightSpace variable in scope by reusing it below
		_ = lightSpace
	}

	// --- MAIN PASS (existing code) -------------------------------------
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

	if rs.OrbitalEnabled && rs.LightEntity != nil {
		angle := float32(glfw.GetTime())
		rs.LightDir[0] = float32(math.Cos(float64(angle)))
		rs.LightDir[2] = float32(math.Sin(float64(angle)))
		rs.LightDir[1] = -0.7
	}

	// Drive the visual light gizmo (LightEntity / LightArrow) from LightDir
	if rs.LightEntity != nil {
		if t, ok := rs.LightEntity.GetComponent((*Transform)(nil)).(*Transform); ok {
			t.Position[0] = rs.LightDir[0] * 5
			t.Position[1] = rs.LightDir[1] * 5
			t.Position[2] = rs.LightDir[2] * 5
		}
	}
	if rs.LightArrow != nil {
		if t, ok := rs.LightArrow.GetComponent((*Transform)(nil)).(*Transform); ok {
			t.Scale = [3]float32{rs.LightDir[0] * 5, rs.LightDir[1] * 5, rs.LightDir[2] * 5}
		}
	}

	rs.Renderer.LightColor = [3]float32{1, 1, 1}
	rs.Renderer.LightIntensity = 1.0

	// Find first LightComponent in the scene
	// Collect all lights in the scene
	lights := make([]engine.LightData, 0, 8)

	for _, e := range entities {
		lc, ok := e.GetComponent((*LightComponent)(nil)).(*LightComponent)
		if !ok {
			continue
		}

		tr, _ := e.GetComponent((*Transform)(nil)).(*Transform)

		// Defaults
		dir := [3]float32{0, 0, -1}
		pos := [3]float32{0, 0, 0}

		if tr != nil {
			// Position
			pos = tr.Position

			// Direction from rotation
			q := mgl32.Quat{
				W: tr.Rotation[0],
				V: mgl32.Vec3{tr.Rotation[1], tr.Rotation[2], tr.Rotation[3]},
			}
			fwd := q.Rotate(mgl32.Vec3{0, 0, -1})
			dir = [3]float32{fwd.X(), fwd.Y(), fwd.Z()}
		}

		// Special case: legacy orbital gizmo light
		if rs.LightEntity != nil && e == rs.LightEntity && lc.Type == LightDirectional {
			dir = rs.LightDir
		}

		lights = append(lights, engine.LightData{
			Type:      int32(lc.Type),
			Color:     lc.Color,
			Intensity: lc.Intensity,
			Direction: dir,
			Position:  pos,
			Range:     lc.Range,
			Angle:     lc.Angle,
		})
	}

	// Upload light count
	// Upload light count
	gl.Uniform1i(rs.Renderer.LocLightCount, int32(len(lights)))

	// Upload each light
	for i, L := range lights {
		gl.Uniform3f(rs.Renderer.LocLightColor[i], L.Color[0], L.Color[1], L.Color[2])
		gl.Uniform1f(rs.Renderer.LocLightIntensity[i], L.Intensity)
		gl.Uniform3f(rs.Renderer.LocLightDir[i], L.Direction[0], L.Direction[1], L.Direction[2])

		gl.Uniform3f(rs.Renderer.LocLightPos[i], L.Position[0], L.Position[1], L.Position[2])
		gl.Uniform1f(rs.Renderer.LocLightRange[i], L.Range)
		gl.Uniform1f(rs.Renderer.LocLightAngle[i], L.Angle)
		gl.Uniform1i(rs.Renderer.LocLightType[i], L.Type)
	}

	// Upload camera position once
	camPos := rs.CameraSystem.Position
	gl.Uniform3fv(rs.Renderer.LocViewPos, 1, &camPos[0])
	for _, e := range entities {
		var t *Transform
		var mesh *Mesh
		var mat *Material
		var tex *Texture

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
