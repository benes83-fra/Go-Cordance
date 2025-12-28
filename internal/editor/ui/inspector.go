package ui

import (
	state "go-engine/Go-Cordance/internal/editor/state"
	"go-engine/Go-Cordance/internal/editorlink"
	"math"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/go-gl/mathgl/mgl32"
)

func parse32(s string) float32 {
	f, err := strconv.ParseFloat(s, 32)
	if err != nil {
		return 0
	}

	return float32(f)
}

// NewInspectorPanel returns the inspector container and a rebuild function.
func NewInspectorPanel() (fyne.CanvasObject, func(world interface{}, st *state.EditorState, hierarchy *widget.List)) {
	// Position entries
	posX := widget.NewEntry()
	posY := widget.NewEntry()
	posZ := widget.NewEntry()
	posBox := container.NewVBox(
		widget.NewLabel("Position"),
		container.NewHBox(widget.NewLabel("X"), posX),
		container.NewHBox(widget.NewLabel("Y"), posY),
		container.NewHBox(widget.NewLabel("Z"), posZ),
	)

	// Rotation entries (w,x,y,z) or whatever you prefer
	// Rotation entries (Euler degrees)
	rotX := widget.NewEntry()
	rotY := widget.NewEntry()
	rotZ := widget.NewEntry()

	rotBox := container.NewVBox(
		widget.NewLabel("Rotation (Euler degrees)"),
		container.NewHBox(
			NewDragLabel("X°", func(dx float32) {
				// dx is pixels dragged horizontally
				step := dx * 0.2 // sensitivity
				val := parse32(rotX.Text)
				val += step
				rotX.SetText(strconv.FormatFloat(float64(val), 'f', 3, 32))
				rotX.OnChanged(rotX.Text) // trigger update
			}),
			rotX,
		),

		container.NewHBox(
			NewDragLabel("Y°", func(dx float32) {
				step := dx * 0.2
				val := parse32(rotY.Text)
				val += step
				rotY.SetText(strconv.FormatFloat(float64(val), 'f', 3, 32))
				rotY.OnChanged(rotY.Text)
			}),
			rotY,
		),
		container.NewHBox(
			NewDragLabel("Z°", func(dx float32) {
				step := dx * 0.2
				val := parse32(rotZ.Text)
				val += step
				rotZ.SetText(strconv.FormatFloat(float64(val), 'f', 3, 32))
				rotZ.OnChanged(rotZ.Text)
			}),
			rotZ,
		),
	)

	// Scale entries
	scaleX := widget.NewEntry()
	scaleY := widget.NewEntry()
	scaleZ := widget.NewEntry()

	scaleBox := container.NewVBox(
		widget.NewLabel("Scale"),
		container.NewHBox(widget.NewLabel("X"), scaleX),
		container.NewHBox(widget.NewLabel("Y"), scaleY),
		container.NewHBox(widget.NewLabel("Z"), scaleZ),
	)

	// Layout

	root := container.NewVBox()

	// Rebuild function will be filled by the caller to capture world/st/hierarchy.
	var rebuild func(world interface{}, st *state.EditorState, hierarchy *widget.List)

	// Helper to parse float safely

	// OnChanged handlers will be set inside rebuild so they capture st.SelectedIndex correctly.
	rebuild = func(world interface{}, st *state.EditorState, hierarchy *widget.List) {
		root.Objects = nil

		if st.Foldout == nil {
			st.Foldout = map[string]bool{
				"Position": true,
				"Rotation": true,
				"Scale":    true,
			}
		}
		posFoldout := NewFoldout("Position", posBox, st.Foldout["Position"], theme.ZoomInIcon())
		rotFoldout := NewFoldout("Rotation", rotBox, st.Foldout["Rotation"], theme.VisibilityIcon())
		scaleFoldout := NewFoldout("Scale", scaleBox, st.Foldout["Scale"], theme.ZoomOutIcon())

		posFoldout.SetOnToggle(func(expanded bool) {
			st.Foldout["Position"] = expanded
		})
		rotFoldout.SetOnToggle(func(expanded bool) {
			st.Foldout["Rotation"] = expanded
		})
		scaleFoldout.SetOnToggle(func(expanded bool) {
			st.Foldout["Scale"] = expanded
		})
		root.Add(posFoldout)
		root.Add(rotFoldout)
		root.Add(scaleFoldout)
		root.Refresh()
		// Defensive: no selection
		if st.SelectedIndex < 0 || st.SelectedIndex >= len(st.Entities) {
			posX.SetText("")
			posY.SetText("")
			posZ.SetText("")

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

		// Convert quaternion → Euler degrees for UI
		// Convert quaternion → Euler degrees
		q := mgl32.Quat{
			W: ent.Rotation[0],
			V: mgl32.Vec3{ent.Rotation[1], ent.Rotation[2], ent.Rotation[3]},
		}
		pitch, yaw, roll := quatToEuler(q)

		rotX.SetText(strconv.FormatFloat(float64(pitch*180/math.Pi), 'f', 3, 32))
		rotY.SetText(strconv.FormatFloat(float64(yaw*180/math.Pi), 'f', 3, 32))
		rotZ.SetText(strconv.FormatFloat(float64(roll*180/math.Pi), 'f', 3, 32))

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
		applyEuler := func() {
			if st.SelectedIndex < 0 || st.SelectedIndex >= len(st.Entities) {
				return
			}

			ex := parse32(rotX.Text) * (math.Pi / 180)
			ey := parse32(rotY.Text) * (math.Pi / 180)
			ez := parse32(rotZ.Text) * (math.Pi / 180)

			// Convert Euler → quaternion
			q := mgl32.AnglesToQuat(ex, ey, ez, mgl32.ZYX)

			st.Entities[st.SelectedIndex].Rotation[0] = q.W
			st.Entities[st.SelectedIndex].Rotation[1] = q.V[0]
			st.Entities[st.SelectedIndex].Rotation[2] = q.V[1]
			st.Entities[st.SelectedIndex].Rotation[3] = q.V[2]

			sendTransformIfConnected(st, st.SelectedIndex)
		}

		rotX.OnChanged = func(_ string) { applyEuler() }
		rotY.OnChanged = func(_ string) { applyEuler() }
		rotZ.OnChanged = func(_ string) { applyEuler() }

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
	return root, func(world interface{}, st *state.EditorState, hierarchy *widget.List) {
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

// Convert quaternion → Euler (pitch=X, yaw=Y, roll=Z), radians
func quatToEuler(q mgl32.Quat) (float32, float32, float32) {
	// Reference: https://en.wikipedia.org/wiki/Conversion_between_quaternions_and_Euler_angles
	w, x, y, z := q.W, q.V[0], q.V[1], q.V[2]

	// Pitch (X axis)
	sinp := 2 * (w*x + y*z)
	cosp := 1 - 2*(x*x+y*y)
	pitch := float32(math.Atan2(float64(sinp), float64(cosp)))

	// Yaw (Y axis)
	siny := 2 * (w*y - z*x)
	var yaw float32
	if math.Abs(float64(siny)) >= 1 {
		yaw = float32(math.Copysign(math.Pi/2, float64(siny)))
	} else {
		yaw = float32(math.Asin(float64(siny)))
	}

	// Roll (Z axis)
	sinr := 2 * (w*z + x*y)
	cosr := 1 - 2*(y*y+z*z)
	roll := float32(math.Atan2(float64(sinr), float64(cosr)))

	return pitch, yaw, roll
}

// Convert Euler radians → quaternion
func eulerToQuat(pitch, yaw, roll float32) mgl32.Quat {
	// mgl32.AnglesToQuat uses intrinsic rotations, ZYX order
	return mgl32.AnglesToQuat(pitch, yaw, roll, mgl32.ZYX)
}
