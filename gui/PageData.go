package gui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"stuartdd.com/lib"
	"stuartdd.com/theme2"
)

var EditEntryList = make(map[string]*EditEntry)

type EditEntry struct {
	Path         string
	Title        string
	Old          string
	New          string
	Url          string
	Wid          *widget.Entry
	Lab          *widget.Label
	UnDo         *widget.Button
	Link         *widget.Button
	Remove       *widget.Button
	Rename       *widget.Button
	OnChangeFunc func(input string, path string)
	UnDoFunc     func(path string)
	ActionFunc   func(action string, path string)
}

type DetailPage struct {
	Uid, Heading, Title, User string
	ViewFunc                  func(w fyne.Window, details DetailPage, actionFunc func(string, string)) fyne.CanvasObject
	CntlFunc                  func(w fyne.Window, details DetailPage, actionFunc func(string, string)) fyne.CanvasObject
	DataRootMap               *map[string]interface{}
}

func NewEditEntry(path string, title string, old string, onChangeFunc func(s string, path string), unDoFunc func(path string), actionFunc func(action string, uid string)) *EditEntry {
	var w *widget.Entry
	if strings.Contains(strings.ToLower(title), "note") {
		w = widget.NewMultiLineEntry()
	} else {
		w = widget.NewEntry()
	}
	w.OnChanged = func(input string) {
		onChangeFunc(input, path)
	}
	w.SetText(old)
	l := widget.NewLabel(fmt.Sprintf(" %s ", title))
	u := widget.NewButtonWithIcon("", theme.ContentUndoIcon(), func() {
		unDoFunc(path)
	})
	i := widget.NewButtonWithIcon("", theme2.LinkToWebIcon(), func() {
		actionFunc(ACTION_LINK, "")
	})
	r := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		actionFunc(ACTION_REMOVE, path)
	})
	n := widget.NewButtonWithIcon("", theme2.EditIcon(), func() {
		actionFunc(ACTION_RENAME, path)
	})
	u.Disable()
	i.Disable()
	return &EditEntry{Path: path, Title: title, Wid: w, Lab: l, UnDo: u, Link: i, Remove: r, Rename: n, Old: old, New: "", OnChangeFunc: onChangeFunc, UnDoFunc: unDoFunc, ActionFunc: actionFunc}
}

func (p *EditEntry) RefreshButtons() {
	p.UnDo = widget.NewButtonWithIcon("", theme.ContentUndoIcon(), func() {
		p.UnDoFunc(p.Path)
	})
	p.Link = widget.NewButtonWithIcon("", theme2.LinkToWebIcon(), func() {
		p.ActionFunc(ACTION_LINK, p.Url)
	})
	p.Remove = widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		p.ActionFunc(ACTION_REMOVE, p.Path)
	})
	p.Rename = widget.NewButtonWithIcon("", theme2.EditIcon(), func() {
		p.ActionFunc(ACTION_RENAME, p.Path)
	})
	p.updateButtons()
}

func (p *EditEntry) SetNew(s string) {
	if p.Old == s {
		p.New = ""
	} else {
		p.New = s
	}
	p.updateButtons()
}

func (p *EditEntry) CommitEdit(data *map[string]interface{}) bool {
	m, _ := lib.GetMapForUid(p.Path, data)
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

func (p *EditEntry) String() string {
	if p.IsChanged() {
		return fmt.Sprintf("Item:'%s' Is updated from '%s' to '%s'", p.Path, p.Old, p.New)
	} else {
		return fmt.Sprintf("Item:'%s' Is unchanged", p.Path)
	}
}

func (p *EditEntry) IsChanged() bool {
	return p.New != ""
}

func (p *EditEntry) HasLink() (string, bool) {
	lnk, ok := parseStringForLink(p.GetCurrentText())
	return lnk, ok
}

func (p *EditEntry) updateButtons() {
	l, ok := p.HasLink()
	if ok {
		p.Url = l
		p.Link.Enable()
	} else {
		p.Url = ""
		p.Link.Disable()
	}
	if p.IsChanged() {
		p.UnDo.Enable()
	} else {
		p.UnDo.Disable()
	}
}

func (p *EditEntry) GetCurrentText() string {
	if p.IsChanged() {
		return p.New
	} else {
		return p.Old
	}
}

func NewDetailPage(uid string, title string, user string, viewFunc func(w fyne.Window, details DetailPage, actionFunc func(string, string)) fyne.CanvasObject, cntlFunc func(w fyne.Window, details DetailPage, actionFunc func(string, string)) fyne.CanvasObject, dataRootMap *map[string]interface{}) *DetailPage {
	heading := fmt.Sprintf("User:  %s", title)
	if user != "" {
		heading = fmt.Sprintf("User:  %s:  %s", user, title)
	}
	return &DetailPage{Uid: uid, Heading: heading, Title: title, User: user, ViewFunc: viewFunc, CntlFunc: cntlFunc, DataRootMap: dataRootMap}
}

func (p *DetailPage) GetMapForUid() *map[string]interface{} {
	m, _ := lib.GetMapForUid(p.Uid, p.DataRootMap)
	return m
}

func parseStringForLink(s string) (string, bool) {
	var sb strings.Builder
	lc := strings.ToLower(s)
	pos := strings.Index(lc, "http://")
	if pos == -1 {
		pos = strings.Index(lc, "https://")
	}
	if pos == -1 {
		return "", false
	}
	count := 0
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
