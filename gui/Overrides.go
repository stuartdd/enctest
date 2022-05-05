/*
 * Copyright (C) 2021 Stuart Davies (stuartdd)
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
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"stuartdd.com/lib"
)

var Padding15 = container.New(NewFixedWLayout(15))
var Padding25 = container.New(NewFixedWLayout(25))
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

func NewFloatFieldRight(f float64, w int) *widget.Label {
	num := fmt.Sprintf("%9.2f", f)
	return NewStringFieldRight(num, w)
}

func NewStringFieldRight(s string, w int) *widget.Label {
	ts := fyne.TextStyle{Bold: true, Italic: false, Monospace: true, Symbol: false, TabWidth: 2}
	return widget.NewLabelWithStyle(lib.PadLeft(s, w), fyne.TextAlignLeading, ts)
}

func NewStringFieldLeft(s string, w int) *widget.Label {
	ts := fyne.TextStyle{Bold: true, Italic: false, Monospace: true, Symbol: false, TabWidth: 2}
	return widget.NewLabelWithStyle(lib.PadRight(s, w), fyne.TextAlignLeading, ts)
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

// --------------------------------------------------
type MyButton struct {
	widget.Button
	statusMessage string
	statusDisplay *StatusDisplay
	isBlank       bool
	myTapped      func(string, string)
	userData1     string
	userData2     string
}

func NewMyIconButton(label string, icon fyne.Resource, tapped func(string, string), ud1, ud2 string, sd *StatusDisplay, sm string) *MyButton {
	mybutton := &MyButton{statusDisplay: sd, userData1: ud1, userData2: ud2, statusMessage: sm}
	mybutton.ExtendBaseWidget(mybutton)
	mybutton.SetIcon(icon)
	mybutton.myTapped = tapped
	mybutton.OnTapped = func() {
		mybutton.myTapped(mybutton.userData1, mybutton.userData2)
	}
	mybutton.SetText(label)
	mybutton.isBlank = false
	return mybutton
}

func NewMyBlankButton(icon fyne.Resource, sd *StatusDisplay, sm string) *MyButton {
	mybutton := &MyButton{statusDisplay: sd, statusMessage: sm}
	mybutton.ExtendBaseWidget(mybutton)
	mybutton.SetIcon(icon)
	mybutton.OnTapped = nil
	mybutton.isBlank = true
	mybutton.Disable()
	return mybutton
}

func (t *MyButton) SetStatusMessage(message string) {
	t.statusMessage = message
	t.statusDisplay.Reset()
}

func (t *MyButton) MouseIn(me *desktop.MouseEvent) {
	t.statusDisplay.PushStatus(t.statusMessage)
}

func (t *MyButton) MouseOut() {
	t.statusDisplay.PopStatus()
}

func (t *MyButton) MyEnable() {
	if t.isBlank {
		t.Disable()
	} else {
		t.Enable()
	}
}

type StatusDisplay struct {
	statusLabel     *widget.Label
	updatedText     *widget.Label
	StatusContainer *fyne.Container
	statusStack     *StringStack
	initialText     string
	prefix          string
	current         string
}

func NewStatusDisplay(initialText, updateText, prefix string) *StatusDisplay {
	s1 := widget.NewLabel(initialText)
	s1.TextStyle.Monospace = true
	s2 := widget.NewLabel(updateText)
	s2.TextStyle.Monospace = true
	c1 := container.New(layout.NewHBoxLayout(), s2, widget.NewSeparator(), s1)
	c2 := container.NewVBox(widget.NewSeparator(), c1)
	ss := NewStringStack()
	return &StatusDisplay{statusLabel: s1, updatedText: s2, StatusContainer: c2, statusStack: ss, initialText: initialText, prefix: prefix, current: ""}
}

func (sd *StatusDisplay) SetUpdated(dt string) {
	sd.updatedText.SetText("Last Updated: " + dt)
}

func (sd *StatusDisplay) Reset() {
	sd.statusStack = NewStringStack()
	sd.PopStatus()
}

func (sd *StatusDisplay) PushStatus(m string) {
	sd.statusStack.Push(sd.current)
	sd.current = m
	sd.statusLabel.SetText(fmt.Sprintf("%s: %s", sd.prefix, m))
	sd.StatusContainer.Refresh()
}

func (sd *StatusDisplay) PopStatus() {
	m := sd.statusStack.Pop()
	if m == "" {
		m = sd.initialText
	}
	sd.current = m
	sd.statusLabel.SetText(fmt.Sprintf("%s: %s", sd.prefix, m))
	sd.StatusContainer.Refresh()
}

type StringStack struct {
	stack []string
}

func NewStringStack() *StringStack {
	return &StringStack{make([]string, 0)}
}

// IsEmpty: check if stack is empty
func (s *StringStack) IsEmpty() bool {
	return len(s.stack) == 0
}

// Push a new value onto the stack
func (s *StringStack) Push(str string) {
	s.stack = append(s.stack, str) // Simply append the new value to the end of the stack
}

// Remove and return top element of stack. Return false if stack is empty.
func (s *StringStack) Pop() string {
	if s.IsEmpty() {
		return ""
	} else {
		index := len(s.stack) - 1   // Get the index of the top most element.
		element := (s.stack)[index] // Index into the slice and obtain the element.
		s.stack = (s.stack)[:index] // Remove it from the stack by slicing it off.
		return element
	}
}
