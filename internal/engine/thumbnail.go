package engine

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"image"
	"image/png"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type ThumbnailRenderer struct {
	r        *Renderer
	mm       *MeshManager
	fbo      uint32
	colorTex uint32
	depthRb  uint32
	width    int
	height   int
}

// engine/thumbnail.go

var globalThumbRenderer *ThumbnailRenderer

func InitThumbnailRenderer(r *Renderer, mm *MeshManager, width, height int) {
	globalThumbRenderer = NewThumbnailRenderer(r, mm, width, height)
}

// RenderMeshThumbnail is the engine-level hook used by thumbnails.
func RenderMeshThumbnail(meshID string, size int) ([]byte, string, error) {
	if globalThumbRenderer == nil {
		return nil, "", fmt.Errorf("thumbnail renderer not initialized")
	}
	return globalThumbRenderer.RenderMeshThumbnail(meshID, size)
}

func NewThumbnailRenderer(r *Renderer, mm *MeshManager, width, height int) *ThumbnailRenderer {
	tr := &ThumbnailRenderer{
		r:      r,
		mm:     mm,
		width:  width,
		height: height,
	}
	tr.initFBO()
	return tr
}

func (tr *ThumbnailRenderer) initFBO() {
	gl.GenFramebuffers(1, &tr.fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, tr.fbo)

	// color texture
	gl.GenTextures(1, &tr.colorTex)
	gl.BindTexture(gl.TEXTURE_2D, tr.colorTex)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(tr.width), int32(tr.height), 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, tr.colorTex, 0)

	// depth renderbuffer
	gl.GenRenderbuffers(1, &tr.depthRb)
	gl.BindRenderbuffer(gl.RENDERBUFFER, tr.depthRb)
	gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH_COMPONENT24, int32(tr.width), int32(tr.height))
	gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.RENDERBUFFER, tr.depthRb)

	if status := gl.CheckFramebufferStatus(gl.FRAMEBUFFER); status != gl.FRAMEBUFFER_COMPLETE {
		fmt.Printf("Thumbnail FBO incomplete: 0x%X\n", status)
	}

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
}

func (tr *ThumbnailRenderer) RenderMeshThumbnail(meshID string, size int) ([]byte, string, error) {
	if tr.fbo == 0 {
		return nil, "", fmt.Errorf("thumbnail FBO not initialized")
	}

	// Save current viewport
	var vp [4]int32
	gl.GetIntegerv(gl.VIEWPORT, &vp[0])

	gl.BindFramebuffer(gl.FRAMEBUFFER, tr.fbo)
	gl.Viewport(0, 0, int32(size), int32(size))
	gl.ClearColor(0.2, 0.2, 0.2, 1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	// Simple camera: look at origin from +Z
	view := mgl32.LookAtV(
		mgl32.Vec3{0, 0, 3},
		mgl32.Vec3{0, 0, 0},
		mgl32.Vec3{0, 1, 0},
	)
	proj := mgl32.Perspective(mgl32.DegToRad(45), 1.0, 0.1, 100.0)
	model := mgl32.Ident4()

	gl.UseProgram(tr.r.Program)
	gl.UniformMatrix4fv(tr.r.LocView, 1, false, &view[0])
	gl.UniformMatrix4fv(tr.r.LocProj, 1, false, &proj[0])
	gl.UniformMatrix4fv(tr.r.LocModel, 1, false, &model[0])

	// basic material: white
	base := [4]float32{1, 1, 1, 1}
	gl.Uniform4fv(tr.r.LocBaseCol, 1, &base[0])
	gl.Uniform1i(tr.r.LocUseTexture, 0)

	vao := tr.mm.GetVAO(meshID)
	count := tr.mm.GetCount(meshID)
	indexType := tr.mm.GetIndexType(meshID)
	vertexCount := tr.mm.GetVertexCount(meshID)
	ebo := tr.mm.GetEBO(meshID)

	if vao == 0 {
		gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
		gl.Viewport(vp[0], vp[1], vp[2], vp[3])
		return nil, "", fmt.Errorf("mesh %s has VAO=0", meshID)
	}

	gl.BindVertexArray(vao)
	if count > 0 && ebo != 0 {
		gl.DrawElements(gl.TRIANGLES, count, indexType, gl.PtrOffset(0))
	} else {
		gl.DrawArrays(gl.TRIANGLES, 0, vertexCount)
	}
	gl.BindVertexArray(0)

	// Read back pixels
	buf := make([]uint8, size*size*4)
	gl.ReadPixels(0, 0, int32(size), int32(size), gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(buf))

	// Restore framebuffer + viewport
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.Viewport(vp[0], vp[1], vp[2], vp[3])

	// Convert to image (flip vertically)
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	rowStride := size * 4
	for y := 0; y < size; y++ {
		srcY := size - 1 - y
		copy(img.Pix[y*rowStride:(y+1)*rowStride], buf[srcY*rowStride:(srcY+1)*rowStride])
	}

	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		return nil, "", err
	}
	data := out.Bytes()
	h := sha1.Sum(data)
	hash := hex.EncodeToString(h[:])

	return data, hash, nil
}
