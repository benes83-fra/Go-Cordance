package scene

import (
	"encoding/json"
	"go-engine/Go-Cordance/internal/assets"
	"go-engine/Go-Cordance/internal/ecs"
	"go-engine/Go-Cordance/internal/engine"
	"os"
)

type SerializedScene struct {
	Entities []SerializedEntity `json:"entities"`
}

type SerializedEntity struct {
	ID         int64                  `json:"id"`
	ParentID   int64                  `json:"parent,omitempty"`
	Components map[string]interface{} `json:"components"`
}

func serializeTransform(t *ecs.Transform) map[string]interface{} {
	return map[string]interface{}{"position": t.Position, "rotation": t.Rotation, "scale": t.Scale}
}
func serializeMesh(m *ecs.Mesh) map[string]interface{} { return map[string]interface{}{"id": m.ID} }

func serializeMaterial(m *ecs.Material) map[string]interface{} {
	out := map[string]interface{}{
		"baseColor": m.BaseColor,
		"ambient":   m.Ambient,
		"diffuse":   m.Diffuse,
		"specular":  m.Specular,
		"shininess": m.Shininess,
		"metallic":  m.Metallic,
		"roughness": m.Roughness,
	}
	// optional fields
	out["type"] = m.Type
	out["useTexture"] = m.UseTexture
	out["textureID"] = m.TextureID
	out["useNormal"] = m.UseNormal
	out["normalID"] = m.NormalID
	out["textureAsset"] = m.TextureAsset
	out["normalAsset"] = m.NormalAsset
	out["shaderName"] = m.ShaderName
	out["shader"] = m.Shader
	out["diffuseTexturePath"] = m.DiffuseTexturePath
	out["normalTexturePath"] = m.NormalTexturePath
	out["occlusionTexturePath"] = m.OcclusionTexturePath
	out["metallicRoughnessTexturePath"] = m.MetallicRoughnessTexturePath

	out["texCoordMap"] = m.TexCoordMap
	out["uvScale"] = m.UVScale
	out["uvOffset"] = m.UVOffset

	out["normalScale"] = m.NormalScale
	out["sheenColor"] = m.SheenColor
	out["sheenRoughness"] = m.SheenRoughness
	out["specularFactor"] = m.SpecularFactor
	out["occlusionAsset"] = m.OcclusionAsset
	out["occlusionID"] = m.OcclusionID

	out["metallicRoughnessAsset"] = m.MetallicRoughnessAsset
	out["metallicRoughnessID"] = m.MetallicRoughnessID
	out["useIBL"] = m.UseIBL
	out["irradianceTex"] = m.IrradianceTex
	out["prefilteredEnvTex"] = m.PrefilteredEnvTex
	out["bRDFLUTTex"] = m.BRDFLUTTex
	out["clearcoatFactor"] = m.ClearcoatFactor
	out["clearcoatRoughness"] = m.ClearcoatRoughTex
	out["clearcoatTexture"] = m.ClearcoatTexture
	out["clearcoatRoughTex"] = m.ClearcoatRoughTex
	out["clearcoatNormalTex"] = m.ClearcoatNormalTex
	out["useClearcoat"] = m.UseClearcoat
	out["transmissionFactor"] = m.TransmissionFactor
	out["useTransmission"] = m.UseTransmission
	out["transMissionTex"] = m.TransmissionTex
	out["dirty"] = m.Dirty

	return out
}

func serializeDiffuseTexture(t *ecs.DiffuseTexture) map[string]interface{} {
	return map[string]interface{}{"id": t.ID}
}
func serializeNormalMap(t *ecs.NormalMap) map[string]interface{} {
	return map[string]interface{}{"id": t.ID}
}

func serializeParent(p *ecs.Parent) map[string]interface{} {
	return map[string]interface{}{
		"parent": p.Entity.ID,
	}
}

