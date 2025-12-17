package main

import (
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"go-engine/Go-Cordance/internal/ecs"
	"go-engine/Go-Cordance/internal/engine"
	"go-engine/Go-Cordance/internal/scene"
)

const (
	width  = 800
	height = 600
)

func init() {
	// OpenGL/GLFW require the main OS thread
	runtime.LockOSThread()
}

func main() {
	// Basic logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Graceful shutdown on SIGINT/SIGTERM
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Initialize GLFW + GL
	window, err := engine.InitGLFW(width, height, "Go3D Prototype")
	if err != nil {
		log.Fatalf("InitGLFW: %v", err)
	}
	defer func() {
		window.Destroy()
		// engine.InitGLFW calls glfw.Init; caller must call Terminate
		// but keep termination centralized
		engine.TerminateGLFW()
	}()

	// Load shaders in background to avoid blocking startup
	shaderCh := make(chan struct {
		vert string
		frag string
		err  error
	}, 1)
	go func() {
		vert, err1 := engine.LoadShaderSource("assets/shaders/triangle.vert")
		frag, err2 := engine.LoadShaderSource("assets/shaders/triangle.frag")
		if err1 != nil {
			shaderCh <- struct {
				vert string
				frag string
				err  error
			}{"", "", err1}
			return
		}
		if err2 != nil {
			shaderCh <- struct {
				vert string
				frag string
				err  error
			}{"", "", err2}
			return
		}
		shaderCh <- struct {
			vert string
			frag string
			err  error
		}{vert, frag, nil}
	}()
	meshMgr := engine.NewMeshManager()
	meshMgr.RegisterTriangle("triangle")
	// Create renderer and scene
	renderer := engine.NewRenderer()
	scene := scene.New()

	camSys := ecs.NewCameraSystem()
	scene.Systems().AddSystem(camSys)
	scene.Systems().AddSystem(ecs.NewRenderSystem(renderer, meshMgr))
	scene.Systems().AddSystem(ecs.NewForceSystem(0, -9.8, 0))
	scene.Systems().AddSystem(ecs.NewPhysicsSystem())
	scene.Systems().AddSystem(ecs.NewCollisionSystem())
	scene.Systems().AddSystem(ecs.NewTorqueSystem(0, 1.0, 0)) // torque around Y

	// Add a camera entity
	camEntity := scene.AddEntity()
	camEntity.AddComponent(ecs.NewCamera())

	e := scene.AddEntity()
	t := ecs.NewTransform()
	rb := ecs.NewRigidBody(1.0)
	av := ecs.NewAngularVelocity(0, 0, 0)
	am := ecs.NewAngularMass(1.0, 2.0, 3.0) // inertia values
	ad := ecs.NewAngularDamping(0.98)
	e.AddComponent(ecs.NewMesh("triangle"))
	e.AddComponent(ecs.NewMaterial("basicShader", "none", [4]float32{1, 1, 1, 1}))
	e.AddComponent(t)
	e.AddComponent(rb)
	e.AddComponent(av)
	e.AddComponent(am)
	e.AddComponent(ad)

	e.AddComponent(&ecs.Renderable{MeshID: "triangle", MaterialID: "basic"})

	// Wait for shaders (or timeout)
	go func() {
		for {
			rb.ApplyForce(0, -9.8, 0) // gravity along Y
			time.Sleep(time.Second / 60)
		}
	}()
	select {
	case s := <-shaderCh:
		if s.err != nil {
			log.Fatalf("shader load: %v", s.err)
		}
		if err := renderer.Init(s.vert, s.frag); err != nil {
			log.Fatalf("renderer init: %v", err)
		}
	case <-time.After(5 * time.Second):
		log.Fatal("timed out loading shaders")
	}
	defer renderer.Shutdown()

	// Main loop: fixed timestep update + render
	const targetFPS = 60
	frameDur := time.Second / time.Duration(targetFPS)
	ticker := time.NewTicker(frameDur)
	defer ticker.Stop()

	//last := time.Now()
loop:
	for {
		select {
		case <-stop:
			break loop
		case <-ticker.C:
			// Poll events via engine wrapper
			engine.PollEvents()

			// Fixed timestep update
			scene.Update(float32(1.0 / float32(targetFPS)))

			// Render
			renderer.Draw()

			// Swap buffers
			window.SwapBuffers()
		}
	}

	// final cleanup happens via deferred calls
}
