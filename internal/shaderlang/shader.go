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

	lines := strings.Split(src, "\n")

	// Find #version line
	versionIndex := -1
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "#version") {
			versionIndex = i
			break
		}
	}

	// Build define block
	var defBlock strings.Builder
	for k, v := range defs {
		defBlock.WriteString(fmt.Sprintf("#define %s %v\n", k, v))
	}
	defBlock.WriteString("\n")

	// If no #version found, prepend defines normally
	if versionIndex == -1 {
		return defBlock.String() + src
	}

	// Insert defines *after* #version
	out := make([]string, 0, len(lines)+len(defs)+2)
	out = append(out, lines[versionIndex]) // keep #version first
	out = append(out, defBlock.String())   // then defines
	out = append(out, lines[versionIndex+1:]...)

	return strings.Join(out, "\n")
}
