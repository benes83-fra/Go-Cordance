package engine

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"path/filepath"

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

type gltfTextureInfo struct {
	Index int `json:"index"`
}

type gltfPBR struct {
	BaseColorFactor  []float32        `json:"baseColorFactor"`
	BaseColorTexture *gltfTextureInfo `json:"baseColorTexture"`
}

type gltfMaterial struct {
	Name          string           `json:"name"`
	PBR           gltfPBR          `json:"pbrMetallicRoughness"`
	NormalTexture *gltfTextureInfo `json:"normalTexture"`
}

type gltfImage struct {
	URI string `json:"uri"`
}

type gltfTexture struct {
	Source int `json:"source"`
}

type gltfNode struct {
	Name        string    `json:"name"`
	Mesh        int       `json:"mesh"`
	Children    []int     `json:"children"`
	Translation []float32 `json:"translation"`
	Rotation    []float32 `json:"rotation"` // quaternion
	Scale       []float32 `json:"scale"`
	Matrix      []float32 `json:"matrix"` // 16 floats
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

	Nodes  []gltfNode  `json:"nodes"`
	Scenes []gltfScene `json:"scenes"`
	Scene  int         `json:"scene"` // default scene index
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
	default:
		panic(fmt.Sprintf("unsupported accessor type: %s", typ))
	}
}

func bytesToFloat32(b []byte) float32 {
	return math.Float32frombits(
		uint32(b[0]) |
			uint32(b[1])<<8 |
			uint32(b[2])<<16 |
			uint32(b[3])<<24)
}

// ---------------------------
// Core accessor reader
// ---------------------------

type accessorData struct {
	acc    gltfAccessor
	bv     gltfBufferView
	buf    []byte
	base   int
	stride int
}

