package gui

import (
	"fmt"
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/stuartdd2/JsonParser4go/parser"
)

type SearchData struct {
	desc string
	path *parser.Path
}

type SearchDataWindow struct {
	paths          map[string]*SearchData
	checks         map[string]*widget.Button
	canSelect      bool
	selectFunction func(string, *parser.Path)
	searchWindow   fyne.Window
}

func newSearchData(path *parser.Path, desc string) *SearchData {
	return &SearchData{desc: desc, path: path}
}

func (sd *SearchData) String() string {
	return sd.path.String()
}

func NewSearchDataWindow(selectFunction func(string, *parser.Path)) *SearchDataWindow {
	return &SearchDataWindow{selectFunction: selectFunction, canSelect: true, paths: make(map[string]*SearchData)}
}

func (lw *SearchDataWindow) Add(desc string, path *parser.Path) {
	lw.paths[desc] = newSearchData(path, desc)
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

func (lw *SearchDataWindow) IsShowing() bool {
	return lw.searchWindow != nil
}

func (lw *SearchDataWindow) Select(path *parser.Path) {
	lw.canSelect = false
	defer func() {
		lw.canSelect = true
	}()
	if lw.searchWindow != nil {
		for _, v := range lw.checks {
			v.SetIcon(theme.CheckButtonIcon())
			v.Disable()
		}
		if !path.IsEmpty() {
			c := lw.checks[path.String()]
			if c != nil {
				c.SetIcon(theme.CheckButtonCheckedIcon())
				c.Enable()
			}
		}
	}
}

func (lw *SearchDataWindow) createRow(sd *SearchData, duplicate bool) *fyne.Container {
	c := container.NewHBox()
	w := widget.NewButtonWithIcon("", theme.MailForwardIcon(), func() {})
	w.SetIcon(theme.CheckButtonIcon())
	b := widget.NewButtonWithIcon("", theme.MailForwardIcon(), func() {
		if lw.canSelect {
			go lw.selectFunction(sd.desc, sd.path)
		}
	})
	if !duplicate {
		lw.checks[sd.path.String()] = w
		hb := container.NewHBox()
		hb.Add(b)
		hb.Add(w)
		c.Add(container.New(NewFixedWLayout(85), hb))
	} else {
		c.Add(container.New(NewFixedWLayout(85), widget.NewLabel("And...")))
	}
	c.Add(widget.NewLabel(sd.desc))
	return c
}

func (lw *SearchDataWindow) Show(w, h float32, searchFor string) {
	pathMap := make(map[string]string)
	lw.canSelect = false
	defer func() {
		lw.canSelect = true
	}()

	pathList := make([]string, 0)
	for k := range lw.paths {
		pathList = append(pathList, k)
	}
	sort.Strings(pathList)
	if !lw.IsShowing() {
		lw.searchWindow = fyne.CurrentApp().NewWindow("Search Results")
	}
	vc := container.NewVBox()
	hb := container.NewHBox()
	hb.Add(widget.NewButtonWithIcon("Close", theme.CancelIcon(), func() {
		lw.Close()
	}))
	hb.Add(widget.NewLabel(fmt.Sprintf("Results for \"%s\"", searchFor)))
	vc.Add(hb)
	lw.checks = make(map[string]*widget.Button)
	for _, v := range pathList {
		sd := lw.paths[v]
		_, ok := pathMap[sd.path.String()]
		if ok {
			vc.Add(lw.createRow(sd, true))
		} else {
			pathMap[sd.path.String()] = ""
			vc.Add(lw.createRow(sd, false))
		}
	}
	c := container.NewScroll(vc)
	lw.searchWindow.SetContent(c)
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
