package ecs

import "strconv"

// RigidBody is a physics component with mass, velocity, and accumulated force.
type RigidBody struct {
	Mass  float32
	Vel   [3]float32
	Force [3]float32
}

// NewRigidBody creates a rigid body with given mass.
func NewRigidBody(mass float32) *RigidBody {
	return &RigidBody{Mass: mass}
}

// ApplyForce adds a force vector to the body (accumulated until next update).
func (rb *RigidBody) ApplyForce(fx, fy, fz float32) {
	rb.Force[0] += fx
	rb.Force[1] += fy
	rb.Force[2] += fz
}

// ClearForce resets accumulated forces (called after integration).
func (rb *RigidBody) ClearForce() {
	rb.Force = [3]float32{0, 0, 0}
}

// Update is a no-op; integration is handled by PhysicsSystem.
func (rb *RigidBody) Update(dt float32) {
	_ = dt
}

func (rb *RigidBody) EditorName() string { return "RigidBody" }

func (rb *RigidBody) EditorFields() map[string]any {
	return map[string]any{
		"Mass":  rb.Mass,
		"Vel":   rb.Vel,
		"Force": rb.Force,
	}
}

func (rb *RigidBody) SetEditorField(name string, value any) {
	switch name {
	case "Mass":
		rb.Mass = toFloat32(value)
	case "Vel":
		rb.Vel = toVec3(value)
	case "Force":
		rb.Force = toVec3(value)
	}
}

func toFloat32(v any) float32 {
	switch n := v.(type) {
	case float32:
		return n
	case float64:
		return float32(n)
	case int:
		return float32(n)
	case string:
		f, _ := strconv.ParseFloat(n, 32)
		return float32(f)
	default:
		return 0
	}
}

func toVec3(v any) [3]float32 {
	var out [3]float32

	switch arr := v.(type) {
	case [3]float32:
		return arr
	case []float32:
		copy(out[:], arr)
		return out
	case []float64:
		for i := 0; i < len(arr) && i < 3; i++ {
			out[i] = float32(arr[i])
		}
		return out
	case []any:
		for i := 0; i < len(arr) && i < 3; i++ {
			out[i] = toFloat32(arr[i])
		}
		return out
	default:
		return out
	}
}
