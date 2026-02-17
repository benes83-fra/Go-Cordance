package engine

import (
	"log"

	"github.com/go-gl/gl/v4.1-core/gl"
)

// All helpers are no-ops if loc == -1.

func SetInt(loc int32, v int32) {
	if loc == -1 {
		return
	}
	gl.Uniform1i(loc, v)
}

func SetFloat(loc int32, v float32) {
	if loc == -1 {
		return
	}
	gl.Uniform1f(loc, v)
}

func SetVec2(loc int32, x, y float32) {
	if loc == -1 {
		return
	}
	gl.Uniform2f(loc, x, y)
}

func SetVec3(loc int32, x, y, z float32) {
	if loc == -1 {
		return
	}
	gl.Uniform3f(loc, x, y, z)
}

func SetVec4fv(loc int32, v *float32) {
	if loc == -1 {
		return
	}
	gl.Uniform4fv(loc, 1, v)
}

func SetMat4(loc int32, m *float32) {
	if loc == -1 {
		return
	}
	gl.UniformMatrix4fv(loc, 1, false, m)
}

func UseProgramChecked(label string, prog uint32) bool {
	if prog == 0 {
		log.Printf("%s: prog == 0, skipping gl.UseProgram", label)
		return false
	}

	var linked int32
	gl.GetProgramiv(prog, gl.LINK_STATUS, &linked)
	if linked == gl.FALSE {
		log.Printf("%s: prog %d not linked, skipping", label, prog)
		return false
	}

	gl.UseProgram(prog)
	return true
}
