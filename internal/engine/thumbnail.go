package engine

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"image"
	"image/png"
	"log"
	"math"
	"runtime"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

type ThumbnailRenderer struct {
	r  *Renderer
	mm *MeshManager

	// GL objects (live in the thumbnail GL context)
	fbo      uint32
	colorTex uint32
	depthRb  uint32
	width    int
	height   int

	program    uint32
	locModel   int32
	locView    int32
	locProj    int32
	locBaseCol int32
	locUseTex  int32
	locDiffuse int32

	// dedicated GL thread + shared hidden window
	win   *glfw.Window
	reqCh chan thumbRequest
}

type PreviewMaterial struct {
	BaseColor  [4]float32
	UseTexture bool
	TextureID  uint32
	UseNormal  bool
	NormalID   uint32
	// later: roughness, metallic, etc.
}

type thumbRequest struct {
	meshID   string
	meshIDs  []string
	size     int
	material *PreviewMaterial
	resp     chan thumbResponse
}

type thumbResponse struct {
	data []byte
	hash string
	err  error
}

var globalThumbRenderer *ThumbnailRenderer

// Simple thumbnail vertex shader: position/normal/uv, basic MVP.
const thumbVertexShaderSrc = `
#version 330 core

layout(location = 0) in vec3 aPos;
layout(location = 1) in vec3 aNormal;
layout(location = 2) in vec2 aTexCoord;

uniform mat4 uModel;
uniform mat4 uView;
uniform mat4 uProj;

out vec3 FragPos;
out vec3 Normal;
out vec2 TexCoord;

void main()
{
    FragPos = vec3(uModel * vec4(aPos, 1.0));
    Normal  = mat3(transpose(inverse(uModel))) * aNormal;
    TexCoord = aTexCoord;
    gl_Position = uProj * uView * vec4(FragPos, 1.0);
}
`

// Simple thumbnail fragment shader: optional texture, no lights/shadows.
const thumbFragmentShaderSrc = `
#version 330 core

in vec3 FragPos;
in vec3 Normal;
in vec2 TexCoord;

out vec4 FragColor;

uniform vec4 uBaseColor;
uniform sampler2D uDiffuseTex;
uniform bool uUseTexture;

void main()
{
    vec3 base = uBaseColor.rgb;
    if (uUseTexture) {
        base = texture(uDiffuseTex, TexCoord).rgb;
    }

    // Simple directional light
    vec3 lightDir = normalize(vec3(0.4, 0.6, 1.0));
    float diff = max(dot(normalize(Normal), lightDir), 0.0);

    // Ambient + diffuse
    vec3 color = base * (0.25 + diff * 0.75);

    FragColor = vec4(color, uBaseColor.a);
}

`

// InitThumbnailRenderer must be called on the main GL thread
// while the main window/context is current.
func InitThumbnailRenderer(r *Renderer, mm *MeshManager, width, height int) {
	mm.RegisterSphere("__preview_sphere", 64, 64)

	globalThumbRenderer = NewThumbnailRenderer(r, mm, width, height)
}

// Public API: synchronous thumbnail request.
// This runs on any goroutine; actual GL work happens on the dedicated GL thread.
func RenderMeshThumbnail(meshID string, size int) ([]byte, string, error) {
	if globalThumbRenderer == nil {
		return nil, "", fmt.Errorf("thumbnail renderer not initialized")
	}
	return globalThumbRenderer.RenderMeshThumbnail(meshID, size)
}

// Public API: multi-mesh thumbnail (same GL thread, same FBO).
func RenderMeshGroupThumbnail(meshIDs []string, size int) ([]byte, string, error) {
	if globalThumbRenderer == nil {
		return nil, "", fmt.Errorf("thumbnail renderer not initialized")
	}
	return globalThumbRenderer.RenderMeshGroupThumbnail(meshIDs, size)
}

// Public API: material thumbnail
func RenderMaterialThumbnail(mat *PreviewMaterial, size int) ([]byte, string, error) {
	if globalThumbRenderer == nil {
		return nil, "", fmt.Errorf("thumbnail renderer not initialized")
	}
	return globalThumbRenderer.RenderMaterialThumbnail(mat, size)
}

