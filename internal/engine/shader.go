package engine

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
)

type ShaderProgram struct {
	ID               uint32
	Uniforms         map[string]int32
	HasMaterialBlock bool
}

func LoadShaderProgram(name, vertSrc, fragSrc string) (*ShaderProgram, error) {
	vs := compileShader(vertSrc, gl.VERTEX_SHADER)
	fs := compileShader(fragSrc, gl.FRAGMENT_SHADER)

	prog, err := buildProgram(vs, fs)
	if err != nil {
		return nil, err
	}

	gl.DeleteShader(vs)
	gl.DeleteShader(fs)

	sp := &ShaderProgram{
		ID:       prog,
		Uniforms: map[string]int32{},
	}

	// --- NEW: detect MaterialBlock ---
	blockIndex := gl.GetUniformBlockIndex(prog, gl.Str("MaterialBlock\x00"))
	sp.HasMaterialBlock = blockIndex != gl.INVALID_INDEX

	return sp, nil
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

func buildProgram(vs, fs uint32) (uint32, error) {
	prog := gl.CreateProgram()
	gl.AttachShader(prog, vs)
	gl.AttachShader(prog, fs)
	gl.LinkProgram(prog)

	var status int32
	gl.GetProgramiv(prog, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLen int32
		gl.GetProgramiv(prog, gl.INFO_LOG_LENGTH, &logLen)
		logBuf := make([]byte, logLen+1)
		gl.GetProgramInfoLog(prog, logLen, nil, &logBuf[0])
		gl.DeleteProgram(prog)
		return 0, fmt.Errorf("program link error: %s", string(logBuf))
	}
	return prog, nil
}

// LoadShaderSource reads a shader file, strips a UTF-8 BOM if present,
// normalizes CRLF to LF, and returns a null-terminated string suitable for gl.Strs.
func LoadShaderSource(path string) (string, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	// Strip UTF-8 BOM if present
	if len(b) >= 3 && b[0] == 0xEF && b[1] == 0xBB && b[2] == 0xBF {
		b = b[3:]
	}
	s := string(b)
	// Normalize CRLF -> LF
	s = strings.ReplaceAll(s, "\r\n", "\n")
	// Trim leading nulls or spaces to ensure #version is first token
	s = strings.TrimLeft(s, "\u0000 \t\n\r")
	return s + "\x00", nil
}
