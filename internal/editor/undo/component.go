package undo

import (
	"go-engine/Go-Cordance/internal/assets"
	"go-engine/Go-Cordance/internal/ecs"
)

// getComponentByName returns the component instance on the entity
// using the ECS ComponentRegistry to resolve the type.
func getComponentByName(ent *ecs.Entity, name string) ecs.Component {
	constructor, ok := ecs.ComponentRegistry[name]
	if !ok {
		return nil
	}
	proto := constructor()         // zero-value instance
	return ent.GetComponent(proto) // ECS matches by reflect.TypeOf
}
func ApplyComponentFields(ent *ecs.Entity, name string, fields map[string]any) {
	if fields == nil {
		return
	}

	comp := getComponentByName(ent, name)
	if comp == nil {
		return
	}

	switch c := comp.(type) {

	case *ecs.Material:
		if v, ok := fields["BaseColor"].([4]float32); ok {
			c.BaseColor = v
		}
		if v, ok := fields["UseTexture"].(bool); ok {
			c.UseTexture = v
		}
		if v, ok := fields["TextureAsset"].(assets.AssetID); ok {
			c.TextureAsset = v
		}
		if v, ok := fields["TextureID"].(uint32); ok {
			c.TextureID = v
		}
		if v, ok := fields["ShaderName"].(string); ok {
			c.ShaderName = v
		}
		if v, ok := fields["UseNormal"].(bool); ok {
			c.UseNormal = v
		}
		if v, ok := fields["NormalAsset"].(assets.AssetID); ok {
			c.NormalAsset = v
		}
		if v, ok := fields["NormalID"].(uint32); ok {
			c.NormalID = v
		}
		c.Dirty = true

	case *ecs.Mesh:
		if v, ok := fields["MeshID"].(string); ok {
			c.ID = v
		}
	}
}

func SnapshotComponent(ent *ecs.Entity, name string) map[string]any {
	comp := getComponentByName(ent, name)
	if comp == nil {
		return nil
	}

	switch c := comp.(type) {

	case *ecs.Material:
		return map[string]any{
			"BaseColor":    c.BaseColor,
			"UseTexture":   c.UseTexture,
			"TextureAsset": c.TextureAsset,
			"TextureID":    c.TextureID,
			"ShaderName":   c.ShaderName,
			"UseNormal":    c.UseNormal,
			"NormalAsset":  c.NormalAsset,
			"NormalID":     c.NormalID,
		}

	case *ecs.Mesh:
		return map[string]any{
			"MeshID": c.ID,
		}
	}

	return nil
}
