package gltf

import (
	"go-engine/Go-Cordance/internal/ecs"
	"go-engine/Go-Cordance/internal/engine"
)

func LoadGLTFAnimations(path string) (map[string]*ecs.AnimationClip, error) {
	g, buffers, err := engine.LoadGLTFOrGLB(path)
	if err != nil {
		return nil, err
	}

	clips := map[string]*ecs.AnimationClip{}

	for _, anim := range g.Animations {
		clip := &ecs.AnimationClip{
			Name: anim.Name,
		}

		var duration float32
		// nodeIndex -> []TransformKeyframe
		perNode := map[int][]ecs.TransformKeyframe{}

		for _, ch := range anim.Channels {
			sampler := anim.Samplers[ch.Sampler]

			inputAcc, _ := engine.GetAccessor(g, buffers, sampler.Input)
			outputAcc, _ := engine.GetAccessor(g, buffers, sampler.Output)

			times := make([]float32, inputAcc.Acc.Count)
			for i := 0; i < inputAcc.Acc.Count; i++ {
				off := inputAcc.Base + i*inputAcc.Stride
				t := engine.BytesToFloat32(inputAcc.Buf[off:])
				times[i] = t
				if t > duration {
					duration = t
				}
			}

			nodeIndex := ch.Target.Node
			kfs := perNode[nodeIndex]
			if len(kfs) < len(times) {
				// grow and preserve existing data
				newKfs := make([]ecs.TransformKeyframe, len(times))
				copy(newKfs, kfs)
				kfs = newKfs
			}

			for i := 0; i < len(times); i++ {
				t := times[i]
				kf := kfs[i]
				kf.Time = t

				off := outputAcc.Base + i*outputAcc.Stride

				switch ch.Target.Path {
				case "translation":
					kf.Position = [3]float32{
						engine.BytesToFloat32(outputAcc.Buf[off+0:]),
						engine.BytesToFloat32(outputAcc.Buf[off+4:]),
						engine.BytesToFloat32(outputAcc.Buf[off+8:]),
					}
				case "rotation":
					kf.Rotation = [4]float32{
						engine.BytesToFloat32(outputAcc.Buf[off+0:]),
						engine.BytesToFloat32(outputAcc.Buf[off+4:]),
						engine.BytesToFloat32(outputAcc.Buf[off+8:]),
						engine.BytesToFloat32(outputAcc.Buf[off+12:]),
					}
				case "scale":
					kf.Scale = [3]float32{
						engine.BytesToFloat32(outputAcc.Buf[off+0:]),
						engine.BytesToFloat32(outputAcc.Buf[off+4:]),
						engine.BytesToFloat32(outputAcc.Buf[off+8:]),
					}
				}

				kfs[i] = kf
			}

			perNode[nodeIndex] = kfs
		}

		clip.Duration = duration

		for nodeIdx, kfs := range perNode {
			if len(kfs) == 0 {
				continue
			}
			clip.Tracks = append(clip.Tracks, ecs.AnimationTrack{
				NodeIndex: nodeIdx,
				Keyframes: kfs,
			})
		}

		clips[clip.Name] = clip
	}

	return clips, nil
}

func PickFirstClip(m map[string]*ecs.AnimationClip) string {
	for k := range m {
		return k
	}
	return ""
}
