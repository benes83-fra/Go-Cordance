package engine

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
)

// ---------------------------
// glTF 2.0 minimal structs
// ---------------------------

type gltfBuffer struct {
	ByteLength int    `json:"byteLength"`
	URI        string `json:"uri"`
}

type gltfBufferView struct {
	Buffer     int `json:"buffer"`
	ByteOffset int `json:"byteOffset"`
	ByteLength int `json:"byteLength"`
	ByteStride int `json:"byteStride"`
	Target     int `json:"target"`
}

type gltfAccessor struct {
	BufferView    int    `json:"bufferView"`
	ByteOffset    int    `json:"byteOffset"`
	ComponentType int    `json:"componentType"`
	Count         int    `json:"count"`
	Type          string `json:"type"`
}

type gltfPrimitive struct {
	Attributes map[string]int `json:"attributes"`
	Indices    int            `json:"indices"`
	Material   int            `json:"material"`
}

type gltfMesh struct {
	Name       string          `json:"name"`
	Primitives []gltfPrimitive `json:"primitives"`
}

// in engine.go (gltf structs)
type gltfTextureInfo struct {
	Index      int                        `json:"index"`
	TexCoord   int                        `json:"texCoord,omitempty"`
	Scale      float32                    `json:"scale,omitempty"` // normalTexture.scale
	Extensions map[string]json.RawMessage `json:"extensions,omitempty"`
}
type gltfPBR struct {
	BaseColorFactor          []float32        `json:"baseColorFactor"`
	BaseColorTexture         *gltfTextureInfo `json:"baseColorTexture"`
	MetallicRoughnessTexture *gltfTextureInfo `json:"metallicRoughnessTexture"`
	RoughnessFactor          float32          `json:"roughnessFactor,omitempty"`
}

type gltfMaterial struct {
	Name             string                     `json:"name"`
	PBR              gltfPBR                    `json:"pbrMetallicRoughness"`
	NormalTexture    *gltfTextureInfo           `json:"normalTexture"`
	OcclusionTexture *gltfTextureInfo           `json:"occlusionTexture"`
	Extensions       map[string]json.RawMessage `json:"extensions,omitempty"`
	AlphaMode        string                     `json:"alphaMode,omitempty"`
}

type gltfImage struct {
	URI string `json:"uri"`
}

// engine.go (gltf structs)
type gltfTexture struct {
	Source     int                        `json:"source,omitempty"`
	Extensions map[string]json.RawMessage `json:"extensions,omitempty"`
}
type gltfAnimationSampler struct {
	Input         int    `json:"input"`
	Output        int    `json:"output"`
	Interpolation string `json:"interpolation"`
}

type gltfAnimationChannelTarget struct {
	Node int    `json:"node"`
	Path string `json:"path"`
}

type gltfAnimationChannel struct {
	Sampler int                        `json:"sampler"`
	Target  gltfAnimationChannelTarget `json:"target"`
}

type gltfAnimation struct {
	Name     string                 `json:"name"`
	Samplers []gltfAnimationSampler `json:"samplers"`
	Channels []gltfAnimationChannel `json:"channels"`
}

type gltfSkin struct {
	Joints              []int `json:"joints"`
	InverseBindMatrices int   `json:"inverseBindMatrices"`
}

// helper: return the image source index for a texture, checking EXT_texture_webp
func textureSourceIndex(t gltfTexture) int {
	// prefer explicit Source if present
	if t.Source != 0 {
		return t.Source
	}
	// check EXT_texture_webp extension: {"EXT_texture_webp": {"source": <int>}}
	if t.Extensions != nil {
		if raw, ok := t.Extensions["EXT_texture_webp"]; ok {
			var ext struct {
				Source int `json:"source"`
			}
			if err := json.Unmarshal(raw, &ext); err == nil {
				return ext.Source
			}
		}
	}
	// fallback: -1 (not found)
	return -1
}

type gltfNode struct {
	Name        string    `json:"name"`
	Mesh        int       `json:"mesh"`
	Children    []int     `json:"children"`
	Translation []float32 `json:"translation"`
	Rotation    []float32 `json:"rotation"` // quaternion
	Scale       []float32 `json:"scale"`
	Matrix      []float32 `json:"matrix"` // 16 floats
	Skin        int       `json:"skin"`   // NEW: index into gltfRoot.Skins, or -1
}

type gltfScene struct {
	Nodes []int `json:"nodes"`
}

