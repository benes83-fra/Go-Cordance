package ecs

import (
	"go-engine/Go-Cordance/internal/engine"
	"go-engine/Go-Cordance/internal/glutil"
	"log"
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

func (rs *RenderSystem) RenderShadowPass(entities []*Entity) {
	glutil.ClearGLErrors()

	if rs.Renderer.ShadowFBO == 0 || rs.Renderer.ShadowProgram == 0 {
		return
	}

	// --- Compute lightSpace ---
	sceneCenter := mgl32.Vec3{
		rs.CameraSystem.Position[0],
		rs.CameraSystem.Position[1],
		rs.CameraSystem.Position[2],
	}
	lightDir := mgl32.Vec3{
		rs.LightDir[0],
		rs.LightDir[1],
		rs.LightDir[2],
	}
	extent := float32(20.0)
	lightSpace := engine.ComputeDirectionalLightSpaceMatrix(lightDir, sceneCenter, extent)

	// --- Bind FBO ---
	gl.Viewport(0, 0, int32(rs.Renderer.ShadowWidth), int32(rs.Renderer.ShadowHeight))
	gl.BindFramebuffer(gl.FRAMEBUFFER, rs.Renderer.ShadowFBO)

	if status := gl.CheckFramebufferStatus(gl.FRAMEBUFFER); status != gl.FRAMEBUFFER_COMPLETE {
		log.Printf("Shadow FBO incomplete: 0x%X", status)
	}

	gl.Clear(gl.DEPTH_BUFFER_BIT)

	// --- Use shadow program + uniforms ---
	glutil.RunGLChecked("ShadowPass: UseProgram+Uniforms", func() {
		gl.UseProgram(rs.Renderer.ShadowProgram)

		// Only set sampler uniforms if the main program is bound
		var curProg int32
		gl.GetIntegerv(gl.CURRENT_PROGRAM, &curProg)
		if curProg == int32(rs.Renderer.Program) && rs.Renderer.LocShadowMap != -1 {
			gl.Uniform1i(rs.Renderer.LocShadowMap, 2)
		}

		locLS := gl.GetUniformLocation(rs.Renderer.ShadowProgram, gl.Str("lightSpaceMatrix\x00"))
		gl.UniformMatrix4fv(locLS, 1, false, &lightSpace[0])
	})

	// --- Draw all meshes into depth map ---
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

		// Build model matrix
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

		locModel := gl.GetUniformLocation(rs.Renderer.ShadowProgram, gl.Str("model\x00"))
		gl.UniformMatrix4fv(locModel, 1, false, &model[0])

		// Draw
		vao := rs.MeshManager.GetVAO(mesh.ID)
		indexCount := rs.MeshManager.GetCount(mesh.ID)
		indexType := rs.MeshManager.GetIndexType(mesh.ID)
		vertexCount := rs.MeshManager.GetVertexCount(mesh.ID)
		ebo := rs.MeshManager.GetEBO(mesh.ID)

		bytesPerIndex := int32(4)
		if indexType == gl.UNSIGNED_SHORT {
			bytesPerIndex = 2
		}

		var eboSize int32
		if ebo != 0 {
			gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
			gl.GetBufferParameteriv(gl.ELEMENT_ARRAY_BUFFER, gl.BUFFER_SIZE, &eboSize)
		}

		glutil.RunGLChecked("ShadowPass: draw "+mesh.ID, func() {
			gl.BindVertexArray(vao)

			if ebo != 0 {
				gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
			}

			if indexCount > 0 && ebo != 0 && eboSize >= int32(indexCount)*bytesPerIndex {
				if mesh.ID == "line" {
					gl.DrawElements(gl.LINES, indexCount, indexType, gl.PtrOffset(0))
				} else {
					gl.DrawElements(gl.TRIANGLES, indexCount, indexType, gl.PtrOffset(0))
				}
			} else {
				if mesh.ID == "line" {
					gl.DrawArrays(gl.LINES, 0, vertexCount)
				} else {
					gl.DrawArrays(gl.TRIANGLES, 0, vertexCount)
				}
			}

			gl.BindVertexArray(0)
		})
	}

	// Restore default framebuffer + viewport
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.Viewport(0, 0, int32(rs.Renderer.ScreenWidth), int32(rs.Renderer.ScreenHeight))

	// Bind shadow texture for main pass
	gl.ActiveTexture(gl.TEXTURE2)
	gl.BindTexture(gl.TEXTURE_2D, rs.Renderer.ShadowTex)
}

