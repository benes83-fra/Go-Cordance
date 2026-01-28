package ui

import (
	"encoding/json"
	"go-engine/Go-Cordance/internal/assets"
	"strconv"
	// other imports...
)

// parseAnyInt converts a variety of numeric representations to int.
// Returns (value, true) if conversion succeeded, otherwise (0, false).
func parseAnyInt(v any) (int, bool) {
	switch t := v.(type) {
	case int:
		return t, true
	case int8:
		return int(t), true
	case int16:
		return int(t), true
	case int32:
		return int(t), true
	case int64:
		return int(t), true
	case uint:
		return int(t), true
	case uint8:
		return int(t), true
	case uint16:
		return int(t), true
	case uint32:
		return int(t), true
	case uint64:
		return int(t), true
	case float32:
		return int(t), true
	case float64:
		return int(t), true
	case assets.AssetID:
		// AssetID is an alias for uint64; convert safely
		return int(t), true
	case json.Number:
		if i, err := t.Int64(); err == nil {
			return int(i), true
		}
		if f, err := t.Float64(); err == nil {
			return int(f), true
		}
		return 0, false
	case string:
		if i, err := strconv.Atoi(t); err == nil {
			return i, true
		}
		if f, err := strconv.ParseFloat(t, 64); err == nil {
			return int(f), true
		}
		return 0, false
	default:
		return 0, false
	}
}
