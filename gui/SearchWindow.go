package gui

import (
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type SearchData struct {
	desc string
	path string
}

type SearchDataWindow struct {
	paths          map[string]*SearchData
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
	return &SearchDataWindow{closeIntercept: closeIntercept, selectFunction: selectFunction, paths: make(map[string]*SearchData)}
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

func (lw *SearchDataWindow) Show(w, h float32) {
	go func(w, h float32) {
		pathList := make([]string, 0)
		for k, _ := range lw.paths {
			pathList = append(pathList, k)
		}
		sort.Strings(pathList)

		list := widget.NewList(
			func() int { return len(lw.paths) },
			func() fyne.CanvasObject {
				return widget.NewLabel("")
			},
			func(lii widget.ListItemID, co fyne.CanvasObject) {
				co.(*widget.Label).SetText(lw.paths[pathList[lii]].desc)
			},
		)
		list.OnSelected = func(id widget.ListItemID) {
			data := lw.paths[pathList[id]]
			go lw.selectFunction(data.desc, data.path)
		}
		c := container.NewScroll(list)
		lw.searchWindow = fyne.CurrentApp().NewWindow("Search List")
		lw.searchWindow.SetCloseIntercept(lw.closeIntercept)
		lw.searchWindow.SetContent(c)
		lw.searchWindow.Resize(fyne.NewSize(w, h))
		lw.searchWindow.SetFixedSize(true)
		lw.searchWindow.Show()
	}(w, h)
}

func (lw *SearchDataWindow) Close() {
	if lw.searchWindow != nil {
		lw.searchWindow.Close()
		lw.searchWindow = nil
	}
}
