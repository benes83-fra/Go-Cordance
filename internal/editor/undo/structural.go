package undo

import (
	"go-engine/Go-Cordance/internal/ecs"
	"go-engine/Go-Cordance/internal/editor/bridge"
)

//
// ──────────────────────────────────────────────────────────────
//   STRUCTURAL COMMAND INTERFACE
// ──────────────────────────────────────────────────────────────
//

type StructuralCommand interface {
	Undo(world *ecs.World)
	Redo(world *ecs.World)
}

//
// ──────────────────────────────────────────────────────────────
//   DELETE ENTITY COMMAND
// ──────────────────────────────────────────────────────────────
//

type DeleteEntityCommand struct {
	Entity bridge.EntityInfo
}

func (c DeleteEntityCommand) Undo(world *ecs.World) {
	// Recreate entity
	ent := ecs.NewEntity(c.Entity.ID)

	// Recreate components
	for _, cname := range c.Entity.Components {
		constructor := ecs.ComponentRegistry[cname]
		comp := constructor()
		ent.AddComponent(comp)
	}

	// Reapply transform
	if tr := ecs.GetTransform(ent); tr != nil {
		tr.Position = c.Entity.Position
		tr.Rotation = c.Entity.Rotation
		tr.Scale = c.Entity.Scale
	}

	world.AddEntity(ent)
}

func (c DeleteEntityCommand) Redo(world *ecs.World) {
	world.RemoveEntityByID(c.Entity.ID)
}

type CreateEntityCommand struct {
	Entity bridge.EntityInfo
}

func (c CreateEntityCommand) Undo(world *ecs.World) {
	// Undo creation = delete the entity
	world.RemoveEntityByID(c.Entity.ID)
}

func (c CreateEntityCommand) Redo(world *ecs.World) {
	// Redo creation = recreate the entity
	ent := ecs.NewEntity(c.Entity.ID)

	for _, cname := range c.Entity.Components {
		constructor := ecs.ComponentRegistry[cname]
		comp := constructor()
		ent.AddComponent(comp)
	}

	if tr := ecs.GetTransform(ent); tr != nil {
		tr.Position = c.Entity.Position
		tr.Rotation = c.Entity.Rotation
		tr.Scale = c.Entity.Scale
	}

	world.AddEntity(ent)
}

//
// ──────────────────────────────────────────────────────────────
//   STRUCTURAL UNDO STACK
// ──────────────────────────────────────────────────────────────
//

type StructuralUndoStack struct {
	stack []StructuralCommand
	idx   int
}

func NewStructuralUndoStack() *StructuralUndoStack {
	return &StructuralUndoStack{
		stack: []StructuralCommand{},
		idx:   -1,
	}
}

func (u *StructuralUndoStack) Push(cmd StructuralCommand) {
	// Drop redo history
	if u.idx < len(u.stack)-1 {
		u.stack = u.stack[:u.idx+1]
	}
	u.stack = append(u.stack, cmd)
	u.idx = len(u.stack) - 1
}

func (u *StructuralUndoStack) CanUndo() bool { return u.idx >= 0 }
func (u *StructuralUndoStack) CanRedo() bool { return u.idx < len(u.stack)-1 }

func (u *StructuralUndoStack) Undo(world *ecs.World) {
	if !u.CanUndo() {
		return
	}
	u.stack[u.idx].Undo(world)
	u.idx--
}

func (u *StructuralUndoStack) Redo(world *ecs.World) {
	if !u.CanRedo() {
		return
	}
	u.idx++
	u.stack[u.idx].Redo(world)
}

//
// ──────────────────────────────────────────────────────────────
//   UNIFIED GLOBAL UNDO STACK
// ──────────────────────────────────────────────────────────────
//

// This is what the user interacts with.
// It stores a timeline of actions, each pointing to either a transform or structural command.

type ActionType int

const (
	ActionTransform ActionType = iota
	ActionStructural
)

type GlobalAction struct {
	Type       ActionType
	Transform  *TransformCommand
	Structural StructuralCommand
}

type GlobalUndoStack struct {
	actions []GlobalAction
	idx     int

	TransformUndo  *UndoStack
	StructuralUndo *StructuralUndoStack
}

func NewGlobalUndoStack() *GlobalUndoStack {
	return &GlobalUndoStack{
		actions:        []GlobalAction{},
		idx:            -1,
		TransformUndo:  NewUndoStack(),
		StructuralUndo: NewStructuralUndoStack(),
	}
}

func (g *GlobalUndoStack) PushTransform(cmd *TransformCommand) {
	g.TransformUndo.Push(cmd)
	g.push(GlobalAction{Type: ActionTransform, Transform: cmd})
}

func (g *GlobalUndoStack) PushStructural(cmd StructuralCommand) {
	g.StructuralUndo.Push(cmd)
	g.push(GlobalAction{Type: ActionStructural, Structural: cmd})
}

func (g *GlobalUndoStack) push(a GlobalAction) {
	// Drop redo history
	if g.idx < len(g.actions)-1 {
		g.actions = g.actions[:g.idx+1]
	}
	g.actions = append(g.actions, a)
	g.idx = len(g.actions) - 1
}

func (g *GlobalUndoStack) CanUndo() bool { return g.idx >= 0 }
func (g *GlobalUndoStack) CanRedo() bool { return g.idx < len(g.actions)-1 }

func (g *GlobalUndoStack) Undo(world *ecs.World) {
	if !g.CanUndo() {
		return
	}

	a := g.actions[g.idx]

	switch a.Type {
	case ActionTransform:
		g.TransformUndo.Undo(world)
	case ActionStructural:
		g.StructuralUndo.Undo(world)
	}

	g.idx--
}

func (g *GlobalUndoStack) Redo(world *ecs.World) {
	if !g.CanRedo() {
		return
	}

	g.idx++
	a := g.actions[g.idx]

	switch a.Type {
	case ActionTransform:
		g.TransformUndo.Redo(world)
	case ActionStructural:
		g.StructuralUndo.Redo(world)
	}
}
