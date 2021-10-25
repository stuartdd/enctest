package gui

import (
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type SearchData struct {
	desc string
	path string
}

type SearchDataWindow struct {
	paths          map[string]*SearchData
	checks         map[string]*widget.Button
	canSelect      bool
	closeIntercept func()
	selectFunction func(string, string)
	searchWindow   fyne.Window
}

func newSearchData(path, desc string) *SearchData {
	return &SearchData{desc: desc, path: path}
}

func (sd *SearchData) String() string {
	return sd.path
}

func NewSearchDataWindow(closeIntercept func(), selectFunction func(string, string)) *SearchDataWindow {
	return &SearchDataWindow{closeIntercept: closeIntercept, selectFunction: selectFunction, canSelect: true, paths: make(map[string]*SearchData)}
}

func (lw *SearchDataWindow) Add(desc, path string) {
	lw.paths[path] = newSearchData(path, desc)
}

func (lw *SearchDataWindow) Reset() {
	lw.paths = make(map[string]*SearchData)
}

func (lw *SearchDataWindow) Len() int {
	return len(lw.paths)
}

func (lw *SearchDataWindow) Width() float32 {
	if lw.searchWindow != nil {
		return lw.searchWindow.Canvas().Size().Width
	}
	return 500
}

func (lw *SearchDataWindow) Height() float32 {
	if lw.searchWindow != nil {
		return lw.searchWindow.Canvas().Size().Height
	}
	return 500
}

func (lw *SearchDataWindow) Showing() bool {
	return lw.searchWindow != nil
}

func (lw *SearchDataWindow) Select(path string) {
	lw.canSelect = false
	defer func() {
		lw.canSelect = true
	}()
	if lw.searchWindow != nil {
		for _, v := range lw.checks {
			v.SetIcon(theme.CheckButtonIcon())
			v.Disable()
		}
		if path != "" {
			c := lw.checks[path]
			if c != nil {
				c.SetIcon(theme.CheckButtonCheckedIcon())
				c.Enable()
			}
		}
	}
}
func (lw *SearchDataWindow) createRow(sd *SearchData) *fyne.Container {
	c := container.NewHBox()
	w := widget.NewButtonWithIcon("", theme.MailForwardIcon(), func() {})
	w.SetIcon(theme.CheckButtonIcon())
	lw.checks[sd.path] = w
	b := widget.NewButtonWithIcon("", theme.MailForwardIcon(), func() {
		if lw.canSelect {
			go lw.selectFunction(sd.desc, sd.path)
		}
	})
	c.Add(b)
	c.Add(w)
	c.Add(widget.NewLabel(sd.desc))
	return c
}

func (lw *SearchDataWindow) Show(w, h float32) {
	lw.canSelect = false
	defer func() {
		lw.canSelect = true
	}()

	pathList := make([]string, 0)
	for k := range lw.paths {
		pathList = append(pathList, k)
	}
	sort.Strings(pathList)

	lw.searchWindow = fyne.CurrentApp().NewWindow("Search List")
	vc := container.NewVBox()
	hb := container.NewHBox()
	hb.Add(widget.NewButtonWithIcon("Close", theme.CancelIcon(), func() {
		lw.Close()
	}))
	hb.Add(widget.NewLabel("Select from below to navigate to the item"))
	vc.Add(hb)
	lw.checks = make(map[string]*widget.Button)
	for _, v := range pathList {
		vc.Add(lw.createRow(lw.paths[v]))
	}
	c := container.NewScroll(vc)
	lw.searchWindow.SetContent(c)
	lw.searchWindow.SetCloseIntercept(lw.closeIntercept)
	lw.searchWindow.Resize(fyne.NewSize(w, h))
	lw.searchWindow.SetFixedSize(true)
	lw.searchWindow.Show()
}

func (lw *SearchDataWindow) Close() {
	if lw.searchWindow != nil {
		lw.searchWindow.Close()
		lw.searchWindow = nil
	}
}
