package gui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type LogData struct {
	text           strings.Builder
	grid           *widget.TextGrid
	scroll         *container.Scroll
	logWindow      fyne.Window
	closeIntercept func()
}

func NewLogData(closeIntercept func()) *LogData {
	ld := &LogData{grid: nil, logWindow: nil, closeIntercept: closeIntercept}
	ld.Reset()
	return ld
}

func (lw *LogData) Reset() *LogData {
	lw.text.Reset()
	if lw.grid != nil {
		lw.grid.SetText(lw.text.String())
	}
	return lw
}

func (lw *LogData) Log(l string) *LogData {
	lw.text.WriteString(l)
	lw.text.WriteString("\n")
	if lw.grid != nil {
		lw.grid.SetText(lw.text.String())
	}
	if lw.scroll != nil {
		lw.scroll.ScrollToBottom()
	}
	return lw
}

func (lw *LogData) Width() float32 {
	if lw.logWindow != nil {
		return lw.logWindow.Canvas().Size().Width
	}
	return 500
}

func (lw *LogData) Height() float32 {
	if lw.logWindow != nil {
		return lw.logWindow.Canvas().Size().Height
	}
	return 500
}

func (lw *LogData) Showing() bool {
	return lw.logWindow != nil
}

func (lw *LogData) Show(w, h float32) {
	go func(w, h float32) {
		if lw.logWindow == nil {
			win := fyne.CurrentApp().NewWindow("Log Window")
			win.Resize(fyne.NewSize(w, h))
			win.SetCloseIntercept(lw.closeIntercept)
			tg := widget.NewTextGridFromString(lw.text.String())
			sb := container.NewScroll(tg)
			win.SetContent(sb)
			lw.grid = tg
			lw.logWindow = win
			lw.scroll = sb
		}
		lw.logWindow.Show()
	}(w, h)
}

func (lw *LogData) Close() {
	if lw.logWindow != nil {
		lw.grid.SetText("")
		lw.logWindow.Close()
		lw.logWindow = nil
	}
}
