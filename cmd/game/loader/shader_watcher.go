package loader

import (
	"log"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

var ReloadQueue = make(chan string, 8)

type ShaderMeta struct {
	Name     string                 `json:"name"`
	Vertex   string                 `json:"vertex"`
	Fragment string                 `json:"fragment"`
	Defines  map[string]interface{} `json:"defines"`
}

var ShaderMetaMap = map[string]ShaderMeta{} // key = shaderName
var FileToShader = map[string]string{}      // key = filename.glsl → shaderName

func StartShaderWatcher() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println("shader watcher create error:", err)
		return
	}

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

	if err := watcher.Add("assets/shaders"); err != nil {
		log.Println("shader watcher add error:", err)
	}
}

func onShaderFileChanged(path string) {
	file := filepath.Base(path)

	shaderName, ok := FileToShader[file]
	if !ok {
		return // not a shader we track
	}

	log.Printf("[ShaderWatcher] Queuing reload for %s due to change in %s", shaderName, file)

	select {
	case ReloadQueue <- shaderName:
	default:
		log.Printf("[ShaderWatcher] ReloadQueue full, dropping %s", shaderName)
	}
}
