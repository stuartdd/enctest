package gui

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

var _ fyne.WidgetRenderer = (*widget002Renderer)(nil)

const strokeWidth = 2

type widget002Renderer struct {
	minSize  fyne.Size
	rect1    *canvas.Rectangle
	t1       *canvas.Text
	s1       fyne.Size
	rect2    *canvas.Rectangle
	t2       *canvas.Text
	s2       fyne.Size
	txtColor color.Color
}

func newWidget002Renderer(ch rune, pos int, txtSize float32, txtColor color.Color, lineColor color.Color) *widget002Renderer {
	r := &widget002Renderer{
		rect1: &canvas.Rectangle{
			StrokeWidth: strokeWidth,
			StrokeColor: lineColor,
			FillColor:   color.Transparent,
		},
		rect2: &canvas.Rectangle{
			StrokeWidth: strokeWidth,
			StrokeColor: lineColor,
			FillColor:   color.Transparent,
		},
		txtColor: txtColor,
	}
	r.t1, r.s1 = r.MakePos(pos, txtSize)
	r.t2, r.s2 = r.MakeRune(ch, pos, txtSize)
	if pos > 9 {
		r.minSize = fyne.Size{Width: r.s1.Width + (strokeWidth * 2) + (r.s1.Width / 4.0), Height: r.s1.Height * 2.5}
	} else {
		r.minSize = fyne.Size{Width: r.s1.Width + (strokeWidth * 2) + (r.s1.Width / 1.0), Height: r.s1.Height * 2.5}
	}
	return r
}

func (r *widget002Renderer) Layout(s fyne.Size) {
	s1 := fyne.Size{Width: s.Width, Height: (s.Height / 2) - strokeWidth}
	if s1.Width <= 5 || s1.Height <= 5 {
		return
	}
	r.rect1.Resize(s1)
	r.rect2.Resize(s1)
	r.rect2.Move(fyne.Position{X: 0, Y: s.Height / 2})
	r.t1.Move(fyne.Position{X: (s.Width / 2) - (r.s1.Width / 2), Y: strokeWidth})
	r.t2.Move(fyne.Position{X: (s.Width / 2) - (r.s2.Width / 2), Y: s.Height / 2})
}

func (r *widget002Renderer) MinSize() fyne.Size {
	return r.minSize
}

func (r *widget002Renderer) Refresh() {
}

func (r *widget002Renderer) Objects() []fyne.CanvasObject {
	o := make([]fyne.CanvasObject, 0)
	o = append(o, r.rect1, r.t1, r.t2, r.rect2)
	return o
}

func (r *widget002Renderer) MakePos(pos int, txtSize float32) (*canvas.Text, fyne.Size) {
	t := canvas.NewText(fmt.Sprintf("%d", pos), r.txtColor)
	t.TextSize = txtSize
	return t, fyne.MeasureText(t.Text, t.TextSize, t.TextStyle)
}

func (r *widget002Renderer) MakeRune(ch rune, pos int, txtSize float32) (*canvas.Text, fyne.Size) {
	t := canvas.NewText(fmt.Sprintf("%c", ch), r.txtColor)
	t.TextSize = txtSize
	return t, fyne.MeasureText(t.Text, t.TextSize, t.TextStyle)
}

func (r *widget002Renderer) Destroy() {
}

var _ fyne.CanvasObject = (*Widget002)(nil)
var _ fyne.Widget = (*Widget002)(nil)

type Widget002 struct {
	widget.BaseWidget
	minSize   fyne.Size
	txtColor  color.Color
	lineColor color.Color
	txtSize   float32
	posit     int
	char      rune
}

func NewPositional(text string, txtSize float32, txtColor color.Color, lineColor color.Color) *fyne.Container {
	xL := container.New(layout.NewHBoxLayout())
	for p, r := range text {
		w := NewWidget002(r, p+1, txtSize, txtColor, lineColor)
		xL.Add(w)
	}
	return xL
}

func NewWidget002(ch rune, pos int, txtSize float32, txtColor color.Color, lineColor color.Color) *Widget002 {
	w := &Widget002{
		posit:     pos,
		char:      ch,
		txtSize:   txtSize,
		txtColor:  txtColor,
		lineColor: lineColor,
	}
	w.ExtendBaseWidget(w)
	return w
}

func (w *Widget002) CreateRenderer() fyne.WidgetRenderer {
	r := newWidget002Renderer(w.char, w.posit, w.txtSize, w.txtColor, w.lineColor)
	w.minSize = r.MinSize()
	return r
}

func (w *Widget002) MinSize() fyne.Size {
	return w.minSize
}

func (w *Widget002) Resize(s fyne.Size) {
	w.BaseWidget.Resize(s)
}