func NewThumbnailRenderer(r *Renderer, mm *MeshManager, width, height int) *ThumbnailRenderer {
	if width <= 0 {
		width = 128
	}
	if height <= 0 {
		height = 128
	}

	tr := &ThumbnailRenderer{
		r:      r,
		mm:     mm,
		width:  width,
		height: height,
		reqCh:  make(chan thumbRequest),
	}

	// Create hidden shared window while main context is current.
	win, err := createHiddenSharedWindow()
	if err != nil {
		log.Printf("thumbnail: failed to create hidden window: %v", err)
		return tr
	}
	tr.win = win

	// Spawn dedicated GL thread that owns this window/context.
	go tr.glThread()

	return tr
}

// createHiddenSharedWindow creates a 1x1 invisible window whose context
// shares objects with the currently-current context (main window).
func createHiddenSharedWindow() (*glfw.Window, error) {
	share := glfw.GetCurrentContext()
	if share == nil {
		return nil, fmt.Errorf("no current GLFW context to share with")
	}

	glfw.WindowHint(glfw.Visible, glfw.False)
	w, err := glfw.CreateWindow(1, 1, "thumb-hidden", nil, share)
	if err != nil {
		return nil, err
	}
	return w, nil
}

// glThread owns the thumbnail GL context and processes all thumbnail requests.
func (tr *ThumbnailRenderer) glThread() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if tr.win == nil {
		log.Printf("thumbnail: no hidden window, GL thread exiting")
		return
	}

	tr.win.MakeContextCurrent()

	// gl.Init is safe to call once per context; main has already called it,
	// but calling again here is harmless in go-gl.
	if err := gl.Init(); err != nil {
		log.Printf("thumbnail: gl.Init failed: %v", err)
		return
	}

	tr.initFBO()
	if err := tr.initProgram(); err != nil {
		log.Printf("thumbnail shader init failed: %v", err)
	}

	log.Printf("thumbnail GL thread started, FBO=%d", tr.fbo)

	for req := range tr.reqCh {
		var (
			data []byte
			hash string
			err  error
		)
		if req.material != nil {
			data, hash, err = tr.renderMaterial(req.material, req.size)

		} else if len(req.meshIDs) > 0 {
			data, hash, err = tr.renderGroup(req.meshIDs, req.size)
		} else {
			data, hash, err = tr.renderOne(req.meshID, req.size)
		}

		req.resp <- thumbResponse{data: data, hash: hash, err: err}
	}

}

// --- FBO / program setup (same as before, but used only on GL thread) ---

func (tr *ThumbnailRenderer) initFBO() {
	gl.GenFramebuffers(1, &tr.fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, tr.fbo)

	// color texture
	gl.GenTextures(1, &tr.colorTex)
	gl.BindTexture(gl.TEXTURE_2D, tr.colorTex)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(tr.width), int32(tr.height), 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, tr.colorTex, 0)

	// depth renderbuffer
	gl.GenRenderbuffers(1, &tr.depthRb)
	gl.BindRenderbuffer(gl.RENDERBUFFER, tr.depthRb)
	gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH_COMPONENT24, int32(tr.width), int32(tr.height))
	gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.RENDERBUFFER, tr.depthRb)

	attachments := []uint32{gl.COLOR_ATTACHMENT0}
	gl.DrawBuffers(1, &attachments[0])
	gl.ReadBuffer(gl.COLOR_ATTACHMENT0)

	status := gl.CheckFramebufferStatus(gl.FRAMEBUFFER)
	if status != gl.FRAMEBUFFER_COMPLETE {
		log.Printf("Thumbnail FBO incomplete: 0x%X", status)
	} else {
		log.Printf("Thumbnail FBO complete, colorTex=%d depthRb=%d", tr.colorTex, tr.depthRb)
	}

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
}

func (tr *ThumbnailRenderer) ensureFBOSize(size int) {
	if size <= 0 {
		return
	}
	if tr.width == size && tr.height == size {
		return
	}
	tr.width = size
	tr.height = size

	gl.BindFramebuffer(gl.FRAMEBUFFER, tr.fbo)

	gl.BindTexture(gl.TEXTURE_2D, tr.colorTex)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(tr.width), int32(tr.height), 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)

	gl.BindRenderbuffer(gl.RENDERBUFFER, tr.depthRb)
	gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH_COMPONENT24, int32(tr.width), int32(tr.height))

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
}

