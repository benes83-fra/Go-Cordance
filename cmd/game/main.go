package main

import (
	"fmt"
	"log"
	"os"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"

	"go-engine/Go-Cordance/internal/ecs"
	"go-engine/Go-Cordance/internal/ecs/gizmo"
	"go-engine/Go-Cordance/internal/ecs/gizmo/bridge"
	"go-engine/Go-Cordance/internal/editor/state"
	"go-engine/Go-Cordance/internal/editorlink"
	"go-engine/Go-Cordance/internal/engine"
	"go-engine/Go-Cordance/internal/scene"
)

const (
	width  = 800
	height = 600
)

func main() {
	// Initialize window / GL context (game runtime only)
	window, err := engine.InitGLFW(width, height, "Go Cordance")
	if err != nil {
		log.Fatal(err)
	}

	// Compile shaders and create renderer (runtime)
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

	// load shadow shader sources
	shadowVertSrc, err := engine.LoadShaderSource("assets/shaders/shadow_vertex.glsl")
	if err != nil {
		log.Fatal(err)
	}
	shadowFragSrc, err := engine.LoadShaderSource("assets/shaders/shadow_fragment.glsl")
	if err != nil {
		log.Fatal(err)
	}

	// initialize shadow resources (choose resolution)
	shadowW, shadowH := 2048, 2048
	renderer.InitShadow(shadowVertSrc, shadowFragSrc, shadowW, shadowH)

	// Resize callback updates viewport
	window.SetFramebufferSizeCallback(func(_ *glfw.Window, w, h int) {
		if h == 0 {
			h = 1
		}
		gl.Viewport(0, 0, int32(w), int32(h))
	})

	// Mesh manager and registrations (runtime)
	meshMgr := engine.NewMeshManager()
	meshMgr.RegisterTriangle("triangle")
	meshMgr.RegisterCube8("Cube8")
	meshMgr.RegisterCube("cube")
	meshMgr.RegisterPlane("plane")
	meshMgr.RegisterCube("cube24")
	meshMgr.RegisterWireCube("wire_cube")
	meshMgr.RegisterWireSphere("wire_sphere", 16, 16)
	meshMgr.RegisterSphere("sphere", 32, 16)
	meshMgr.RegisterLine("line")
	meshMgr.RegisterGizmoArrow("gizmo_arrow")
	meshMgr.RegisterGizmoPlane("gizmo_plane")
	meshMgr.RegisterGizmoCircle("gizmo_circle", 64)
	meshMgr.RegisterBillboardQuad("billboardQuad")

	// Load GLTF meshes that require runtime resources
	if err := meshMgr.RegisterGLTF("teapot", "assets/models/teapot/teapot.gltf"); err != nil {
		log.Fatal("Failed to load glTF:", err)
	}
	if err := meshMgr.RegisterGLTFMulti("assets/models/sofa/sofa.gltf"); err != nil {
		log.Fatal(err)
	}

	// Load shader sources for debug renderer
	debugVertexSrc, err := engine.LoadShaderSource("assets/shaders/debug_vertex.glsl")
	if err != nil {
		log.Fatal(err)
	}
	debugFragmentSrc, err := engine.LoadShaderSource("assets/shaders/debug_fragment.glsl")
	if err != nil {
		log.Fatal(err)
	}

	// Load textures (runtime GPU resources)
	texID, err := engine.LoadTexture("assets/textures/crate.png")
	if err != nil {
		log.Fatal(err)
	}
	ecs.RegisterTexture("Crate", texID)
	texID2, err := engine.LoadTexture("assets/textures/teapot_diffuse.png")
	if err != nil {
		fmt.Println("Could not load :", texID2)
		log.Fatal(err)
	}
	ecs.RegisterTexture("Teapot", texID2)
	// Load GLTF materials info (runtime)
	mats, err := engine.LoadGLTFMaterials("sofa", "assets/models/sofa/sofa.gltf")
	if err != nil {
		log.Fatal(err)
	}
	matInfo := mats[0]

	// Create runtime wrappers for textures (ecs.Texture holds GPU id)
	crateTex := ecs.NewTexture(texID)
	teaTex := ecs.NewTexture(texID2)

	// Create renderers / debug systems that require runtime resources
	debugRenderer := engine.NewDebugRenderer(debugVertexSrc, debugFragmentSrc)
	debugSys := ecs.NewDebugRenderSystem(debugRenderer, meshMgr, nil) // camSys set later
	lightDebug := ecs.NewLightDebugRenderSystem(debugRenderer, meshMgr, nil)
	gizmoSys := gizmo.NewGizmoRenderSystem(debugRenderer, meshMgr, nil)
	// later, after camera system exists, call gizmoSys.SetCameraSystem(camSys)

	lightDebug.Enabled = true

	// Build the logical scene (entities + components) only.
	// BootstrapScene returns the Scene and a map of named entities so we can
	// bind runtime-only resources (textures, set LightEntity, etc).
	sc, named := scene.BootstrapScene()
	gizmoSys.SetWorld(sc.World())
	gizmo.RegisterGlobalGizmo(gizmoSys)
	// Create runtime systems that need the window/renderer/meshMgr
	camSys := ecs.NewCameraSystem(window)
	renderSys := ecs.NewRenderSystem(renderer, meshMgr, camSys)
	camCtrl := ecs.NewCameraControllerSystem(window)

	// Now that we have camSys, set it on debug systems that need it
	debugSys.SetCameraSystem(camSys)
	lightDebug.SetCameraSystem(camSys)
	gizmoSys.SetCameraSystem(camSys)

	// Register systems on the scene
	sc.Systems().AddSystem(ecs.NewForceSystem(0, -9.8, 0))
	sc.Systems().AddSystem(ecs.NewPhysicsSystem())
	sc.Systems().AddSystem(ecs.NewCollisionSystem())
	sc.Systems().AddSystem(ecs.NewTransformSystem())

	sc.Systems().AddSystem(camCtrl)
	sc.Systems().AddSystem(camSys)
	sc.Systems().AddSystem(renderSys)
	sc.Systems().AddSystem(debugSys)
	sc.Systems().AddSystem(lightDebug)
	cursorDisabled := false

	window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Press {
			switch key {
			case glfw.KeyLeft:
				renderSys.LightDir[0] -= 0.1
			case glfw.KeyRight:
				renderSys.LightDir[0] += 0.1
			case glfw.KeyUp:
				renderSys.LightDir[1] += 0.1
			case glfw.KeyDown:
				renderSys.LightDir[1] -= 0.1
			case glfw.KeySpace:
				renderSys.OrbitalEnabled = !renderSys.OrbitalEnabled
				log.Printf("Light orbit: %v", renderSys.OrbitalEnabled)
			case glfw.KeyF1:
				debugSys.Enabled = !debugSys.Enabled
				log.Printf("Debug rendering: %v", debugSys.Enabled)
			case glfw.KeyF2:
				lightDebug.Enabled = !lightDebug.Enabled
				log.Printf("Light Debug rendering: %v", lightDebug.Enabled)
			case glfw.KeyEscape:
				os.Exit(0)
				//further debug options
			case glfw.Key1:
				renderSys.DebugShowMode = 0 // final
			case glfw.Key2:
				renderSys.DebugShowMode = 1 // normal map raw
			case glfw.Key3:
				renderSys.DebugShowMode = 2 // tangent
			case glfw.Key4:
				renderSys.DebugShowMode = 3 // bitangent
			case glfw.Key5:
				renderSys.DebugShowMode = 4 // normal
			case glfw.Key6:
				renderSys.DebugShowMode = 5 // tangentW
			case glfw.Key7:
				renderSys.DebugShowMode = 6 // uv
			case glfw.KeyG:
				renderSys.DebugFlipGreen = !renderSys.DebugFlipGreen
			case glfw.KeyTab:
				cursorDisabled = !cursorDisabled
				if cursorDisabled {
					window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
					log.Println("Cursor disabled (camera mode)")
				} else {
					window.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
					log.Println("Cursor normal (editor mode)")
				}
			case glfw.KeyL:
				gizmoSys.LocalRotation = !gizmoSys.LocalRotation
				fmt.Println("Local rotation:", gizmoSys.LocalRotation)

			case glfw.KeyW:
				if !cursorDisabled {
					gizmoSys.Mode = gizmo.GizmoMove
					fmt.Println("Gizmo mode: Move")
				}
			case glfw.KeyE:
				if !cursorDisabled {
					gizmoSys.Mode = gizmo.GizmoRotate
					fmt.Println("Gizmo mode: Rotate")
				}
			case glfw.KeyR:
				if !cursorDisabled {
					gizmoSys.Mode = gizmo.GizmoScale
					fmt.Println("Gizmo mode: Scale")
				}
			case glfw.KeyQ:
				if !cursorDisabled {
					gizmoSys.Mode = gizmo.GizmoCombined
					fmt.Println("Gizmo mode: Combined")
				}
			case glfw.KeyP:
				if gizmoSys.PivotMode == state.PivotModePivot {
					gizmoSys.SetPivotMode(state.PivotModeCenter)

				} else {
					gizmoSys.SetPivotMode(state.PivotModePivot)

				}
			case glfw.KeyZ:
				log.Printf("Undo Gizmo action")
				gizmoSys.Undo.Undo(sc.World())
			case glfw.KeyY:
				log.Printf("Redo Gizmo action")
				gizmoSys.Undo.Redo(sc.World())
			}

		}
	})
	// Bind runtime-only resources to entities created by the bootstrap.
	// We look up entities by name in the map returned by BootstrapScene.
	if e, ok := named["cube1"]; ok {
		//e.AddComponent(crateTex)
		mat := e.GetComponent((*ecs.Material)(nil)).(*ecs.Material)
		mat.UseTexture = true
		mat.TextureID = crateTex.ID

	}
	if e, ok := named["cube2"]; ok {
		//	e.AddComponent(teaTex)
		mat := e.GetComponent((*ecs.Material)(nil)).(*ecs.Material)
		mat.UseTexture = true
		mat.TextureID = teaTex.ID

	}
	if _, ok := named["metalCube"]; ok {
		// metalCube used a material already in bootstrap; optionally add textures
		if matInfo.DiffuseTexturePath != "" {
			// load and attach diffuse texture if desired (example)
			// texID3, _ := engine.LoadTexture(matInfo.DiffuseTexturePath)
			// e.AddComponent(ecs.NewDiffuseTexture(texID3))
		}
	}
	// Attach textures to teapot if present
	if e, ok := named["teapot"]; ok {
		mat := e.GetComponent((*ecs.Material)(nil)).(*ecs.Material)
		mat.UseTexture = true
		mat.TextureID = teaTex.ID

		// optionally add normal map later if available
	}

	// Set render system light entity and light debug tracking if present
	if light, ok := named["lightGizmo"]; ok {
		renderSys.LightEntity = light
		lightDebug.Track(light)
		lightDebug.SetColor(light, [4]float32{1.0, 1.0, 0.2, 1.0})
	}
	if arrow, ok := named["lightArrow"]; ok {
		lightDebug.Track(arrow)
		lightDebug.SetColor(arrow, [4]float32{1.0, 0.5, 0.0, 1.0})
		renderSys.LightArrow = arrow
	}
	// Force select cube1 for debugging (do this once after named map is available)
	var selected *ecs.Entity
	selected = sc.Selected
	fmt.Printf("Initial selected entity: %v\n", selected)
	vao := meshMgr.GetVAO("gizmo_arrow")
	count := meshMgr.GetCount("gizmo_arrow")
	log.Printf("gizmo VAO=%d count=%d", vao, count)

	// Optionally save the scene (pure data) to disk
	sc.Save("my_scene.json")
	go editorlink.StartServer(":7777", sc)
	bridge.SendTransformToEditor = func(
		id int64,
		pos [3]float32,
		rot [4]float32,
		scale [3]float32,
	) {
		if editorlink.EditorConn != nil {
			go editorlink.WriteTransformFromGame(
				editorlink.EditorConn,
				int64(id),
				pos,
				rot,
				scale,
			)
		}
	}
	bridge.SendTransformToEditorFinal = func(id int64, pos [3]float32, rot [4]float32, scale [3]float32) {
		if editorlink.EditorConn != nil {
			msg := editorlink.MsgSetTransform{
				ID:       uint64(id),
				Position: pos,
				Rotation: rot,
				Scale:    scale,
			}
			go editorlink.WriteSetTransformFinal(editorlink.EditorConn, msg)
		}
	}

	// Main loop
	last := glfw.GetTime()
	for !window.ShouldClose() {
		now := glfw.GetTime()
		dt := float32(now - last)
		last = now
		if dt > 0.05 {
			dt = 0.05
		}

		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		// determine selected entity pointer as you already do for other editor features

		// ... set selected appropriately ...

		sc.Update(dt)

		// debug: draw gizmo on top (disable depth to rule out occlusion)
		gl.Disable(gl.DEPTH_TEST)
		selected = sc.Selected
		gizmoSys.Update(dt, sc.Entities(), selected)
		gl.Enable(gl.DEPTH_TEST)
		err := gl.GetError()
		if err != gl.NO_ERROR {
			log.Printf("GL error after gizmo draw: 0x%X", err)
		}

		// Swap buffers / poll events
		window.SwapBuffers()
		engine.PollEvents()
	}

	// Cleanup
	meshMgr.Delete()
	engine.TerminateGLFW()
}
