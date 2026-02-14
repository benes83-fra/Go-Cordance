package engine

import (
	"fmt"

	"github.com/go-gl/gl/v4.1-core/gl"
)

type ShaderProgram struct {
	ID       uint32
	Uniforms map[string]int32
}

func LoadShaderProgram(name, vertSrc, fragSrc string) (*ShaderProgram, error) {
	vs := compileShader(vertSrc, gl.VERTEX_SHADER)

	fs := compileShader(fragSrc, gl.FRAGMENT_SHADER)

	prog, err := linkProgram(vs, fs)
	if err != nil {
		return nil, err
	}

	gl.DeleteShader(vs)
	gl.DeleteShader(fs)

	return &ShaderProgram{
		ID:       prog,
		Uniforms: map[string]int32{},
	}, nil
}

// global registry of compiled shader programs
var shaderPrograms = map[string]*ShaderProgram{}

// RegisterShaderProgram stores a compiled shader program under a name.
func RegisterShaderProgram(name string, prog *ShaderProgram) {
	shaderPrograms[name] = prog
}

// GetShaderProgram retrieves a compiled shader program by name.
func GetShaderProgram(name string) (*ShaderProgram, error) {
	p, ok := shaderPrograms[name]
	if !ok {
		return nil, fmt.Errorf("shader program %q not found", name)
	}
	return p, nil
}

// MustGetShaderProgram panics if missing (useful for render.go)
func MustGetShaderProgram(name string) *ShaderProgram {
	p, ok := shaderPrograms[name]
	if !ok {
		panic("shader program not registered: " + name)
	}
	return p
}
