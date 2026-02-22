package loader

import (
	"fmt"
	"log"
	"strings"
	"time"

	"go-engine/Go-Cordance/internal/engine"
	"go-engine/Go-Cordance/internal/shaderlang"
)

func ReloadShader(shaderName string) error {
	meta, ok := ShaderMetaMap[shaderName]
	if !ok {
		return fmt.Errorf("no ShaderMeta for %s", shaderName)
	}

	var vert, frag string

	for i := 0; i < 5; i++ {
		v, err1 := shaderlang.LoadGLSL(meta.Vertex)
		f, err2 := shaderlang.LoadGLSL(meta.Fragment)

		if err1 == nil && err2 == nil &&
			len(strings.TrimSpace(v)) > 0 &&
			len(strings.TrimSpace(f)) > 0 {

			vert = shaderlang.ApplyDefines(v, meta.Defines)
			frag = shaderlang.ApplyDefines(f, meta.Defines)
			break
		}

		time.Sleep(120 * time.Millisecond)
	}

	if len(strings.TrimSpace(vert)) == 0 || len(strings.TrimSpace(frag)) == 0 {
		return fmt.Errorf("shader %s still empty after retries", shaderName)
	}

	log.Printf("[ShaderReload] Compiling %s (vert=%d, frag=%d)", shaderName, len(vert), len(frag))

	sp := engine.MustGetShaderProgram(shaderName)
	if err := sp.Reload(vert, frag); err != nil {
		return fmt.Errorf("reload failed for %s: %w", shaderName, err)
	}

	return nil
}
