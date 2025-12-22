package scene

import (
	"encoding/json"
	"os"

	"go-engine/Go-Cordance/internal/ecs"
)

type Prefab struct {
	Root SerializedEntity   `json:"root"`
	All  []SerializedEntity `json:"all"`
}

func (s *Scene) SavePrefab(path string, root *ecs.Entity) error {
	// First: collect the entire subtree
	entities := collectSubtree(root)

	// Convert to serialized form
	serialized := Prefab{
		Root: serializeEntity(root),
		All:  make([]SerializedEntity, 0, len(entities)),
	}

	for _, e := range entities {
		serialized.All = append(serialized.All, serializeEntity(e))
	}

	// Write JSON
	data, err := json.MarshalIndent(serialized, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func collectSubtree(root *ecs.Entity) []*ecs.Entity {
	var out []*ecs.Entity

	var walk func(e *ecs.Entity)
	walk = func(e *ecs.Entity) {
		out = append(out, e)

		if ch := e.GetComponent((*ecs.Children)(nil)); ch != nil {
			for _, c := range ch.(*ecs.Children).Entities {
				walk(c)
			}
		}
	}

	walk(root)
	return out
}
func serializeEntity(e *ecs.Entity) SerializedEntity {
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

	return se
}

func (s *Scene) InstantiatePrefab(path string) (*ecs.Entity, []*ecs.Entity, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}

	var prefab Prefab
	if err := json.Unmarshal(data, &prefab); err != nil {
		return nil, nil, err
	}

	// Map old IDs → new entities
	idMap := make(map[int64]*ecs.Entity)

	// First pass: create entities
	for _, se := range prefab.All {
		e := ecs.NewEntity(s.nextID)
		s.AddExisting(e)
		idMap[se.ID] = e
	}

	// Second pass: add components
	for _, se := range prefab.All {
		e := idMap[se.ID]

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
	for _, se := range prefab.All {
		if se.ParentID != 0 {
			child := idMap[se.ID]
			parent := idMap[se.ParentID]

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

	// Return the new root entity
	root := idMap[prefab.Root.ID]
	return root, idMapToSlice(idMap), nil
}

func idMapToSlice(m map[int64]*ecs.Entity) []*ecs.Entity {
	out := make([]*ecs.Entity, 0, len(m))
	for _, e := range m {
		out = append(out, e)
	}
	return out
}
