package bridge

// in gizmo/bridge.go or system.go

// Called whenever a transform changes in the game.
// The editorlink layer will assign this.
var SendTransformToEditor func(
	id int64,
	pos [3]float32,
	rot [4]float32,
	scale [3]float32,
)

func NotifyEditorOfTransform(id int64, pos [3]float32, rot [4]float32, scale [3]float32) {
	if SendTransformToEditor != nil {
		SendTransformToEditor(id, pos, rot, scale)
	}
}
