package undo

import (
	"go-engine/Go-Cordance/internal/ecs"
)

// TransformSnapshot uses entity ID only (no engine/editor imports)
type TransformSnapshot struct {
	EntityID int64
	Position [3]float32
	Rotation [4]float32
	Scale    [3]float32
}

type TransformCommand struct {
	Before []TransformSnapshot
	After  []TransformSnapshot
}

type UndoStack struct {
	stack []*TransformCommand
	idx   int
}

func NewUndoStack() *UndoStack { return &UndoStack{stack: []*TransformCommand{}, idx: -1} }

func (u *UndoStack) Push(cmd *TransformCommand) {
	if u.idx < len(u.stack)-1 {
		u.stack = u.stack[:u.idx+1]
	}
	u.stack = append(u.stack, cmd)
	u.idx = len(u.stack) - 1
}

func (u *UndoStack) CanUndo() bool { return u.idx >= 0 }
func (u *UndoStack) CanRedo() bool { return u.idx < len(u.stack)-1 }

func (u *UndoStack) Undo(world *ecs.World) {
	if !u.CanUndo() {
		return
	}
	cmd := u.stack[u.idx]
	applySnapshots(world, cmd.Before)
	u.idx--
}

func (u *UndoStack) Redo(world *ecs.World) {
	if !u.CanRedo() {
		return
	}
	u.idx++
	cmd := u.stack[u.idx]
	applySnapshots(world, cmd.After)
}

func applySnapshots(world *ecs.World, snaps []TransformSnapshot) {
	for _, s := range snaps {
		if e := world.FindByID(s.EntityID); e != nil {
			if tr := ecs.GetTransform(e); tr != nil {
				tr.Position = s.Position
				tr.Rotation = s.Rotation
				tr.Scale = s.Scale
			}
		}
	}
}
