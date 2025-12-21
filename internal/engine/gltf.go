package engine

import (
	"encoding/json"
	"io/ioutil"
	"math"
	"path/filepath"

	"github.com/go-gl/gl/v4.1-core/gl"
)

type gltf struct {
	Buffers []struct {
		URI string `json:"uri"`
	} `json:"buffers"`
	BufferViews []struct {
		Buffer     int `json:"buffer"`
		ByteOffset int `json:"byteOffset"`
		ByteLength int `json:"byteLength"`
		ByteStride int `json:"byteStride"`
	} `json:"bufferViews"`
	Accessors []struct {
		BufferView    int    `json:"bufferView"`
		ByteOffset    int    `json:"byteOffset"`
		ComponentType int    `json:"componentType"`
		Count         int    `json:"count"`
		Type          string `json:"type"`
	} `json:"accessors"`
	Meshes []struct {
		Primitives []struct {
			Attributes map[string]int `json:"attributes"`
			Indices    int            `json:"indices"`
		} `json:"primitives"`
	} `json:"meshes"`
}

func readAccessor(g *gltf, baseDir string, accessorIndex int) ([]byte, int, error) {
	acc := g.Accessors[accessorIndex]
	bv := g.BufferViews[acc.BufferView]
	buf := g.Buffers[bv.Buffer]

	// load .bin file
	binPath := filepath.Join(baseDir, buf.URI)
	data, err := ioutil.ReadFile(binPath)
	if err != nil {
		return nil, 0, err
	}

	start := bv.ByteOffset + acc.ByteOffset
	end := start + acc.Count*componentSize(acc.Type, acc.ComponentType)
	return data[start:end], acc.Count, nil
}

func componentSize(typ string, comp int) int {
	// comp type sizes
	var csize int
	switch comp {
	case 5123: // UNSIGNED_SHORT
		csize = 2
	case 5125: // UNSIGNED_INT
		csize = 4
	case 5126: // FLOAT
		csize = 4
	default:
		panic("unsupported component type")
	}

	// type multiplier
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
		panic("unsupported accessor type")
	}
}

func (mm *MeshManager) RegisterGLTF(id, path string) error {
	baseDir := filepath.Dir(path)

	// load JSON
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	var g gltf
	if err := json.Unmarshal(raw, &g); err != nil {
		return err
	}

	// assume 1 mesh, 1 primitive
	prim := g.Meshes[0].Primitives[0]

	// load attributes
	posData, posCount, _ := readAccessor(&g, baseDir, prim.Attributes["POSITION"])
	norData, _, _ := readAccessor(&g, baseDir, prim.Attributes["NORMAL"])
	uvData, _, _ := readAccessor(&g, baseDir, prim.Attributes["TEXCOORD_0"])

	// tangents optional
	tanData := []byte{}
	if tIndex, ok := prim.Attributes["TANGENT"]; ok {
		tanData, _, _ = readAccessor(&g, baseDir, tIndex)
	}

	// load indices
	idxData, idxCount, _ := readAccessor(&g, baseDir, prim.Indices)

	// build interleaved vertex buffer
	vertices := make([]float32, 0, posCount*12)

	for i := 0; i < posCount; i++ {
		// pos
		px := bytesToFloat32(posData[i*12+0:])
		py := bytesToFloat32(posData[i*12+4:])
		pz := bytesToFloat32(posData[i*12+8:])

		// normal
		nx := bytesToFloat32(norData[i*12+0:])
		ny := bytesToFloat32(norData[i*12+4:])
		nz := bytesToFloat32(norData[i*12+8:])

		// uv
		u := bytesToFloat32(uvData[i*8+0:])
		v := bytesToFloat32(uvData[i*8+4:])

		// tangent (vec4)
		tx, ty, tz, tw := float32(1), float32(0), float32(0), float32(1)
		if len(tanData) > 0 {
			tx = bytesToFloat32(tanData[i*16+0:])
			ty = bytesToFloat32(tanData[i*16+4:])
			tz = bytesToFloat32(tanData[i*16+8:])
			tw = bytesToFloat32(tanData[i*16+12:])
		}

		vertices = append(vertices,
			px, py, pz,
			nx, ny, nz,
			u, v,
			tx, ty, tz, tw,
		)
	}

	// convert indices
	indices := make([]uint32, idxCount)
	for i := 0; i < idxCount; i++ {
		indices[i] = uint32(idxData[i*4+0]) |
			uint32(idxData[i*4+1])<<8 |
			uint32(idxData[i*4+2])<<16 |
			uint32(idxData[i*4+3])<<24
	}

	// upload to GL
	var vao, vbo, ebo uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)
	gl.GenBuffers(1, &ebo)

	gl.BindVertexArray(vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	stride := int32(12 * 4)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, stride, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, stride, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(2)
	gl.VertexAttribPointer(2, 2, gl.FLOAT, false, stride, gl.PtrOffset(6*4))
	gl.EnableVertexAttribArray(3)
	gl.VertexAttribPointer(3, 4, gl.FLOAT, false, stride, gl.PtrOffset(8*4))

	gl.BindVertexArray(0)

	mm.vaos[id] = vao
	mm.vbos[id] = vbo
	mm.ebos[id] = ebo
	mm.counts[id] = int32(len(indices))

	return nil
}

func bytesToFloat32(b []byte) float32 {
	return math.Float32frombits(
		uint32(b[0]) |
			uint32(b[1])<<8 |
			uint32(b[2])<<16 |
			uint32(b[3])<<24)
}
