package engine

import (
	"fmt"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
)

// Renderer manages simple GL resources for the prototype.
type Renderer struct {
	Program uint32
	vao     uint32
	vbo     uint32
}

// NewRenderer returns an empty renderer. Call Init after GL is initialized.
func NewRenderer() *Renderer {
	return &Renderer{}
}

// Init compiles shaders (from source) and sets up a simple triangle VAO/VBO.
// vertSrc and fragSrc must be null-terminated C strings (use gl.Str).
func (r *Renderer) Init(vertSrc, fragSrc string) error {
	prog, err := newProgram(vertSrc, fragSrc)
	if err != nil {
		return err
	}
	r.Program = prog

	// Triangle data: positions and colors
	vertices := []float32{
		// X, Y, Z,    R, G, B
		0.0, 0.5, 0.0, 1.0, 0.0, 0.0,
		-0.5, -0.5, 0.0, 0.0, 1.0, 0.0,
		0.5, -0.5, 0.0, 0.0, 0.0, 1.0,
	}

	gl.GenVertexArrays(1, &r.vao)
	gl.BindVertexArray(r.vao)

	gl.GenBuffers(1, &r.vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// position attribute
	posAttrib := uint32(gl.GetAttribLocation(r.Program, gl.Str("vp\x00")))
	gl.EnableVertexAttribArray(posAttrib)
	gl.VertexAttribPointer(posAttrib, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(0))

	// color attribute
	colAttrib := uint32(gl.GetAttribLocation(r.Program, gl.Str("color\x00")))
	gl.EnableVertexAttribArray(colAttrib)
	gl.VertexAttribPointer(colAttrib, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(3*4))

	return nil
}

// Draw issues the draw call for the triangle.
func (r *Renderer) Draw() {
	gl.UseProgram(r.Program)
	gl.BindVertexArray(r.vao)
	gl.DrawArrays(gl.TRIANGLES, 0, 3)
}

// Shutdown frees GL resources owned by the renderer.
func (r *Renderer) Shutdown() {
	if r.vbo != 0 {
		gl.DeleteBuffers(1, &r.vbo)
		r.vbo = 0
	}
	if r.vao != 0 {
		gl.DeleteVertexArrays(1, &r.vao)
		r.vao = 0
	}
	if r.Program != 0 {
		gl.DeleteProgram(r.Program)
		r.Program = 0
	}
}

// Helpers: compile shader and link program

func compileShader(src string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)
	csources, free := gl.Strs(src)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)
		log := make([]byte, logLength+1)
		gl.GetShaderInfoLog(shader, logLength, nil, &log[0])

		// Provide a short preview of the source for debugging
		srcPreview := src
		if len(srcPreview) > 1024 {
			srcPreview = srcPreview[:1024] + "...(truncated)"
		}
		// Remove null terminator for readability
		srcPreview = strings.TrimSuffix(srcPreview, "\x00")

		return 0, fmt.Errorf("shader compile error (%s):\n--- source preview ---\n%s\n--- info log ---\n%s",
			shaderTypeString(shaderType), srcPreview, string(log))
	}
	return shader, nil
}

func shaderTypeString(t uint32) string {
	switch t {
	case gl.VERTEX_SHADER:
		return "VERTEX_SHADER"
	case gl.FRAGMENT_SHADER:
		return "FRAGMENT_SHADER"
	default:
		return fmt.Sprintf("SHADER(%d)", t)
	}
}

func newProgram(vertSrc, fragSrc string) (uint32, error) {
	vert, err := compileShader(vertSrc, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}
	frag, err := compileShader(fragSrc, gl.FRAGMENT_SHADER)
	if err != nil {
		gl.DeleteShader(vert)
		return 0, err
	}

	prog := gl.CreateProgram()
	gl.AttachShader(prog, vert)
	gl.AttachShader(prog, frag)
	gl.LinkProgram(prog)

	var status int32
	gl.GetProgramiv(prog, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(prog, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(prog, logLength, nil, gl.Str(log))
		gl.DeleteShader(vert)
		gl.DeleteShader(frag)
		gl.DeleteProgram(prog)
		return 0, fmt.Errorf("failed to link program: %s", log)
	}

	// shaders can be deleted after linking
	gl.DeleteShader(vert)
	gl.DeleteShader(frag)
	return prog, nil
}
