package gui

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/stuartdd2/JsonParser4go/parser"
	"stuartdd.com/lib"
	"stuartdd.com/theme2"
)

type EditEntryList struct {
	editEntryList map[string]*EditEntry
}

type EditEntry struct {
	Path           *parser.Path
	Title          string
	OldTxt         string
	NewTxt         string
	Url            string
	We             *widget.Entry
	Lab            *widget.Label
	UnDo           *MyButton
	Link           *MyButton
	Remove         *MyButton
	Rename         *MyButton
	NodeAnnotation lib.NodeAnnotationEnum
	NodeType       parser.NodeType
	OnChangeFunc   func(input string, path *parser.Path)
	UnDoFunc       func(path *parser.Path)
	ActionFunc     func(string, *parser.Path, string)
	StatusDisplay  *StatusDisplay
}

func NewEditEntryList() *EditEntryList {
	return &EditEntryList{editEntryList: make(map[string]*EditEntry)}
}

func (p *EditEntryList) Clear() {
	p.editEntryList = make(map[string]*EditEntry)
}

func (p *EditEntryList) Add(ee *EditEntry) {
	p.editEntryList[ee.Path.String()] = ee
}

func (p *EditEntryList) Get(path *parser.Path) (*EditEntry, bool) {
	ee := p.editEntryList[path.String()]
	if ee == nil {
		return nil, false
	}
	return ee, true
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

func NewEditEntry(node parser.NodeI, path *parser.Path, titleWithAnnotation string, currentTxt string, onChangeFunc func(s string, path *parser.Path), unDoFunc func(path *parser.Path), actionFunc func(string, *parser.Path, string), statusData *StatusDisplay) *EditEntry {
	nodeAnnotation, title := lib.GetNodeAnnotationTypeAndName(titleWithAnnotation)
	lab := widget.NewLabel(fmt.Sprintf(" %s ", title))
	nType := node.GetNodeType()
	undo := NewMyIconButton("", theme.ContentUndoIcon(), func() {
		unDoFunc(path)
	}, statusData, fmt.Sprintf("Undo changes to '%s'", title))
	remove := NewMyIconButton("", theme.DeleteIcon(), func() {
		actionFunc(ACTION_REMOVE, path, "")
	}, statusData, fmt.Sprintf("Delete '%s'", title))
	rename := NewMyIconButton("", theme2.EditIcon(), func() {
		actionFunc(ACTION_RENAME, path, "")
	}, statusData, fmt.Sprintf("Rename '%s'", title))
	if nodeAnnotation == lib.NODE_REQUIRED {
		remove = NewMyBlankButton(theme.DeleteIcon(), statusData, fmt.Sprintf("Cannot delete a Required field: '%s'", title))
		rename = NewMyBlankButton(theme2.EditIcon(), statusData, fmt.Sprintf("Cannot rename a Required field: '%s'", title))
	}
	undo.Disable()
	ee := &EditEntry{Path: path, Title: title, NodeAnnotation: nodeAnnotation, NodeType: nType, We: nil, Lab: lab, UnDo: undo, Link: nil, Remove: remove, Rename: rename, OldTxt: currentTxt, NewTxt: currentTxt, OnChangeFunc: onChangeFunc, UnDoFunc: unDoFunc, ActionFunc: actionFunc, StatusDisplay: statusData}
	link := NewMyIconButton("", theme2.LinkToWebIcon(), func() {
		actionFunc(ACTION_LINK, path, ee.Url)
	}, statusData, fmt.Sprintf("Follow the link in '%s'. Launches a seperate browser.", title))
	link.Disable()
	ee.Link = link
	return ee
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
		oldV := m.String()
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
		newV := m.String()
		p.ActionFunc(ACTION_LOG, p.Path, fmt.Sprintf("CommitEdit Path:%s Len:%d --> Len:%d --->%s<--+-->%s<---", p.Path, len(oldV), len(newV), oldV, newV))
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
	return lib.ParseStringForLink(p.GetCurrentText())
}

func (p *EditEntry) GetCurrentText() string {
	return p.NewTxt
}
