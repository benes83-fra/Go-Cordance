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
	vertexSrc, err := engine.LoadShaderSource("assets/shaders/vertex.glsl")
	if err != nil {
		log.Fatal(err)
	}
	fragmentSrc, err := engine.LoadShaderSource("assets/shaders/fragment.glsl")
	if err != nil {
		log.Fatal(err)
	}
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
	meshMgr.RegisterSphere("sphere", 32, 16) // slices, stacks

	// optionally: meshMgr.RegisterWireSphere("wire_sphere")

	scene := scene.New()

	camSys := ecs.NewCameraSystem(window)
	renderSys := ecs.NewRenderSystem(renderer, meshMgr, camSys)

	camCtrl := ecs.NewCameraControllerSystem(window)
	debugVertexSrc, err := engine.LoadShaderSource("assets/shaders/debug_vertex.glsl")
	if err != nil {
		log.Fatal(err)
	}
	debugFragmentSrc, err := engine.LoadShaderSource("assets/shaders/debug_fragment.glsl")
	if err != nil {
		log.Fatal(err)
	}

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
	sphere.AddComponent(ecs.NewMesh("sphere"))
	sphere.AddComponent(ecs.NewMaterial([4]float32{0.0, 1.0, 0.0, 1.0}))
	sphere.AddComponent(ecs.NewRigidBody(1.0))
	sphere.AddComponent(ecs.NewColliderSphere(0.5))

	// Shiny metal cube
	metalCube := scene.AddEntity()
	metalCube.AddComponent(ecs.NewTransform([3]float32{1.5, 4.0, 0.0}))
	metalCube.AddComponent(ecs.NewMesh("cube"))
	metalCube.AddComponent(&ecs.Material{
		BaseColor: [4]float32{0.8, 0.8, 0.9, 1.0}, // light gray
		Ambient:   0.1,
		Diffuse:   0.6,
		Specular:  1.0,  // strong specular
		Shininess: 64.0, // tight highlight
	})
	metalCube.AddComponent(ecs.NewRigidBody(1.0))
	metalCube.AddComponent(ecs.NewColliderAABB([3]float32{0.5, 0.5, 0.5}))

	// Matte plastic cube
	plasticCube := scene.AddEntity()
	plasticCube.AddComponent(ecs.NewTransform([3]float32{-1.5, 4.0, 0.0}))
	plasticCube.AddComponent(ecs.NewMesh("cube"))
	plasticCube.AddComponent(&ecs.Material{
		BaseColor: [4]float32{0.2, 0.7, 0.2, 1.0}, // green
		Ambient:   0.3,
		Diffuse:   0.8,
		Specular:  0.1, // weak specular
		Shininess: 8.0, // broad, dull highlight
	})
	plasticCube.AddComponent(ecs.NewRigidBody(1.0))
	plasticCube.AddComponent(ecs.NewColliderAABB([3]float32{0.5, 0.5, 0.5}))

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
