package gui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var EditEntryList = make(map[string]*EditEntry, 0)

type EditEntry struct {
	Path         string
	Title        string
	Old          string
	New          string
	Wid          *widget.Entry
	Lab          *widget.Label
	UnDo         *widget.Button
	OnChangeFunc func(input string, path string)
	UnDoFunc     func(path string)
}

type DetailPage struct {
	Uid, Title, Intro string
	View              func(w fyne.Window, details DetailPage) fyne.CanvasObject
	DataRootMap       *map[string]interface{}
}

func NewDetailEdit(path string, title string, old string, onChangeFunc func(s string, path string), unDoFunc func(path string)) *EditEntry {
	w := widget.NewMultiLineEntry()
	w.OnChanged = func(input string) {
		onChangeFunc(input, path)
	}
	w.SetText(old)
	l := widget.NewLabel(fmt.Sprintf(" %s ", title))
	b := widget.NewButtonWithIcon("", theme.ContentUndoIcon(), func() {
		unDoFunc(path)
	})
	b.Disable()

	return &EditEntry{Path: path, Title: title, Wid: w, Lab: l, UnDo: b, Old: old, New: "", OnChangeFunc: onChangeFunc, UnDoFunc: unDoFunc}
}

func (p *EditEntry) SetNew(s string) {
	if p.Old == s {
		p.New = ""
		p.UnDo.Disable()
	} else {
		p.New = s
		p.UnDo.Enable()
	}
}

func (p *EditEntry) RevertEdit() {
	p.SetNew(p.Old)
	p.Wid.SetText(p.Old)
}

func (p *EditEntry) IsChanged() bool {
	return p.New != ""
}

func (p *EditEntry) String() string {
	if p.IsChanged() {
		return fmt.Sprintf("Item:'%s' Is updated from '%s' to '%s'", p.Path, p.Old, p.New)
	} else {
		return fmt.Sprintf("Item:'%s' Is unchanged", p.Path)
	}
}

func NewDetailPage(uid string, title, intro string, view func(w fyne.Window, details DetailPage) fyne.CanvasObject, dataRootMap *map[string]interface{}) *DetailPage {
	return &DetailPage{Uid: uid, Title: title, Intro: intro, View: view, DataRootMap: dataRootMap}
}

func (p *DetailPage) GetMapForUid() *map[string]interface{} {
	nodes := strings.Split(p.Uid, ".")
	m := *p.DataRootMap
	n := m["groups"]
	for _, v := range nodes {
		n = n.(map[string]interface{})[v]
	}
	o := n.(map[string]interface{})
	return &o
}
