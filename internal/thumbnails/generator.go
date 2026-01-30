package thumbnails

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"

	_ "image/jpeg"
	_ "image/png"

	"github.com/nfnt/resize"
)

// Replace this with your asset lookup
func assetPathForID(assetID uint64) (string, bool) {
	// TODO: map assetID to file path or return false
	return fmt.Sprintf("assets/textures/%d.png", assetID), true
}

func GenerateThumbnailBytes(assetID uint64, size int) ([]byte, string, error) {
	path, ok := assetPathForID(assetID)
	if !ok {
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

	return data, hash, nil
}