type gltfRoot struct {
	Buffers     []gltfBuffer     `json:"buffers"`
	BufferViews []gltfBufferView `json:"bufferViews"`
	Accessors   []gltfAccessor   `json:"accessors"`
	Meshes      []gltfMesh       `json:"meshes"`
	Materials   []gltfMaterial   `json:"materials"`
	Images      []gltfImage      `json:"images"`
	Textures    []gltfTexture    `json:"textures"`
	Animations  []gltfAnimation  `json:"animations"`
	Nodes       []gltfNode       `json:"nodes"`
	Scenes      []gltfScene      `json:"scenes"`
	Scene       int              `json:"scene"`
	Skins       []gltfSkin       `json:"skins"` // default scene index
}

// ---------------------------
// Helpers
// ---------------------------

func componentByteSize(typ string, comp int) int {
	var csize int
	switch comp {
	case 5123: // UNSIGNED_SHORT
		csize = 2
	case 5125: // UNSIGNED_INT
		csize = 4
	case 5126: // FLOAT
		csize = 4
	default:
		panic(fmt.Sprintf("unsupported component type: %d", comp))
	}

	switch typ {
	case "SCALAR":
		return csize
	case "VEC2":
		return csize * 2
	case "VEC3":
		return csize * 3
	case "VEC4":
		return csize * 4
	case "MAT4":
		return csize * 16 // <‑‑ ADD THIS
	default:
		panic(fmt.Sprintf("unsupported accessor type: %s", typ))
	}
}

func BytesToFloat32(b []byte) float32 {
	return math.Float32frombits(
		uint32(b[0]) |
			uint32(b[1])<<8 |
			uint32(b[2])<<16 |
			uint32(b[3])<<24)
}

// ---------------------------
// Core accessor reader
// ---------------------------

type AccessorData struct {
	Acc    gltfAccessor
	Bv     gltfBufferView
	Buf    []byte
	Base   int
	Stride int
}

func GetAccessor(g *gltfRoot, buffers [][]byte, idx int) (AccessorData, error) {
	if idx < 0 || idx >= len(g.Accessors) {
		return AccessorData{}, fmt.Errorf("accessor index out of range: %d", idx)
	}
	Acc := g.Accessors[idx]

	if Acc.BufferView < 0 || Acc.BufferView >= len(g.BufferViews) {
		return AccessorData{}, fmt.Errorf("bufferView index out of range: %d", Acc.BufferView)
	}
	Bv := g.BufferViews[Acc.BufferView]

	if Bv.Buffer < 0 || Bv.Buffer >= len(buffers) {
		return AccessorData{}, fmt.Errorf("buffer index out of range: %d", Bv.Buffer)
	}
	Buf := buffers[Bv.Buffer]

	elemSize := componentByteSize(Acc.Type, Acc.ComponentType)
	Stride := Bv.ByteStride
	if Stride == 0 {
		Stride = elemSize
	}

	Base := Bv.ByteOffset + Acc.ByteOffset
	end := Base + Acc.Count*Stride
	if end > len(Buf) {
		return AccessorData{}, fmt.Errorf("accessor out of range: end=%d len=%d", end, len(Buf))
	}

	return AccessorData{Acc, Bv, Buf, Base, Stride}, nil
}

// ---------------------------
// Upload to OpenGL
// ---------------------------

func uploadMeshToGL(mm *MeshManager, id string, vertices []float32, indices []uint32) {
	var vao, vbo, ebo uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)
	gl.GenBuffers(1, &ebo)

	gl.BindVertexArray(vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	// The importer builds vertices with this interleaved layout:
	// pos(3), normal(3), uv(2), tangent(4) => 12 floats per vertex
	Stride := int32(12 * 4)

	gl.VertexAttribPointerWithOffset(0, 3, gl.FLOAT, false, Stride, 0)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointerWithOffset(1, 3, gl.FLOAT, false, Stride, 3*4)
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointerWithOffset(2, 2, gl.FLOAT, false, Stride, 6*4)
	gl.EnableVertexAttribArray(2)
	gl.VertexAttribPointerWithOffset(3, 4, gl.FLOAT, false, Stride, 8*4)
	gl.EnableVertexAttribArray(3)
	gl.BindVertexArray(0)

	// Store GL objects and counts in MeshManager
	mm.vaos[id] = vao
	mm.vbos[id] = vbo
	mm.ebos[id] = ebo
	mm.counts[id] = int32(len(indices))

	// Bookkeeping: index type and vertex count for reliable DrawElements
	mm.indexTypes[id] = gl.UNSIGNED_INT
	// 12 floats per vertex as above
	mm.vertexCounts[id] = int32(len(vertices) / 12)
	mm.layoutType[id] = 12

	// Verify EBO size (warn if mismatch)
	var eboSize int32
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.GetBufferParameteriv(gl.ELEMENT_ARRAY_BUFFER, gl.BUFFER_SIZE, &eboSize)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, 0)
	expected := int32(len(indices) * 4) // uint32 indices -> 4 bytes each
	if eboSize != expected {
		// lightweight warning; replace with log.Printf if you prefer
		println("Warning: EBO size mismatch for mesh", id, "got", eboSize, "expected", expected)
	}
	// compute vertexCount from interleaved layout (12 floats per vertex for importer)
	vertexCount := int32(len(vertices) / 12)

	// validate max index
	var maxIdx uint32 = 0
	for _, idx := range indices {
		if idx > maxIdx {
			maxIdx = idx
		}
	}
	if int(maxIdx) >= int(vertexCount) {
		log.Printf("ERROR: mesh %s has max index %d >= vertexCount %d — aborting upload", id, maxIdx, vertexCount)
		// Option A: return an error so caller can fix the mesh
		// return fmt.Errorf("mesh %s: max index %d >= vertexCount %d", id, maxIdx, vertexCount)
		// Option B: clamp or convert indices (not recommended silently). For now, abort upload.
	}

}

