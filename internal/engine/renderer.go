// internal/engine/renderer.go
package engine

import (
	"fmt"
	"log"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type Renderer struct {
	Program    uint32
	LocModel   int32
	LocView    int32
	LocProj    int32
	LocBaseCol int32

	LocViewPos    int32
	LocAmbient    int32
	LocDiffuse    int32
	LocSpecular   int32
	LocShininess  int32
	LocDiffuseTex int32
	LocUseTexture int32

	LocLightCount     int32
	LocLightDir       [8]int32
	LocLightColor     [8]int32
	LocLightIntensity [8]int32

	LightColor     [3]float32
	LightIntensity float32
	LocLightPos    [8]int32
	LocLightRange  [8]int32
	LocLightAngle  [8]int32
	LocLightType   [8]int32

	// new debug / normal map uniforms
	LocNormalMap    int32
	LocUseNormalMap int32
	LocFlipNormalG  int32
	LocShowMode     int32

	// shadow map
	ShadowFBO    uint32
	ShadowTex    uint32
	ShadowWidth  int
	ShadowHeight int

	// shadow shader program (depth-only)
	ShadowProgram uint32

	// uniform locations
	LocLightSpace       int32
	LocShadowMap        int32
	LocShadowMapSize    int32
	LocShadowLightIndex int32
	// store screen size for viewport restore
	ScreenWidth  int
	ScreenHeight int
}

func (r *Renderer) InitUniforms() {
	r.LocModel = gl.GetUniformLocation(r.Program, gl.Str("model\x00"))
	r.LocView = gl.GetUniformLocation(r.Program, gl.Str("view\x00"))
	r.LocProj = gl.GetUniformLocation(r.Program, gl.Str("projection\x00"))
	r.LocBaseCol = gl.GetUniformLocation(r.Program, gl.Str("BaseColor\x00"))

	r.LocViewPos = gl.GetUniformLocation(r.Program, gl.Str("viewPos\x00"))
	r.LocAmbient = gl.GetUniformLocation(r.Program, gl.Str("matAmbient\x00"))
	r.LocDiffuse = gl.GetUniformLocation(r.Program, gl.Str("matDiffuse\x00"))
	r.LocSpecular = gl.GetUniformLocation(r.Program, gl.Str("matSpecular\x00"))
	r.LocShininess = gl.GetUniformLocation(r.Program, gl.Str("matShininess\x00"))
	r.LocDiffuseTex = gl.GetUniformLocation(r.Program, gl.Str("diffuseTex\x00"))
	r.LocUseTexture = gl.GetUniformLocation(r.Program, gl.Str("useTexture\x00"))

	r.LocLightCount = gl.GetUniformLocation(r.Program, gl.Str("lightCount\x00"))
	for i := 0; i < 8; i++ {
		nameDir := fmt.Sprintf("lightDir[%d]\x00", i)
		nameCol := fmt.Sprintf("lightColor[%d]\x00", i)
		nameInt := fmt.Sprintf("lightIntensity[%d]\x00", i)
		namePos := fmt.Sprintf("lightPos[%d]\x00", i)
		nameRange := fmt.Sprintf("lightRange[%d]\x00", i)
		nameAngle := fmt.Sprintf("lightAngle[%d]\x00", i)
		nameType := fmt.Sprintf("lightType[%d]\x00", i)

		r.LocLightDir[i] = gl.GetUniformLocation(r.Program, gl.Str(nameDir))
		r.LocLightColor[i] = gl.GetUniformLocation(r.Program, gl.Str(nameCol))
		r.LocLightIntensity[i] = gl.GetUniformLocation(r.Program, gl.Str(nameInt))
		r.LocLightPos[i] = gl.GetUniformLocation(r.Program, gl.Str(namePos))
		r.LocLightRange[i] = gl.GetUniformLocation(r.Program, gl.Str(nameRange))
		r.LocLightAngle[i] = gl.GetUniformLocation(r.Program, gl.Str(nameAngle))
		r.LocLightType[i] = gl.GetUniformLocation(r.Program, gl.Str(nameType))
	}
	// main shader will sample shadow map and receive lightSpaceMatrix
	r.LocLightSpace = gl.GetUniformLocation(r.Program, gl.Str("lightSpaceMatrix\x00"))
	r.LocShadowMap = gl.GetUniformLocation(r.Program, gl.Str("shadowMap\x00"))
	r.LocShadowMapSize = gl.GetUniformLocation(r.Program, gl.Str("uShadowMapSize\x00"))

	//for debugging purpose
	// renderer.InitUniforms (add these lines)
	r.LocNormalMap = gl.GetUniformLocation(r.Program, gl.Str("normalMap\x00"))
	r.LocUseNormalMap = gl.GetUniformLocation(r.Program, gl.Str("useNormalMap\x00"))
	r.LocFlipNormalG = gl.GetUniformLocation(r.Program, gl.Str("flipNormalGreen\x00"))
	r.LocShowMode = gl.GetUniformLocation(r.Program, gl.Str("showMode\x00"))
	r.LocShadowLightIndex = gl.GetUniformLocation(r.Program, gl.Str("shadowLightIndex\x00"))

	names := map[string]int32{
		"model": r.LocModel, "view": r.LocView, "projection": r.LocProj,
		"baseColor": r.LocBaseCol, "viewPos": r.LocViewPos,
		"matAmbient": r.LocAmbient, "matDiffuse": r.LocDiffuse,
		"matSpecular": r.LocSpecular, "matShininess": r.LocShininess,
		"diffuseTex": r.LocDiffuseTex, "useTexture": r.LocUseTexture,
	}
	log.Println("LocLightSpace =", r.LocLightSpace, "LocShadowLightIndex =", r.LocShadowLightIndex)

	for n, loc := range names {
		if loc == -1 {
			fmt.Printf("WARN: uniform %s not found in program\n", n)
		}
	}

}

func newRendererBase(width, height int) *Renderer {
	r := &Renderer{
		ScreenWidth:    width,
		ScreenHeight:   height,
		LightColor:     [3]float32{1.0, 1.0, 1.0},
		LightIntensity: 1.0,
	}

	gl.Enable(gl.DEPTH_TEST)
	gl.ClearColor(0.1, 0.1, 0.1, 1.0)
	gl.Viewport(0, 0, int32(width), int32(height))

	return r
}

// existing constructor: unchanged from the outside
func NewRenderer(vertexSrc, fragmentSrc string, width, height int) *Renderer {
	r := newRendererBase(width, height)
	r.Program = compileProgram(vertexSrc, fragmentSrc)
	// legacy path can choose to call InitUniforms explicitly where it always did before
	return r
}

func NewRendererWithProgram(program uint32, width, height int) *Renderer {
	r := newRendererBase(width, height)
	r.Program = program
	gl.UseProgram(program)
	r.InitUniforms()
	return r
}

// optional: switch program on an existing renderer (for hot-reload)
func (r *Renderer) SetProgram(program uint32) {
	r.Program = program
	gl.UseProgram(program)
	r.InitUniforms()
}
func compileShader(src string, t uint32) uint32 {
	shader := gl.CreateShader(t)
	csrc, free := gl.Strs(src + "\x00")
	gl.ShaderSource(shader, 1, csrc, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLen int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLen)
		log := make([]byte, logLen)
		gl.GetShaderInfoLog(shader, logLen, nil, &log[0])
		panic(fmt.Sprintf("shader compile error: %s", log))
	}
	return shader
}

