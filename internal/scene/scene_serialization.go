package scene

import (
	"encoding/json"
	"go-engine/Go-Cordance/internal/ecs"
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
	return map[string]interface{}{"baseColor": m.BaseColor, "ambient": m.Ambient, "diffuse": m.Diffuse, "specular": m.Specular, "shininess": m.Shininess}
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

		// DiffuseTexture
		if dt := e.GetComponent((*ecs.DiffuseTexture)(nil)); dt != nil {
			se.Components["DiffuseTexture"] = serializeDiffuseTexture(dt.(*ecs.DiffuseTexture))
		}

		// NormalMap
		if nm := e.GetComponent((*ecs.NormalMap)(nil)); nm != nil {
			se.Components["NormalMap"] = serializeNormalMap(nm.(*ecs.NormalMap))
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

	var ss SerializedScene
	if err := json.Unmarshal(data, &ss); err != nil {
		return nil, err
	}

	scene := New()

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

			case "Mesh":
				var m struct{ ID string }
				b, _ := json.Marshal(raw)
				json.Unmarshal(b, &m)
				e.AddComponent(ecs.NewMesh(m.ID))

			case "Material":
				var m struct {
					BaseColor [4]float32
					Ambient   float32
					Diffuse   float32
					Specular  float32
					Shininess float32
				}
				b, _ := json.Marshal(raw)
				json.Unmarshal(b, &m)
				mat := ecs.NewMaterial(m.BaseColor)
				mat.Ambient = m.Ambient
				mat.Diffuse = m.Diffuse
				mat.Specular = m.Specular
				mat.Shininess = m.Shininess
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
