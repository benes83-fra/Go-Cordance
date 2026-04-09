package ecs

type AnimationClip struct {
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
