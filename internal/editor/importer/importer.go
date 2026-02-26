package importer

import (
	"io"
	"os"
	"path/filepath"
)

type AssetType string

const (
	Texture  AssetType = "texture"
	Mesh     AssetType = "mesh"
	Material AssetType = "material"
)

func CopyToAssetFolder(srcPath string, t AssetType) (string, error) {
	base := filepath.Base(srcPath)

	var dstDir string
	switch t {
	case Texture:
		dstDir = "assets/textures"
	case Mesh:
		dstDir = "assets/models"
	case Material:
		dstDir = "assets/materials"
	}

	dstPath := filepath.Join(dstDir, base)

	// ensure directory exists
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return "", err
	}

	// copy file
	in, err := os.Open(srcPath)
	if err != nil {
		return "", err
	}
	defer in.Close()

	out, err := os.Create(dstPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return "", err
	}

	return dstPath, nil
}
