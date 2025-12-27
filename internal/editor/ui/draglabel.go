package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type DragLabel struct {
	widget.BaseWidget
	Text   string
	OnDrag func(delta float32)
}

func NewDragLabel(text string, onDrag func(delta float32)) *DragLabel {
	dl := &DragLabel{Text: text, OnDrag: onDrag}
	dl.ExtendBaseWidget(dl)
	return dl
}

func (d *DragLabel) CreateRenderer() fyne.WidgetRenderer {
	text := canvas.NewText(d.Text, theme.ForegroundColor())
	return widget.NewSimpleRenderer(text)
}

func (d *DragLabel) Dragged(ev *fyne.DragEvent) {
	if d.OnDrag != nil {
		d.OnDrag(float32(ev.Dragged.DX))
	}
}

func (d *DragLabel) DragEnd() {}
