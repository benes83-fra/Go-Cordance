package engine

/*
import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/go-gl/gl/v4.1-core/gl"
)

type gltf_multi struct {
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
		Name       string `json:"name"`
		Primitives []struct {
			Attributes map[string]int `json:"attributes"`
			Indices    int            `json:"indices"`
		} `json:"primitives"`
	} `json:"meshes"`
}

func readAccessor_multi(g *gltf_multi, baseDir string, accessorIndex int) ([]byte, int, error) {
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

func (mm *MeshManager) RegisterGLTFMulti(path string) error {
	base := filepath.Dir(path)

	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	var g gltf_multi
	if err := json.Unmarshal(raw, &g); err != nil {
		return err
	}

	for mi, mesh := range g.Meshes {
		meshName := mesh.Name
		if meshName == "" {
			meshName = fmt.Sprintf("mesh_%d", mi)
		}

		for pi, prim := range mesh.Primitives {
			id := fmt.Sprintf("%s/%d", meshName, pi)

			// load attributes
			posData, count, _ := readAccessor_multi(&g, base, prim.Attributes["POSITION"])
			norData, _, _ := readAccessor_multi(&g, base, prim.Attributes["NORMAL"])
			uvData, _, _ := readAccessor_multi(&g, base, prim.Attributes["TEXCOORD_0"])

			tanData := []byte{}
			if tIndex, ok := prim.Attributes["TANGENT"]; ok {
				tanData, _, _ = readAccessor_multi(&g, base, tIndex)
			}

			// load indices
			idxData, idxCount, _ := readAccessor_multi(&g, base, prim.Indices)

			// build interleaved vertex buffer
			vertices := make([]float32, 0, count*12)

			for i := 0; i < count; i++ {
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

				// tangent
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

			fmt.Println("Loaded primitive:", id)
		}
	}

	return nil
}
*/