func (tr *ThumbnailRenderer) initProgram() error {
	vs, err := thumbCompileShader(thumbVertexShaderSrc, gl.VERTEX_SHADER)
	if err != nil {
		return fmt.Errorf("compile thumbnail vertex shader: %w", err)
	}
	fs, err := thumbCompileShader(thumbFragmentShaderSrc, gl.FRAGMENT_SHADER)
	if err != nil {
		gl.DeleteShader(vs)
		return fmt.Errorf("compile thumbnail fragment shader: %w", err)
	}
	prog, err := linkProgram(vs, fs)
	gl.DeleteShader(vs)
	gl.DeleteShader(fs)
	if err != nil {
		return fmt.Errorf("link thumbnail program: %w", err)
	}

	tr.program = prog
	tr.locModel = gl.GetUniformLocation(prog, gl.Str("uModel\x00"))
	tr.locView = gl.GetUniformLocation(prog, gl.Str("uView\x00"))
	tr.locProj = gl.GetUniformLocation(prog, gl.Str("uProj\x00"))
	tr.locBaseCol = gl.GetUniformLocation(prog, gl.Str("uBaseColor\x00"))
	tr.locUseTex = gl.GetUniformLocation(prog, gl.Str("uUseTexture\x00"))
	tr.locDiffuse = gl.GetUniformLocation(prog, gl.Str("uDiffuseTex\x00"))

	gl.UseProgram(tr.program)
	gl.Uniform1i(tr.locDiffuse, 0)
	gl.UseProgram(0)

	return nil
}

func (tr *ThumbnailRenderer) RenderMeshGroupThumbnail(meshIDs []string, size int) ([]byte, string, error) {
	if tr.fbo == 0 || tr.program == 0 || tr.reqCh == nil {
		return nil, "", fmt.Errorf("thumbnail renderer not ready (FBO/program/reqCh)")
	}
	if size <= 0 {
		size = 128
	}
	if len(meshIDs) == 0 {
		return nil, "", fmt.Errorf("no meshIDs for group thumbnail")
	}

	respCh := make(chan thumbResponse, 1)
	tr.reqCh <- thumbRequest{
		meshIDs: meshIDs,
		size:    size,
		resp:    respCh,
	}
	resp := <-respCh
	return resp.data, resp.hash, resp.err
}

// --- Public entry: send request to GL thread ---

func (tr *ThumbnailRenderer) RenderMeshThumbnail(meshID string, size int) ([]byte, string, error) {
	if tr.fbo == 0 || tr.program == 0 || tr.reqCh == nil {
		return nil, "", fmt.Errorf("thumbnail renderer not ready (FBO/program/reqCh)")
	}
	if size <= 0 {
		size = 128
	}

	respCh := make(chan thumbResponse, 1)
	tr.reqCh <- thumbRequest{
		meshID: meshID,
		size:   size,
		resp:   respCh,
	}
	resp := <-respCh
	return resp.data, resp.hash, resp.err
}

func (tr *ThumbnailRenderer) RenderMaterialThumbnail(mat *PreviewMaterial, size int) ([]byte, string, error) {
	if tr.fbo == 0 || tr.program == 0 || tr.reqCh == nil {
		return nil, "", fmt.Errorf("thumbnail renderer not ready")
	}

	respCh := make(chan thumbResponse, 1)
	tr.reqCh <- thumbRequest{
		meshID: "__preview_sphere",
		size:   size,
		resp:   respCh,
		// NEW: include material
		material: mat,
	}
	resp := <-respCh
	return resp.data, resp.hash, resp.err
}

// --- Actual GL work, runs only on GL thread ---

