package ecs

import (
	"go-engine/Go-Cordance/internal/assets"
	"go-engine/Go-Cordance/internal/engine"
	"go-engine/Go-Cordance/internal/glutil"
	"log"
	"math"
	"unsafe"

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

	DebugShowMode    int32
	DebugFlipGreen   bool
	ActiveShader     *engine.ShaderProgram
	shadowLightIndex int
	DefaultShader    *engine.ShaderProgram

	// --- NEW: GPU material UBO ---
	materialUBO     uint32
	materialBinding uint32
}

// std140-compatible layout for the MaterialBlock
type gpuMaterial struct {
	BaseColor    [4]float32
	Ambient      float32
	Diffuse      float32
	Specular     float32
	Shininess    float32
	Metallic     float32
	Roughness    float32
	MaterialType int32   // must be int32 for std140
	_Pad0        float32 // padding to 16‑byte alignment
}

type MeshDrawItem struct {
	MeshID    string
	Material  *Material
	NormalMap *NormalMap
}

func NewRenderSystem(r *engine.Renderer, mm *engine.MeshManager, cs *CameraSystem) *RenderSystem {
	rs := &RenderSystem{
		Renderer:       r,
		MeshManager:    mm,
		CameraSystem:   cs,
		LightDir:       [3]float32{1.0, -0.7, -0.3},
		OrbitalEnabled: true,
		DefaultShader:  engine.MustGetShaderProgram("default_shader"),
	}

	// --- NEW: create material UBO ---
	rs.materialBinding = 1 // must match GLSL binding = 1
	var ubo uint32
	gl.GenBuffers(1, &ubo)
	gl.BindBuffer(gl.UNIFORM_BUFFER, ubo)
	gl.BufferData(gl.UNIFORM_BUFFER, int(unsafe.Sizeof(gpuMaterial{})), nil, gl.DYNAMIC_DRAW)
	gl.BindBuffer(gl.UNIFORM_BUFFER, 0)
	rs.materialUBO = ubo

	return rs
}

func (rs *RenderSystem) computeShadowLightSpace(entities []*Entity) (mgl32.Mat4, int, bool) {
	var shadowLight *LightComponent
	var shadowTransform *Transform
	shadowIndex := -1

	// FIRST PASS: explicit shadow-casting light
	for i, e := range entities {
		lc, ok := e.GetComponent((*LightComponent)(nil)).(*LightComponent)
		if !ok || !lc.CastsShadows {
			continue
		}
		tr, _ := e.GetComponent((*Transform)(nil)).(*Transform)
		shadowLight = lc
		shadowTransform = tr
		shadowIndex = i
		break
	}

	if shadowLight == nil || shadowTransform == nil {
		return mgl32.Ident4(), -1, false
	}

	// Compute light-space matrix (unchanged)
	var lightSpace mgl32.Mat4

	switch shadowLight.Type {
	case LightDirectional:
		// derive direction from the shadow light's transform, not rs.LightDir
		q := mgl32.Quat{
			W: shadowTransform.Rotation[3],
			V: mgl32.Vec3{
				shadowTransform.Rotation[0],
				shadowTransform.Rotation[1],
				shadowTransform.Rotation[2],
			},
		}
		fwd := q.Rotate(mgl32.Vec3{0, 0, -1})
		lightDir := fwd.Normalize()

		camPos := mgl32.Vec3{
			rs.CameraSystem.Position[0],
			rs.CameraSystem.Position[1],
			rs.CameraSystem.Position[2],
		}
		// you can keep your offset if you like it:
		sceneCenter := camPos.Sub(lightDir.Mul(20))
		extent := float32(50.0)

		lightSpace = engine.ComputeDirectionalLightSpaceMatrix(lightDir, sceneCenter, extent)

	case LightSpot:
		pos := mgl32.Vec3{
			shadowTransform.Position[0],
			shadowTransform.Position[1],
			shadowTransform.Position[2],
		}
		q := mgl32.Quat{
			W: shadowTransform.Rotation[3],
			V: mgl32.Vec3{
				shadowTransform.Rotation[0],
				shadowTransform.Rotation[1],
				shadowTransform.Rotation[2],
			},
		}
		dir := q.Rotate(mgl32.Vec3{0, 0, -1})

		fov := shadowLight.Angle * (math.Pi / 180.0) * 2
		proj := mgl32.Perspective(fov, 1.0, 0.1, shadowLight.Range)
		view := mgl32.LookAtV(pos, pos.Add(dir), mgl32.Vec3{0, 1, 0})

		lightSpace = proj.Mul4(view)
	}

	return lightSpace, shadowIndex, true
}

