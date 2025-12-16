package resources

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestReadTextFile(t *testing.T) {
	dir, err := ioutil.TempDir("", "res-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "hello.txt")
	content := "hello world\n"
	if err := ioutil.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := ReadTextFile(path)
	if err != nil {
		t.Fatalf("ReadTextFile error: %v", err)
	}
	if got != content {
		t.Fatalf("expected %q got %q", content, got)
	}
}

func TestAssetManagerResolve(t *testing.T) {
	am := NewAssetManager("/root/assets")
	got := am.Resolve("models/cube.obj")
	want := filepath.Join("/root/assets", "models/cube.obj")
	if got != want {
		t.Fatalf("expected %q got %q", want, got)
	}
}
