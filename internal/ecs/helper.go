package ecs

import (
	"math"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

func RayCapsuleIntersect(rayOrigin, rayDir, a, b mgl32.Vec3, radius float32) (hit bool, dist float32) {
	ab := b.Sub(a)
	ao := rayOrigin.Sub(a)

	abDotAb := ab.Dot(ab)
	abDotAo := ab.Dot(ao)
	abDotDir := ab.Dot(rayDir)

	m := abDotAb
	n := abDotDir
	k := abDotAo

	// Solve quadratic for distance to infinite cylinder
	q := ao.Sub(ab.Mul(k / m))
	r := rayDir.Sub(ab.Mul(n / m))

	A := r.Dot(r)
	B := 2 * q.Dot(r)
	C := q.Dot(q) - radius*radius

	disc := B*B - 4*A*C
	if disc < 0 {
		return false, 0
	}

	t := (-B - float32(math.Sqrt(float64(disc)))) / (2 * A)
	if t < 0 {
		return false, 0
	}

	// Check if projected point lies within segment
	proj := k + t*n
	if proj < 0 || proj > m {
		return false, 0
	}

	return true, t
}

func RayFromMouse(window *glfw.Window, cam *CameraSystem) (origin, dir mgl32.Vec3) {
	w, h := window.GetSize()
	mx, my := window.GetCursorPos()

	// Convert to Normalized Device Coordinates
	x := float32((2.0*mx)/float64(w) - 1.0)
	y := float32(1.0 - (2.0*my)/float64(h)) // flip Y

	ndc := mgl32.Vec4{x, y, -1, 1}

	invProj := cam.Projection.Inv()
	invView := cam.View.Inv()

	// View space
	viewSpace := invProj.Mul4x1(ndc)
	viewSpace = mgl32.Vec4{viewSpace.X(), viewSpace.Y(), -1, 0}

	// World space
	worldSpace := invView.Mul4x1(viewSpace)

	dir = mgl32.Vec3{worldSpace.X(), worldSpace.Y(), worldSpace.Z()}.Normalize()
	origin = cam.Position

	return
}
func projectRayOntoAxis(rayOrigin, rayDir, axisOrigin, axisDir mgl32.Vec3) float32 {
	// Solve for t where ray intersects axis direction
	// t = dot((rayOrigin - axisOrigin), axisDir) / dot(rayDir, axisDir)
	denom := rayDir.Dot(axisDir)
	if float32(math.Abs(float64(denom))) < 1e-6 {
		return 0
	}
	return (rayOrigin.Sub(axisOrigin)).Dot(axisDir) / denom
}

func RayPlaneIntersection(rayOrigin, rayDir, planePoint, planeNormal mgl32.Vec3) (hit bool, t float32) {
	denom := planeNormal.Dot(rayDir)
	if float32(math.Abs(float64(denom))) < 1e-6 {
		return false, 0
	}
	t = planePoint.Sub(rayOrigin).Dot(planeNormal) / denom
	return t >= 0, t
}
func RayCircleIntersect(rayOrigin, rayDir, center, normal mgl32.Vec3, radius, thickness float32) (bool, float32) {
	denom := normal.Dot(rayDir)
	if float32(math.Abs(float64(denom))) < 1e-6 {
		return false, 0
	}

	t := center.Sub(rayOrigin).Dot(normal) / denom
	if t < 0 {
		return false, 0
	}

	hitPoint := rayOrigin.Add(rayDir.Mul(t))
	dist := hitPoint.Sub(center).Len()

	if float32(math.Abs(float64(dist-radius))) <= thickness {
		return true, t
	}
	return false, 0
}

func degToRad(d float32) float32 {
	return d * (math.Pi / 180)
}
func radToDeg(r float32) float32 {
	return r * (180 / math.Pi)
}
func SnapAngle(angle float32, increment float32) float32 {
	return float32(math.Round(float64(angle)/float64(increment))) * increment
}
func SnapPosition(pos mgl32.Vec3, increment float32) mgl32.Vec3 {
	return mgl32.Vec3{
		float32(math.Round(float64(pos.X())/float64(increment))) * increment,
		float32(math.Round(float64(pos.Y())/float64(increment))) * increment,
		float32(math.Round(float64(pos.Z())/float64(increment))) * increment,
	}
}
func MinFloat32(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}
func MaxFloat32(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}
func ClampFloat32(val, min, max float32) float32 {
	return MaxFloat32(min, MinFloat32(max, val))
}
func rotationFromAxis(axis mgl32.Vec3) mgl32.Mat4 {
	z := mgl32.Vec3{0, 0, 1}
	dot := z.Dot(axis)
	if dot > 0.9999 {
		return mgl32.Ident4()
	}
	if dot < -0.9999 {
		return mgl32.HomogRotate3D(float32(math.Pi), mgl32.Vec3{1, 0, 0})
	}
	rotAxis := z.Cross(axis).Normalize()
	angle := float32(math.Acos(float64(dot)))
	return mgl32.HomogRotate3D(angle, rotAxis)
}
