package gui

import (
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/stuartdd/jsonParserGo/parser"
	"stuartdd.com/lib"
	"stuartdd.com/pref"
	"stuartdd.com/theme2"
	"stuartdd.com/types"
)

type EditEntryList struct {
	editEntryList map[string]*EditEntry
}

type EditEntry struct {
	Path           string
	Title          string
	OldTxt         string
	NewTxt         string
	Url            string
	We             *widget.Entry
	Lab            *widget.Label
	UnDo           *widget.Button
	Link           *widget.Button
	Remove         *widget.Button
	Rename         *widget.Button
	NodeAnnotation types.NodeAnnotationEnum
	OnChangeFunc   func(input string, path string)
	UnDoFunc       func(path string)
	ActionFunc     func(action string, path string, extra string)
}

type DetailPage struct {
	Uid, Heading, Title, User string
	ViewFunc                  func(w fyne.Window, details DetailPage, actionFunc func(string, string, string)) fyne.CanvasObject
	CntlFunc                  func(w fyne.Window, details DetailPage, actionFunc func(string, string, string)) fyne.CanvasObject
	DataRootMap               parser.NodeI
	Preferences               pref.PrefData
}

func NewEditEntryList() *EditEntryList {
	return &EditEntryList{editEntryList: make(map[string]*EditEntry)}
}

func (p *EditEntryList) Clear() {
	p.editEntryList = make(map[string]*EditEntry)
}

func (p *EditEntryList) Add(ee *EditEntry) {
	p.editEntryList[ee.Path] = ee
}

func (p *EditEntryList) Get(path string) (*EditEntry, bool) {
	ee := p.editEntryList[path]
	if ee == nil {
		return nil, false
	}
	return p.editEntryList[path], true
}

func (p *EditEntryList) Commit(dataRoot parser.NodeI) int {
	count := 0
	for _, v := range p.editEntryList {
		if v.IsChanged() {
			if v.CommitEdit(dataRoot) {
				count++
			}
		}
	}
	return count
}

func (p *EditEntryList) Count() int {
	count := 0
	for _, v := range p.editEntryList {
		if v.IsChanged() {
			count++
		}
	}
	return count
}

func NewEditEntry(path string, combinedTitle string, currentTxt string, onChangeFunc func(s string, path string), unDoFunc func(path string), actionFunc func(action string, uid string, extra string)) *EditEntry {
	nodeAnnotation, title := types.GetNodeAnnotationTypeAndName(combinedTitle)
	l := widget.NewLabel(fmt.Sprintf(" %s ", title))
	u := widget.NewButtonWithIcon("", theme.ContentUndoIcon(), func() {
		unDoFunc(path)
	})
	i := widget.NewButtonWithIcon("", theme2.LinkToWebIcon(), func() {
		actionFunc(ACTION_LINK, "", "")
	})
	r := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		actionFunc(ACTION_REMOVE, path, "")
	})
	n := widget.NewButtonWithIcon("", theme2.EditIcon(), func() {
		actionFunc(ACTION_RENAME, path, "")
	})
	u.Disable()
	i.Disable()
	return &EditEntry{Path: path, Title: title, NodeAnnotation: nodeAnnotation, We: nil, Lab: l, UnDo: u, Link: i, Remove: r, Rename: n, OldTxt: currentTxt, NewTxt: currentTxt, OnChangeFunc: onChangeFunc, UnDoFunc: unDoFunc, ActionFunc: actionFunc}
}

func (p *EditEntry) SetNew(s string) {
	p.NewTxt = s
	p.RefreshData()
}

func (p *EditEntry) RefreshData() {
	if p.We != nil {
		p.We.SetText(p.NewTxt)
	}
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

func (p *EditEntry) CommitEdit(data parser.NodeI) bool {
	m, _ := lib.GetUserDataForUid(data, p.Path)
	if m != nil {
		switch m.GetNodeType() {
		case parser.NT_STRING:
			m.(*parser.JsonString).SetValue(p.NewTxt)
		case parser.NT_BOOL:
			m.(*parser.JsonBool).SetValue(p.NewTxt == "true")
		case parser.NT_NUMBER:
			f, err := strconv.ParseFloat(p.NewTxt, 64)
			if err != nil {
				return false
			}
			m.(*parser.JsonNumber).SetValue(f)
		}
		p.OldTxt = p.NewTxt
		p.RefreshData()
		return true
	}
	return false
}

func (p *EditEntry) RevertEdit() {
	p.SetNew(p.OldTxt)
}

func (p *EditEntry) String() string {
	if p.IsChanged() {
		return fmt.Sprintf("Item:'%s' Is updated from '%s' to '%s'", p.Path, p.OldTxt, p.NewTxt)
	} else {
		return fmt.Sprintf("Item:'%s' Is unchanged", p.Path)
	}
}

func (p *EditEntry) IsChanged() bool {
	return p.NewTxt != p.OldTxt
}

func (p *EditEntry) HasLink() (string, bool) {
	lnk, ok := parseStringForLink(p.GetCurrentText())
	return lnk, ok
}

func (p *EditEntry) GetCurrentText() string {
	return p.NewTxt
}

func NewDetailPage(
	uid string,
	title string,
	user string,
	viewFunc func(w fyne.Window, details DetailPage, actionFunc func(string, string, string)) fyne.CanvasObject,
	cntlFunc func(w fyne.Window, details DetailPage, actionFunc func(string, string, string)) fyne.CanvasObject,
	dataRootMap parser.NodeI,
	preferences pref.PrefData) *DetailPage {

	heading := fmt.Sprintf("User:  %s", title)
	if user != "" {
		heading = fmt.Sprintf("User:  %s:  %s", user, title)
	}
	return &DetailPage{Uid: uid, Heading: heading, Title: title, User: user, ViewFunc: viewFunc, CntlFunc: cntlFunc, DataRootMap: dataRootMap, Preferences: preferences}
}

func (p *DetailPage) GetObjectsForUid() *parser.JsonObject {
	m, err := lib.GetUserDataForUid(p.DataRootMap, p.Uid)
	if err != nil {
		panic(fmt.Sprintf("DetailPage.GetMapForUid. Uid '%s' not found. %s", p.Uid, err.Error()))
	}
	if m.GetNodeType() == parser.NT_OBJECT {
		return m.(*parser.JsonObject)
	}
	panic("DetailPage.GetMapForUid must only return JsonObject types")
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
