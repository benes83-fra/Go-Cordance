package engine

import (
	"io/ioutil"
	"strings"
)

// LoadShaderSource reads a shader file, strips a UTF-8 BOM if present,
// normalizes CRLF to LF, and returns a null-terminated string suitable for gl.Strs.
func LoadShaderSource(path string) (string, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	// Strip UTF-8 BOM if present
	if len(b) >= 3 && b[0] == 0xEF && b[1] == 0xBB && b[2] == 0xBF {
		b = b[3:]
	}
	s := string(b)
	// Normalize CRLF -> LF
	s = strings.ReplaceAll(s, "\r\n", "\n")
	// Trim leading nulls or spaces to ensure #version is first token
	s = strings.TrimLeft(s, "\u0000 \t\n\r")
	return s + "\x00", nil
}
