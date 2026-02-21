package loader

import (
	"log"
	"path/filepath"

	"go-engine/Go-Cordance/internal/engine"

	"github.com/fsnotify/fsnotify"
)

var ReloadQueue = make(chan string, 8)

type ShaderMeta struct {
	Name     string `json:"name"`
	Vertex   string `json:"vertex"`
	Fragment string `json:"fragment"`
}

var ShaderMetaMap = map[string]ShaderMeta{} // key = shaderName
var FileToShader = map[string]string{}      // key = filename.glsl → shaderName

func StartShaderWatcher() {
	watcher, _ := fsnotify.NewWatcher()

	go func() {
		for {
			select {
			case ev := <-watcher.Events:
				if ev.Op&(fsnotify.Write|fsnotify.Create) != 0 {
					onShaderFileChanged(ev.Name)
				}
			case err := <-watcher.Errors:
				log.Println("shader watcher error:", err)
			}
		}
	}()

	watcher.Add("assets/shaders")
}

func onShaderFileChanged(path string) {
	file := filepath.Base(path)

	shaderName, ok := FileToShader[file]
	if !ok {
		return // not a shader we track
	}

	meta := ShaderMetaMap[shaderName]

	log.Printf("[ShaderWatcher] Reloading %s due to change in %s", shaderName, file)

	vert, err := engine.LoadShaderSource(meta.Vertex)
	if err != nil {
		log.Printf("[ShaderWatcher] Failed to load vertex shader: %v", err)
		return
	}

	frag, err := engine.LoadShaderSource(meta.Fragment)
	if err != nil {
		log.Printf("[ShaderWatcher] Failed to load fragment shader: %v", err)
		return
	}

	sp := engine.MustGetShaderProgram(shaderName)
	if err := sp.Reload(vert, frag); err != nil {
		log.Printf("[ShaderWatcher] Reload failed: %v", err)
		return
	}

	ReloadQueue <- shaderName
	log.Printf("[ShaderWatcher] Reloaded %s successfully", shaderName)
}
