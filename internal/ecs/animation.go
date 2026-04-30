package ecs

import (
	"math"
)

type AnimationTrack struct {
	NodeIndex int
	Keyframes []TransformKeyframe
}

type AnimationClip struct {
	Name     string
	Duration float32
	Tracks   []AnimationTrack
}

type TransformKeyframe struct {
	Time     float32
	Position [3]float32
	Rotation [4]float32
	Scale    [3]float32
}

type AnimationPlayer struct {
	Clips        map[string]*AnimationClip
	Current      string
	Time         float32
	Speed        float32
	Playing      bool
	NodeEntities []*Entity
}

func (ap *AnimationPlayer) Update(dt float32) {
	if !ap.Playing {
		return
	}

	clip := ap.Clips[ap.Current]
	if clip == nil || len(clip.Tracks) == 0 {
		return
	}

	ap.Time += dt * ap.Speed
	// temporary debug
	// log.Printf("Animation %s time=%.3f", ap.Current, ap.Time)

	if ap.Time > clip.Duration {
		ap.Time = float32(math.Mod(float64(ap.Time), float64(clip.Duration)))
	}

	for _, track := range clip.Tracks {
		kf := sampleTrack(track, ap.Time)

		// bounds check to avoid panics
		if track.NodeIndex < 0 || track.NodeIndex >= len(ap.NodeEntities) {
			continue
		}
		ent := ap.NodeEntities[track.NodeIndex]
		if ent == nil {
			continue
		}
		if track.NodeIndex == 0 { // root bone
			kf.Position[1] += 0.5 // bob up and down
		}

		t := ent.GetTransform()
		t.Position = kf.Position
		t.Rotation = kf.Rotation
		t.Scale = kf.Scale
	}
}

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
func sampleTrack(track AnimationTrack, t float32) TransformKeyframe {
	kfs := track.Keyframes
	if len(kfs) == 0 {
		return TransformKeyframe{}
	}

	// find the two keyframes around t
	for i := 0; i < len(kfs)-1; i++ {
		if t >= kfs[i].Time && t <= kfs[i+1].Time {
			a := kfs[i]
			b := kfs[i+1]
			alpha := (t - a.Time) / (b.Time - a.Time)
			return lerpKeyframe(a, b, alpha)
		}
	}

	return kfs[len(kfs)-1]
}
func lerpKeyframe(a, b TransformKeyframe, alpha float32) TransformKeyframe {
	return TransformKeyframe{
		Time:     a.Time + alpha*(b.Time-a.Time),
		Position: lerp3(a.Position, b.Position, alpha),
		Rotation: slerp(a.Rotation, b.Rotation, alpha),
		Scale:    lerp3(a.Scale, b.Scale, alpha),
	}
}
func slerp(a, b [4]float32, t float32) [4]float32 {
	dot := a[0]*b[0] + a[1]*b[1] + a[2]*b[2] + a[3]*b[3]

	if dot < 0 {
		b = [4]float32{-b[0], -b[1], -b[2], -b[3]}
		dot = -dot
	}

	if dot > 0.9995 {
		return lerpQuat(a, b, t)
	}

	theta0 := float32(math.Acos(float64(dot)))
	theta := theta0 * t

	sinTheta := float32(math.Sin(float64(theta)))
	sinTheta0 := float32(math.Sin(float64(theta0)))

	s0 := float32(math.Cos(float64(theta))) - dot*sinTheta/sinTheta0
	s1 := sinTheta / sinTheta0

	return [4]float32{
		s0*a[0] + s1*b[0],
		s0*a[1] + s1*b[1],
		s0*a[2] + s1*b[2],
		s0*a[3] + s1*b[3],
	}
}

func lerpQuat(a, b [4]float32, t float32) [4]float32 {
	return [4]float32{
		a[0] + (b[0]-a[0])*t,
		a[1] + (b[1]-a[1])*t,
		a[2] + (b[2]-a[2])*t,
		a[3] + (b[3]-a[3])*t,
	}
}
func lerp3(a, b [3]float32, t float32) [3]float32 {
	return [3]float32{
		a[0] + (b[0]-a[0])*t,
		a[1] + (b[1]-a[1])*t,
		a[2] + (b[2]-a[2])*t,
	}
}