func getAccessor(g *gltfRoot, buffers [][]byte, idx int) (accessorData, error) {
	if idx < 0 || idx >= len(g.Accessors) {
		return accessorData{}, fmt.Errorf("accessor index out of range: %d", idx)
	}
	acc := g.Accessors[idx]

	if acc.BufferView < 0 || acc.BufferView >= len(g.BufferViews) {
		return accessorData{}, fmt.Errorf("bufferView index out of range: %d", acc.BufferView)
	}
	bv := g.BufferViews[acc.BufferView]

	if bv.Buffer < 0 || bv.Buffer >= len(buffers) {
		return accessorData{}, fmt.Errorf("buffer index out of range: %d", bv.Buffer)
	}
	buf := buffers[bv.Buffer]

	elemSize := componentByteSize(acc.Type, acc.ComponentType)
	stride := bv.ByteStride
	if stride == 0 {
		stride = elemSize
	}

	base := bv.ByteOffset + acc.ByteOffset
	end := base + acc.Count*stride
	if end > len(buf) {
		return accessorData{}, fmt.Errorf("accessor out of range: end=%d len=%d", end, len(buf))
	}

	return accessorData{acc, bv, buf, base, stride}, nil
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
	stride := int32(12 * 4)

	gl.VertexAttribPointerWithOffset(0, 3, gl.FLOAT, false, stride, 0)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointerWithOffset(1, 3, gl.FLOAT, false, stride, 3*4)
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointerWithOffset(2, 2, gl.FLOAT, false, stride, 6*4)
	gl.EnableVertexAttribArray(2)
	gl.VertexAttribPointerWithOffset(3, 4, gl.FLOAT, false, stride, 8*4)
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

// ---------------------------
// Single-mesh loader (default)
// ---------------------------

func (mm *MeshManager) RegisterGLTF(id, path string) error {
	return mm.loadGLTFInternal(id, path, false)
}

// ---------------------------
// Multi-mesh loader (optional)
// ---------------------------

func (mm *MeshManager) RegisterGLTFMulti(path string) error {
	return mm.loadGLTFInternal("", path, true)
}

// ---------------------------
// Shared geometry loader
// ---------------------------

func (mm *MeshManager) loadGLTFInternal(id, path string, multi bool) error {
	baseDir := filepath.Dir(path)

	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	var g gltfRoot
	if err := json.Unmarshal(raw, &g); err != nil {
		return err
	}

	// Load buffers
	buffers := make([][]byte, len(g.Buffers))
	for i, b := range g.Buffers {
		data, err := ioutil.ReadFile(filepath.Join(baseDir, b.URI))
		if err != nil {
			return err
		}
		buffers[i] = data
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

			// POSITION
			posA, err := getAccessor(&g, buffers, prim.Attributes["POSITION"])
			if err != nil {
				return err
			}
			count := posA.acc.Count

			// NORMAL
			norA, err := getAccessor(&g, buffers, prim.Attributes["NORMAL"])
			if err != nil {
				return err
			}

			// UV (optional)
			var uvA accessorData
			hasUV := false
			if uvIdx, ok := prim.Attributes["TEXCOORD_0"]; ok {
				uvA, err = getAccessor(&g, buffers, uvIdx)
				if err != nil {
					return err
				}
				hasUV = true
			}

			// TANGENT (optional)
			var tanA accessorData
			hasTan := false
			if tanIdx, ok := prim.Attributes["TANGENT"]; ok {
				tanA, err = getAccessor(&g, buffers, tanIdx)
				if err != nil {
					return err
				}
				hasTan = true
			}

			// INDICES
			idxA, err := getAccessor(&g, buffers, prim.Indices)
			if err != nil {
				return err
			}

			// Decode indices
			indices := make([]uint32, idxA.acc.Count)
			switch idxA.acc.ComponentType {
			case 5123: // UNSIGNED_SHORT
				for i := 0; i < idxA.acc.Count; i++ {
					off := idxA.base + i*idxA.stride
					b := idxA.buf[off : off+2]
					indices[i] = uint32(b[0]) | uint32(b[1])<<8
				}
			case 5125: // UNSIGNED_INT
				for i := 0; i < idxA.acc.Count; i++ {
					off := idxA.base + i*idxA.stride
					b := idxA.buf[off : off+4]
					indices[i] = uint32(b[0]) |
						uint32(b[1])<<8 |
						uint32(b[2])<<16 |
						uint32(b[3])<<24
				}
			default:
				return fmt.Errorf("unsupported index type: %d", idxA.acc.ComponentType)
			}

			// Build interleaved vertices
			vertices := make([]float32, 0, count*12)

			for i := 0; i < count; i++ {
				// POSITION
				pOff := posA.base + i*posA.stride
				px := bytesToFloat32(posA.buf[pOff+0:])
				py := bytesToFloat32(posA.buf[pOff+4:])
				pz := bytesToFloat32(posA.buf[pOff+8:])

				// NORMAL
				nOff := norA.base + i*norA.stride
				nx := bytesToFloat32(norA.buf[nOff+0:])
				ny := bytesToFloat32(norA.buf[nOff+4:])
				nz := bytesToFloat32(norA.buf[nOff+8:])

				// UV
				var u, v float32
				if hasUV {
					uvOff := uvA.base + i*uvA.stride
					u = bytesToFloat32(uvA.buf[uvOff+0:])
					v = bytesToFloat32(uvA.buf[uvOff+4:])
				}

				// TANGENT
				tx, ty, tz, tw := float32(1), float32(0), float32(0), float32(1)
				if hasTan {
					tOff := tanA.base + i*tanA.stride
					tx = bytesToFloat32(tanA.buf[tOff+0:])
					ty = bytesToFloat32(tanA.buf[tOff+4:])
					tz = bytesToFloat32(tanA.buf[tOff+8:])
					tw = bytesToFloat32(tanA.buf[tOff+12:])
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

	return nil
}

// ---------------------------
// Material metadata helpers
// ---------------------------

// LoadedMeshMaterial holds material info per meshID,
// to be mapped onto ecs.Material + texture components by the caller.
type LoadedMeshMaterial struct {
	MeshID             string
	BaseColor          [4]float32
	DiffuseTexturePath string
	NormalTexturePath  string
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
	baseDir := filepath.Dir(path)

	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var g gltfRoot
	if err := json.Unmarshal(raw, &g); err != nil {
		return nil, err
	}

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

			if prim.Material >= 0 && prim.Material < len(g.Materials) {
				gm := g.Materials[prim.Material]

				if len(gm.PBR.BaseColorFactor) == 4 {
					m.BaseColor = [4]float32{
						gm.PBR.BaseColorFactor[0],
						gm.PBR.BaseColorFactor[1],
						gm.PBR.BaseColorFactor[2],
						gm.PBR.BaseColorFactor[3],
					}
				}

				if gm.PBR.BaseColorTexture != nil {
					ti := gm.PBR.BaseColorTexture
					if ti.Index >= 0 && ti.Index < len(g.Textures) {
						imgIndex := g.Textures[ti.Index].Source
						if imgIndex >= 0 && imgIndex < len(g.Images) {
							m.DiffuseTexturePath = filepath.Join(baseDir, g.Images[imgIndex].URI)
						}
					}
				}

				if gm.NormalTexture != nil {
					ti := gm.NormalTexture
					if ti.Index >= 0 && ti.Index < len(g.Textures) {
						imgIndex := g.Textures[ti.Index].Source
						if imgIndex >= 0 && imgIndex < len(g.Images) {
							m.NormalTexturePath = filepath.Join(baseDir, g.Images[imgIndex].URI)
						}
					}
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
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var g gltfRoot
	if err := json.Unmarshal(raw, &g); err != nil {
		return nil, err
	}
	return &g, nil
}
