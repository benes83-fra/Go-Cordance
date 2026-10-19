package ecs

// AnimationPath is the node property an AnimationTrack drives.
type AnimationPath int

const (
	PathTranslation AnimationPath = iota
	PathRotation
	PathScale
)

// Interpolation mirrors glTF sampler interpolation modes.
type Interpolation int

const (
	InterpLinear Interpolation = iota
	InterpStep
	InterpCubicSpline
)

// AnimationTrack drives one property of one node (one glTF channel).
// Values hold vec3 (xyz, w unused) or quaternions (x, y, z, w).
// For InterpCubicSpline there are 3 values per key: in-tangent, value, out-tangent.
type AnimationTrack struct {
	NodeIndex     int
	Path          AnimationPath
	Interpolation Interpolation
	Times         []float32
	Values        [][4]float32
}

type AnimationClip struct {
	Name     string
	Duration float32
	Tracks   []AnimationTrack
}

type AnimationPlayer struct {
	Clips        map[string]*AnimationClip
	Current      string
	Time         float32
	Speed        float32
	Playing      bool
	NodeEntities []*Entity
}

// Update is intentionally a no-op: AnimationSystem advances time and applies
// tracks. Doing it here too ran the animation twice per frame and zeroed bone
// positions for rotation-only tracks.
func (ap *AnimationPlayer) Update(dt float32) { _ = dt }

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
