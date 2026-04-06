package scene

import (
	"encoding/json"
	"os"
	"path/filepath"

	"go-engine/Go-Cordance/internal/ecs"
)

// -----------------------------------------------------------------------------
// Modern prefab format: reuse full scene serialization
// -----------------------------------------------------------------------------

type Prefab struct {
	RootID int64           `json:"root"`
	Scene  SerializedScene `json:"scene"`
}

// -----------------------------------------------------------------------------
// SavePrefab: serialize subtree using the same logic as full scenes
// -----------------------------------------------------------------------------

func (s *Scene) SavePrefab(path string, root *ecs.Entity) error {
	// Collect subtree
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	ents := collectSubtree(root)

	// Build SerializedScene
	ser := SerializedScene{
		Entities: make([]SerializedEntity, 0, len(ents)),
	}

	for _, e := range ents {
		ser.Entities = append(ser.Entities, serializeEntity(e))
	}

	// Wrap into Prefab
	prefab := Prefab{
		RootID: root.ID,
		Scene:  ser,
	}

	// Write JSON
	data, err := json.MarshalIndent(prefab, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// -----------------------------------------------------------------------------
// InstantiatePrefab: reuse scene deserialization logic
// -----------------------------------------------------------------------------

func (s *Scene) InstantiatePrefab(path string) (*ecs.Entity, []*ecs.Entity, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}

	var prefab Prefab
	if err := json.Unmarshal(data, &prefab); err != nil {
		return nil, nil, err
	}

	// 1. Create empty entities (new IDs)
	idMap := make(map[int64]*ecs.Entity)
	for _, se := range prefab.Scene.Entities {
		e := ecs.NewEntity(s.nextID)
		s.AddExisting(e)
		idMap[se.ID] = e
	}

	// 2. Add components
	for _, se := range prefab.Scene.Entities {
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

			case "Name":
				var n struct{ Value string }
				b, _ := json.Marshal(raw)
				json.Unmarshal(b, &n)
				e.AddComponent(ecs.NewName(n.Value))

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

	// 3. Restore hierarchy
	for _, se := range prefab.Scene.Entities {
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

	// Return new root
	root := idMap[prefab.RootID]
	return root, idMapToSlice(idMap), nil
}

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

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

func idMapToSlice(m map[int64]*ecs.Entity) []*ecs.Entity {
	out := make([]*ecs.Entity, 0, len(m))
	for _, e := range m {
		out = append(out, e)
	}
	return out
}
