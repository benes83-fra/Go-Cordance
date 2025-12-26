package ui

import (
	state "go-engine/Go-Cordance/internal/editor/state"
	"go-engine/Go-Cordance/internal/editorlink"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// NewInspectorPanel returns the inspector container and a rebuild function.
func NewInspectorPanel() (fyne.CanvasObject, func(world interface{}, st *state.EditorState, hierarchy *widget.List)) {
	// Position entries
	posX := widget.NewEntry()
	posY := widget.NewEntry()
	posZ := widget.NewEntry()

	// Rotation entries (w,x,y,z) or whatever you prefer
	rotW := widget.NewEntry()
	rotX := widget.NewEntry()
	rotY := widget.NewEntry()
	rotZ := widget.NewEntry()

	// Scale entries
	scaleX := widget.NewEntry()
	scaleY := widget.NewEntry()
	scaleZ := widget.NewEntry()

	// Layout
	posBox := container.NewVBox(widget.NewLabel("Position"), posX, posY, posZ)
	rotBox := container.NewVBox(widget.NewLabel("Rotation"), rotW, rotX, rotY, rotZ)
	scaleBox := container.NewVBox(widget.NewLabel("Scale"), scaleX, scaleY, scaleZ)

	containerObj := container.NewVBox(posBox, rotBox, scaleBox)

	// Rebuild function will be filled by the caller to capture world/st/hierarchy.
	var rebuild func(world interface{}, st *state.EditorState, hierarchy *widget.List)

	// Helper to parse float safely
	parse32 := func(s string) float32 {
		f, _ := strconv.ParseFloat(s, 32)
		return float32(f)
	}

	// OnChanged handlers will be set inside rebuild so they capture st.SelectedIndex correctly.
	rebuild = func(world interface{}, st *state.EditorState, hierarchy *widget.List) {
		// Defensive: no selection
		if st.SelectedIndex < 0 || st.SelectedIndex >= len(st.Entities) {
			posX.SetText("")
			posY.SetText("")
			posZ.SetText("")
			rotW.SetText("")
			rotX.SetText("")
			rotY.SetText("")
			rotZ.SetText("")
			scaleX.SetText("")
			scaleY.SetText("")
			scaleZ.SetText("")
			return
		}

		ent := st.Entities[st.SelectedIndex]

		// Fill UI fields from st.Entities (bridge.EntityInfo must have Position/Rotation/Scale)
		posX.SetText(strconv.FormatFloat(float64(ent.Position[0]), 'f', 4, 32))
		posY.SetText(strconv.FormatFloat(float64(ent.Position[1]), 'f', 4, 32))
		posZ.SetText(strconv.FormatFloat(float64(ent.Position[2]), 'f', 4, 32))

		rotW.SetText(strconv.FormatFloat(float64(ent.Rotation[0]), 'f', 4, 32))
		rotX.SetText(strconv.FormatFloat(float64(ent.Rotation[1]), 'f', 4, 32))
		rotY.SetText(strconv.FormatFloat(float64(ent.Rotation[2]), 'f', 4, 32))
		rotZ.SetText(strconv.FormatFloat(float64(ent.Rotation[3]), 'f', 4, 32))

		scaleX.SetText(strconv.FormatFloat(float64(ent.Scale[0]), 'f', 4, 32))
		scaleY.SetText(strconv.FormatFloat(float64(ent.Scale[1]), 'f', 4, 32))
		scaleZ.SetText(strconv.FormatFloat(float64(ent.Scale[2]), 'f', 4, 32))

		// Set OnChanged handlers to update state and send SetTransform
		posX.OnChanged = func(val string) {
			if st.SelectedIndex < 0 || st.SelectedIndex >= len(st.Entities) {
				return
			}
			st.Entities[st.SelectedIndex].Position[0] = parse32(val)
			sendTransformIfConnected(st, st.SelectedIndex)
		}
		posY.OnChanged = func(val string) {
			if st.SelectedIndex < 0 || st.SelectedIndex >= len(st.Entities) {
				return
			}
			st.Entities[st.SelectedIndex].Position[1] = parse32(val)
			sendTransformIfConnected(st, st.SelectedIndex)
		}
		posZ.OnChanged = func(val string) {
			if st.SelectedIndex < 0 || st.SelectedIndex >= len(st.Entities) {
				return
			}
			st.Entities[st.SelectedIndex].Position[2] = parse32(val)
			sendTransformIfConnected(st, st.SelectedIndex)
		}

		// Rotation handlers (same pattern)
		rotW.OnChanged = func(val string) {
			if st.SelectedIndex < 0 || st.SelectedIndex >= len(st.Entities) {
				return
			}
			st.Entities[st.SelectedIndex].Rotation[0] = parse32(val)
			sendTransformIfConnected(st, st.SelectedIndex)
		}
		rotX.OnChanged = func(val string) {
			if st.SelectedIndex < 0 || st.SelectedIndex >= len(st.Entities) {
				return
			}
			st.Entities[st.SelectedIndex].Rotation[1] = parse32(val)
			sendTransformIfConnected(st, st.SelectedIndex)
		}
		rotY.OnChanged = func(val string) {
			if st.SelectedIndex < 0 || st.SelectedIndex >= len(st.Entities) {
				return
			}
			st.Entities[st.SelectedIndex].Rotation[2] = parse32(val)
			sendTransformIfConnected(st, st.SelectedIndex)
		}
		rotZ.OnChanged = func(val string) {
			if st.SelectedIndex < 0 || st.SelectedIndex >= len(st.Entities) {
				return
			}
			st.Entities[st.SelectedIndex].Rotation[3] = parse32(val)
			sendTransformIfConnected(st, st.SelectedIndex)
		}

		// Scale handlers
		scaleX.OnChanged = func(val string) {
			if st.SelectedIndex < 0 || st.SelectedIndex >= len(st.Entities) {
				return
			}
			st.Entities[st.SelectedIndex].Scale[0] = parse32(val)
			sendTransformIfConnected(st, st.SelectedIndex)
		}
		scaleY.OnChanged = func(val string) {
			if st.SelectedIndex < 0 || st.SelectedIndex >= len(st.Entities) {
				return
			}
			st.Entities[st.SelectedIndex].Scale[1] = parse32(val)
			sendTransformIfConnected(st, st.SelectedIndex)
		}
		scaleZ.OnChanged = func(val string) {
			if st.SelectedIndex < 0 || st.SelectedIndex >= len(st.Entities) {
				return
			}
			st.Entities[st.SelectedIndex].Scale[2] = parse32(val)
			sendTransformIfConnected(st, st.SelectedIndex)
		}
	}

	// Return the UI and the rebuild function
	return containerObj, func(world interface{}, st *state.EditorState, hierarchy *widget.List) {
		rebuild(world, st, hierarchy)
	}
}

// helper to send transform
func sendTransformIfConnected(st *state.EditorState, idx int) {
	if editorlink.EditorConn == nil {
		return
	}
	ent := st.Entities[idx]
	// Build message with proper casts
	msg := editorlink.MsgSetTransform{
		ID:       uint64(ent.ID),
		Position: [3]float32{ent.Position[0], ent.Position[1], ent.Position[2]},
		Rotation: [4]float32{ent.Rotation[0], ent.Rotation[1], ent.Rotation[2], ent.Rotation[3]},
		Scale:    [3]float32{ent.Scale[0], ent.Scale[1], ent.Scale[2]},
	}
	go func(m editorlink.MsgSetTransform) {
		if err := editorlink.WriteSetTransform(editorlink.EditorConn, m); err != nil {
			// optional: log error
			// log.Printf("editor: WriteSetTransform error: %v", err)
		}
	}(msg)
}
