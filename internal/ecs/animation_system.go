package ecs

type AnimationSystem struct{}

func NewAnimationSystem() *AnimationSystem {
	return &AnimationSystem{}
}

func (sys *AnimationSystem) Update(dt float32, ents []*Entity) {
	for _, ent := range ents {
		apc := ent.GetComponent((*AnimationPlayer)(nil))
		if apc == nil {
			continue
		}
		player := apc.(*AnimationPlayer)

		if !player.Playing || player.Current == "" {
			continue
		}

		clip := player.Clips[player.Current]
		if clip == nil || len(clip.Keyframes) == 0 {
			continue
		}

		// Advance time
		player.Time += dt * player.Speed
		if player.Time > clip.Duration {
			player.Time = 0 // loop
		}

		// Find keyframes
		kf1, kf2 := findKeyframePair(clip, player.Time)
		if kf1 == nil || kf2 == nil {
			continue
		}

		// Interpolate
		t := (player.Time - kf1.Time) / (kf2.Time - kf1.Time)

		pos := lerpVec3(kf1.Position, kf2.Position, t)
		rot := slerpQuat(kf1.Rotation, kf2.Rotation, t)
		scl := lerpVec3(kf1.Scale, kf2.Scale, t)

		// Apply to Transform
		if tr := ent.GetComponent(&Transform{}); tr != nil {
			transform := tr.(*Transform)
			transform.Position = pos
			transform.Rotation = rot
			transform.Scale = scl
		}
	}
}

func findKeyframePair(clip *AnimationClip, time float32) (*TransformKeyframe, *TransformKeyframe) {
	for i := 0; i < len(clip.Keyframes)-1; i++ {
		k1 := &clip.Keyframes[i]
		k2 := &clip.Keyframes[i+1]
		if time >= k1.Time && time <= k2.Time {
			return k1, k2
		}
	}
	return nil, nil
}

func lerpVec3(a, b [3]float32, t float32) [3]float32 {
	return [3]float32{
		a[0] + (b[0]-a[0])*t,
		a[1] + (b[1]-a[1])*t,
		a[2] + (b[2]-a[2])*t,
	}
}

func slerpQuat(a, b [4]float32, t float32) [4]float32 {
	// simple linear fallback for now (safe)
	return [4]float32{
		a[0] + (b[0]-a[0])*t,
		a[1] + (b[1]-a[1])*t,
		a[2] + (b[2]-a[2])*t,
		a[3] + (b[3]-a[3])*t,
	}
}
