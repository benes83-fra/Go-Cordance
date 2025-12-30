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
