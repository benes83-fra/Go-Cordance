package main

import (
	"log"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"

	"go-engine/Go-Cordance/internal/ecs"
	"go-engine/Go-Cordance/internal/engine"
	"go-engine/Go-Cordance/internal/scene"
)

const (
	width  = 800
	height = 600
)

func main() {
	/*if err := glfw.Init(); err != nil {
		log.Fatal(err)
	}
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(width, height, "Go-Cordance Prototype", nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	window.MakeContextCurrent()
	glfw.SwapInterval(1)

	if err := gl.Init(); err != nil {
		log.Fatal(err)
	}*/

	window, err := engine.InitGLFW(width, height, "Go Cordance")
	if err != nil {
		log.Fatal(err)
	}
	// Compile shaders and set viewport
	vertexSrc := `#version 330 core
		    layout(location = 0) in vec3 position;
		    uniform mat4 model;
		    uniform mat4 view;
		    uniform mat4 projection;
		    void main() {
		        gl_Position = projection * view * model * vec4(position, 1.0);
		    }`
	fragmentSrc := `#version 330 core
			out vec4 FragColor;
			uniform vec4 baseColor;
			void main() {
				FragColor = baseColor;
			}`

	renderer := engine.NewRenderer(vertexSrc, fragmentSrc, width, height)
	renderer.InitUniforms()
	// Resize callback updates viewport

	window.SetFramebufferSizeCallback(func(_ *glfw.Window, w, h int) {
		if h == 0 {
			h = 1
		}
		gl.Viewport(0, 0, int32(w), int32(h))
	})
	window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)

	meshMgr := engine.NewMeshManager()
	meshMgr.RegisterTriangle("triangle")

	scene := scene.New()

	camSys := ecs.NewCameraSystem(window)
	renderSys := ecs.NewRenderSystem(renderer, meshMgr, camSys)

	camCtrl := ecs.NewCameraControllerSystem(window)

	scene.Systems().AddSystem(ecs.NewForceSystem(0, -9.8, 0)) // gravity
	scene.Systems().AddSystem(ecs.NewPhysicsSystem())
	scene.Systems().AddSystem(camCtrl) // updates Camera component
	scene.Systems().AddSystem(camSys)  // computes view/projection from Camera
	scene.Systems().AddSystem(renderSys)

	// Camera entity
	cam := scene.AddEntity()
	cam.AddComponent(ecs.NewCamera()) // default at (0,0,3) looking at origin

	// Triangle entity
	// Triangle entity with material
	tri := scene.AddEntity()
	tri.AddComponent(ecs.NewTransform([3]float32{0, 2, 0})) // start above ground
	tri.AddComponent(ecs.NewMesh("triangle"))
	tri.AddComponent(ecs.NewMaterial([4]float32{0.0, 1.0, 0.0, 1.0}))
	tri.AddComponent(ecs.NewRigidBody(0.1)) // mass = 1

	last := glfw.GetTime()
	for !window.ShouldClose() {
		now := glfw.GetTime()
		dt := float32(now - last)
		last = now
		if dt > 0.05 {
			dt = 0.05
		} // clamp to ~20 FPS max step

		last = now

		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		scene.Update(dt)

		window.SwapBuffers()
		glfw.PollEvents()
	}
	glfw.Terminate()
}
