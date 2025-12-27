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
}

func NewFoldout(title string, content fyne.CanvasObject, expanded bool) *Foldout {
	f := &Foldout{
		Title:    title,
		Content:  content,
		Expanded: expanded,
	}
	f.ExtendBaseWidget(f)
	return f
}

func (f *Foldout) CreateRenderer() fyne.WidgetRenderer {
	arrow := canvas.NewText("", theme.ForegroundColor())
	title := canvas.NewText(f.Title, theme.ForegroundColor())
	title.TextStyle.Bold = true

	header := container.NewHBox(arrow, title)

	// Make header clickable
	tap := widget.NewButton("", func() {
		f.Expanded = !f.Expanded
		f.Refresh()
	})
	tap.Importance = widget.LowImportance
	tapContainer := container.NewMax(header, tap)

	// Build layout
	var objects []fyne.CanvasObject
	objects = append(objects, tapContainer)
	if f.Expanded {
		objects = append(objects, f.Content)
	}

	// Renderer
	r := &foldoutRenderer{
		foldout: f,
		arrow:   arrow,
		title:   title,
		objects: objects,
		box:     container.NewVBox(objects...),
	}
	return r
}

type foldoutRenderer struct {
	foldout *Foldout
	arrow   *canvas.Text
	title   *canvas.Text
	objects []fyne.CanvasObject
	box     *fyne.Container
}

func (r *foldoutRenderer) Layout(size fyne.Size) {
	r.box.Resize(size)
}

func (r *foldoutRenderer) MinSize() fyne.Size {
	return r.box.MinSize()
}

func (r *foldoutRenderer) Refresh() {
	if r.foldout.Expanded {
		r.arrow.Text = "▼"
	} else {
		r.arrow.Text = "▶"
	}
	r.arrow.Refresh()

	// Rebuild children
	r.box.Objects = nil
	header := container.NewHBox(r.arrow, r.title)
	tap := widget.NewButton("", func() {
		r.foldout.Expanded = !r.foldout.Expanded
		r.foldout.Refresh()
	})
	tap.Importance = widget.LowImportance
	tapContainer := container.NewMax(header, tap)
	r.box.Add(tapContainer)

	if r.foldout.Expanded {
		r.box.Add(r.foldout.Content)
	}

	r.box.Refresh()
}

func (r *foldoutRenderer) Objects() []fyne.CanvasObject {
	return r.box.Objects
}

func (r *foldoutRenderer) Destroy() {}
