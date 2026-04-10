package ecs

type AnimationClip struct {
	Name      string
	Duration  float32
	Keyframes []TransformKeyframe
}

type TransformKeyframe struct {
	Time     float32
	Position [3]float32
	Rotation [4]float32
	Scale    [3]float32
}

type AnimationPlayer struct {
	Clips   map[string]*AnimationClip
	Current string
	Time    float32
	Speed   float32
	Playing bool
}

func (ap *AnimationPlayer) Update(dt float32) {}

func (ap *AnimationPlayer) EditorName() string {
	return "AnimationPlayer"
}

func (ap *AnimationPlayer) EditorFields() map[string]any {
	fields := map[string]any{
		"Current": ap.Current,
		"Speed":   ap.Speed,
		"Playing": ap.Playing,
		"Time":    ap.Time,
	}

	// expose clip names for UI
	clipNames := make([]string, 0, len(ap.Clips))
	for name := range ap.Clips {
		clipNames = append(clipNames, name)
	}
	fields["Clips"] = clipNames

	return fields
}

func (ap *AnimationPlayer) SetEditorField(name string, value any) {
	switch name {
	case "Current":
		if s, ok := value.(string); ok {
			ap.Current = s
			ap.Time = 0
		}
	case "Speed":
		if f, ok := value.(float32); ok {
			ap.Speed = f
		}
	case "Playing":
		if b, ok := value.(bool); ok {
			ap.Playing = b
		}
	case "Time":
		if f, ok := value.(float32); ok {
			ap.Time = f
		}
	}
}
