package loader

import (
	"go-engine/Go-Cordance/internal/assets"
	"go-engine/Go-Cordance/internal/engine"
	"go-engine/Go-Cordance/internal/shaderlang"
	"log"
	"os"
	"path/filepath"
)

func LoadMaterials() {

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
		if assets.FindAssetByPath(full) != nil {
			log.Printf("Skipping already-loaded Material: %s", full)
			continue
		}
		id, err := assets.LoadMaterial(full)
		if err != nil {
			log.Printf("Failed to load material %s: %v", full, err)
			continue
		}

		log.Printf("Loaded material asset %d from %s", id, full)
	}

}

func LoadTextures() {
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

func LoadShaders() {
	shaderDir := "assets/shaders"
	entries, err := os.ReadDir(shaderDir)
	if err != nil {
		log.Fatal(err)

	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if filepath.Ext(e.Name()) != ".json" { // your JSON shader asset extension
			continue
		}

		full := filepath.Join(shaderDir, e.Name())

		// Skip if already loaded
		if assets.FindAssetByPath(full) != nil {
			log.Printf("Skipping already-loaded shader: %s", full)
			continue
		}

		id, err := assets.LoadShader(full)
		if err != nil {
			log.Printf("Failed to load shader %s: %v", full, err)
			continue
		}
		// After: id, err := assets.LoadShader(full)

		a := assets.Get(id) // <-- FIX: fetch the asset struct

		src := a.Data.(shaderlang.ShaderSource)

		// Register metadata
		ShaderMetaMap[src.Name] = ShaderMeta{
			Name:     src.Name,
			Vertex:   src.VertexPath,
			Fragment: src.FragmentPath,
			Defines:  src.Defines, // <-- REQUIRED
		}

		// Map GLSL filenames → shader name
		FileToShader[filepath.Base(src.VertexPath)] = src.Name
		FileToShader[filepath.Base(src.FragmentPath)] = src.Name

		log.Printf("Loaded shader asset %d from %s", id, full)

	}

}

func LoadAllShaders() error {
	for _, a := range assets.All() {
		if a.Type != assets.AssetShader {
			continue
		}

		src := a.Data.(shaderlang.ShaderSource)

		// Load GLSL text
		vert, err := shaderlang.LoadGLSL(src.VertexPath)
		if err != nil {
			return err
		}
		frag, err := shaderlang.LoadGLSL(src.FragmentPath)
		if err != nil {
			return err
		}

		// Apply defines
		vert = shaderlang.ApplyDefines(vert, src.Defines)
		frag = shaderlang.ApplyDefines(frag, src.Defines)

		// Compile in engine
		prog, err := engine.LoadShaderProgram(src.Name, vert, frag)
		if err != nil {
			return err
		}

		engine.RegisterShaderProgram(src.Name, prog)
	}

	return nil
}
func LoadMeshes(meshMgr *engine.MeshManager) {
	meshDir := "assets/models"
	entries, err := os.ReadDir(meshDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		ext := filepath.Ext(e.Name())
		if ext != ".gltf" && ext != ".glb" && ext != ".obj" {
			log.Printf("Skipping non-mesh file: %s", e.Name())
			continue
		}

		full := filepath.Join(meshDir, e.Name())

		// --- Skip if already loaded ---
		if assets.FindAssetByPath(full) != nil {
			log.Printf("Skipping already-loaded mesh: %s", full)
			continue
		}

		// --- Load mesh asset ---
		var assetID assets.AssetID
		var meshIDs []string

		switch ext {
		case ".gltf", ".glb":
			id, ids, err := assets.ImportGLTFMulti(full, meshMgr)
			if err != nil {
				log.Printf("Failed to load GLTF mesh %s: %v", full, err)
				continue
			}
			assetID = id
			meshIDs = ids

		case ".obj":
			id, meshID, err := assets.ImportOBJ(full, meshMgr)
			if err != nil {
				log.Printf("Failed to load OBJ mesh %s: %v", full, err)
				continue
			}
			assetID = id
			meshIDs = []string{meshID}
		}

		log.Printf("Loaded mesh asset %d from %s (submeshes: %v)", assetID, full, meshIDs)
	}
}