func compileProgram(vert, frag string) uint32 {
	vs := compileShader(vert, gl.VERTEX_SHADER)
	fs := compileShader(frag, gl.FRAGMENT_SHADER)

	prog := gl.CreateProgram()
	gl.AttachShader(prog, vs)
	gl.AttachShader(prog, fs)
	gl.LinkProgram(prog)

	var status int32
	gl.GetProgramiv(prog, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLen int32
		gl.GetProgramiv(prog, gl.INFO_LOG_LENGTH, &logLen)
		log := make([]byte, logLen)
		gl.GetProgramInfoLog(prog, logLen, nil, &log[0])
		panic(fmt.Sprintf("program link error: %s", log))
	}

	gl.DeleteShader(vs)
	gl.DeleteShader(fs)
	return prog
}

type DebugRenderer struct {
	Program  uint32
	LocModel int32
	LocView  int32
	LocProj  int32
	LocColor int32

	lineVAO uint32
	lineVBO uint32
}

func NewDebugRenderer(vertexSrc, fragmentSrc string) *DebugRenderer {
	prog := compileProgram(vertexSrc, fragmentSrc)

	dr := &DebugRenderer{
		Program:  prog,
		LocModel: gl.GetUniformLocation(prog, gl.Str("model\x00")),
		LocView:  gl.GetUniformLocation(prog, gl.Str("view\x00")),
		LocProj:  gl.GetUniformLocation(prog, gl.Str("projection\x00")),
		LocColor: gl.GetUniformLocation(prog, gl.Str("debugColor\x00")),
	}

	// Create VAO/VBO for a single line segment
	gl.GenVertexArrays(1, &dr.lineVAO)
	gl.GenBuffers(1, &dr.lineVBO)

	gl.BindVertexArray(dr.lineVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, dr.lineVBO)

	// allocate space for 2 vec3 positions (start + end)
	gl.BufferData(gl.ARRAY_BUFFER, 6*4, nil, gl.DYNAMIC_DRAW)

	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointerWithOffset(0, 3, gl.FLOAT, false, 3*4, 0)

	gl.BindVertexArray(0)

	return dr
}
func (dr *DebugRenderer) DrawLine(start, end mgl32.Vec3, color mgl32.Vec3, view, proj mgl32.Mat4) {
	gl.UseProgram(dr.Program)

	// Upload uniforms
	model := mgl32.Ident4()
	gl.UniformMatrix4fv(dr.LocView, 1, false, &view[0])
	gl.UniformMatrix4fv(dr.LocProj, 1, false, &proj[0])
	gl.UniformMatrix4fv(dr.LocModel, 1, false, &model[0])
	gl.Uniform3fv(dr.LocColor, 1, &color[0])

	// Upload line vertices
	verts := []float32{
		start.X(), start.Y(), start.Z(),
		end.X(), end.Y(), end.Z(),
	}

	gl.BindBuffer(gl.ARRAY_BUFFER, dr.lineVBO)
	gl.BufferSubData(gl.ARRAY_BUFFER, 0, len(verts)*4, gl.Ptr(verts))

	gl.BindVertexArray(dr.lineVAO)
	gl.DrawArrays(gl.LINES, 0, 2)
	gl.BindVertexArray(0)
}

