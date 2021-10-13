/*
 * Copyright (C) 2018 Stuart Davies (stuartdd)
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */
package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

var Padding50 = container.New(NewFixedWLayout(50))
var Padding100 = container.New(NewFixedWLayout(100))

//-----------------------------------------------------------------------------

type FixedHLayout struct {
	minW float32
	h    float32
}

func NewFixedHLayout(minW, h float32) *FixedHLayout {
	return &FixedHLayout{minW: minW, h: h}
}

func (d *FixedHLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(d.minW, d.h)
}

func (d *FixedHLayout) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	for _, o := range objects {
		if containerSize.Width < d.minW {
			o.Resize(fyne.NewSize(d.minW, d.h))
		} else {
			o.Resize(fyne.NewSize(containerSize.Width, d.h))
		}
	}
}

//-----------------------------------------------------------------------------

type FixedWLayout struct {
	w float32
	h float32
}

func NewFixedWLayout(w float32) *FixedWLayout {
	return &FixedWLayout{w: w, h: 0}
}

func (d *FixedWLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(d.w, d.h)
}

func (d *FixedWLayout) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	for _, o := range objects {
		d.h = containerSize.Height
		o.Resize(fyne.NewSize(d.w, d.h))
	}
}

//-----------------------------------------------------------------------------

type FixedLayout struct {
	w       float32
	yOffset float32
}

func NewFixedLayout(w float32, yOffset float32) *FixedLayout {
	return &FixedLayout{w: w, yOffset: yOffset}
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

//-----------------------------------------------------------------------------

type FixedWHLayout struct {
	w float32
	h float32
}

func NewFixedWHLayout(w float32, h float32) *FixedWHLayout {
	return &FixedWHLayout{w: w, h: h}
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