func LoadGLTFOrGLB(path string) (*gltfRoot, [][]byte, error) {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".glb":
		// Use your existing GLB loader
		g, buffers, err := loadGLB(path)
		if err != nil {
			return nil, nil, err
		}

		// GLB always has exactly one BIN buffer
		return g, buffers, nil

	case ".gltf":
		// Standard JSON glTF
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, err
		}

		var g gltfRoot
		if err := json.Unmarshal(raw, &g); err != nil {
			return nil, nil, err
		}

		// Load external buffers
		baseDir := filepath.Dir(path)
		buffers := make([][]byte, len(g.Buffers))

		for i, b := range g.Buffers {
			bufPath := filepath.Join(baseDir, b.URI)
			data, err := os.ReadFile(bufPath)
			if err != nil {
				return nil, nil, err
			}
			buffers[i] = data
		}

		return &g, buffers, nil

	default:
		return nil, nil, fmt.Errorf("unsupported mesh format: %s", ext)
	}
}

// ---------------------------
// Single-mesh loader (default)
// ---------------------------

func (mm *MeshManager) RegisterGLTF(id, path string) ([]string, error) {
	return mm.loadGLTFInternal(id, path, false)
}

// ---------------------------
// Multi-mesh loader (optional)
// ---------------------------

func (mm *MeshManager) RegisterGLTFMulti(path string) ([]string, error) {
	return mm.loadGLTFInternal("", path, true)
}

// ---------------------------
// Shared geometry loader
// ---------------------------

