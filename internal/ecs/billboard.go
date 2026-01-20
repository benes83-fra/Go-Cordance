package ecs

type BillboardMode int

const (
	BillboardSpherical BillboardMode = iota
	BillboardCylindrical
	BillboardAxial
)

type Billboard struct {
	Mode BillboardMode
	Axis [3]float32 // used only for Axial mode
}

func NewBillboard() *Billboard { return &Billboard{} }

func (b *Billboard) Update(dt float32) { _ = dt }

func (b *Billboard) EditorFields() map[string]any {
	return map[string]any{
		"Mode": int(b.Mode),
		"Axis": b.Axis,
	}
}

func (b *Billboard) SetEditorField(key string, val any) {
	switch key {
	case "Mode":
		b.Mode = BillboardMode(val.(float64))
	case "Axis":
		arr := val.([]any)
		b.Axis = [3]float32{
			float32(arr[0].(float64)),
			float32(arr[1].(float64)),
			float32(arr[2].(float64)),
		}
	}
}