func (tr *ThumbnailRenderer) renderOne(meshID string, size int) ([]byte, string, error) {
	if tr.fbo == 0 {
		return nil, "", fmt.Errorf("thumbnail FBO not initialized")
	}
	if tr.program == 0 {
		return nil, "", fmt.Errorf("thumbnail shader program not initialized")
	}

	tr.ensureFBOSize(size)

	var vp [4]int32
	gl.GetIntegerv(gl.VIEWPORT, &vp[0])

	gl.BindFramebuffer(gl.FRAMEBUFFER, tr.fbo)
	gl.ReadBuffer(gl.COLOR_ATTACHMENT0)

	var fb int32
	gl.GetIntegerv(gl.FRAMEBUFFER_BINDING, &fb)

	gl.Viewport(0, 0, int32(size), int32(size))

	gl.Disable(gl.BLEND)
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)

	gl.ClearColor(0.0, 0.0, 0.0, 1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	gl.Finish()
	test := make([]uint8, 4)
	gl.ReadPixels(0, 0, 1, 1, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(test))

	vao := tr.mm.GetVAO(meshID)

	// VAOs are NOT shared → rebuild if needed
	if vao == 0 || !tr.vaoExistsInThisContext(vao) {
		vao = tr.rebuildVAO(meshID)
	}

	count := tr.mm.GetCount(meshID)
	indexType := tr.mm.GetIndexType(meshID)
	vertexCount := tr.mm.GetVertexCount(meshID)
	ebo := tr.mm.GetEBO(meshID)

	if vao == 0 {
		gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
		gl.Viewport(vp[0], vp[1], vp[2], vp[3])
		return nil, "", fmt.Errorf("mesh %s has VAO=0", meshID)
	}

	gl.UseProgram(tr.program)

	_ = [16]float32{
		0.5, 0, 0, 0,
		0, 0.4, 0, 0,
		0, 0, 0.2, 0,
		0, 0, 0, 0.5,
	}

	// --- Simple thumbnail camera ---
	// Projection: 45° FOV, square aspect, near=0.1, far=100
	proj := perspective(45.0*(math.Pi/180.0), 1.0, 0.1, 100.0)

	// View: camera at (0,0,3), looking at origin
	view := lookAt(
		[3]float32{0, 0, 5}, // was 3.0
		[3]float32{0, 0, 0},
		[3]float32{0, 1, 0},
	)

	model := scale(0.60) // was 0.8

	gl.UniformMatrix4fv(tr.locModel, 1, false, &model[0])
	gl.UniformMatrix4fv(tr.locView, 1, false, &view[0])
	gl.UniformMatrix4fv(tr.locProj, 1, false, &proj[0])

	base := [4]float32{1, 1, 1, 1}
	gl.Uniform4fv(tr.locBaseCol, 1, &base[0])
	gl.Uniform1i(tr.locUseTex, 0)

	gl.BindVertexArray(vao)
	if count > 0 && ebo != 0 {
		gl.DrawElements(gl.TRIANGLES, count, indexType, gl.PtrOffset(0))
	} else {
		gl.DrawArrays(gl.TRIANGLES, 0, vertexCount)
	}
	gl.BindVertexArray(0)

	gl.Finish()

	buf := make([]uint8, size*size*4)
	gl.BindFramebuffer(gl.FRAMEBUFFER, tr.fbo)
	gl.ReadBuffer(gl.COLOR_ATTACHMENT0)
	gl.ReadPixels(0, 0, int32(size), int32(size), gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(buf))

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.Viewport(vp[0], vp[1], vp[2], vp[3])

	img := image.NewRGBA(image.Rect(0, 0, size, size))
	rowStride := size * 4
	for y := 0; y < size; y++ {
		srcY := size - 1 - y
		copy(img.Pix[y*rowStride:(y+1)*rowStride], buf[srcY*rowStride:(srcY+1)*rowStride])
	}
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i+3] = 255
	}

	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		return nil, "", err
	}
	data := out.Bytes()
	h := sha1.Sum(data)
	return data, hex.EncodeToString(h[:]), nil
}

func (tr *ThumbnailRenderer) renderGroup(meshIDs []string, size int) ([]byte, string, error) {
	if tr.fbo == 0 {
		return nil, "", fmt.Errorf("thumbnail FBO not initialized")
	}
	if tr.program == 0 {
		return nil, "", fmt.Errorf("thumbnail shader program not initialized")
	}

	tr.ensureFBOSize(size)

	var vp [4]int32
	gl.GetIntegerv(gl.VIEWPORT, &vp[0])

	gl.BindFramebuffer(gl.FRAMEBUFFER, tr.fbo)
	gl.ReadBuffer(gl.COLOR_ATTACHMENT0)

	gl.Viewport(0, 0, int32(size), int32(size))

	gl.Disable(gl.BLEND)
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)

	gl.ClearColor(0.0, 0.0, 0.0, 1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	gl.UseProgram(tr.program)

	proj := perspective(45.0*(math.Pi/180.0), 1.0, 0.1, 100.0)
	view := lookAt(
		[3]float32{0, 0, 5},
		[3]float32{0, 0, 0},
		[3]float32{0, 1, 0},
	)
	model := scale(0.60)

	gl.UniformMatrix4fv(tr.locModel, 1, false, &model[0])
	gl.UniformMatrix4fv(tr.locView, 1, false, &view[0])
	gl.UniformMatrix4fv(tr.locProj, 1, false, &proj[0])

	base := [4]float32{1, 1, 1, 1}
	gl.Uniform4fv(tr.locBaseCol, 1, &base[0])
	gl.Uniform1i(tr.locUseTex, 0)
	gl.Disable(gl.DEPTH_TEST)

	for _, meshID := range meshIDs {
		vao := tr.mm.GetVAO(meshID)
		if vao == 0 || !tr.vaoExistsInThisContext(vao) {
			vao = tr.rebuildVAO(meshID)
		}
		if vao == 0 {
			continue
		}

		count := tr.mm.GetCount(meshID)
		indexType := tr.mm.GetIndexType(meshID)
		vertexCount := tr.mm.GetVertexCount(meshID)
		ebo := tr.mm.GetEBO(meshID)

		gl.BindVertexArray(vao)
		if count > 0 && ebo != 0 {
			gl.DrawElements(gl.TRIANGLES, count, indexType, gl.PtrOffset(0))
		} else {
			gl.DrawArrays(gl.TRIANGLES, 0, vertexCount)
		}
	}

	gl.BindVertexArray(0)
	gl.Finish()

	buf := make([]uint8, size*size*4)
	gl.ReadPixels(0, 0, int32(size), int32(size), gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(buf))

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.Viewport(vp[0], vp[1], vp[2], vp[3])

	img := image.NewRGBA(image.Rect(0, 0, size, size))
	rowStride := size * 4
	for y := 0; y < size; y++ {
		srcY := size - 1 - y
		copy(img.Pix[y*rowStride:(y+1)*rowStride], buf[srcY*rowStride:(srcY+1)*rowStride])
	}
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i+3] = 255
	}

	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		return nil, "", err
	}
	data := out.Bytes()
	h := sha1.Sum(data)
	return data, hex.EncodeToString(h[:]), nil
}

