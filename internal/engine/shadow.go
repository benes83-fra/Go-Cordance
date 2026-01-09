package engine

import "github.com/go-gl/mathgl/mgl32"

// ComputeDirectionalLightSpaceMatrix computes an orthographic light-space matrix.
// lightDir is world-space direction (pointing from light toward scene, e.g. {0,-1,0}).
// sceneCenter is the center of the region you want to shadow.
// extent is half-size of the orthographic box (shadowExtent).
func ComputeDirectionalLightSpaceMatrix(lightDir mgl32.Vec3, sceneCenter mgl32.Vec3, extent float32) mgl32.Mat4 {
	dir := lightDir.Normalize()
	// place the light camera some distance back along the light direction
	lightDistance := float32(50.0)
	lightPos := sceneCenter.Sub(dir.Mul(lightDistance))

	lightView := mgl32.LookAtV(lightPos, sceneCenter, mgl32.Vec3{0, 1, 0})

	left, right := -extent, extent
	bottom, top := -extent, extent
	near, far := float32(1.0), float32(200.0)

	lightProj := mgl32.Ortho(left, right, bottom, top, near, far)

	return lightProj.Mul4(lightView)
}
