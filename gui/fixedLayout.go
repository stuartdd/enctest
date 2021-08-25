package gui

import "fyne.io/fyne/v2"

type FixedLayout struct {
	w       float32
	yOffset float32
}

type BoxLayout struct {
	w float32
	h float32
}

func NewBoxLayout(w float32, h float32) *BoxLayout {
	return &BoxLayout{w: w, h: h}
}

func NewFixedLayout(w float32, yOffset float32) *FixedLayout {
	return &FixedLayout{w: w, yOffset: yOffset}
}

func (d *BoxLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(d.w, d.h)
}

func (d *BoxLayout) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	pos := fyne.NewPos(0, 0)
	for _, o := range objects {
		o.Resize(containerSize)
		o.Move(pos)
	}
}

func (d *FixedLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	h := float32(0)
	for _, o := range objects {
		childSize := o.MinSize()
		if childSize.Width > d.w {
			d.w = childSize.Width
		}
		h += childSize.Height
	}
	return fyne.NewSize(d.w, h)
}

func (d *FixedLayout) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	pos := fyne.NewPos(0, d.yOffset)
	for _, o := range objects {
		size := o.MinSize()
		o.Resize(size)
		o.Move(pos)
		pos = pos.Add(fyne.NewPos(size.Width, size.Height))
	}
}
