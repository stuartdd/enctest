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
	logDataRequest func(string)
}

func NewLogData(closeIntercept func(), logDataRequest func(string)) *LogData {
	ld := &LogData{grid: nil, logWindow: nil, closeIntercept: closeIntercept, logDataRequest: logDataRequest}
	ld.Reset()
	return ld
}

func (lw *LogData) Reset() {
	lw.text.Reset()
	if lw.grid != nil {
		lw.grid.SetText(lw.text.String())
	}
}

func (lw *LogData) Log(l string) {
	lw.text.WriteString(l)
	lw.text.WriteString("\n")
	go func() {
		if lw.grid != nil {
			lw.grid.SetText(lw.text.String())
		}
		if lw.scroll != nil {
			lw.scroll.ScrollToBottom()
		}
	}()
}

func (lw *LogData) Window() fyne.Window {
	return lw.logWindow
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

			bClear := widget.NewButton("Clear", func() { lw.Reset() })
			bCopy := widget.NewButton("Copy", func() {
				go func() {
					lw.logWindow.Clipboard().SetContent(lw.text.String())
					lw.logDataRequest("copy")
				}()
			})
			bClose := widget.NewButton("Close", func() {
				go func() {
					lw.logDataRequest("close")
				}()
			})
			bSelect := widget.NewButton("Selected", func() {
				go func() {
					lw.logDataRequest("select")
				}()
			})
			hb := container.NewHBox(bClose, bClear, bSelect, bCopy)

			bl := container.NewBorder(hb, nil, nil, nil, sb)
			win.SetContent(bl)
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
