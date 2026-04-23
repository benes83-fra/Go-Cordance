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

		if !player.Playing /*|| player.Current == "" */ {
			// fmt.Println("  plyer playing not found on ent", player, player.Playing, player.Current)
			continue
		}

		clip := player.Clips[player.Current]
		if clip == nil || len(clip.Tracks) == 0 {
			continue
		}
		// advance time
		player.Time += dt * player.Speed
		if player.Time > clip.Duration {
			player.Time = 0
		}

		// find skeleton
		skc := ent.GetComponent((*Skeleton)(nil))
		if skc == nil {
			// fmt.Println("  BUT: no Skeleton on ent", ent.ID)
			continue
		}
		skeleton := skc.(*Skeleton)
		// fmt.Println("  Skeleton found on ent", ent.ID, "nodes:", len(skeleton.Nodes))

		// apply each track to its node entity
		for _, track := range clip.Tracks {
			if track.NodeIndex < 0 || track.NodeIndex >= len(skeleton.Nodes) {
				continue
			}
			nodeEnt := skeleton.Nodes[track.NodeIndex]
			if nodeEnt == nil {
				continue
			}

			kf1, kf2 := findKeyframePairTrack(track.Keyframes, player.Time)
			if kf1 == nil || kf2 == nil {
				continue
			}

			t := (player.Time - kf1.Time) / (kf2.Time - kf1.Time)

			pos := lerpVec3(kf1.Position, kf2.Position, t)
			rot := slerpQuat(kf1.Rotation, kf2.Rotation, t)
			// Inside AnimationSystem.Update
			scl := lerpVec3(kf1.Scale, kf2.Scale, t)

			// Add this safety check:
			if math.Abs(float64(scl[0]-1.0)) < 1e-5 &&
				math.Abs(float64(scl[1]-1.0)) < 1e-5 &&
				math.Abs(float64(scl[2]-1.0)) < 1e-5 {
				scl = [3]float32{1, 1, 1}
			}

			// fmt.Printf("Track node=%d t=%.3f pos=%v rot=%v scl=%v\n",
			// 	track.NodeIndex, player.Time, pos, rot, scl)

			if tr := nodeEnt.GetComponent((*Transform)(nil)); tr != nil {
				transform := tr.(*Transform)
				// only overwrite if channel actually had data
				if kf1.Position != [3]float32{} || kf2.Position != [3]float32{} {
					transform.Position = pos
				}
				if kf1.Rotation != [4]float32{} || kf2.Rotation != [4]float32{} {
					transform.Rotation = rot
				}
				if kf1.Scale != [3]float32{} || kf2.Scale != [3]float32{} {
					transform.Scale = scl
				}
				transform.Dirty = true
			}
			if clip == nil || len(clip.Tracks) == 0 {
				// fmt.Printf("Anim: ent %d has clip %q but no tracks or nil clip\n", ent.ID, player.Current)
				continue
			}

			// fmt.Printf("Anim: ent %d time=%.3f duration=%.3f tracks=%d\n",
			// 	ent.ID, player.Time, clip.Duration, len(clip.Tracks))

		}
	}
}

func findKeyframePairTrack(kfs []TransformKeyframe, time float32) (*TransformKeyframe, *TransformKeyframe) {
	for i := 0; i < len(kfs)-1; i++ {
		k1 := &kfs[i]
		k2 := &kfs[i+1]
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

// func slerpQuat(a, b [4]float32, t float32) [4]float32 {
// 	// simple linear fallback for now (safe)
// 	return [4]float32{
// 		a[0] + (b[0]-a[0])*t,
// 		a[1] + (b[1]-a[1])*t,
// 		a[2] + (b[2]-a[2])*t,
// 		a[3] + (b[3]-a[3])*t,
// 	}
// }

func slerpQuat(a, b [4]float32, t float32) [4]float32 {
	q1 := mgl32.Quat{W: a[3], V: mgl32.Vec3{a[0], a[1], a[2]}}
	q2 := mgl32.Quat{W: b[3], V: mgl32.Vec3{b[0], b[1], b[2]}}

	// Use mathgl's Slerp
	res := mgl32.QuatSlerp(q1, q2, t).Normalize()

	return [4]float32{res.V[0], res.V[1], res.V[2], res.W}
}
