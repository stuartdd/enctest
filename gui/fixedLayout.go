package gui

import (
	"fyne.io/fyne/v2"
)

type FixedLayout struct {
	w       float32
	yOffset float32
}

type FixedWHLayout struct {
	w float32
	h float32
}

func NewFixedLayout(w float32, yOffset float32) *FixedLayout {
	return &FixedLayout{w: w, yOffset: yOffset}
}

func (d *FixedWHLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(d.w, d.h)
}

func (d *FixedWHLayout) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	pos := fyne.NewPos(-8, -(d.h/2 + 4))
	for _, o := range objects {
		size := fyne.NewSize(d.w, d.h)
		o.Resize(size)
		o.Move(pos)
		pos = pos.Add(fyne.NewPos(size.Width, size.Height))
	}

}

func NewFixedWHLayout(w float32, h float32) *FixedWHLayout {
	return &FixedWHLayout{w: w, h: h}
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
