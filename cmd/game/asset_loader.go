package main

import (
	"go-engine/Go-Cordance/internal/assets"
	"log"
	"os"
	"path/filepath"
)

func load_materials() {

	// Load material assets
	materialDir := "assets/materials"
	entries, err := os.ReadDir(materialDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if filepath.Ext(e.Name()) != ".mat" {
			continue
		}

		full := filepath.Join(materialDir, e.Name())
		id, err := assets.LoadMaterial(full)
		if err != nil {
			log.Printf("Failed to load material %s: %v", full, err)
			continue
		}

		log.Printf("Loaded material asset %d from %s", id, full)
	}

}

func load_textures() {

	// Load material assets
	textureDir := "assets/textures"
	entries, err := os.ReadDir(textureDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if (filepath.Ext(e.Name()) != ".png") && (filepath.Ext(e.Name()) != ".jpeg") && (filepath.Ext(e.Name()) != ".jpg") {
			log.Printf("File not allowed as Texture %+v", e.Name())
			continue
		}
		log.Printf("We found textures")
		full := filepath.Join(textureDir, e.Name())
		id, _, err := assets.ImportTexture(full)
		if err != nil {
			log.Printf("Failed to load Texture %s: %v", full, err)
			continue
		}

		log.Printf("Loaded Texture asset %d from %s", id, full)
	}

}
