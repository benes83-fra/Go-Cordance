package engine

type LightData struct {
	Type      int32
	Color     [3]float32
	Intensity float32
	Direction [3]float32
	Position  [3]float32
	Range     float32
	Angle     float32
}
