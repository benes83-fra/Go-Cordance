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
	textureDir := "assets/textures"
	entries, err := os.ReadDir(textureDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		ext := filepath.Ext(e.Name())
		if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
			log.Printf("File not allowed as Texture: %s", e.Name())
			continue
		}

		full := filepath.Join(textureDir, e.Name())

		// --- NEW: skip if already loaded manually ---
		if assets.FindAssetByPath(full) != nil {
			log.Printf("Skipping already-loaded texture: %s", full)
			continue
		}

		id, _, err := assets.ImportTexture(full)
		if err != nil {
			log.Printf("Failed to load Texture %s: %v", full, err)
			continue
		}

		log.Printf("Loaded Texture asset %d from %s", id, full)
	}
}
