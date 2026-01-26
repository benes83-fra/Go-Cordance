package undo

import (
	"log"

	"go-engine/Go-Cordance/internal/ecs"
	"go-engine/Go-Cordance/internal/scene"
)

//
// ──────────────────────────────────────────────────────────────
//   STRUCTURAL COMMAND INTERFACE  (UPDATED TO USE SCENE)
// ──────────────────────────────────────────────────────────────
//

type StructuralCommand interface {
	Undo(sc *scene.Scene)
	Redo(sc *scene.Scene)
}

//
// ──────────────────────────────────────────────────────────────
//   DELETE ENTITY COMMAND
// ──────────────────────────────────────────────────────────────
//

type DeleteEntityCommand struct {
	Entity *ecs.Entity
}

func (c DeleteEntityCommand) Undo(sc *scene.Scene) {
	world := sc.World()

	// If the entity already exists, don't duplicate it.
	if existing := world.FindByID(c.Entity.ID); existing != nil {
		log.Printf("undo: DeleteEntityCommand.Undo: entity %d already exists, skipping recreate", c.Entity.ID)
		return
	}

	// Add to world + scene
	world.AddEntity(c.Entity)
	sc.AddExisting(c.Entity)

	log.Printf("undo: DeleteEntityCommand.Undo: recreated entity %d with components %v",
		c.Entity.ID, c.Entity.Components)
}

func (c DeleteEntityCommand) Redo(sc *scene.Scene) {
	world := sc.World()

	world.RemoveEntityByID(c.Entity.ID)
	sc.DeleteEntityByID(c.Entity.ID)

	log.Printf("undo: DeleteEntityCommand.Redo: removed entity %d", c.Entity.ID)
}

//
// ──────────────────────────────────────────────────────────────
//   CREATE ENTITY COMMAND
// ──────────────────────────────────────────────────────────────
//

type CreateEntityCommand struct {
	Entity *ecs.Entity
}

func (c CreateEntityCommand) Undo(sc *scene.Scene) {
	world := sc.World()

	world.RemoveEntityByID(c.Entity.ID)
	sc.DeleteEntityByID(c.Entity.ID)

	log.Printf("undo: CreateEntityCommand.Undo: removed entity %d", c.Entity.ID)
}

func (c CreateEntityCommand) Redo(sc *scene.Scene) {
	world := sc.World()

	if existing := world.FindByID(c.Entity.ID); existing != nil {
		log.Printf("undo: CreateEntityCommand.Redo: entity %d already exists, skipping recreate", c.Entity.ID)
		return
	}

	world.AddEntity(c.Entity)
	sc.AddExisting(c.Entity)

	log.Printf("undo: CreateEntityCommand.Redo: recreated entity %d with components %v",
		c.Entity.ID, c.Entity.Components)
}

//
// ──────────────────────────────────────────────────────────────
//   STRUCTURAL UNDO STACK (UPDATED TO PASS SCENE)
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

func (u *StructuralUndoStack) Undo(sc *scene.Scene) {
	if !u.CanUndo() {
		return
	}
	u.stack[u.idx].Undo(sc)
	u.idx--
}

func (u *StructuralUndoStack) Redo(sc *scene.Scene) {
	if !u.CanRedo() {
		return
	}
	u.idx++
	u.stack[u.idx].Redo(sc)
}

//
// ──────────────────────────────────────────────────────────────
//   GLOBAL UNDO STACK (UPDATED TO PASS SCENE)
// ──────────────────────────────────────────────────────────────
//

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

func (g *GlobalUndoStack) Undo(sc *scene.Scene) {
	if !g.CanUndo() {
		return
	}

	a := g.actions[g.idx]

	switch a.Type {
	case ActionTransform:
		g.TransformUndo.Undo(sc.World())
	case ActionStructural:
		g.StructuralUndo.Undo(sc)
	}

	g.idx--
}

func (g *GlobalUndoStack) Redo(sc *scene.Scene) {
	if !g.CanRedo() {
		return
	}

	g.idx++
	a := g.actions[g.idx]

	switch a.Type {
	case ActionTransform:
		g.TransformUndo.Redo(sc.World())
	case ActionStructural:
		g.StructuralUndo.Redo(sc)
	}
}

var Global = NewGlobalUndoStack()
