package resources

import (
	"io/ioutil"
	"path/filepath"
)

// ReadTextFile reads a file and returns its contents as a string.
// It does not modify the contents; callers that need null-termination
// should do so explicitly or use engine.LoadShaderSource for shaders.
func ReadTextFile(path string) (string, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// AssetManager is a tiny stub that demonstrates how to centralize asset loading.
// Expand this with caching, reference counting, and background loading as needed.
type AssetManager struct {
	root string
}

// NewAssetManager returns a manager rooted at the given directory.
func NewAssetManager(root string) *AssetManager {
	return &AssetManager{root: root}
}

// Resolve joins the root with a relative path and returns the absolute path.
func (am *AssetManager) Resolve(rel string) string {
	return filepath.Join(am.root, rel)
}
