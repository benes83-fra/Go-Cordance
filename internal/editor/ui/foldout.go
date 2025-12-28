package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Foldout struct {
	widget.BaseWidget
	Title    string
	Content  fyne.CanvasObject
	Expanded bool
	onToggle func(bool)
	Icon     fyne.Resource
}

func NewFoldout(title string, content fyne.CanvasObject, expanded bool, icon fyne.Resource) *Foldout {
	f := &Foldout{
		Title:    title,
		Content:  content,
		Expanded: expanded,
		Icon:     icon,
	}
	f.ExtendBaseWidget(f)
	return f
}

func (f *Foldout) SetOnToggle(fn func(bool)) {
	f.onToggle = fn
}

func (f *Foldout) CreateRenderer() fyne.WidgetRenderer {
	arrow := canvas.NewText("", theme.ForegroundColor())
	title := canvas.NewText(f.Title, theme.ForegroundColor())
	title.TextStyle.Bold = true

	box := container.NewVBox()

	r := &foldoutRenderer{
		foldout: f,
		arrow:   arrow,
		title:   title,
		box:     box,
	}

	r.Refresh()
	return r
}

type foldoutRenderer struct {
	foldout *Foldout
	arrow   *canvas.Text
	title   *canvas.Text
	box     *fyne.Container
}

func (r *foldoutRenderer) Layout(size fyne.Size) {
	r.box.Resize(size)
}

func (r *foldoutRenderer) MinSize() fyne.Size {
	return r.box.MinSize()
}

func (r *foldoutRenderer) Refresh() {
	// Update arrow
	if r.foldout.Expanded {
		r.arrow.Text = "▼"
	} else {
		r.arrow.Text = "▶"
	}
	r.arrow.Refresh()

	// Build header (WITH ICON SUPPORT)
	var header fyne.CanvasObject
	if r.foldout.Icon != nil {
		img := canvas.NewImageFromResource(r.foldout.Icon)
		img.SetMinSize(fyne.NewSize(16, 16))
		header = container.NewHBox(r.arrow, img, r.title)
	} else {
		header = container.NewHBox(r.arrow, r.title)
	}

	// Clickable overlay
	tap := widget.NewButton("", func() {
		r.foldout.Expanded = !r.foldout.Expanded
		if r.foldout.onToggle != nil {
			r.foldout.onToggle(r.foldout.Expanded)
		}
		r.foldout.Refresh()
	})
	tap.Importance = widget.LowImportance

	headerClickable := container.NewMax(header, tap)

	// Rebuild layout
	r.box.Objects = []fyne.CanvasObject{headerClickable}

	if r.foldout.Expanded {
		r.box.Add(r.foldout.Content)
	}

	r.box.Refresh()
}

func (r *foldoutRenderer) Objects() []fyne.CanvasObject {
	return r.box.Objects
}

func (r *foldoutRenderer) Destroy() {}
