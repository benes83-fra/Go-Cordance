// internal/engine/objloader.go (new file)
package engine

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
)

type objVertex struct {
	pIdx, tIdx, nIdx int
}

func parseIndex(s string) (int, int, int, error) {
	// face element formats: v, v/t, v//n, v/t/n
	parts := strings.Split(s, "/")
	var vi, ti, ni int
	var err error
	if len(parts) >= 1 && parts[0] != "" {
		vi, err = strconv.Atoi(parts[0])
		if err != nil {
			return 0, 0, 0, err
		}
	}
	if len(parts) >= 2 && parts[1] != "" {
		ti, err = strconv.Atoi(parts[1])
		if err != nil {
			return 0, 0, 0, err
		}
	}
	if len(parts) >= 3 && parts[2] != "" {
		ni, err = strconv.Atoi(parts[2])
		if err != nil {
			return 0, 0, 0, err
		}
	}
	return vi, ti, ni, nil
}

func (mm *MeshManager) RegisterOBJ(id, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	var positions [][]float32
	var normals [][]float32
	var uvs [][]float32

	// map of unique combined vertex -> index
	vertMap := make(map[string]uint32)
	var vertices []float32
	var indices []uint32
	var nextIndex uint32 = 0

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		switch fields[0] {
		case "v":
			if len(fields) < 4 {
				continue
			}
			x, _ := strconv.ParseFloat(fields[1], 32)
			y, _ := strconv.ParseFloat(fields[2], 32)
			z, _ := strconv.ParseFloat(fields[3], 32)
			positions = append(positions, []float32{float32(x), float32(y), float32(z)})
		case "vt":
			if len(fields) < 3 {
				continue
			}
			u, _ := strconv.ParseFloat(fields[1], 32)
			v, _ := strconv.ParseFloat(fields[2], 32)
			uvs = append(uvs, []float32{float32(u), float32(v)})
		case "vn":
			if len(fields) < 4 {
				continue
			}
			nx, _ := strconv.ParseFloat(fields[1], 32)
			ny, _ := strconv.ParseFloat(fields[2], 32)
			nz, _ := strconv.ParseFloat(fields[3], 32)
			normals = append(normals, []float32{float32(nx), float32(ny), float32(nz)})
		case "f":
			// faces can be triangles or quads; triangulate quads
			faceElems := fields[1:]
			// triangulate polygon fan
			for i := 1; i < len(faceElems)-1; i++ {
				tri := []string{faceElems[0], faceElems[i], faceElems[i+1]}
				for _, elem := range tri {
					vi, ti, ni, err := parseIndex(elem)
					if err != nil {
						return fmt.Errorf("parseIndex error: %v", err)
					}
					// OBJ indices are 1-based; convert to 0-based
					key := fmt.Sprintf("%d/%d/%d", vi, ti, ni)
					idx, ok := vertMap[key]
					if !ok {
						// fetch position
						var px, py, pz float32
						if vi != 0 {
							p := positions[vi-1]
							px, py, pz = p[0], p[1], p[2]
						}
						var tx, ty float32
						if ti != 0 {
							t := uvs[ti-1]
							tx, ty = t[0], t[1]
						}
						var nx, ny, nz float32
						if ni != 0 {
							n := normals[ni-1]
							nx, ny, nz = n[0], n[1], n[2]
						}
						// append interleaved vertex: pos(3), normal(3), uv(2)
						vertices = append(vertices, px, py, pz, nx, ny, nz, tx, ty)
						idx = nextIndex
						vertMap[key] = idx
						nextIndex++
					}
					indices = append(indices, idx)
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	// create GL buffers
	var vao, vbo, ebo uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)
	gl.GenBuffers(1, &ebo)

	gl.BindVertexArray(vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	stride := int32(8 * 4)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, stride, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, stride, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(2)
	gl.VertexAttribPointer(2, 2, gl.FLOAT, false, stride, gl.PtrOffset(6*4))

	gl.BindVertexArray(0)

	mm.vaos[id] = vao
	mm.vbos[id] = vbo
	mm.ebos[id] = ebo
	mm.counts[id] = int32(len(indices))

	return nil
}