// --- local shader helpers, confined to this file ---

func thumbCompileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)
	csources, free := gl.Strs(source + "\x00")
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLen int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLen)
		logBuf := make([]byte, logLen+1)
		gl.GetShaderInfoLog(shader, logLen, nil, &logBuf[0])
		gl.DeleteShader(shader)
		return 0, fmt.Errorf("shader compile error: %s", string(logBuf))
	}
	return shader, nil
}

func linkProgram(vs, fs uint32) (uint32, error) {
	prog := gl.CreateProgram()
	gl.AttachShader(prog, vs)
	gl.AttachShader(prog, fs)
	gl.LinkProgram(prog)

	var status int32
	gl.GetProgramiv(prog, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLen int32
		gl.GetProgramiv(prog, gl.INFO_LOG_LENGTH, &logLen)
		logBuf := make([]byte, logLen+1)
		gl.GetProgramInfoLog(prog, logLen, nil, &logBuf[0])
		gl.DeleteProgram(prog)
		return 0, fmt.Errorf("program link error: %s", string(logBuf))
	}
	return prog, nil
}

func (tr *ThumbnailRenderer) rebuildVAO(meshID string) uint32 {
	vbo := tr.mm.GetVBO(meshID)
	ebo := tr.mm.GetEBO(meshID)
	layout := tr.mm.GetLayout(meshID)
	if vbo == 0 {
		return 0
	}
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	if ebo != 0 {
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	}

	stride := int32(layout * 4)

	// pos
	gl.VertexAttribPointerWithOffset(0, 3, gl.FLOAT, false, stride, 0)
	gl.EnableVertexAttribArray(0)

	// normal
	gl.VertexAttribPointerWithOffset(1, 3, gl.FLOAT, false, stride, 3*4)
	gl.EnableVertexAttribArray(1)

	// uv
	gl.VertexAttribPointerWithOffset(2, 2, gl.FLOAT, false, stride, 6*4)
	gl.EnableVertexAttribArray(2)

	if layout == 12 {
		gl.VertexAttribPointerWithOffset(3, 4, gl.FLOAT, false, stride, 8*4)
		gl.EnableVertexAttribArray(3)
	}

	gl.BindVertexArray(0)
	return vao
}
func (tr *ThumbnailRenderer) vaoExistsInThisContext(vao uint32) bool {
	var tmp int32
	gl.GetVertexArrayiv(vao, gl.VERTEX_ARRAY_BINDING, &tmp)
	return gl.GetError() == gl.NO_ERROR
}

func scale(s float32) [16]float32 {
	return [16]float32{
		s, 0, 0, 0,
		0, s, 0, 0,
		0, 0, s, 0,
		0, 0, 0, 1,
	}
}

func perspective(fovY, aspect, near, far float32) [16]float32 {
	f := 1.0 / float32(math.Tan(float64(fovY)/2))
	return [16]float32{
		f / aspect, 0, 0, 0,
		0, f, 0, 0,
		0, 0, (far + near) / (near - far), -1,
		0, 0, (2 * far * near) / (near - far), 0,
	}
}