func (s *Scene) Save(path string) error {
	out := SerializedScene{
		Entities: make([]SerializedEntity, 0, len(s.entities)),
	}

	for _, e := range s.entities {
		se := SerializedEntity{
			ID:         e.ID,
			Components: make(map[string]interface{}),
		}

		// Parent
		if p := e.GetComponent((*ecs.Parent)(nil)); p != nil {
			se.ParentID = p.(*ecs.Parent).Entity.ID
		}

		// Transform
		if t := e.GetTransform(); t != nil {
			se.Components["Transform"] = serializeTransform(t)
		}

		// Mesh
		if m := e.GetComponent((*ecs.Mesh)(nil)); m != nil {
			se.Components["Mesh"] = serializeMesh(m.(*ecs.Mesh))
		}

		// Material
		if m := e.GetComponent((*ecs.Material)(nil)); m != nil {
			se.Components["Material"] = serializeMaterial(m.(*ecs.Material))
		}
		if mm := e.GetComponent((*ecs.MultiMesh)(nil)); mm != nil {
			se.Components["MultiMesh"] = map[string]any{
				"meshes": mm.(*ecs.MultiMesh).Meshes,
			}
		}

		if rb := e.GetComponent((*ecs.RigidBody)(nil)); rb != nil {
			se.Components["RigidBody"] = map[string]any{
				"Mass":  rb.(*ecs.RigidBody).Mass,
				"Vel":   rb.(*ecs.RigidBody).Vel,
				"Force": rb.(*ecs.RigidBody).Force,
			}
		}

		if c := e.GetComponent((*ecs.ColliderSphere)(nil)); c != nil {
			cs := c.(*ecs.ColliderSphere)
			se.Components["ColliderSphere"] = map[string]any{
				"radius":      cs.Radius,
				"layer":       cs.Layer,
				"mask":        cs.Mask,
				"restitution": cs.Restitution,
				"friction":    cs.Friction,
			}
		}

		if c := e.GetComponent((*ecs.ColliderAABB)(nil)); c != nil {
			ca := c.(*ecs.ColliderAABB)
			se.Components["ColliderAABB"] = map[string]any{
				"halfExtents": ca.HalfExtents,
				"layer":       ca.Layer,
				"mask":        ca.Mask,
				"restitution": ca.Restitution,
				"friction":    ca.Friction,
			}
		}

		if c := e.GetComponent((*ecs.ColliderPlane)(nil)); c != nil {
			cp := c.(*ecs.ColliderPlane)
			se.Components["ColliderPlane"] = map[string]any{
				"y":           cp.Y,
				"layer":       cp.Layer,
				"mask":        cp.Mask,
				"restitution": cp.Restitution,
				"friction":    cp.Friction,
			}
		}

		// DiffuseTexture
		if dt := e.GetComponent((*ecs.DiffuseTexture)(nil)); dt != nil {
			se.Components["DiffuseTexture"] = serializeDiffuseTexture(dt.(*ecs.DiffuseTexture))
		}

		// NormalMap
		if nm := e.GetComponent((*ecs.NormalMap)(nil)); nm != nil {
			se.Components["NormalMap"] = serializeNormalMap(nm.(*ecs.NormalMap))
		}
		if n := e.GetComponent((*ecs.Name)(nil)); n != nil {
			se.Components["Name"] = map[string]interface{}{
				"value": n.(*ecs.Name).Value,
			}
		}
		// example Camera component
		// example Camera component
		if c := e.GetComponent((*ecs.Camera)(nil)); c != nil {
			cam := c.(*ecs.Camera)
			se.Components["Camera"] = map[string]interface{}{
				"position": cam.Position,
				"target":   cam.Target,
				"up":       cam.Up,
				"fov":      cam.Fov,
				"near":     cam.Near,
				"far":      cam.Far,
				"aspect":   cam.Aspect,
				"active":   cam.Active,
			}
		}

		out.Entities = append(out.Entities, se)
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func Load(path string) (*Scene, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	// at top of Load, before second pass:
	var firstCam *ecs.Camera
	var activeCam *ecs.Camera

	var ss SerializedScene
	if err := json.Unmarshal(data, &ss); err != nil {
		return nil, err
	}

	scene := &Scene{
		entities: make([]*ecs.Entity, 0, 16),
		world:    ecs.NewWorld(),
		nextID:   1,
		sysMgr:   ecs.NewSystemManager(),
	}
	scene.sysMgr.AddSystem(ecs.NewTransformSystem())

	entityByID := make(map[int64]*ecs.Entity)

	// First pass: create entities
	for _, se := range ss.Entities {
		e := ecs.NewEntity(se.ID)
		scene.AddExisting(e)
		entityByID[se.ID] = e
	}

	// Second pass: add components
	for _, se := range ss.Entities {
		e := entityByID[se.ID]

		for name, raw := range se.Components {
			switch name {
			case "Transform":
				var t struct {
					Position [3]float32
					Rotation [4]float32
					Scale    [3]float32
				}
				b, _ := json.Marshal(raw)
				json.Unmarshal(b, &t)

				tr := ecs.NewTransform(t.Position)
				tr.Rotation = t.Rotation
				tr.Scale = t.Scale
				tr.RecalculateLocal()
				e.AddComponent(tr)

			case "RigidBody":
				var rb struct {
					Mass  float32
					Vel   [3]float32
					Force [3]float32
				}
				b, _ := json.Marshal(raw)
				json.Unmarshal(b, &rb)
				rbody := ecs.NewRigidBody(rb.Mass)
				rbody.Vel = rb.Vel
				rbody.Force = rb.Force
				e.AddComponent(rbody)

			case "ColliderSphere":
				var c struct {
					Radius      float32
					Layer       int
					Mask        uint32
					Restitution float32
					Friction    float32
				}
				b, _ := json.Marshal(raw)
				json.Unmarshal(b, &c)
				col := ecs.NewColliderSphere(c.Radius)
				col.Layer = c.Layer
				col.Mask = c.Mask
				col.Restitution = c.Restitution
				col.Friction = c.Friction
				e.AddComponent(col)

			case "ColliderAABB":
				var c struct {
					HalfExtents [3]float32
					Layer       int
					Mask        uint32
					Restitution float32
					Friction    float32
				}
				b, _ := json.Marshal(raw)
				json.Unmarshal(b, &c)
				col := ecs.NewColliderAABB(c.HalfExtents)
				col.Layer = c.Layer
				col.Mask = c.Mask
				col.Restitution = c.Restitution
				col.Friction = c.Friction
				e.AddComponent(col)

			case "ColliderPlane":
				var c struct {
					Y           float32
					Layer       int
					Mask        uint32
					Restitution float32
					Friction    float32
				}
				b, _ := json.Marshal(raw)
				json.Unmarshal(b, &c)
				col := ecs.NewColliderPlane(c.Y)
				col.Layer = c.Layer
				col.Mask = c.Mask
				col.Restitution = c.Restitution
				col.Friction = c.Friction
				e.AddComponent(col)

			case "Camera":
				var c struct {
					Position [3]float32
					Target   [3]float32
					Up       [3]float32
					Fov      float32
					Near     float32
					Far      float32
					Aspect   float32
					Active   bool
				}
				b, _ := json.Marshal(raw)
				json.Unmarshal(b, &c)

				cam := ecs.NewCamera()
				cam.Position = c.Position
				cam.Target = c.Target
				cam.Up = c.Up
				cam.Fov = c.Fov
				cam.Near = c.Near
				cam.Far = c.Far
				cam.Aspect = c.Aspect
				cam.Active = c.Active

				e.AddComponent(cam)
				if firstCam == nil {
					firstCam = cam
				}
				if cam.Active {
					activeCam = cam
				}
			case "MultiMesh":
				var mm struct{ Meshes []string }
				b, _ := json.Marshal(raw)
				json.Unmarshal(b, &mm)
				e.AddComponent(ecs.NewMultiMesh(mm.Meshes))

			case "Mesh":
				var m struct{ ID string }
				b, _ := json.Marshal(raw)
				json.Unmarshal(b, &m)
				e.AddComponent(ecs.NewMesh(m.ID))
			case "Name":
				var n struct{ Value string }
				b, _ := json.Marshal(raw)
				json.Unmarshal(b, &n)
				e.AddComponent(ecs.NewName(n.Value))

				// in Load, inside case "Material":
			case "Material":
				var m struct {
					BaseColor [4]float32 // RGBA
					Ambient   float32
					Diffuse   float32
					Specular  float32
					Shininess float32
					Metallic  float32
					Roughness float32

					Type int

					// --- Existing inspector workflow (kept intact) ---
					UseTexture bool
					TextureID  uint32 // raw GL texture ID (inspector uses this)
					UseNormal  bool
					NormalID   uint32 // raw GL normal map ID

					// --- New asset pipeline fields (optional, non-breaking) ---
					TextureAsset assets.AssetID // future: replace TextureID
					NormalAsset  assets.AssetID // future: replace NormalID
					ShaderName   string
					Shader       *engine.ShaderProgram

					DiffuseTexturePath           string
					NormalTexturePath            string
					OcclusionTexturePath         string
					MetallicRoughnessTexturePath string

					TexCoordMap map[string]int        // e.g. "baseColor":0, "occlusion":1
					UVScale     map[string][2]float32 // per-texture uv scale
					UVOffset    map[string][2]float32 // per-texture uv offset

					NormalScale    float32
					SheenColor     [3]float32
					SheenRoughness float32
					SpecularFactor float32
					OcclusionAsset assets.AssetID
					OcclusionID    uint32

					MetallicRoughnessAsset assets.AssetID
					MetallicRoughnessID    uint32
					UseIBL                 bool
					IrradianceTex          uint32
					PrefilteredEnvTex      uint32
					BRDFLUTTex             uint32
					ClearcoatFactor        float32
					ClearcoatRoughness     float32
					ClearcoatTexture       uint32
					ClearcoatRoughTex      uint32
					ClearcoatNormalTex     uint32
					UseClearcoat           bool
					TransmissionFactor     float32
					UseTransmission        bool
					TransmissionTex        uint32 // optional

					// (optional) OcclusionID/MetallicRoughnessID can be zero if not present

					Dirty bool
				}
				b, _ := json.Marshal(raw)
				json.Unmarshal(b, &m)

				mat := ecs.NewMaterial(m.BaseColor)
				mat.Ambient = m.Ambient
				mat.Diffuse = m.Diffuse
				mat.Specular = m.Specular
				mat.Shininess = m.Shininess
				mat.Metallic = m.Metallic
				mat.Roughness = m.Roughness
				mat.Type = m.Type
				mat.UseTexture = m.UseTexture
				mat.TextureID = m.TextureID
				mat.UseNormal = m.UseNormal
				mat.NormalID = m.NormalID
				mat.TextureAsset = m.TextureAsset
				mat.NormalAsset = m.NormalAsset
				mat.ShaderName = m.ShaderName
				mat.Shader = m.Shader

				// NEW: restore extended fields
				mat.DiffuseTexturePath = m.DiffuseTexturePath
				mat.NormalTexturePath = m.NormalTexturePath
				mat.OcclusionTexturePath = m.OcclusionTexturePath
				mat.MetallicRoughnessTexturePath = m.MetallicRoughnessTexturePath

				mat.TexCoordMap = m.TexCoordMap
				mat.UVScale = m.UVScale
				mat.UVOffset = m.UVOffset

				mat.NormalScale = m.NormalScale
				mat.SheenColor = m.SheenColor
				mat.SheenRoughness = m.SheenRoughness
				mat.SpecularFactor = m.SpecularFactor
				mat.OcclusionAsset = m.OcclusionAsset
				mat.OcclusionID = m.OcclusionID

				mat.MetallicRoughnessAsset = m.MetallicRoughnessAsset
				mat.MetallicRoughnessID = m.MetallicRoughnessID
				mat.UseIBL = m.UseIBL
				mat.IrradianceTex = m.IrradianceTex
				mat.PrefilteredEnvTex = m.PrefilteredEnvTex
				mat.BRDFLUTTex = m.BRDFLUTTex
				mat.ClearcoatFactor = m.ClearcoatFactor
				mat.ClearcoatRoughness = m.ClearcoatRoughness
				mat.ClearcoatRoughTex = m.ClearcoatRoughTex
				mat.ClearcoatTexture = m.ClearcoatTexture
				mat.ClearcoatNormalTex = m.ClearcoatNormalTex
				mat.UseClearcoat = m.UseClearcoat
				mat.TransmissionFactor = m.TransmissionFactor
				mat.UseTransmission = m.UseTransmission
				mat.TransmissionTex = m.TransmissionTex
				mat.Dirty = m.Dirty

				e.AddComponent(mat)

			case "DiffuseTexture":
				var t struct{ ID uint32 }
				b, _ := json.Marshal(raw)
				json.Unmarshal(b, &t)
				e.AddComponent(ecs.NewDiffuseTexture(t.ID))

			case "NormalMap":
				var t struct{ ID uint32 }
				b, _ := json.Marshal(raw)
				json.Unmarshal(b, &t)
				e.AddComponent(ecs.NewNormalMap(t.ID))
			}
		}
	}
	// choose a camera for scene.camera
	if activeCam != nil {
		scene.camera.Position = activeCam.Position
		scene.camera.Target = activeCam.Target
		scene.camera.Up = activeCam.Up
		scene.camera.Fov = activeCam.Fov
		scene.camera.Near = activeCam.Near
		scene.camera.Far = activeCam.Far
	} else if firstCam != nil {
		scene.camera.Position = firstCam.Position
		scene.camera.Target = firstCam.Target
		scene.camera.Up = firstCam.Up
		scene.camera.Fov = firstCam.Fov
		scene.camera.Near = firstCam.Near
		scene.camera.Far = firstCam.Far
	}

	// Third pass: restore hierarchy
	for _, se := range ss.Entities {
		if se.ParentID != 0 {
			child := entityByID[se.ID]
			parent := entityByID[se.ParentID]

			child.AddComponent(ecs.NewParent(parent))

			if ch := parent.GetComponent((*ecs.Children)(nil)); ch != nil {
				ch.(*ecs.Children).AddChild(child)
			} else {
				c := ecs.NewChildren()
				c.AddChild(child)
				parent.AddComponent(c)
			}
		}
	}

	return scene, nil
}
