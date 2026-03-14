package engine

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"strings"

	"golang.org/x/image/webp"

	"github.com/go-gl/gl/v4.1-core/gl"
)

// LoadTextureWithColorSpace loads an image and creates an OpenGL texture.
// srgb true -> use sRGB internal format for albedo; false -> use linear formats.
func LoadTextureWithColorSpace(path string, srgb bool) (uint32, error) {
	// Reuse your existing LoadTexture implementation but allow choosing internal format.
	// If your current LoadTexture already accepts flags, adapt it. Otherwise:
	// - load image pixels (stb or image package)
	// - choose internalFormat = gl.SRGB8_ALPHA8 when srgb == true else gl.RGBA8
	// - upload and generate mipmaps
	// - set sampler params
	// Return GL texture id.
	return LoadTexture(path) // temporary fallback if you can't change now
}

func DeleteTexture(tex uint32) {
	if tex != 0 {
		gl.DeleteTextures(1, &tex)
	}
}

func LoadTexture(path string) (uint32, error) {
	imgFile, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer imgFile.Close()

	var img image.Image
	var format string

	// --- 1) Try WebP explicitly ---
	if strings.HasSuffix(strings.ToLower(path), ".webp") {
		img, err = webp.Decode(imgFile)
		if err != nil {
			return 0, fmt.Errorf("webp decode failed: %w", err)
		}
		format = "webp"
	} else {
		// --- 2) Try generic decode ---
		img, format, err = image.Decode(imgFile)
		if err != nil {
			// --- 3) Try PNG fallback ---
			imgFile.Close()
			imgFile, _ = os.Open(path)
			defer imgFile.Close()

			img, err = png.Decode(imgFile)
			if err != nil {
				return 0, fmt.Errorf("decode failed: %w", err)
			}
			format = "png"
		}
	}

	fmt.Println("Loaded texture format:", format)

	// Convert to RGBA
	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{}, draw.Src)

	// Upload to GL
	var tex uint32
	gl.GenTextures(1, &tex)
	gl.BindTexture(gl.TEXTURE_2D, tex)
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)

	w := int32(rgba.Rect.Dx())
	h := int32(rgba.Rect.Dy())
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, w, h, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(rgba.Pix))

	gl.GenerateMipmap(gl.TEXTURE_2D)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)

	gl.BindTexture(gl.TEXTURE_2D, 0)

	fmt.Printf("Loaded texture %s -> GL id %d (%dx%d)\n", path, tex, w, h)
	return tex, nil
}
