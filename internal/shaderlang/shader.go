package shaderlang

import (
	"fmt"
	"os"
	"strings"
)

type ShaderSource struct {
	Name         string         `json:"name"`
	VertexPath   string         `json:"vertex"`
	FragmentPath string         `json:"fragment"`
	Defines      map[string]any `json:"defines"`
}

// LoadGLSL loads the raw GLSL text from disk.
func LoadGLSL(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ApplyDefines injects #define statements at the top of the shader.
func ApplyDefines(src string, defs map[string]any) string {
	if len(defs) == 0 {
		return src
	}

	var b strings.Builder
	for k, v := range defs {
		b.WriteString(fmt.Sprintf("#define %s %v\n", k, v))
	}
	b.WriteString("\n")
	b.WriteString(src)
	return b.String()
}
