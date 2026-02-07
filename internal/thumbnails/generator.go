package thumbnails

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"

	_ "image/jpeg"
	_ "image/png"

	"go-engine/Go-Cordance/internal/assets"
	"go-engine/Go-Cordance/internal/engine"

	"github.com/nfnt/resize"
)

// assetPathForID returns the canonical source path for an asset ID.
// It looks up the engine's asset registry rather than assuming numeric filenames.
func assetPathForID(assetID uint64) (string, bool) {
	for _, a := range assets.All() {
		if uint64(a.ID) == assetID {
			return a.Path, true
		}
	}
	return "", false
}

// generator.go
func generateTextureThumbnail(a *assets.Asset, size int) ([]byte, string, error) {
	path := a.Path
	if path == "" {
		return nil, "", fmt.Errorf("asset %d has empty path", a.ID)
	}

	cacheDir := filepath.Join("cache", "thumbs")
	os.MkdirAll(cacheDir, 0755)

	f, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, "", err
	}

	thumb := resize.Thumbnail(uint(size), uint(size), img, resize.Lanczos3)

	var buf bytes.Buffer
	if err := png.Encode(&buf, thumb); err != nil {
		return nil, "", err
	}
	data := buf.Bytes()
	hash := sha1Hex(data)

	fname := filepath.Join(cacheDir, fmt.Sprintf("%d-%s.png", a.ID, hash))
	if _, err := os.Stat(fname); os.IsNotExist(err) {
		_ = os.WriteFile(fname, data, 0644)
	}

	return data, hash, nil
}

func GenerateThumbnailBytes(assetID uint64, size int) ([]byte, string, error) {
	a := assets.Get(assets.AssetID(assetID))
	if a == nil {
		return nil, "", fmt.Errorf("asset %d not found", assetID)
	}

	switch a.Type {
	case assets.AssetTexture:
		return generateTextureThumbnail(a, size)
	case assets.AssetMesh:
		return generateMeshThumbnail(a, size)
	default:
		return nil, "", fmt.Errorf("no thumbnail generator for asset type %v", a.Type)
	}
}

func sha1Hex(data []byte) string {
	h := sha1.Sum(data)
	return hex.EncodeToString(h[:])
}
func generateMeshThumbnail(a *assets.Asset, size int) ([]byte, string, error) {
	switch v := a.Data.(type) {
	case string:
		// single mesh
		data, hash, err := engine.RenderMeshThumbnail(v, size)
		if err != nil {
			return nil, "", err
		}
		return cacheMeshThumb(a.ID, data, hash)

	case []string:
		if len(v) == 0 {
			return nil, "", fmt.Errorf("mesh asset %d has no meshIDs", a.ID)
		}
		// multi-mesh: render all submeshes together
		log.Printf("We need to render these meshes %+v", v)
		data, hash, err := engine.RenderMeshGroupThumbnail(v, size)
		if err != nil {
			return nil, "", err
		}
		return cacheMeshThumb(a.ID, data, hash)

	default:
		return nil, "", fmt.Errorf("mesh asset %d has unexpected Data type %T", a.ID, v)
	}
}

func cacheMeshThumb(id assets.AssetID, data []byte, hash string) ([]byte, string, error) {
	cacheDir := filepath.Join("cache", "thumbs")
	os.MkdirAll(cacheDir, 0755)

	fname := filepath.Join(cacheDir, fmt.Sprintf("%d-%s.png", id, hash))
	if _, err := os.Stat(fname); os.IsNotExist(err) {
		_ = os.WriteFile(fname, data, 0644)
	}
	return data, hash, nil
}

// GenerateMeshSubThumbnailBytes renders ONE submesh by meshID.
func GenerateMeshSubThumbnailBytes(assetID uint64, meshID string, size int) ([]byte, string, error) {
	data, hash, err := engine.RenderMeshThumbnail(meshID, size)
	if err != nil {
		return nil, "", err
	}
	return cacheMeshThumb(assets.AssetID(assetID), data, hash)

}