func (rs *RenderSystem) RenderShadowPass(entities []*Entity) {
	glutil.ClearGLErrors()

	if rs.Renderer.ShadowFBO == 0 || rs.Renderer.ShadowProgram == 0 {
		return
	}
	// Pick a shadow-casting light (first found for now)
	lightSpace, _, ok := rs.computeShadowLightSpace(entities)
	if !ok {
		return
	}

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
		if t.Rotation != [4]float32{0, 0, 0, 1} {
			q := mgl32.Quat{
				W: t.Rotation[3],
				V: mgl32.Vec3{t.Rotation[0], t.Rotation[1], t.Rotation[2]},
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

		//engine.SetMat4(rs.Renderer.LocModel, &t.WorldMatrix[0])

		//locModel := gl.GetUniformLocation(rs.Renderer.ShadowProgram, gl.Str("model\x00"))
		//gl.UniformMatrix4fv(locModel, 1, false, &t.WorldMatrix[0])

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
	// 1) Bind baseline shader for the frame.
	//    If you later want a global override, you can use rs.ActiveShader here.
	// 1) Determine baseline shader for this frame
	var drawItems []MeshDrawItem
	drawItems = drawItems[:0]
	base := rs.DefaultShader
	if rs.ActiveShader != nil {
		base = rs.ActiveShader
	}

	if base == nil {
		// Fallback: keep whatever renderer.Program currently is.
	} else {
		rs.Renderer.SwitchProgram(base)
	}

	// 2) Upload globals for the currently bound program.

	view := rs.CameraSystem.View
	proj := rs.CameraSystem.Projection

	// Track which shader is currently bound so we only switch when needed.
	currentShader := base
	rs.uploadGlobals(entities, currentShader)

	// 3) Draw all meshes
	for _, e := range entities {
		var t *Transform
		var mesh *Mesh
		var mat *Material
		var normalMapComp *NormalMap
		var multi *MultiMesh
		var multiMat *MultiMaterial

		for _, c := range e.Components {
			switch v := c.(type) {
			case *Transform:
				t = v
			case *Mesh:
				mesh = v
			case *Material:
				mat = v
			case *NormalMap:
				normalMapComp = v
			case *MultiMesh:
				multi = v
			case *MultiMaterial:
				multiMat = v
			}
		}
		if t == nil || mat == nil {
			continue
		}

		// --- Per-material shader selection (additive) ---
		resolveMaterialShader(mat)
		desiredShader := base
		if mat.Shader != nil {
			desiredShader = mat.Shader
		}

		if desiredShader != currentShader {
			currentShader = desiredShader
			if currentShader != nil {
				rs.Renderer.SwitchProgram(currentShader)
				// After switching program, all uniform locations changed,
				// so we re-upload the global state once for this shader.
				rs.uploadGlobals(entities, currentShader)
			}
		}

		// Build model matrix
		model := mgl32.Translate3D(t.Position[0], t.Position[1], t.Position[2])
		if t.Rotation != [4]float32{0, 0, 0, 1} {
			q := mgl32.Quat{
				W: t.Rotation[3],
				V: mgl32.Vec3{t.Rotation[0], t.Rotation[1], t.Rotation[2]},
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
		engine.SetMat4(rs.Renderer.LocModel, &model[0])
		// New Logic useing the World Matrix
		//engine.SetMat4(rs.Renderer.LocModel, &t.WorldMatrix[0])
		engine.SetMat4(rs.Renderer.LocView, &view[0])
		engine.SetMat4(rs.Renderer.LocProj, &proj[0])
		// --- draw each mesh ---

		if rs.materialUBO != 0 {

			matType := mat.Type

			// Default Blinn/Phong shader
			if currentShader == rs.DefaultShader {
				matType = int(MaterialBlinnPhong)
			}

			// PBR shader
			if sp, err := engine.GetShaderProgram("pbr_shader"); err == nil && currentShader == sp {
				matType = int(MaterialPBR)
			}

			// Toon shader
			if sp, err := engine.GetShaderProgram("toon_shader"); err == nil && currentShader == sp {
				matType = int(MaterialToon)
			}

			m := gpuMaterial{
				BaseColor:    mat.BaseColor,
				Ambient:      mat.Ambient,
				Diffuse:      mat.Diffuse,
				Specular:     mat.Specular,
				Shininess:    mat.Shininess,
				Metallic:     mat.Metallic,
				Roughness:    mat.Roughness,
				MaterialType: int32(matType),
			}
			// Selection highlight: override base color if needed
			if uint64(e.ID) == rs.SelectedEntity {
				m.BaseColor = [4]float32{1, 1, 0, 1}
			}

			gl.BindBuffer(gl.UNIFORM_BUFFER, rs.materialUBO)
			gl.BufferSubData(gl.UNIFORM_BUFFER, 0, int(unsafe.Sizeof(m)), unsafe.Pointer(&m))
			gl.BindBuffer(gl.UNIFORM_BUFFER, 0)
		}

		engine.SetVec4fv(rs.Renderer.LocBaseCol, &mat.BaseColor[0])
		if uint64(e.ID) == rs.SelectedEntity {
			highlight := [4]float32{1, 1, 0, 1}
			engine.SetVec4fv(rs.Renderer.LocBaseCol, &highlight[0])
		}
		engine.SetFloat(rs.Renderer.LocAmbient, mat.Ambient)
		engine.SetFloat(rs.Renderer.LocDiffuse, mat.Diffuse)
		engine.SetFloat(rs.Renderer.LocSpecular, mat.Specular)
		engine.SetFloat(rs.Renderer.LocShininess, mat.Shininess)

		// Diffuse texture
		if mat.TextureAsset != 0 {
			textureID := assets.ResolveTextureGLID(assets.AssetID(mat.TextureAsset))
			gl.ActiveTexture(gl.TEXTURE0)
			gl.BindTexture(gl.TEXTURE_2D, textureID)
			engine.SetInt(rs.Renderer.LocDiffuseTex, 0)
			engine.SetInt(rs.Renderer.LocUseTexture, 1)
		} else {
			engine.SetInt(rs.Renderer.LocUseTexture, 0)
		}

		// Normal map
		if normalMapComp != nil && normalMapComp.ID != 0 && rs.MeshManager.HasTangents(mesh.ID) {
			gl.ActiveTexture(gl.TEXTURE1)
			gl.BindTexture(gl.TEXTURE_2D, normalMapComp.ID)
			engine.SetInt(rs.Renderer.LocNormalMap, 1)
			engine.SetInt(rs.Renderer.LocUseNormalMap, 1)
		} else {
			engine.SetInt(rs.Renderer.LocUseNormalMap, 0)
		}
		drawItems = drawItems[:0]
		drawItems = rs.collectMeshes(mesh, multi, mat, multiMat, normalMapComp, drawItems)

		for _, item := range drawItems {
			rs.drawMesh(item.MeshID, item.Material, normalMapComp)
		}
	}
}

func selectMaterialShader(mat *Material) {
	switch mat.ShaderName {
	case "pbr_shade":
		mat.Type = int(MaterialPBR)
	case "toon_shader":
		mat.Type = int(MaterialToon)
	case "default_shader":
		mat.Type = int(MaterialBlinnPhong)
	}
}

// uploadGlobals binds the current rs.Renderer.Program and uploads
// shadow map, lights, camera position, lightSpace and debug flags.
// It assumes rs.Renderer.Program already points to the active shader.
func (rs *RenderSystem) uploadGlobals(entities []*Entity, shader *engine.ShaderProgram) {
	glutil.RunGLChecked("MainPass: UseProgram+Uniforms", func() {

		if !engine.UseProgramChecked("MainPass", rs.Renderer.Program) {
			return
		}

		// Bind shadow map
		gl.ActiveTexture(gl.TEXTURE2)
		gl.BindTexture(gl.TEXTURE_2D, rs.Renderer.ShadowTex)

		if rs.Renderer.LocShadowMap != -1 {
			gl.Uniform1i(rs.Renderer.LocShadowMap, 2)
		}
		if rs.Renderer.LocShadowMapSize != -1 {
			engine.SetVec2(
				rs.Renderer.LocShadowMapSize,
				float32(rs.Renderer.ShadowWidth),
				float32(rs.Renderer.ShadowHeight),
			)
		}

		// --- Rebuild and upload lights ---
		rs.Renderer.LightColor = [3]float32{1, 1, 1}
		rs.Renderer.LightIntensity = 1.0

		lights := make([]engine.LightData, 0, 8)
		shadowLightIdx := -1
		var shadowLight *LightComponent
		var shadowTransform *Transform

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
					W: tr.Rotation[3],
					V: mgl32.Vec3{tr.Rotation[0], tr.Rotation[1], tr.Rotation[2]},
				}
				fwd := q.Rotate(mgl32.Vec3{0, 0, -1})
				dir = [3]float32{fwd.X(), fwd.Y(), fwd.Z()}
			}

			// legacy orbital gizmo light override
			if rs.LightEntity != nil && e == rs.LightEntity && lc.Type == LightDirectional {
				dir = rs.LightDir
			}
			idx := len(lights)
			lights = append(lights, engine.LightData{
				Type:      int32(lc.Type),
				Color:     lc.Color,
				Intensity: lc.Intensity,
				Direction: dir,
				Position:  pos,
				Range:     lc.Range,
				Angle:     lc.Angle,
			})
			if shadowLightIdx == -1 && lc.CastsShadows && (lc.Type == LightDirectional || lc.Type == LightSpot) {
				shadowLightIdx = idx
				shadowLight = lc
				shadowTransform = tr
			}
		}

		engine.SetInt(rs.Renderer.LocLightCount, int32(len(lights)))
		for i, L := range lights {
			if rs.Renderer.LocLightColor[i] != -1 {
				engine.SetVec3(rs.Renderer.LocLightColor[i], L.Color[0], L.Color[1], L.Color[2])
			}
			if rs.Renderer.LocLightIntensity[i] != -1 {
				engine.SetFloat(rs.Renderer.LocLightIntensity[i], L.Intensity)
			}
			if rs.Renderer.LocLightDir[i] != -1 {
				engine.SetVec3(rs.Renderer.LocLightDir[i], L.Direction[0], L.Direction[1], L.Direction[2])
			}
			if rs.Renderer.LocLightPos[i] != -1 {
				engine.SetVec3(rs.Renderer.LocLightPos[i], L.Position[0], L.Position[1], L.Position[2])
			}
			if rs.Renderer.LocLightRange[i] != -1 {
				engine.SetFloat(rs.Renderer.LocLightRange[i], L.Range)
			}
			if rs.Renderer.LocLightAngle[i] != -1 {
				engine.SetFloat(rs.Renderer.LocLightAngle[i], L.Angle)
			}
			if rs.Renderer.LocLightType[i] != -1 {
				engine.SetInt(rs.Renderer.LocLightType[i], L.Type)
			}
		}

		// Upload camera position
		camPos := rs.CameraSystem.Position
		gl.Uniform3fv(rs.Renderer.LocViewPos, 1, &camPos[0])

		// Upload lightSpace + shadow light index
		if shadowLightIdx >= 0 && shadowLight != nil && shadowTransform != nil {
			lightSpace, _, ok := rs.computeShadowLightSpace(entities)
			if ok {
				//log.Printf("Shadow light index = %d", shadowLightIdx)
				if rs.Renderer.LocLightSpace != -1 {
					gl.UniformMatrix4fv(rs.Renderer.LocLightSpace, 1, false, &lightSpace[0])
				}
				if rs.Renderer.LocShadowLightIndex != -1 {
					gl.Uniform1i(rs.Renderer.LocShadowLightIndex, int32(shadowLightIdx))
				}
			}
		}

		// Debug flags
		engine.SetInt(rs.Renderer.LocShowMode, rs.DebugShowMode)
		engine.SetInt(rs.Renderer.LocFlipNormalG, boolToInt(rs.DebugFlipGreen))
		if shader.HasMaterialBlock {
			blockIndex := gl.GetUniformBlockIndex(shader.ID, gl.Str("MaterialBlock\x00"))
			gl.UniformBlockBinding(shader.ID, blockIndex, rs.materialBinding)
			gl.BindBufferBase(gl.UNIFORM_BUFFER, rs.materialBinding, rs.materialUBO)
		}

	})
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

func (rs *RenderSystem) selectShaderForPass(entities []*Entity) {
	var chosen *engine.ShaderProgram

	// Example: use selected entity’s material shader
	for _, e := range entities {
		if uint64(e.ID) != rs.SelectedEntity {
			continue
		}
		if mat, ok := e.GetComponent((*Material)(nil)).(*Material); ok {
			if mat.Shader != nil {
				chosen = mat.Shader
			}
		}
	}

	// fallback to default renderer program
	if chosen == nil {
		gl.UseProgram(rs.Renderer.Program)
		return
	}

	// bind chosen shader ONCE
	gl.UseProgram(chosen.ID)
	rs.Renderer.Program = chosen.ID
	rs.Renderer.InitUniforms()
}

func (rs *RenderSystem) BindMaterialUBO(sp *engine.ShaderProgram) {
	if !sp.HasMaterialBlock {
		return
	}
	blockIndex := gl.GetUniformBlockIndex(sp.ID, gl.Str("MaterialBlock\x00"))
	gl.UniformBlockBinding(sp.ID, blockIndex, rs.materialBinding)
	gl.BindBufferBase(gl.UNIFORM_BUFFER, rs.materialBinding, rs.materialUBO)
}

func (rs *RenderSystem) SetGlobalShader(p *engine.ShaderProgram) {
	rs.ActiveShader = p
	rs.Renderer.SwitchProgram(p)
}

func (rs *RenderSystem) collectMeshes(
	mesh *Mesh,
	multi *MultiMesh,
	mat *Material,
	multiMat *MultiMaterial,
	normalMap *NormalMap,
	out []MeshDrawItem,
) []MeshDrawItem {

	if multi != nil {
		for _, meshID := range multi.Meshes {
			m := mat
			if multiMat != nil {
				if mm, ok := multiMat.Materials[meshID]; ok {
					m = mm
				}
			}
			out = append(out, MeshDrawItem{
				MeshID:    meshID,
				Material:  m,
				NormalMap: normalMap,
			})
		}
		return out
	}

	if mesh != nil {
		out = append(out, MeshDrawItem{
			MeshID:    mesh.ID,
			Material:  mat,
			NormalMap: normalMap,
		})
	}

	return out
}

func (rs *RenderSystem) drawMesh(
	meshID string,
	mat *Material,
	normalMap *NormalMap,
) {
	vao := rs.MeshManager.GetVAO(meshID)
	if vao == 0 {
		return
	}

	indexCount := rs.MeshManager.GetCount(meshID)
	indexType := rs.MeshManager.GetIndexType(meshID)
	vertexCount := rs.MeshManager.GetVertexCount(meshID)
	ebo := rs.MeshManager.GetEBO(meshID)

	gl.BindVertexArray(vao)

	if indexCount > 0 {
		if ebo == 0 {
			log.Printf("SKIP indexed draw: mesh=%s has indexCount=%d but no EBO", meshID, indexCount)
			gl.BindVertexArray(0)
			return
		}

		bytesPerIndex := int32(4)
		if indexType == gl.UNSIGNED_SHORT {
			bytesPerIndex = 2
		}

		var eboSize int32
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
		gl.GetBufferParameteriv(gl.ELEMENT_ARRAY_BUFFER, gl.BUFFER_SIZE, &eboSize)

		required := int32(indexCount) * bytesPerIndex
		if eboSize < required {
			log.Printf("SKIP draw: mesh=%s indexCount=%d (%d bytes) but EBO size=%d bytes",
				meshID, indexCount, required, eboSize)
			gl.BindVertexArray(0)
			return
		}

		if meshID == "line" {
			gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
			gl.DrawElements(gl.LINES, indexCount, indexType, gl.PtrOffset(0))
			gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
		} else {
			gl.DrawElements(gl.TRIANGLES, indexCount, indexType, gl.PtrOffset(0))
		}
	} else {
		if meshID == "line" {
			gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
			gl.DrawArrays(gl.LINES, 0, vertexCount)
			gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
		} else {
			gl.DrawArrays(gl.TRIANGLES, 0, vertexCount)
		}
	}
	/*
		if indexCount > 0 && ebo != 0 {
			gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
			gl.DrawElements(gl.TRIANGLES, indexCount, indexType, gl.PtrOffset(0))
		} else {
			gl.DrawArrays(gl.TRIANGLES, 0, vertexCount)
		}*/

	gl.BindVertexArray(0)
}
