package ecs

// AngularVelocity stores rotational velocity (radians/sec) around X,Y,Z axes.
type AngularVelocity struct {
	Vel [3]float32
}

func NewAngularVelocity(vx, vy, vz float32) *AngularVelocity {
	return &AngularVelocity{Vel: [3]float32{vx, vy, vz}}
}

func (av *AngularVelocity) Update(dt float32) {
	_ = dt
}

// AngularAcceleration stores rotational acceleration (radians/sec^2).
type AngularAcceleration struct {
	Acc [3]float32
}

func NewAngularAcceleration(ax, ay, az float32) *AngularAcceleration {
	return &AngularAcceleration{Acc: [3]float32{ax, ay, az}}
}

func (aa *AngularAcceleration) Update(dt float32) {
	_ = dt
}
