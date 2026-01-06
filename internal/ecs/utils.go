package ecs

import "strconv"

func toFloat32(v any) float32 {
	switch n := v.(type) {
	case float32:
		return n
	case float64:
		return float32(n)
	case int:
		return float32(n)
	case string:
		f, _ := strconv.ParseFloat(n, 32)
		return float32(f)
	default:
		return 0
	}
}

func toVec3(v any) [3]float32 {
	var out [3]float32

	switch arr := v.(type) {
	case [3]float32:
		return arr
	case []float32:
		copy(out[:], arr)
		return out
	case []float64:
		for i := 0; i < len(arr) && i < 3; i++ {
			out[i] = float32(arr[i])
		}
		return out
	case []any:
		for i := 0; i < len(arr) && i < 3; i++ {
			out[i] = toFloat32(arr[i])
		}
		return out
	default:
		return out
	}
}
func toVec4(v any) [4]float32 {
	var out [4]float32

	switch arr := v.(type) {
	case [4]float32:
		return arr
	case []float32:
		copy(out[:], arr)
		return out
	case []float64:
		for i := 0; i < len(arr) && i < 4; i++ {
			out[i] = float32(arr[i])
		}
		return out
	case []any:
		for i := 0; i < len(arr) && i < 4; i++ {
			out[i] = toFloat32(arr[i])
		}
		return out
	default:
		return out
	}
}

func min(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func clamp(val, min, max float32) float32 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}
func toInt(v any) int {
	switch n := v.(type) {
	case int64:
		return int(n)
	case int32:
		return int(n)
	case int:
		return int(n)
	case uint:
		return int(n)
	case uint16:
		return int(n)
	case uint32:
		return int(n)
	case uint64:
		return int(n)
	case string:
		i, _ := strconv.ParseInt(n, 0, 0)
		return int(i)
	default:
		return 0

	}

}