func (mm *MeshManager) loadGLTFInternal(id, path string, multi bool) ([]string, error) {

	var meshIDs []string
	g, buffers, err := LoadGLTFOrGLB(path)
	if err != nil {
		return nil, err
	}

	if strings.ToLower(filepath.Ext(path)) == ".gltf" {
		// load external buffers
		baseDir := filepath.Dir(path)
		buffers = make([][]byte, len(g.Buffers))
		for i, b := range g.Buffers {
			data, err := os.ReadFile(filepath.Join(baseDir, b.URI))
			if err != nil {
				return nil, err
			}
			buffers[i] = data
		}
	}
	// Build world transforms for all nodes
	nodeWorld := make([][16]float32, len(g.Nodes))

	var compute func(i int) [16]float32
	compute = func(i int) [16]float32 {
		if nodeWorld[i] != ([16]float32{}) {
			return nodeWorld[i]
		}
		local := composeNodeTransform(g.Nodes[i])

		// parent multiply
		for _, root := range g.Scenes[g.Scene].Nodes {
			if root == i {
				nodeWorld[i] = local
				return local
			}
		}
		// find parent
		for p, n := range g.Nodes {
			for _, c := range n.Children {
				if c == i {
					parent := compute(p)
					nodeWorld[i] = MulMat4(parent, local)
					return nodeWorld[i]
				}
			}
		}
		nodeWorld[i] = local
		return local
	}

	// Loop meshes
	for mi, mesh := range g.Meshes {
		meshName := mesh.Name
		if meshName == "" {
			meshName = fmt.Sprintf("mesh_%d", mi)
		}

		// Loop primitives
		for pi, prim := range mesh.Primitives {

			// If single-mesh mode: only load first primitive
			if !multi && (mi != 0 || pi != 0) {
				continue
			}

			// Build ID
			meshID := id
			if multi {
				meshID = fmt.Sprintf("%s/%d", meshName, pi)
			}
			meshIDs = append(meshIDs, meshID)

			// POSITION
			posA, err := GetAccessor(g, buffers, prim.Attributes["POSITION"])
			if err != nil {
				return nil, err
			}

			count := posA.Acc.Count

			// NORMAL
			norA, err := GetAccessor(g, buffers, prim.Attributes["NORMAL"])
			if err != nil {
				return nil, err
			}

			// UV (optional)
			var uvA AccessorData
			hasUV := false
			if uvIdx, ok := prim.Attributes["TEXCOORD_0"]; ok {
				uvA, err = GetAccessor(g, buffers, uvIdx)
				if err != nil {
					return nil, err
				}
				hasUV = true
			}

			// TANGENT (optional)
			var tanA AccessorData
			hasTan := false
			if tanIdx, ok := prim.Attributes["TANGENT"]; ok {
				tanA, err = GetAccessor(g, buffers, tanIdx)
				if err != nil {
					return nil, err
				}
				hasTan = true
			}
			// JOINTS_0 (optional)
			var jointsA AccessorData
			hasJoints := false
			if jIdx, ok := prim.Attributes["JOINTS_0"]; ok {
				jointsA, err = GetAccessor(g, buffers, jIdx)
				if err == nil {
					hasJoints = true
				}
			}

			// WEIGHTS_0 (optional)
			var weightsA AccessorData
			hasWeights := false
			if wIdx, ok := prim.Attributes["WEIGHTS_0"]; ok {
				weightsA, err = GetAccessor(g, buffers, wIdx)
				if err == nil {
					hasWeights = true
				}
			}

			// INDICES
			idxA, err := GetAccessor(g, buffers, prim.Indices)
			if err != nil {
				return nil, err
			}

			// Decode indices
			indices := make([]uint32, idxA.Acc.Count)
			switch idxA.Acc.ComponentType {
			case 5123: // UNSIGNED_SHORT
				for i := 0; i < idxA.Acc.Count; i++ {
					off := idxA.Base + i*idxA.Stride
					b := idxA.Buf[off : off+2]
					indices[i] = uint32(b[0]) | uint32(b[1])<<8
				}
			case 5125: // UNSIGNED_INT
				for i := 0; i < idxA.Acc.Count; i++ {
					off := idxA.Base + i*idxA.Stride
					b := idxA.Buf[off : off+4]
					indices[i] = uint32(b[0]) |
						uint32(b[1])<<8 |
						uint32(b[2])<<16 |
						uint32(b[3])<<24
				}
			default:
				return nil, fmt.Errorf("unsupported index type: %d", idxA.Acc.ComponentType)
			}

			// Build interleaved vertices
			vertices := make([]float32, 0, count*12)

			// Build interleaved vertices

			for i := 0; i < count; i++ {
				// POSITION
				pOff := posA.Base + i*posA.Stride
				px := BytesToFloat32(posA.Buf[pOff+0:])
				py := BytesToFloat32(posA.Buf[pOff+4:])
				pz := BytesToFloat32(posA.Buf[pOff+8:])

				// NORMAL
				nOff := norA.Base + i*norA.Stride
				nx := BytesToFloat32(norA.Buf[nOff+0:])
				ny := BytesToFloat32(norA.Buf[nOff+4:])
				nz := BytesToFloat32(norA.Buf[nOff+8:])

				// UV
				var u, v float32
				if hasUV {
					uvOff := uvA.Base + i*uvA.Stride
					u = BytesToFloat32(uvA.Buf[uvOff+0:])
					v = BytesToFloat32(uvA.Buf[uvOff+4:])
				}

				// TANGENT
				tx, ty, tz, tw := float32(1), float32(0), float32(0), float32(1)
				if hasTan {
					tOff := tanA.Base + i*tanA.Stride
					tx = BytesToFloat32(tanA.Buf[tOff+0:])
					ty = BytesToFloat32(tanA.Buf[tOff+4:])
					tz = BytesToFloat32(tanA.Buf[tOff+8:])
					tw = BytesToFloat32(tanA.Buf[tOff+12:])
				}
				if hasJoints {
					js := make([][4]uint16, count)
					for i := 0; i < count; i++ {
						off := jointsA.Base + i*jointsA.Stride
						js[i] = [4]uint16{
							uint16(jointsA.Buf[off+0]),
							uint16(jointsA.Buf[off+1]),
							uint16(jointsA.Buf[off+2]),
							uint16(jointsA.Buf[off+3]),
						}
					}
					mm.JointData[meshID] = js
				}
				if hasWeights {
					ws := make([][4]float32, count)
					for i := 0; i < count; i++ {
						off := weightsA.Base + i*weightsA.Stride
						ws[i] = [4]float32{
							BytesToFloat32(weightsA.Buf[off+0:]),
							BytesToFloat32(weightsA.Buf[off+4:]),
							BytesToFloat32(weightsA.Buf[off+8:]),
							BytesToFloat32(weightsA.Buf[off+12:]),
						}
					}
					mm.WeightData[meshID] = ws
				}

				vertices = append(vertices,
					px, py, pz,
					nx, ny, nz,
					u, v,
					tx, ty, tz, tw,
				)
			}

			// Upload
			uploadMeshToGL(mm, meshID, vertices, indices)
		}

	}

	return meshIDs, nil

}

