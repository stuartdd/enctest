package gui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"stuartdd.com/lib"
)

var EditEntryList = make(map[string]*EditEntry)

type EditEntry struct {
	Path         string
	Title        string
	Old          string
	New          string
	Wid          *widget.Entry
	Lab          *widget.Label
	UnDo         *widget.Button
	Link         *widget.Button
	OnChangeFunc func(input string, path string)
	UnDoFunc     func(path string)
	LinkFunc     func(path string)
}

type DetailPage struct {
	Uid, Title, User string
	ViewFunc         func(w fyne.Window, details DetailPage) fyne.CanvasObject
	DataRootMap      *map[string]interface{}
}

func NewDetailEdit(path string, title string, old string, onChangeFunc func(s string, path string), unDoFunc func(path string), linkFunc func(path string)) *EditEntry {
	w := widget.NewMultiLineEntry()
	w.OnChanged = func(input string) {
		onChangeFunc(input, path)
	}
	w.SetText(old)
	l := widget.NewLabel(fmt.Sprintf(" %s ", title))
	t := theme.ContentUndoIcon()
	u := widget.NewButtonWithIcon("", t, func() {
		unDoFunc(path)
	})
	n := widget.NewButtonWithIcon("", theme.ContentRemoveIcon(), func() {
		linkFunc(path)
	})
	u.Disable()
	n.Disable()
	return &EditEntry{Path: path, Title: title, Wid: w, Lab: l, UnDo: u, Link: n, Old: old, New: "", OnChangeFunc: onChangeFunc, UnDoFunc: unDoFunc, LinkFunc: linkFunc}
}

func (p *EditEntry) SetNew(s string) {
	if p.Old == s {
		p.New = ""
		p.UnDo.Disable()
	} else {
		p.New = s
		p.UnDo.Enable()
	}
	_, ok := p.ParseForLink()
	if ok {
		p.Link.Enable()
	} else {
		p.Link.Disable()
	}
}

func (p *EditEntry) CommitEdit(data *map[string]interface{}) bool {
	m := lib.GetMapForUid(p.Path, data)
	if m != nil {
		n := *m
		n[p.Title] = p.New
		p.Old = p.New
		p.SetNew(p.New)
		return true
	}
	return false
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

func (p *EditEntry) ParseForLink() (string, bool) {
	var s string
	if p.IsChanged() {
		s = p.New
	} else {
		s = p.Old
	}
	l, ok := parseForLink(s)
	if ok {
		p.Link.Enable()
		return l, true
	} else {
		p.Link.Disable()
		return "", false
	}
}

func NewDetailPage(uid string, title string, user string, viewFunc func(w fyne.Window, details DetailPage) fyne.CanvasObject, dataRootMap *map[string]interface{}) *DetailPage {
	return &DetailPage{Uid: uid, Title: title, User: user, ViewFunc: viewFunc, DataRootMap: dataRootMap}
}

func (p *DetailPage) GetMapForUid() *map[string]interface{} {
	return lib.GetMapForUid(p.Uid, p.DataRootMap)
}

func parseForLink(s string) (string, bool) {
	var sb strings.Builder
	lc := strings.ToLower(s)
	pos := strings.Index(lc, "http://")
	count := 0
	if pos == -1 {
		pos = strings.Index(lc, "https://")
	}
	if pos == -1 {
		return "", false
	}
	for i := pos; i < len(lc); i++ {
		if lc[i] <= ' ' {
			break
		} else {
			sb.WriteByte(lc[i])
			count++
		}
	}
	if count < 12 {
		return "", false
	}
	return sb.String(), true
}
