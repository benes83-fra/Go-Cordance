// internal/engine/renderer.go
package engine

import (
	"fmt"

	"github.com/go-gl/gl/v4.1-core/gl"
)

type Renderer struct {
	Program    uint32
	LocModel   int32
	LocView    int32
	LocProj    int32
	LocBaseCol int32
}

func (r *Renderer) InitUniforms() {
	r.LocModel = gl.GetUniformLocation(r.Program, gl.Str("model\x00"))
	r.LocView = gl.GetUniformLocation(r.Program, gl.Str("view\x00"))
	r.LocProj = gl.GetUniformLocation(r.Program, gl.Str("projection\x00"))
	r.LocBaseCol = gl.GetUniformLocation(r.Program, gl.Str("baseColor\x00"))
}

func NewRenderer(vertexSrc, fragmentSrc string, width, height int) *Renderer {
	prog := compileProgram(vertexSrc, fragmentSrc)
	gl.Enable(gl.DEPTH_TEST)
	gl.ClearColor(0.1, 0.1, 0.1, 1.0)
	gl.Viewport(0, 0, int32(width), int32(height)) // critical

	return &Renderer{Program: prog}
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