// ---------------------------
// Material metadata helpers
// ---------------------------

// LoadedMeshMaterial holds material info per meshID,
// to be mapped onto ecs.Material + texture components by the caller.
type LoadedMeshMaterial struct {
	MeshID                       string
	BaseColor                    [4]float32
	DiffuseTexturePath           string
	NormalTexturePath            string
	OcclusionTexturePath         string
	MetallicRoughnessTexturePath string

	TexCoordMap map[string]int
	UVScale     map[string][2]float32
	UVOffset    map[string][2]float32

	NormalScale    float32
	SheenColor     [3]float32
	SheenRoughness float32
	SpecularFactor float32
}

// LoadGLTFMaterials returns material info for the first mesh/primitive,
// matching RegisterGLTF(id, path).
func LoadGLTFMaterials(id, path string) ([]LoadedMeshMaterial, error) {
	return loadGLTFMaterialsInternal(id, path, false)
}

// LoadGLTFMaterialsMulti returns material info for all meshes/primitives,
// matching RegisterGLTFMulti(path).
func LoadGLTFMaterialsMulti(path string) ([]LoadedMeshMaterial, error) {
	return loadGLTFMaterialsInternal("", path, true)
}

func loadGLTFMaterialsInternal(id, path string, multi bool) ([]LoadedMeshMaterial, error) {
	g, _, err := LoadGLTFOrGLB(path)
	if err != nil {
		return nil, err
	}

	baseDir := filepath.Dir(path)

	var results []LoadedMeshMaterial

	for mi, mesh := range g.Meshes {
		meshName := mesh.Name
		if meshName == "" {
			meshName = fmt.Sprintf("mesh_%d", mi)
		}

		for pi, prim := range mesh.Primitives {
			if !multi && (mi != 0 || pi != 0) {
				continue
			}

			meshID := id
			if multi {
				meshID = fmt.Sprintf("%s/%d", meshName, pi)
			}

			m := LoadedMeshMaterial{
				MeshID:    meshID,
				BaseColor: [4]float32{1, 1, 1, 1},
			}

			// inside loadGLTFMaterialsInternal, replace the existing "if prim.Material >= 0 ..." block
			if prim.Material >= 0 && prim.Material < len(g.Materials) {
				gm := g.Materials[prim.Material]

				// BaseColorFactor
				if len(gm.PBR.BaseColorFactor) == 4 {
					m.BaseColor = [4]float32{
						gm.PBR.BaseColorFactor[0],
						gm.PBR.BaseColorFactor[1],
						gm.PBR.BaseColorFactor[2],
						gm.PBR.BaseColorFactor[3],
					}
				}

				// helper to resolve texture index -> image path
				// inside loadGLTFMaterialsInternal, replace resolveTex with:
				resolveTex := func(ti *gltfTextureInfo) string {
					if ti == nil {
						return ""
					}
					if ti.Index < 0 || ti.Index >= len(g.Textures) {
						return ""
					}
					tex := g.Textures[ti.Index]
					imgIndex := textureSourceIndex(tex)
					if imgIndex >= 0 && imgIndex < len(g.Images) {
						return filepath.Join(baseDir, g.Images[imgIndex].URI)
					}
					return ""
				}

				// Base color (diffuse)
				if gm.PBR.BaseColorTexture != nil {
					m.DiffuseTexturePath = resolveTex(gm.PBR.BaseColorTexture)
					if m.TexCoordMap == nil {
						m.TexCoordMap = map[string]int{}
					}
					m.TexCoordMap["baseColor"] = gm.PBR.BaseColorTexture.TexCoord
					// parse KHR_texture_transform if present
					if ext, ok := gm.PBR.BaseColorTexture.Extensions["KHR_texture_transform"]; ok {
						off, scale, _, err := parseTextureTransform(ext)
						if err == nil {
							if m.UVScale == nil {
								m.UVScale = map[string][2]float32{}
							}
							if m.UVOffset == nil {
								m.UVOffset = map[string][2]float32{}
							}
							m.UVScale["baseColor"] = scale
							m.UVOffset["baseColor"] = off
						}
					}
				}

				// MetallicRoughness texture
				if gm.PBR.MetallicRoughnessTexture != nil {
					m.MetallicRoughnessTexturePath = resolveTex(gm.PBR.MetallicRoughnessTexture)
					if m.TexCoordMap == nil {
						m.TexCoordMap = map[string]int{}
					}
					m.TexCoordMap["metallicRoughness"] = gm.PBR.MetallicRoughnessTexture.TexCoord
					if ext, ok := gm.PBR.MetallicRoughnessTexture.Extensions["KHR_texture_transform"]; ok {
						off, scale, _, err := parseTextureTransform(ext)
						if err == nil {
							if m.UVScale == nil {
								m.UVScale = map[string][2]float32{}
							}
							if m.UVOffset == nil {
								m.UVOffset = map[string][2]float32{}
							}
							m.UVScale["metallicRoughness"] = scale
							m.UVOffset["metallicRoughness"] = off
						}
					}
				}

				// Normal texture + normal scale
				if gm.NormalTexture != nil {
					m.NormalTexturePath = resolveTex(gm.NormalTexture)
					if m.TexCoordMap == nil {
						m.TexCoordMap = map[string]int{}
					}
					m.TexCoordMap["normal"] = gm.NormalTexture.TexCoord
					// normalTexture.scale (glTF allows a scale on normalTexture)
					if gm.NormalTexture.Scale != 0 {
						m.NormalScale = gm.NormalTexture.Scale
					}
					if ext, ok := gm.NormalTexture.Extensions["KHR_texture_transform"]; ok {
						off, scale, _, err := parseTextureTransform(ext)
						if err == nil {
							if m.UVScale == nil {
								m.UVScale = map[string][2]float32{}
							}
							if m.UVOffset == nil {
								m.UVOffset = map[string][2]float32{}
							}
							m.UVScale["normal"] = scale
							m.UVOffset["normal"] = off
						}
					}
				}

				// Occlusion (AO)
				if gm.OcclusionTexture != nil {
					m.OcclusionTexturePath = resolveTex(gm.OcclusionTexture)
					if m.TexCoordMap == nil {
						m.TexCoordMap = map[string]int{}
					}
					m.TexCoordMap["occlusion"] = gm.OcclusionTexture.TexCoord
					if ext, ok := gm.OcclusionTexture.Extensions["KHR_texture_transform"]; ok {
						off, scale, _, err := parseTextureTransform(ext)
						if err == nil {
							if m.UVScale == nil {
								m.UVScale = map[string][2]float32{}
							}
							if m.UVOffset == nil {
								m.UVOffset = map[string][2]float32{}
							}
							m.UVScale["occlusion"] = scale
							m.UVOffset["occlusion"] = off
						}
					}
				}

				// Roughness factor (fallback if no metallicRoughness texture)
				if gm.PBR.RoughnessFactor != 0 {
					// you may want to expose this later; for now it's available in gm.PBR.RoughnessFactor
				}

				// KHR extensions: specular / sheen
				if gm.Extensions != nil {
					// KHR_materials_specular
					if raw, ok := gm.Extensions["KHR_materials_specular"]; ok {
						var spec struct {
							SpecularFactor float32 `json:"specularFactor"`
						}
						if err := json.Unmarshal(raw, &spec); err == nil {
							m.SpecularFactor = spec.SpecularFactor
						}
					}
					// KHR_materials_sheen
					if raw, ok := gm.Extensions["KHR_materials_sheen"]; ok {
						var sheen struct {
							SheenColorFactor     []float32 `json:"sheenColorFactor"`
							SheenRoughnessFactor float32   `json:"sheenRoughnessFactor"`
						}
						if err := json.Unmarshal(raw, &sheen); err == nil {
							if len(sheen.SheenColorFactor) >= 3 {
								m.SheenColor = [3]float32{
									sheen.SheenColorFactor[0],
									sheen.SheenColorFactor[1],
									sheen.SheenColorFactor[2],
								}
							}
							m.SheenRoughness = sheen.SheenRoughnessFactor
						}
					}
				}

				// Alpha mode (optional)
				if gm.AlphaMode != "" {
					// store if you want to handle transparency later
				}
			}

			results = append(results, m)
		}
	}

	return results, nil
}
func composeNodeTransform(n gltfNode) [16]float32 {
	// If matrix is provided, it overrides everything
	if len(n.Matrix) == 16 {
		var out [16]float32
		copy(out[:], n.Matrix)
		return out
	}

	// Translation
	tx, ty, tz := float32(0), float32(0), float32(0)
	if len(n.Translation) == 3 {
		tx, ty, tz = n.Translation[0], n.Translation[1], n.Translation[2]
	}

	// Scale
	sx, sy, sz := float32(1), float32(1), float32(1)
	if len(n.Scale) == 3 {
		sx, sy, sz = n.Scale[0], n.Scale[1], n.Scale[2]
	}

	// Rotation (quaternion)
	qx, qy, qz, qw := float32(0), float32(0), float32(0), float32(1)
	if len(n.Rotation) == 4 {
		qx, qy, qz, qw = n.Rotation[0], n.Rotation[1], n.Rotation[2], n.Rotation[3]
	}

	xx := qx * qx
	yy := qy * qy
	zz := qz * qz
	xy := qx * qy
	xz := qx * qz
	yz := qy * qz
	wx := qw * qx
	wy := qw * qy
	wz := qw * qz

	m := [16]float32{
		1 - 2*(yy+zz), 2 * (xy - wz), 2 * (xz + wy), 0,
		2 * (xy + wz), 1 - 2*(xx+zz), 2 * (yz - wx), 0,
		2 * (xz - wy), 2 * (yz + wx), 1 - 2*(xx+yy), 0,
		0, 0, 0, 1,
	}

	// Scale
	m[0] *= sx
	m[1] *= sx
	m[2] *= sx
	m[4] *= sy
	m[5] *= sy
	m[6] *= sy
	m[8] *= sz
	m[9] *= sz
	m[10] *= sz

	// Translation
	m[12] = tx
	m[13] = ty
	m[14] = tz

	return m
}

