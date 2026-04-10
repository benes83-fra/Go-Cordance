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
		clip := &ecs.AnimationClip{}
		clip.Name = anim.Name

		var duration float32 = 0
		keyframes := map[int]*ecs.TransformKeyframe{}

		for _, ch := range anim.Channels {
			sampler := anim.Samplers[ch.Sampler]

			inputAcc, _ := engine.GetAccessor(g, buffers, sampler.Input)
			times := make([]float32, inputAcc.Acc.Count)

			for i := 0; i < inputAcc.Acc.Count; i++ {
				off := inputAcc.Base + i*inputAcc.Stride
				times[i] = engine.BytesToFloat32(inputAcc.Buf[off:])
				if times[i] > duration {
					duration = times[i]
				}
			}

			outputAcc, _ := engine.GetAccessor(g, buffers, sampler.Output)

			for i := 0; i < inputAcc.Acc.Count; i++ {
				t := times[i]

				kf := keyframes[i]
				if kf == nil {
					kf = &ecs.TransformKeyframe{Time: t}
					keyframes[i] = kf
				}

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
			}
		}

		for _, kf := range keyframes {
			clip.Keyframes = append(clip.Keyframes, *kf)
		}

		clip.Duration = duration
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
