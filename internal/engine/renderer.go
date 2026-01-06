// internal/engine/renderer.go
package engine

import (
	"fmt"

	"github.com/go-gl/gl/v4.1-core/gl"
)

type Renderer struct {
	Program           uint32
	LocModel          int32
	LocView           int32
	LocProj           int32
	LocBaseCol        int32
	LocLightDir       int32
	LocViewPos        int32
	LocAmbient        int32
	LocDiffuse        int32
	LocSpecular       int32
	LocShininess      int32
	LocDiffuseTex     int32
	LocUseTexture     int32
	LocLightColor     int32
	LocLightIntensity int32

	LightColor     [3]float32
	LightIntensity float32
	// new debug / normal map uniforms
	LocNormalMap    int32
	LocUseNormalMap int32
	LocFlipNormalG  int32
	LocShowMode     int32

	//imgui support
}

func (r *Renderer) InitUniforms() {
	r.LocModel = gl.GetUniformLocation(r.Program, gl.Str("model\x00"))
	r.LocView = gl.GetUniformLocation(r.Program, gl.Str("view\x00"))
	r.LocProj = gl.GetUniformLocation(r.Program, gl.Str("projection\x00"))
	r.LocBaseCol = gl.GetUniformLocation(r.Program, gl.Str("BaseColor\x00"))
	r.LocLightDir = gl.GetUniformLocation(r.Program, gl.Str("lightDir\x00"))
	r.LocViewPos = gl.GetUniformLocation(r.Program, gl.Str("viewPos\x00"))
	r.LocAmbient = gl.GetUniformLocation(r.Program, gl.Str("matAmbient\x00"))
	r.LocDiffuse = gl.GetUniformLocation(r.Program, gl.Str("matDiffuse\x00"))
	r.LocSpecular = gl.GetUniformLocation(r.Program, gl.Str("matSpecular\x00"))
	r.LocShininess = gl.GetUniformLocation(r.Program, gl.Str("matShininess\x00"))
	r.LocDiffuseTex = gl.GetUniformLocation(r.Program, gl.Str("diffuseTex\x00"))
	r.LocUseTexture = gl.GetUniformLocation(r.Program, gl.Str("useTexture\x00"))
	r.LocLightColor = gl.GetUniformLocation(r.Program, gl.Str("lightColor\x00"))
	r.LocLightIntensity = gl.GetUniformLocation(r.Program, gl.Str("lightIntensity\x00"))

	//for debugging purpose
	// renderer.InitUniforms (add these lines)
	r.LocNormalMap = gl.GetUniformLocation(r.Program, gl.Str("normalMap\x00"))
	r.LocUseNormalMap = gl.GetUniformLocation(r.Program, gl.Str("useNormalMap\x00"))
	r.LocFlipNormalG = gl.GetUniformLocation(r.Program, gl.Str("flipNormalGreen\x00"))
	r.LocShowMode = gl.GetUniformLocation(r.Program, gl.Str("showMode\x00"))

	names := map[string]int32{
		"model": r.LocModel, "view": r.LocView, "projection": r.LocProj,
		"baseColor": r.LocBaseCol, "lightDir": r.LocLightDir, "viewPos": r.LocViewPos,
		"matAmbient": r.LocAmbient, "matDiffuse": r.LocDiffuse,
		"matSpecular": r.LocSpecular, "matShininess": r.LocShininess,
		"diffuseTex": r.LocDiffuseTex, "useTexture": r.LocUseTexture,
		"lightColor": r.LocLightColor, "lightIntensity": r.LocLightIntensity,
	}

	for n, loc := range names {
		if loc == -1 {
			fmt.Printf("WARN: uniform %s not found in program\n", n)
		}
	}

}

func NewRenderer(vertexSrc, fragmentSrc string, width, height int) *Renderer {
	r := &Renderer{}

	// Compile shader program
	r.Program = compileProgram(vertexSrc, fragmentSrc)

	// GL state
	gl.Enable(gl.DEPTH_TEST)
	gl.ClearColor(0.1, 0.1, 0.1, 1.0)
	gl.Viewport(0, 0, int32(width), int32(height))
	r.LightColor = [3]float32{1.0, 1.0, 1.0}
	r.LightIntensity = 1.0
	return r
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
}

func NewDebugRenderer(vertexSrc, fragmentSrc string) *DebugRenderer {
	prog := compileProgram(vertexSrc, fragmentSrc) // reuse your shader compile helper
	return &DebugRenderer{
		Program:  prog,
		LocModel: gl.GetUniformLocation(prog, gl.Str("model\x00")),
		LocView:  gl.GetUniformLocation(prog, gl.Str("view\x00")),
		LocProj:  gl.GetUniformLocation(prog, gl.Str("projection\x00")),
		LocColor: gl.GetUniformLocation(prog, gl.Str("debugColor\x00")),
	}
}
