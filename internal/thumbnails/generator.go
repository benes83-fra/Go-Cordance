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

func GenerateThumbnailBytes(assetID uint64, size int) ([]byte, string, error) {

	path, ok := assetPathForID(assetID)
	if !ok {
		// helpful debug: list known texture IDs/paths once
		log.Printf("thumbnail: asset id %d not found in assets registry", assetID)
		for _, a := range assets.All() {
			log.Printf("thumbnail: known asset ID=%d Path=%s Type=%v", a.ID, a.Path, a.Type)
		}
		return nil, "", fmt.Errorf("no path for asset %d", assetID)
	}

	// disk cache check
	cacheDir := filepath.Join("cache", "thumbs")
	os.MkdirAll(cacheDir, 0755)

	// load source image
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

	h := sha1.Sum(data)
	hash := hex.EncodeToString(h[:])

	// optional: write to disk cache
	fname := filepath.Join(cacheDir, fmt.Sprintf("%d-%s.png", assetID, hash))
	if _, err := os.Stat(fname); os.IsNotExist(err) {
		_ = os.WriteFile(fname, data, 0644)
	}
	for _, a := range assets.All() {
		log.Printf("asset-registry: ID=%d Path=%s Type=%v", a.ID, a.Path, a.Type)
	}

	return data, hash, nil
}
