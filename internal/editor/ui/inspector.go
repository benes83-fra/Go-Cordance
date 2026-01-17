package ui

import (
	"fmt"
	"go-engine/Go-Cordance/internal/ecs"
	state "go-engine/Go-Cordance/internal/editor/state"
	"go-engine/Go-Cordance/internal/editorlink"
	"log"
	"math"
	"sort"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/go-gl/mathgl/mgl32"
)

var dlg dialog.Dialog

func parse32(s string) float32 {
	f, err := strconv.ParseFloat(s, 32)
	if err != nil {
		return 0
	}

	return float32(f)
}

// NewInspectorPanel returns the inspector container and a rebuild function.
func NewInspectorPanel() (
	fyne.CanvasObject,
	func(world interface{}, st *state.EditorState, hierarchy *widget.List),

) {
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
		st.IsRebuilding = true
		defer func() { st.IsRebuilding = false }()
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
		// Create left column for transform
		left := container.NewVBox(posFoldout, rotFoldout, scaleFoldout)

		// Create right column for components
		right := container.NewVBox()

		if st.SelectedIndex >= 0 && st.SelectedIndex < len(st.Entities) {
			entInfo := st.Entities[st.SelectedIndex]
			ecsWorld := world.(*ecs.World)
			ecsEnt := ecsWorld.FindByID(entInfo.ID)
			names := append([]string{}, entInfo.Components...)
			sort.Strings(names)

			if ecsEnt != nil {

				sort.Strings(names)
				for _, name := range names {
					constructor, ok := ecs.ComponentRegistry[name]
					if !ok {
						log.Printf("editor: no constructor for component %q in registry", name)
						continue
					}

					comp := ecsEnt.GetComponent(constructor())
					if comp == nil {
						log.Printf("editor: entity %d has %q in snapshot but not in ECS world", entInfo.ID, name)
						continue
					}

					if insp, ok := comp.(ecs.EditorInspectable); ok {
						fold := buildComponentUI(insp, entInfo.ID, func() {
							rebuild(world, st, hierarchy)
						})
						right.Add(fold)
					}
				}

			}
			addBtn := widget.NewButton("Add Component", func() {
				log.Printf("AddButton for entity %d (%s)", entInfo.ID, entInfo.Name)

				showAddComponentDialog(ecsWorld, ecsEnt, entInfo.ID, entInfo.Components, func() {
					rebuild(world, st, hierarchy)
				})

			})

			right.Add(addBtn)
		}
		leftScroll := container.NewScroll(left)
		rightScroll := container.NewScroll(right)

		// Create a resizable splitter
		split := container.NewHSplit(leftScroll, rightScroll)

		// Apply saved offset (default 0.35 if zero)
		if st.SplitOffset > 0 {
			split.Offset = st.SplitOffset
		} else {
			split.Offset = 0.35
		}
		// Save the offset after layout
		go func() {
			// Let Fyne finish layout
			time.Sleep(50 * time.Millisecond)
			st.SplitOffset = split.Offset
		}()

		// Optional: minimum sizes so neither side collapses
		leftScroll.SetMinSize(fyne.NewSize(200, 750))
		rightScroll.SetMinSize(fyne.NewSize(200, 750))

		root.Objects = []fyne.CanvasObject{split}

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
func buildComponentUI(c ecs.EditorInspectable, entityID int64, refresh func()) fyne.CanvasObject {
	fields := c.EditorFields()
	box := container.NewVBox()
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, name := range keys {
		value := fields[name]
		switch v := value.(type) {

		case float32:
			e := widget.NewEntry()
			e.SetText(fmt.Sprintf("%.3f", v))
			e.OnSubmitted = func(s string) {
				if state.Global.IsRebuilding {
					return
				}
				c.SetEditorField(name, parse32(s))
				sendComponentUpdate(entityID, c)
			}

			box.Add(container.NewHBox(widget.NewLabel(name), e))

		case [3]float32:
			x := widget.NewEntry()
			y := widget.NewEntry()
			z := widget.NewEntry()

			x.SetText(fmt.Sprintf("%.3f", v[0]))
			y.SetText(fmt.Sprintf("%.3f", v[1]))
			z.SetText(fmt.Sprintf("%.3f", v[2]))

			x.OnSubmitted = func(s string) {
				if state.Global.IsRebuilding {
					return
				}
				v[0] = parse32(s)
				c.SetEditorField(name, v)
				sendComponentUpdate(entityID, c)
			}

			y.OnSubmitted = func(s string) {
				if state.Global.IsRebuilding {
					return
				}
				v[1] = parse32(s)
				c.SetEditorField(name, v)
				sendComponentUpdate(entityID, c)
			}

			z.OnSubmitted = func(s string) {
				if state.Global.IsRebuilding {
					return
				}
				v[2] = parse32(s)
				c.SetEditorField(name, v)
				sendComponentUpdate(entityID, c)
			}

			box.Add(container.NewVBox(
				widget.NewLabel(name),
				container.NewHBox(widget.NewLabel("X"), x),
				container.NewHBox(widget.NewLabel("Y"), y),
				container.NewHBox(widget.NewLabel("Z"), z),
			))
		case [4]float32:
			r := widget.NewEntry()
			g := widget.NewEntry()
			b := widget.NewEntry()
			a := widget.NewEntry()

			r.SetText(fmt.Sprintf("%.3f", v[0]))
			g.SetText(fmt.Sprintf("%.3f", v[1]))
			b.SetText(fmt.Sprintf("%.3f", v[2]))
			a.SetText(fmt.Sprintf("%.3f", v[3]))

			apply := func() {
				if state.Global.IsRebuilding {
					return
				}
				v[0] = parse32(r.Text)
				v[1] = parse32(g.Text)
				v[2] = parse32(b.Text)
				v[3] = parse32(a.Text)
				c.SetEditorField(name, v)
				sendComponentUpdate(entityID, c)
			}

			r.OnSubmitted = func(_ string) { apply() }
			g.OnSubmitted = func(_ string) { apply() }
			b.OnSubmitted = func(_ string) { apply() }
			a.OnSubmitted = func(_ string) { apply() }

			box.Add(container.NewVBox(
				widget.NewLabel(name),
				container.NewHBox(widget.NewLabel("R"), r),
				container.NewHBox(widget.NewLabel("G"), g),
				container.NewHBox(widget.NewLabel("B"), b),
				container.NewHBox(widget.NewLabel("A"), a),
			))

		case bool:
			initial := v
			chk := widget.NewCheck(name, nil)
			chk.SetChecked(v)
			log.Printf("Checkbox init: %s = %v (entity %d)", name, v, entityID)

			last := initial
			fieldName := name

			chk.OnChanged = func(newVal bool) {
				if state.Global.IsRebuilding {
					return
				}
				log.Printf("Checkbox changed: %s = %v (entity %d)", name, newVal, entityID)

				if newVal == last {
					return
				}
				last = newVal
				c.SetEditorField(fieldName, newVal)
				sendComponentUpdate(entityID, c)

			}

			box.Add(chk)
		case string:
			e := widget.NewEntry()
			e.SetText(v)
			e.OnChanged = func(s string) {
				if state.Global.IsRebuilding {
					return
				}
				c.SetEditorField(name, s)
				sendComponentUpdate(entityID, c)
			}
			box.Add(container.NewHBox(widget.NewLabel(name), e))
		case uint32:
			if name == "TextureID" {
				if fields["UseTexture"].(bool) {
					texName := lookupTextureName(uint32(v))
					options := state.Global.TextureNames

					// 1. Create dropdown with no callback yet
					dropdown := widget.NewSelect(options, nil)

					// 2. Preselect current texture WITHOUT triggering a send
					if texName != "" {
						dropdown.SetSelected(texName)
					}

					// 3. Now wire the callback for real user changes
					currentID := v
					dropdown.OnChanged = func(selected string) {
						if state.Global.IsRebuilding {
							return
						}

						id := lookupTextureID(selected)
						if id == currentID {
							// avoid spamming identical updates
							return
						}
						currentID = id
						c.SetEditorField("TextureID", id)
						sendComponentUpdate(entityID, c)
					}

					box.Add(container.NewHBox(widget.NewLabel("Texture"), dropdown))
				}
			}

		case int:
			// Special case: LightType dropdown
			if name == "Type" {
				options := []string{"Directional", "Point", "Spot"}
				dropdown := widget.NewSelect(options, nil)

				if v >= 0 && v < len(options) {
					dropdown.SetSelected(options[v])

				}
				dropdown.OnChanged = func(selected string) {
					if state.Global.IsRebuilding {
						return
					}
					var idx int
					switch selected {
					case "Directional":
						idx = 0
					case "Point":
						idx = 1
					case "Spot":
						idx = 2

					}
					c.SetEditorField(name, idx)
					sendComponentUpdate(entityID, c)

				}
				box.Add(container.NewHBox(widget.NewLabel("Type"), dropdown))
			}

		}
	}

	removeBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		sendRemoveComponent(entityID, c.EditorName())

		// --- NEW: update editor snapshot ---
		for i := range state.Global.Entities {
			if state.Global.Entities[i].ID == entityID {
				comps := state.Global.Entities[i].Components
				newList := make([]string, 0, len(comps))
				for _, cname := range comps {
					if cname != c.EditorName() {
						newList = append(newList, cname)
					}
				}
				state.Global.Entities[i].Components = newList
				break
			}
		}

		refresh()
	})

	header := container.NewHBox(
		widget.NewLabel(c.EditorName()),
		layout.NewSpacer(),
		removeBtn,
	)

	return NewFoldoutWithHeader(header, box, true)

}

