package gltf

import (
	"fmt"
	"go-engine/Go-Cordance/internal/ecs"
	"go-engine/Go-Cordance/internal/engine"
	"log"
	"sort"
)

// LoadGLTFAnimations loads every animation as a clip, keyed by animation name
// (or "Animation_<i>" if unnamed). Each glTF channel becomes its own track, so
// channels with different key times or counts never get mixed.
func LoadGLTFAnimations(path string) (map[string]*ecs.AnimationClip, error) {
	g, buffers, err := engine.LoadGLTFOrGLB(path)
	if err != nil {
		return nil, err
	}

	clips := map[string]*ecs.AnimationClip{}

	for ai, anim := range g.Animations {
		name := anim.Name
		if name == "" {
			name = fmt.Sprintf("Animation_%d", ai)
		}
		clip := &ecs.AnimationClip{Name: name}

		for ci, ch := range anim.Channels {
			var path ecs.AnimationPath
			var comps int
			switch ch.Target.Path {
			case "translation":
				path, comps = ecs.PathTranslation, 3
			case "rotation":
				path, comps = ecs.PathRotation, 4
			case "scale":
				path, comps = ecs.PathScale, 3
			default:
				// e.g. "weights" (morph targets) – not supported yet
				continue
			}

			if ch.Sampler < 0 || ch.Sampler >= len(anim.Samplers) {
				log.Printf("animation %q channel %d: invalid sampler %d", name, ci, ch.Sampler)
				continue
			}
			sampler := anim.Samplers[ch.Sampler]

			interp := ecs.InterpLinear
			switch sampler.Interpolation {
			case "STEP":
				interp = ecs.InterpStep
			case "CUBICSPLINE":
				interp = ecs.InterpCubicSpline
			}

			inputAcc, err := engine.GetAccessor(g, buffers, sampler.Input)
			if err != nil {
				log.Printf("animation %q channel %d input: %v", name, ci, err)
				continue
			}
			outputAcc, err := engine.GetAccessor(g, buffers, sampler.Output)
			if err != nil {
				log.Printf("animation %q channel %d output: %v", name, ci, err)
				continue
			}

			times := make([]float32, inputAcc.Acc.Count)
			for i := range times {
				times[i] = engine.BytesToFloat32(inputAcc.Buf[inputAcc.Base+i*inputAcc.Stride:])
			}
			if n := len(times); n > 0 && times[n-1] > clip.Duration {
				clip.Duration = times[n-1]
			}

			wantValues := len(times)
			if interp == ecs.InterpCubicSpline {
				wantValues *= 3
			}
			if outputAcc.Acc.Count < wantValues {
				log.Printf("animation %q channel %d: %d output values, expected %d", name, ci, outputAcc.Acc.Count, wantValues)
				continue
			}

			values := make([][4]float32, wantValues)
			for i := range values {
				off := outputAcc.Base + i*outputAcc.Stride
				for c := 0; c < comps; c++ {
					v, ok := readComponent(outputAcc.Buf, off, c, outputAcc.Acc.ComponentType)
					if !ok {
						return nil, fmt.Errorf("animation %q channel %d: unsupported component type %d", name, ci, outputAcc.Acc.ComponentType)
					}
					values[i][c] = v
				}
			}

			clip.Tracks = append(clip.Tracks, ecs.AnimationTrack{
				NodeIndex:     ch.Target.Node,
				Path:          path,
				Interpolation: interp,
				Times:         times,
				Values:        values,
			})
		}

		clips[name] = clip
	}

	return clips, nil
}

// readComponent reads component c of an element at byte offset off.
// Integer types are treated as normalized (the only integer form glTF allows
// for animation outputs).
func readComponent(buf []byte, off, c, componentType int) (float32, bool) {
	switch componentType {
	case 5126: // FLOAT
		return engine.BytesToFloat32(buf[off+4*c:]), true
	case 5120: // BYTE
		v := float32(int8(buf[off+c])) / 127
		if v < -1 {
			v = -1
		}
		return v, true
	case 5121: // UNSIGNED_BYTE
		return float32(buf[off+c]) / 255, true
	case 5122: // SHORT
		o := off + 2*c
		v := float32(int16(uint16(buf[o])|uint16(buf[o+1])<<8)) / 32767
		if v < -1 {
			v = -1
		}
		return v, true
	case 5123: // UNSIGNED_SHORT
		o := off + 2*c
		return float32(uint16(buf[o])|uint16(buf[o+1])<<8) / 65535, true
	}
	return 0, false
}

// PickClip returns the first of the preferred clip names that exists in m.
// If none match, it falls back to the alphabetically first clip so the choice
// is deterministic across runs. Returns "" if m is empty.
func PickClip(m map[string]*ecs.AnimationClip, preferred ...string) string {
	for _, p := range preferred {
		if _, ok := m[p]; ok {
			return p
		}
	}
	names := make([]string, 0, len(m))
	for k := range m {
		names = append(names, k)
	}
	if len(names) == 0 {
		return ""
	}
	sort.Strings(names)
	return names[0]
}

// PickFirstClip returns the alphabetically first clip name (deterministic).
func PickFirstClip(m map[string]*ecs.AnimationClip) string {
	return PickClip(m)
}