// Public wrapper so scene package can use it
func ComposeNodeTransform(n gltfNode) [16]float32 {
	return composeNodeTransform(n)
}

func LoadGLTFRoot(path string) (*gltfRoot, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var g gltfRoot
	if err := json.Unmarshal(raw, &g); err != nil {
		return nil, err
	}
	return &g, nil
}

func loadGLB(path string) (*gltfRoot, [][]byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}

	if len(raw) < 20 || string(raw[0:4]) != "glTF" {
		return nil, nil, fmt.Errorf("not a valid GLB file")
	}

	version := binary.LittleEndian.Uint32(raw[4:8])
	if version != 2 {
		return nil, nil, fmt.Errorf("unsupported GLB version %d", version)
	}

	length := binary.LittleEndian.Uint32(raw[8:12])
	if int(length) != len(raw) {
		return nil, nil, fmt.Errorf("GLB length mismatch")
	}

	offset := 12

	// --- JSON chunk ---
	jsonChunkLen := int(binary.LittleEndian.Uint32(raw[offset : offset+4]))
	jsonChunkType := string(raw[offset+4 : offset+8])
	offset += 8

	if jsonChunkType != "JSON" {
		return nil, nil, fmt.Errorf("first GLB chunk is not JSON")
	}

	jsonBytes := raw[offset : offset+jsonChunkLen]
	offset += jsonChunkLen

	var g gltfRoot
	if err := json.Unmarshal(jsonBytes, &g); err != nil {
		return nil, nil, err
	}

	// --- BIN chunk (optional) ---
	var buffers [][]byte
	if offset < len(raw) {
		binChunkLen := int(binary.LittleEndian.Uint32(raw[offset : offset+4]))
		binChunkType := string(raw[offset+4 : offset+8])
		offset += 8

		if binChunkType != "BIN\x00" {
			return nil, nil, fmt.Errorf("second GLB chunk is not BIN")
		}

		binBytes := raw[offset : offset+binChunkLen]
		buffers = append(buffers, binBytes)
	}

	return &g, buffers, nil
}

