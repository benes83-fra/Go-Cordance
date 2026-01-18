package scene

import "go-engine/Go-Cordance/internal/ecs"

// BootstrapScene constructs the logical scene (entities + components) only.
// It does NOT load GPU resources (textures, GLTFs) or initialize renderer/systems.
// It returns the Scene and a map of named entities so the runtime (game) can
// bind runtime-only resources (textures, mesh buffers, GLTF instances) afterwards.
func BootstrapScene() (*Scene, map[string]*ecs.Entity) {
	sc := New()
	named := make(map[string]*ecs.Entity)

	// Camera
	cam := sc.AddEntity()
	cam.AddComponent(ecs.NewCamera())
	cam.AddComponent(ecs.NewName("Main Camera"))
	named["camera"] = cam

	// Ground
	ground := sc.AddEntity()
	ground.AddComponent(&ecs.Transform{
		Position: [3]float32{0, -2, 0},
		Scale:    [3]float32{50, 0.1, 50},
	})
	ground.AddComponent(ecs.NewColliderPlane(-2.0))
	ground.AddComponent(ecs.NewMesh("plane"))
	ground.AddComponent(ecs.NewName("Ground Floor"))
	ground.AddComponent(&ecs.Material{BaseColor: [4]float32{0.8, 0.8, 0.9, 1},
		Ambient:  0.3,
		Diffuse:  0.8,
		Specular: 0.2,
	})
	named["ground"] = ground
	// Cube 1
	cube1 := sc.AddEntity()
	cube1.AddComponent(ecs.NewTransform([3]float32{0.0, 4.0, 0.0}))
	cube1.AddComponent(ecs.NewMesh("cube24"))
	cube1.AddComponent(ecs.NewMaterial([4]float32{1.0, 0.0, 0.0, 1.0}))
	cube1.AddComponent(ecs.NewRigidBody(1.0))
	cube1.AddComponent(ecs.NewColliderAABB([3]float32{0.5, 0.5, 0.5}))
	cube1.AddComponent(ecs.NewName("First Cube"))
	named["cube1"] = cube1

	// Cube 2
	cube2 := sc.AddEntity()
	cube2.AddComponent(ecs.NewTransform([3]float32{0.2, 6.0, 0.0}))
	cube2.AddComponent(ecs.NewMesh("cube24"))
	cube2.AddComponent(ecs.NewMaterial([4]float32{0.8, 1.0, 0.8, 1.0}))
	cube2.AddComponent(ecs.NewRigidBody(1.0))
	cube2.AddComponent(ecs.NewColliderAABB([3]float32{0.5, 0.5, 0.5}))
	cube2.AddComponent(ecs.NewName("Second Cube"))
	named["cube2"] = cube2

	sphere := sc.AddEntity()
	sphere.AddComponent(ecs.NewTransform([3]float32{1.1, 2.5, 2.5}))
	sphere.AddComponent(ecs.NewMesh("sphere"))
	sphere.AddComponent(ecs.NewRigidBody(1.0))
	sphere.AddComponent(ecs.NewMaterial([4]float32{0.8, 1.0, 0.8, 1.0}))
	sphere.AddComponent(ecs.NewColliderSphere(1))
	sphere.AddComponent(ecs.NewName("Sphere1"))
	named["Sphere1"] = sphere

	// Metal cube
	metalCube := sc.AddEntity()
	metalCube.AddComponent(ecs.NewTransform([3]float32{1.5, 4.0, 0.0}))
	metalCube.AddComponent(ecs.NewMesh("cube24"))
	metalCube.AddComponent(&ecs.Material{
		BaseColor: [4]float32{0.8, 0.8, 0.9, 1.0},
		Ambient:   0.05,
		Diffuse:   0.5,
		Specular:  1.0,
		Shininess: 128.0,
	})
	metalCube.AddComponent(ecs.NewRigidBody(1.0))
	metalCube.AddComponent(ecs.NewColliderAABB([3]float32{0.5, 0.5, 0.5}))
	metalCube.AddComponent(ecs.NewName("Metal Cube"))
	named["metalCube"] = metalCube

	// Plastic cube
	plasticCube := sc.AddEntity()
	plasticCube.AddComponent(ecs.NewTransform([3]float32{-1.5, 4.0, 0.0}))
	plasticCube.AddComponent(ecs.NewMesh("cube24"))
	plasticCube.AddComponent(&ecs.Material{
		BaseColor: [4]float32{0.2, 0.7, 0.2, 1.0},
		Ambient:   0.4,
		Diffuse:   0.6,
		Specular:  0.02,
		Shininess: 2.0,
	})
	plasticCube.AddComponent(ecs.NewRigidBody(1.0))
	plasticCube.AddComponent(ecs.NewColliderAABB([3]float32{0.5, 0.5, 0.5}))
	plasticCube.AddComponent(ecs.NewName("Plastic Cube"))
	named["plasticCube"] = plasticCube

	// Teapot (mesh reference only)
	teapot := sc.AddEntity()
	teapot.AddComponent(ecs.NewTransform([3]float32{0, 2, 0}))
	teapot.AddComponent(ecs.NewMesh("teapot"))
	teapot.AddComponent(&ecs.Material{
		BaseColor: [4]float32{1, 1, 1, 1.0},
		Ambient:   0.4,
		Diffuse:   0.6,
		Specular:  0.02,
		Shininess: 2.0,
	})
	teapot.AddComponent(ecs.NewName("Tea Pot"))
	named["teapot"] = teapot

	// Generic entity with material (example)
	ent := sc.AddEntity()
	ent.AddComponent(ecs.NewMesh("Frame/0"))
	ent.AddComponent(ecs.NewTransform([3]float32{1, 1, 1}))
	ent.AddComponent(ecs.NewMaterial([4]float32{1, 1, 1, 1}))
	ent.AddComponent(ecs.NewName("Sofa Entity"))
	named["ent"] = ent

	// Light gizmo and arrow (for debug / gizmos)
	lightGizmo := sc.AddEntity()
	lightGizmo.AddComponent(ecs.NewTransform([3]float32{5, 5, 0}))
	lightGizmo.AddComponent(ecs.NewMesh("sphere"))
	lightGizmo.AddComponent(ecs.NewName("Light Gizmo"))
	named["lightGizmo"] = lightGizmo

	lightArrow := sc.AddEntity()
	lightArrow.AddComponent(ecs.NewTransform([3]float32{0, 0, 0}))
	lightArrow.AddComponent(ecs.NewMesh("line"))
	lightArrow.AddComponent(ecs.NewName("Light Arrow"))
	named["lightArrow"] = lightArrow

	bill := sc.AddEntity()
	bill.AddComponent(ecs.NewTransform([3]float32{0, 3, 0}))
	bill.AddComponent(ecs.NewMesh("billboardQuad"))
	bill.AddComponent(ecs.NewBillboard())
	bill.AddComponent(ecs.NewMaterial([4]float32{1, 1, 1, 1}))
	bill.AddComponent(ecs.NewName("Billboard"))
	named["Billboard"] = bill
	return sc, named
}
