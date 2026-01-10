package engine

import (
	"runtime"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

// InitGLFW initializes GLFW and creates a window with an OpenGL context.
// Caller must call glfw.Terminate() when done.
func InitGLFW(width, height int, title string) (*glfw.Window, error) {
	runtime.LockOSThread() // required by GLFW / OpenGL
	if err := glfw.Init(); err != nil {
		return nil, err
	}

	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	// On macOS uncomment the following:
	// glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(width, height, title, nil, nil)
	if err != nil {
		glfw.Terminate()
		return nil, err
	}
	window.MakeContextCurrent()
	glfw.SwapInterval(1)

	if err := gl.Init(); err != nil {
		window.Destroy()
		glfw.Terminate()
		return nil, err
	}

	gl.Enable(gl.DEPTH_TEST)
	gl.ClearColor(0.1, 0.1, 0.1, 1.0)
	gl.Viewport(0, 0, int32(width), int32(height))
	return window, nil
}
func PollEvents() {
	glfw.PollEvents()
}

// TerminateGLFW terminates GLFW. Call this once during shutdown.
func TerminateGLFW() {
	glfw.Terminate()
}
