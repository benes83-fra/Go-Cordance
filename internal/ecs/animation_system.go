package ecs

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

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

		if !player.Playing {
			continue
		}

		clip := player.Clips[player.Current]
		if clip == nil || len(clip.Tracks) == 0 {
			continue
		}

		// advance time and loop
		player.Time += dt * player.Speed
		if clip.Duration > 0 {
			player.Time = float32(math.Mod(float64(player.Time), float64(clip.Duration)))
			if player.Time < 0 {
				player.Time += clip.Duration
			}
		}

		// find skeleton
		skc := ent.GetComponent((*Skeleton)(nil))
		if skc == nil {
			continue
		}
		skeleton := skc.(*Skeleton)

		// apply each track to its node entity; only the animated property is touched
		for i := range clip.Tracks {
			track := &clip.Tracks[i]
			if track.NodeIndex < 0 || track.NodeIndex >= len(skeleton.Nodes) {
				continue
			}
			nodeEnt := skeleton.Nodes[track.NodeIndex]
			if nodeEnt == nil {
				continue
			}
			transform := nodeEnt.GetTransform()
			if transform == nil {
				continue
			}

			v, ok := sampleTrack(track, player.Time)
			if !ok {
				continue
			}

			switch track.Path {
			case PathTranslation:
				transform.Position = [3]float32{v[0], v[1], v[2]}
			case PathRotation:
				transform.Rotation = v
			case PathScale:
				transform.Scale = [3]float32{v[0], v[1], v[2]}
			}
			transform.Dirty = true
		}
	}
}

// sampleTrack evaluates a track at time t. Times before the first key clamp to
// the first value, times after the last key clamp to the last value.
func sampleTrack(track *AnimationTrack, t float32) ([4]float32, bool) {
	times := track.Times
	n := len(times)
	if n == 0 {
		return [4]float32{}, false
	}
	cubic := track.Interpolation == InterpCubicSpline
	if (cubic && len(track.Values) < 3*n) || (!cubic && len(track.Values) < n) {
		return [4]float32{}, false
	}

	value := func(i int) [4]float32 {
		if cubic {
			return track.Values[3*i+1]
		}
		return track.Values[i]
	}

	if n == 1 || t <= times[0] {
		return value(0), true
	}
	if t >= times[n-1] {
		return value(n - 1), true
	}

	// find k such that times[k] <= t < times[k+1]
	k := 0
	for k < n-2 && t >= times[k+1] {
		k++
	}

	td := times[k+1] - times[k]
	if td <= 0 {
		return value(k), true
	}
	alpha := (t - times[k]) / td
	isRot := track.Path == PathRotation

	switch track.Interpolation {
	case InterpStep:
		return value(k), true

	case InterpCubicSpline:
		// glTF cubic Hermite: p0 = value k, m0 = out-tangent k, p1 = value k+1, m1 = in-tangent k+1
		p0 := track.Values[3*k+1]
		m0 := track.Values[3*k+2]
		p1 := track.Values[3*(k+1)+1]
		m1 := track.Values[3*(k+1)]
		a2 := alpha * alpha
		a3 := a2 * alpha
		h00 := 2*a3 - 3*a2 + 1
		h10 := a3 - 2*a2 + alpha
		h01 := -2*a3 + 3*a2
		h11 := a3 - a2
		var out [4]float32
		for c := 0; c < 4; c++ {
			out[c] = h00*p0[c] + h10*td*m0[c] + h01*p1[c] + h11*td*m1[c]
		}
		if isRot {
			out = normalizeQuat(out)
		}
		return out, true

	default: // linear
		a, b := value(k), value(k+1)
		if isRot {
			return slerpQuat(a, b, alpha), true
		}
		return [4]float32{
			a[0] + (b[0]-a[0])*alpha,
			a[1] + (b[1]-a[1])*alpha,
			a[2] + (b[2]-a[2])*alpha,
			0,
		}, true
	}
}

func normalizeQuat(q [4]float32) [4]float32 {
	l := float32(math.Sqrt(float64(q[0]*q[0] + q[1]*q[1] + q[2]*q[2] + q[3]*q[3])))
	if l < 1e-8 {
		return [4]float32{0, 0, 0, 1}
	}
	return [4]float32{q[0] / l, q[1] / l, q[2] / l, q[3] / l}
}

func slerpQuat(a, b [4]float32, t float32) [4]float32 {
	q1 := mgl32.Quat{W: a[3], V: mgl32.Vec3{a[0], a[1], a[2]}}
	q2 := mgl32.Quat{W: b[3], V: mgl32.Vec3{b[0], b[1], b[2]}}

	// take the shortest arc (mathgl's Slerp does not flip on its own)
	if q1.Dot(q2) < 0 {
		q2 = q2.Scale(-1)
	}

	// Use mathgl's Slerp
	res := mgl32.QuatSlerp(q1, q2, t).Normalize()

	return [4]float32{res.V[0], res.V[1], res.V[2], res.W}
}
