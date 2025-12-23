package main

import (
	"go-engine/Go-Cordance/internal/editor"
	"go-engine/Go-Cordance/internal/scene"
)

func main() {
	sc, _ := scene.BootstrapScene()
	world := sc.World() // or sc.Entities if you expose them directly
	editor.Run(world)

}
