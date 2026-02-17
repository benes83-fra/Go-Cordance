package glutil

import (
	"log"
	"strconv"

	"github.com/go-gl/gl/v4.1-core/gl"
)

// ClearGLErrors drains the GL error queue.
func ClearGLErrors() {
	for {
		if err := gl.GetError(); err == gl.NO_ERROR {
			return
		}
	}
}

// glErrToString maps common GL error codes to readable strings.
func glErrToString(e uint32) string {
	switch e {
	case gl.NO_ERROR:
		return "GL_NO_ERROR"
	case gl.INVALID_ENUM:
		return "GL_INVALID_ENUM"
	case gl.INVALID_VALUE:
		return "GL_INVALID_VALUE"
	case gl.INVALID_OPERATION:
		return "GL_INVALID_OPERATION"
	case gl.INVALID_FRAMEBUFFER_OPERATION:
		return "GL_INVALID_FRAMEBUFFER_OPERATION"
	case gl.OUT_OF_MEMORY:
		return "GL_OUT_OF_MEMORY"
	default:
		return "GL_UNKNOWN(0x" + strconv.FormatUint(uint64(e), 16) + ")"
	}
}

// LogGLErrors logs all errors currently in the GL error queue with context.
func LogGLErrors(context string) {
	for {
		err := gl.GetError()
		if err == gl.NO_ERROR {
			return
		}
		log.Printf("GLERR [%s]: 0x%X %s", context, err, glErrToString(err))
		panic("Stopped")
	}
}

// RunGLChecked clears errors, runs the block, then logs any errors with context.
// Use this to wrap a small block of GL calls you want to validate.
func RunGLChecked(context string, block func()) {
	ClearGLErrors()
	block()
	LogGLErrors(context)
}
