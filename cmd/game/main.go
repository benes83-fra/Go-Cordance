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
	meshMgr.RegisterCube("cube")
	meshMgr.RegisterWireCube("wire_cube")
	meshMgr.RegisterWireSphere("wire_sphere", 16, 16)
	// optionally: meshMgr.RegisterWireSphere("wire_sphere")

	scene := scene.New()

	camSys := ecs.NewCameraSystem(window)
	renderSys := ecs.NewRenderSystem(renderer, meshMgr, camSys)

	camCtrl := ecs.NewCameraControllerSystem(window)
	debugVertexSrc := `#version 330 core
						layout(location = 0) in vec3 position;
						uniform mat4 model;
						uniform mat4 view;
						uniform mat4 projection;
						void main() {
							gl_Position = projection * view * model * vec4(position, 1.0);
						}`

	debugFragmentSrc := `#version 330 core
						out vec4 FragColor;
						uniform vec4 debugColor;
						void main() {
							FragColor = debugColor;
						}`

	debugRenderer := engine.NewDebugRenderer(debugVertexSrc, debugFragmentSrc)
	debugSys := ecs.NewDebugRenderSystem(debugRenderer, meshMgr, camSys)
	scene.Systems().AddSystem(debugSys)

	scene.Systems().AddSystem(ecs.NewForceSystem(0, -9.8, 0)) // gravity
	scene.Systems().AddSystem(ecs.NewPhysicsSystem())
	scene.Systems().AddSystem(ecs.NewCollisionSystem())
	scene.Systems().AddSystem(camCtrl) // updates Camera component
	scene.Systems().AddSystem(camSys)  // computes view/projection from Camera
	scene.Systems().AddSystem(renderSys)
	scene.Systems().AddSystem(debugSys)

	window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		if key == glfw.KeyF1 && action == glfw.Press {
			debugSys.Enabled = !debugSys.Enabled
			log.Printf("Debug rendering: %v", debugSys.Enabled)
		}
	})

	// Camera entity
	cam := scene.AddEntity()
	cam.AddComponent(ecs.NewCamera()) // default at (0,0,3) looking at origin

	// Ground entity
	ground := scene.AddEntity()
	ground.AddComponent(ecs.NewTransform([3]float32{0, 0, 0}))
	ground.AddComponent(ecs.NewColliderPlane(-1.0)) // y=0 plane

	// Triangle entity
	// Triangle entity with material
	tri := scene.AddEntity()
	tri.AddComponent(ecs.NewTransform([3]float32{0, 2, 0}))
	tri.AddComponent(ecs.NewMesh("triangle"))
	tri.AddComponent(ecs.NewMaterial([4]float32{0.0, 1.0, 0.0, 1.0}))
	tri.AddComponent(ecs.NewRigidBody(1.0))
	tri.AddComponent(ecs.NewColliderSphere(0.5)) // simple bounding sphere

	tri2 := scene.AddEntity()
	tri2.AddComponent(ecs.NewTransform([3]float32{0.5, 3, 0}))
	tri2.AddComponent(ecs.NewMesh("triangle"))
	tri2.AddComponent(ecs.NewMaterial([4]float32{0, 0, 1, 1}))
	tri2.AddComponent(ecs.NewRigidBody(1.0))
	tri2.AddComponent(ecs.NewColliderSphere(0.5))

	cube1 := scene.AddEntity()
	cube1.AddComponent(ecs.NewTransform([3]float32{0.0, 4.0, 0.0}))
	cube1.AddComponent(ecs.NewMesh("cube"))
	cube1.AddComponent(ecs.NewMaterial([4]float32{1.0, 0.0, 0.0, 1.0}))
	cube1.AddComponent(ecs.NewRigidBody(1.0))
	cube1.AddComponent(ecs.NewColliderAABB([3]float32{0.5, 0.5, 0.5}))

	cube2 := scene.AddEntity()
	cube2.AddComponent(ecs.NewTransform([3]float32{0.2, 6.0, 0.0}))
	cube2.AddComponent(ecs.NewMesh("cube"))
	cube2.AddComponent(ecs.NewMaterial([4]float32{0.0, 1.0, 0.0, 1.0}))
	cube2.AddComponent(ecs.NewRigidBody(1.0))
	cube2.AddComponent(ecs.NewColliderAABB([3]float32{0.5, 0.5, 0.5}))

	sphere := scene.AddEntity()
	sphere.AddComponent(ecs.NewTransform([3]float32{0.0, 4.0, 0.0}))
	sphere.AddComponent(ecs.NewMesh("triangle")) // still using triangle mesh for sphere placeholder
	sphere.AddComponent(ecs.NewMaterial([4]float32{0.0, 1.0, 0.0, 1.0}))
	sphere.AddComponent(ecs.NewRigidBody(1.0))
	sphere.AddComponent(ecs.NewColliderSphere(0.5))

	// Sphere–AABB collisions
	last := glfw.GetTime()
	for !window.ShouldClose() {
		now := glfw.GetTime()
		dt := float32(now - last)
		last = now
		if dt > 0.05 {
			dt = 0.05
		} // clamp to ~20 FPS max step

		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		scene.Update(dt)

		window.SwapBuffers()
		engine.PollEvents()
	}
	engine.TerminateGLFW()
}
