package engine

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"

	"github.com/go-gl/gl/v4.1-core/gl"
)

func LoadTexture(path string) (uint32, error) {
	// open file
	imgFile, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer imgFile.Close()

	// try generic decode first
	img, format, err := image.Decode(imgFile)
	if err != nil {
		// generic decode failed; try reopening and forcing PNG decode
		fmt.Println("Decode error:", err)
		fmt.Println("Trying png.Decode fallback")

		// reopen file (rewind alternative)
		imgFile.Close()
		imgFile, err = os.Open(path)
		if err != nil {
			return 0, err
		}
		defer imgFile.Close()

		img, err = png.Decode(imgFile)
		if err != nil {
			fmt.Println("png.Decode error:", err)
			return 0, err
		}
		format = "png"
	}
	fmt.Println("Loaded texture format:", format)

	// convert to RGBA if necessary
	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)

	// generate GL texture
	var tex uint32
	gl.GenTextures(1, &tex)
	if tex == 0 {
		return 0, fmt.Errorf("gl.GenTextures returned 0")
	}

	gl.BindTexture(gl.TEXTURE_2D, tex)

	// ensure correct row alignment for arbitrary widths
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)

	// upload pixels
	width := int32(rgba.Rect.Size().X)
	height := int32(rgba.Rect.Size().Y)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, width, height, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(rgba.Pix))

	// generate mipmaps and set parameters
	gl.GenerateMipmap(gl.TEXTURE_2D)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)

	// restore default alignment
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 4)

	// unbind
	gl.BindTexture(gl.TEXTURE_2D, 0)

	// debug: check for GL errors
	if errCode := gl.GetError(); errCode != gl.NO_ERROR {
		fmt.Printf("GL error after TexImage2D: 0x%X\n", errCode)
	}

	fmt.Printf("Loaded texture %s -> GL id %d (%dx%d)\n", path, tex, width, height)
	return tex, nil
}

func DeleteTexture(tex uint32) {
	if tex != 0 {
		gl.DeleteTextures(1, &tex)
	}
}