func MulMat4(a, b [16]float32) [16]float32 {
	// Column-major: r = a * b
	var r [16]float32
	for col := 0; col < 4; col++ {
		for row := 0; row < 4; row++ {
			r[col*4+row] =
				a[0*4+row]*b[col*4+0] +
					a[1*4+row]*b[col*4+1] +
					a[2*4+row]*b[col*4+2] +
					a[3*4+row]*b[col*4+3]
		}
	}
	return r
}

func TransformPoint(m [16]float32, v [3]float32) [3]float32 {
	return [3]float32{
		m[0]*v[0] + m[4]*v[1] + m[8]*v[2] + m[12],
		m[1]*v[0] + m[5]*v[1] + m[9]*v[2] + m[13],
		m[2]*v[0] + m[6]*v[1] + m[10]*v[2] + m[14],
	}
}
func TransformNormal(m [16]float32, n [3]float32) [3]float32 {
	return [3]float32{
		m[0]*n[0] + m[4]*n[1] + m[8]*n[2],
		m[1]*n[0] + m[5]*n[1] + m[9]*n[2],
		m[2]*n[0] + m[6]*n[1] + m[10]*n[2],
	}
}
func findNodesForMesh(g *gltfRoot, meshIndex int) []int {
	out := []int{}
	for i, n := range g.Nodes {
		if n.Mesh == meshIndex {
			out = append(out, i)
		}
	}
	return out
}
func IdentityMatrix() [16]float32 {
	return [16]float32{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

// DecomposeTRS extracts translation, rotation (quat), and scale from a 4x4 matrix.
// Assumes column-major order (OpenGL style).
func DecomposeTRS(m [16]float32) (pos [3]float32, rot [4]float32, scale [3]float32) {

	// Translation
	pos = [3]float32{m[12], m[13], m[14]}

	// Extract basis vectors
	x := [3]float32{m[0], m[1], m[2]}
	y := [3]float32{m[4], m[5], m[6]}
	z := [3]float32{m[8], m[9], m[10]}

	// Scale = length of basis vectors
	scale[0] = float32(math.Sqrt(float64(x[0]*x[0] + x[1]*x[1] + x[2]*x[2])))
	scale[1] = float32(math.Sqrt(float64(y[0]*y[0] + y[1]*y[1] + y[2]*y[2])))
	scale[2] = float32(math.Sqrt(float64(z[0]*z[0] + z[1]*z[1] + z[2]*z[2])))

	// Normalize basis vectors
	if scale[0] != 0 {
		x[0] /= scale[0]
		x[1] /= scale[0]
		x[2] /= scale[0]
	}
	if scale[1] != 0 {
		y[0] /= scale[1]
		y[1] /= scale[1]
		y[2] /= scale[1]
	}
	if scale[2] != 0 {
		z[0] /= scale[2]
		z[1] /= scale[2]
		z[2] /= scale[2]
	}

	// Convert rotation matrix → quaternion
	trace := x[0] + y[1] + z[2]

	if trace > 0 {
		s := float32(math.Sqrt(float64(trace+1.0)) * 2)
		rot[3] = 0.25 * s
		rot[0] = (y[2] - z[1]) / s
		rot[1] = (z[0] - x[2]) / s
		rot[2] = (x[1] - y[0]) / s
	} else if x[0] > y[1] && x[0] > z[2] {
		s := float32(math.Sqrt(float64(1.0+x[0]-y[1]-z[2])) * 2)
		rot[3] = (y[2] - z[1]) / s
		rot[0] = 0.25 * s
		rot[1] = (y[0] + x[1]) / s
		rot[2] = (z[0] + x[2]) / s
	} else if y[1] > z[2] {
		s := float32(math.Sqrt(float64(1.0+y[1]-x[0]-z[2])) * 2)
		rot[3] = (z[0] - x[2]) / s
		rot[0] = (y[0] + x[1]) / s
		rot[1] = 0.25 * s
		rot[2] = (z[1] + y[2]) / s
	} else {
		s := float32(math.Sqrt(float64(1.0+z[2]-x[0]-y[1])) * 2)
		rot[3] = (x[1] - y[0]) / s
		rot[0] = (z[0] + x[2]) / s
		rot[1] = (z[1] + y[2]) / s
		rot[2] = 0.25 * s
	}

	return
}

type MeshTRS struct {
	Position [3]float32
	Rotation [4]float32
	Scale    [3]float32
}

func ExtractGLTFMeshTRS(path string) (map[string]MeshTRS, error) {
	g, _, err := LoadGLTFOrGLB(path)
	if err != nil {
		return nil, err
	}

	nodeWorld := make([][16]float32, len(g.Nodes))

	var compute func(i int) [16]float32
	compute = func(i int) [16]float32 {
		if nodeWorld[i] != ([16]float32{}) {
			return nodeWorld[i]
		}
		local := composeNodeTransform(g.Nodes[i])

		// root nodes
		for _, root := range g.Scenes[g.Scene].Nodes {
			if root == i {
				nodeWorld[i] = local
				return local
			}
		}

		// find parent
		for p, n := range g.Nodes {
			for _, c := range n.Children {
				if c == i {
					parent := compute(p)
					nodeWorld[i] = MulMat4(parent, local)
					return nodeWorld[i]
				}
			}
		}

		nodeWorld[i] = local
		return local
	}

	result := make(map[string]MeshTRS)

	for mi, mesh := range g.Meshes {
		meshName := mesh.Name
		if meshName == "" {
			meshName = fmt.Sprintf("mesh_%d", mi)
		}

		nodes := findNodesForMesh(g, mi)
		world := IdentityMatrix()
		if len(nodes) > 0 {
			world = compute(nodes[0])
		}

		pos, rot, scl := DecomposeTRS(world)

		for pi := range mesh.Primitives {
			meshID := fmt.Sprintf("%s/%d", meshName, pi)
			result[meshID] = MeshTRS{
				Position: pos,
				Rotation: rot,
				Scale:    scl,
			}
		}
	}

	return result, nil
}

// helper to read KHR_texture_transform
func parseTextureTransform(ext json.RawMessage) (offset [2]float32, scale [2]float32, rotation float32, err error) {
	var t struct {
		Offset   [2]float32 `json:"offset"`
		Scale    [2]float32 `json:"scale"`
		Rotation float32    `json:"rotation"`
	}
	if err = json.Unmarshal(ext, &t); err != nil {
		return
	}
	offset = t.Offset
	scale = t.Scale
	rotation = t.Rotation
	return
}