func sendComponentUpdate(entityID int64, c ecs.EditorInspectable) {
	if editorlink.EditorConn == nil {
		return
	}
	log.Printf("Component: %d  with Name: %s set to: %v", entityID, c.EditorName(), c.EditorFields())
	msg := editorlink.MsgSetComponent{
		EntityID: uint64(entityID),
		Name:     c.EditorName(),
		Fields:   c.EditorFields(),
	}

	go editorlink.WriteSetComponent(editorlink.EditorConn, msg)
}

func showAddComponentDialog(world *ecs.World, ent *ecs.Entity, entityID int64, components []string, refresh func()) {
	items := []string{}
	log.Printf("showAddComponentDialog for entityID=%d", entityID)
	// Build a set of existing components from the snapshot
	existing := map[string]bool{}
	for _, cname := range components {
		existing[cname] = true
	}

	// Show only components NOT in the snapshot
	for name := range ecs.ComponentRegistry {
		if !existing[name] {
			items = append(items, name)
		}
	}

	log.Printf(" -> len(items)=%d", len(items))

	if len(items) == 0 {
		dialog.ShowInformation("Add Component", "All components already added.", fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}

	// Build a clean vertical list of buttons
	buttons := container.NewVBox()
	for _, name := range items {
		compName := name
		btn := widget.NewButton(compName, func() {
			constructor := ecs.ComponentRegistry[compName]
			newComp := constructor()
			ent.AddComponent(newComp)

			// --- NEW: update editor snapshot ---
			for i := range state.Global.Entities {
				if state.Global.Entities[i].ID == entityID {
					state.Global.Entities[i].Components = append(
						state.Global.Entities[i].Components,
						compName,
					)
					break
				}
			}

			if insp, ok := newComp.(ecs.EditorInspectable); ok {
				sendComponentUpdate(entityID, insp)
			}

			dlg.Hide()
			refresh()
		})

		buttons.Add(btn)
	}

	dlg = dialog.NewCustom(
		"Add Component",
		"Close",
		container.NewVBox(
			widget.NewLabel("Choose a component to add:"),
			buttons,
		),
		fyne.CurrentApp().Driver().AllWindows()[0],
	)

	dlg.Resize(fyne.NewSize(300, 400)) // prevents scrollbars
	dlg.Show()
}

func sendRemoveComponent(entityID int64, name string) {
	if editorlink.EditorConn == nil {
		return
	}
	log.Printf("Removing Component %d: %s", entityID, name)

	msg := editorlink.MsgRemoveComponent{
		EntityID: uint64(entityID),
		Name:     name,
	}

	go editorlink.WriteRemoveComponent(editorlink.EditorConn, msg)

}
func lookupTextureID(name string) uint32 {
	for i, n := range state.Global.TextureNames {
		if n == name {
			return state.Global.TextureIDs[i]
		}
	}
	return 0
}
func lookupTextureName(id uint32) string {
	for i, tid := range state.Global.TextureIDs {
		if tid == id {
			return state.Global.TextureNames[i]
		}
	}
	return ""
}