func lookAt(eye, center, up [3]float32) [16]float32 {
	f := normalize(sub(center, eye))
	s := normalize(cross(f, up))
	u := cross(s, f)

	return [16]float32{
		s[0], u[0], -f[0], 0,
		s[1], u[1], -f[1], 0,
		s[2], u[2], -f[2], 0,
		-dot(s, eye), -dot(u, eye), dot(f, eye), 1,
	}
}

func sub(a, b [3]float32) [3]float32 {
	return [3]float32{a[0] - b[0], a[1] - b[1], a[2] - b[2]}
}

func dot(a, b [3]float32) float32 {
	return a[0]*b[0] + a[1]*b[1] + a[2]*b[2]
}

func cross(a, b [3]float32) [3]float32 {
	return [3]float32{
		a[1]*b[2] - a[2]*b[1],
		a[2]*b[0] - a[0]*b[2],
		a[0]*b[1] - a[1]*b[0],
	}
}

func normalize(v [3]float32) [3]float32 {
	l := float32(math.Sqrt(float64(v[0]*v[0] + v[1]*v[1] + v[2]*v[2])))
	return [3]float32{v[0] / l, v[1] / l, v[2] / l}
}

func (tr *ThumbnailRenderer) renderMaterial(mat *PreviewMaterial, size int) ([]byte, string, error) {
	gl.Uniform4fv(tr.locBaseCol, 1, &mat.BaseColor[0])

	if mat.UseTexture && mat.TextureID != 0 {
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, mat.TextureID)
		gl.Uniform1i(tr.locUseTex, 1)
	} else {
		gl.Uniform1i(tr.locUseTex, 0)
	}

	// same sphere rendering as before

	tr.ensureFBOSize(size)

	gl.BindFramebuffer(gl.FRAMEBUFFER, tr.fbo)
	gl.Viewport(0, 0, int32(size), int32(size))

	gl.Disable(gl.BLEND)
	gl.Enable(gl.DEPTH_TEST)
	gl.ClearColor(0.1, 0.1, 0.1, 1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	gl.UseProgram(tr.program)

	// Camera
	proj := perspective(45*(math.Pi/180), 1.0, 0.1, 100.0)

	// Zoomed out + tilted down 20 degrees
	eye := [3]float32{0, 0.35, 3.7}
	center := [3]float32{0, 0, 0}
	up := [3]float32{0, 1, 0}

	view := lookAt(eye, center, up)
	model := scale(1.25)

	gl.UniformMatrix4fv(tr.locModel, 1, false, &model[0])
	gl.UniformMatrix4fv(tr.locView, 1, false, &view[0])
	gl.UniformMatrix4fv(tr.locProj, 1, false, &proj[0])

	// Material uniforms
	gl.Uniform4fv(tr.locBaseCol, 1, &mat.BaseColor[0])

	if mat.UseTexture && mat.TextureID != 0 {
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, uint32(mat.TextureID))
		gl.Uniform1i(tr.locUseTex, 1)
	} else {
		gl.Uniform1i(tr.locUseTex, 0)
	}
	vao := tr.mm.GetVAO("__preview_sphere")
	if vao == 0 || !tr.vaoExistsInThisContext(vao) {
		vao = tr.rebuildVAO("__preview_sphere")
	}

	// Draw sphere

	count := tr.mm.GetCount("__preview_sphere")
	indexType := tr.mm.GetIndexType("__preview_sphere")
	_ = tr.mm.GetEBO("__preview_sphere")

	gl.BindVertexArray(vao)
	gl.DrawElements(gl.TRIANGLES, count, indexType, gl.PtrOffset(0))
	gl.BindVertexArray(0)

	// Read pixels → PNG
	buf := make([]uint8, size*size*4)
	gl.ReadPixels(0, 0, int32(size), int32(size), gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(buf))

	// Flip vertically
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	row := size * 4
	for y := 0; y < size; y++ {
		copy(img.Pix[y*row:(y+1)*row], buf[(size-1-y)*row:(size-y)*row])
	}

	var out bytes.Buffer
	png.Encode(&out, img)
	data := out.Bytes()
	hash := sha1Hex(data)

	return data, hash, nil
}
func sha1Hex(data []byte) string {
	h := sha1.Sum(data)
	return hex.EncodeToString(h[:])
}