func (rs *RenderSystem) Update(dt float32, entities []*Entity) {
	rs.UpdateLightGizmos()
	rs.RenderShadowPass(entities)
	rs.RenderMainPass(entities)
}

func (rs *RenderSystem) RenderMainPass(entities []*Entity) {
	glutil.RunGLChecked("MainPass: UseProgram+Uniforms", func() {
		gl.UseProgram(rs.Renderer.Program)

		// Bind shadow map
		gl.ActiveTexture(gl.TEXTURE2)
		gl.BindTexture(gl.TEXTURE_2D, rs.Renderer.ShadowTex)

		if rs.Renderer.LocShadowMap != -1 {
			gl.Uniform1i(rs.Renderer.LocShadowMap, 2)
		}
		// --- Rebuild and upload lights (from old Update) ---
		rs.Renderer.LightColor = [3]float32{1, 1, 1}
		rs.Renderer.LightIntensity = 1.0

		// Collect lights
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
				pos = tr.Position

				q := mgl32.Quat{
					W: tr.Rotation[0],
					V: mgl32.Vec3{tr.Rotation[1], tr.Rotation[2], tr.Rotation[3]},
				}
				fwd := q.Rotate(mgl32.Vec3{0, 0, -1})
				dir = [3]float32{fwd.X(), fwd.Y(), fwd.Z()}
			}

			// legacy orbital gizmo light override
			if rs.LightEntity != nil && e == rs.LightEntity && lc.Type == LightDirectional {
				dir = rs.LightDir
			}
			if len(lights) == 0 {
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

		}

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

		// Upload camera position
		camPos := rs.CameraSystem.Position
		gl.Uniform3fv(rs.Renderer.LocViewPos, 1, &camPos[0])

		// Upload lightSpace
		if rs.Renderer.LocLightSpace != -1 {
			sceneCenter := mgl32.Vec3{
				rs.CameraSystem.Position[0],
				rs.CameraSystem.Position[1],
				rs.CameraSystem.Position[2],
			}
			lightDir := mgl32.Vec3{
				rs.LightDir[0],
				rs.LightDir[1],
				rs.LightDir[2],
			}
			extent := float32(20.0)
			lightSpace := engine.ComputeDirectionalLightSpaceMatrix(lightDir, sceneCenter, extent)
			gl.UniformMatrix4fv(rs.Renderer.LocLightSpace, 1, false, &lightSpace[0])
		}

		// Debug flags
		gl.Uniform1i(rs.Renderer.LocShowMode, rs.DebugShowMode)
		gl.Uniform1i(rs.Renderer.LocFlipNormalG, boolToInt(rs.DebugFlipGreen))
	})

	view := rs.CameraSystem.View
	proj := rs.CameraSystem.Projection

	// Draw all meshes
	for _, e := range entities {
		var t *Transform
		var mesh *Mesh
		var mat *Material
		var tex *Texture
		var normalMapComp *NormalMap

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
			case *NormalMap:
				normalMapComp = v
			}
		}
		if t == nil || mesh == nil || mat == nil {
			continue
		}

		// Build model matrix
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

		gl.UniformMatrix4fv(rs.Renderer.LocModel, 1, false, &model[0])
		gl.UniformMatrix4fv(rs.Renderer.LocView, 1, false, &view[0])
		gl.UniformMatrix4fv(rs.Renderer.LocProj, 1, false, &proj[0])

		// Material
		gl.Uniform4fv(rs.Renderer.LocBaseCol, 1, &mat.BaseColor[0])
		if uint64(e.ID) == rs.SelectedEntity {
			highlight := [4]float32{1, 1, 0, 1}
			gl.Uniform4fv(rs.Renderer.LocBaseCol, 1, &highlight[0])
		}
		gl.Uniform1f(rs.Renderer.LocAmbient, mat.Ambient)
		gl.Uniform1f(rs.Renderer.LocDiffuse, mat.Diffuse)
		gl.Uniform1f(rs.Renderer.LocSpecular, mat.Specular)
		gl.Uniform1f(rs.Renderer.LocShininess, mat.Shininess)

		// Diffuse texture
		if tex != nil {
			gl.ActiveTexture(gl.TEXTURE0)
			gl.BindTexture(gl.TEXTURE_2D, tex.ID)
			gl.Uniform1i(rs.Renderer.LocDiffuseTex, 0)
			gl.Uniform1i(rs.Renderer.LocUseTexture, 1)
		} else {
			gl.Uniform1i(rs.Renderer.LocUseTexture, 0)
		}

		// Normal map
		if normalMapComp != nil && normalMapComp.ID != 0 {
			gl.ActiveTexture(gl.TEXTURE1)
			gl.BindTexture(gl.TEXTURE_2D, normalMapComp.ID)
			gl.Uniform1i(rs.Renderer.LocNormalMap, 1)
			gl.Uniform1i(rs.Renderer.LocUseNormalMap, 1)
		} else {
			gl.Uniform1i(rs.Renderer.LocUseNormalMap, 0)
		}

		// Draw
		vao := rs.MeshManager.GetVAO(mesh.ID)
		indexCount := rs.MeshManager.GetCount(mesh.ID)
		indexType := rs.MeshManager.GetIndexType(mesh.ID)
		vertexCount := rs.MeshManager.GetVertexCount(mesh.ID)

		gl.BindVertexArray(vao)

		if indexCount > 0 {
			if mesh.ID == "line" {
				gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
				gl.DrawElements(gl.LINES, indexCount, indexType, gl.PtrOffset(0))
				gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
			} else {
				gl.DrawElements(gl.TRIANGLES, indexCount, indexType, gl.PtrOffset(0))
			}
		} else {
			if mesh.ID == "line" {
				gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
				gl.DrawArrays(gl.LINES, 0, vertexCount)
				gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
			} else {
				gl.DrawArrays(gl.TRIANGLES, 0, vertexCount)
			}
		}

		gl.BindVertexArray(0)
	}
}

func boolToInt(b bool) int32 {
	if b {
		return 1
	}
	return 0
}

func (rs *RenderSystem) UpdateLightGizmos() {
	// Orbital light motion
	if rs.OrbitalEnabled && rs.LightEntity != nil {
		angle := float32(glfw.GetTime())
		rs.LightDir[0] = float32(math.Cos(float64(angle)))
		rs.LightDir[2] = float32(math.Sin(float64(angle)))
		rs.LightDir[1] = -0.7
	}

	// Move light gizmo
	if rs.LightEntity != nil {
		if t, ok := rs.LightEntity.GetComponent((*Transform)(nil)).(*Transform); ok {
			t.Position = [3]float32{
				rs.LightDir[0] * 5,
				rs.LightDir[1] * 5,
				rs.LightDir[2] * 5,
			}
		}
	}

	// Scale arrow gizmo
	if rs.LightArrow != nil {
		if t, ok := rs.LightArrow.GetComponent((*Transform)(nil)).(*Transform); ok {
			t.Scale = [3]float32{
				rs.LightDir[0] * 5,
				rs.LightDir[1] * 5,
				rs.LightDir[2] * 5,
			}
		}
	}
}