// Create depth-only FBO + texture and compile shadow shader program.
// Call this from main after NewRenderer.
func (r *Renderer) InitShadow(shadowVertSrc, shadowFragSrc string, width, height int) {
	r.ShadowWidth = width
	r.ShadowHeight = height

	// create depth texture
	var tex uint32
	gl.GenTextures(1, &tex)
	gl.BindTexture(gl.TEXTURE_2D, tex)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.DEPTH_COMPONENT, int32(width), int32(height), 0, gl.DEPTH_COMPONENT, gl.FLOAT, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_BORDER)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_BORDER)
	border := []float32{1.0, 1.0, 1.0, 1.0}
	gl.TexParameterfv(gl.TEXTURE_2D, gl.TEXTURE_BORDER_COLOR, &border[0])

	// create FBO
	var fbo uint32
	gl.GenFramebuffers(1, &fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.TEXTURE_2D, tex, 0)
	gl.DrawBuffer(gl.NONE)
	gl.ReadBuffer(gl.NONE)
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	r.ShadowFBO = fbo
	r.ShadowTex = tex

	// compile shadow shader program
	r.ShadowProgram = compileProgram(shadowVertSrc, shadowFragSrc)

	// get uniform locations for shadow shader and main shader usage

	// For main shader, we will set LocShadowMap on the main program (use InitUniforms or set later)
	// But store a location name for convenience (we'll get it from main program in InitUniforms)
}
