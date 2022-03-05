package gui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	ERROR_GOOD = "  "
	ERROR_BAD  = "<"
)

type InpuFieldData struct {
	Id         string
	Label      []string
	Value      []string
	Selection  int
	Validator  func(string) error
	errorLabel *widget.Label
}

type InputDataWindow struct {
	inputDataPopUp *widget.PopUp
	title          string
	info           string
	entries        map[string]*InpuFieldData
	cancelFunction func()
	selectFunction func(map[string]*InpuFieldData)
	longestLabel   int
	okButton       *widget.Button
	infoLabel      *widget.Label
}

func newInpuFieldData(id string, label []string, value []string, validator func(string) error) *InpuFieldData {
	return &InpuFieldData{Id: id, Label: label, Value: value, Selection: 0, Validator: validator, errorLabel: widget.NewLabel(ERROR_GOOD)}
}

func (ifd *InpuFieldData) String() string {
	return fmt.Sprintf("FieldData: ID: %s, Label: %s, Value: %s", ifd.Id, ifd.Label, ifd.Value)
}

func NewInputDataWindow(title string, info string, cancelFunction func(), selectFunction func(map[string]*InpuFieldData)) *InputDataWindow {
	return &InputDataWindow{
		entries:        make(map[string]*InpuFieldData),
		title:          title,
		info:           info,
		cancelFunction: cancelFunction,
		selectFunction: selectFunction,
		inputDataPopUp: nil,
		longestLabel:   10,
		okButton:       nil,

		infoLabel: widget.NewLabel(info),
	}
}

func (idl *InputDataWindow) AddOptions(id string, label []string, value []string, validator func(string) error) {
	if idl.longestLabel < len(label) {
		idl.longestLabel = len(label)
	}
	idl.entries[id] = newInpuFieldData(id, label, value, validator)
}

func (idl *InputDataWindow) Add(id string, label string, value string, validator func(string) error) {
	if idl.longestLabel < len(label) {
		idl.longestLabel = len(label)
	}
	idl.entries[id] = newInpuFieldData(id, append(make([]string, 0), label), append(make([]string, 0), value), validator)
}

func (idl *InputDataWindow) Len() int {
	return len(idl.entries)
}

func (idl *InputDataWindow) Show(w fyne.Window) {
	vc := container.NewVBox()
	hb := container.NewHBox()

	idl.okButton = &widget.Button{Text: "OK", Icon: theme.ConfirmIcon(), Importance: widget.HighImportance,
		OnTapped: idl.confirmFunc,
	}
	hb.Add(&widget.Button{Text: "Close", Icon: theme.ConfirmIcon(),
		OnTapped: idl.cancelFunc,
	})
	hb.Add(idl.okButton)
	vc.Add(container.NewCenter(widget.NewLabel(idl.title)))
	vc.Add(widget.NewSeparator())
	for _, v := range idl.entries {
		vc.Add(idl.createRow(v))
	}
	vc.Add(widget.NewSeparator())
	idl.infoLabel.SetText(idl.info)
	vc.Add(container.NewCenter(idl.infoLabel))
	vc.Add(widget.NewSeparator())

	vc.Add(container.NewCenter(hb))
	idl.inputDataPopUp = widget.NewModalPopUp(
		vc,
		w.Canvas(),
	)
	idl.inputDataPopUp.Resize(fyne.NewSize(400, -1))
	idl.inputDataPopUp.Show()
}

func (idl *InputDataWindow) cancelFunc() {
	idl.close()
	idl.cancelFunction()
}

func (idl *InputDataWindow) confirmFunc() {
	idl.close()
	idl.selectFunction(idl.entries)
}

func (idl *InputDataWindow) close() {
	if idl.inputDataPopUp != nil {
		idl.inputDataPopUp.Hide()
		idl.inputDataPopUp = nil
	}
}

func (idl *InputDataWindow) createRow(ifd *InpuFieldData) *fyne.Container {
	e := widget.NewEntry()
	e.SetText(strings.TrimSpace(ifd.Value[ifd.Selection]))
	l := NewStringFieldRight(ifd.Label[ifd.Selection]+":", idl.longestLabel+2)
	e.OnChanged = func(s string) {
		ifd.Value[ifd.Selection] = strings.TrimSpace(s)
		idl.validateAll()
	}
	c := container.NewBorder(nil, nil, l, ifd.errorLabel, e)
	return c
}

func (idl *InputDataWindow) validateAll() {
	errStr := ""
	field := ""
	for _, v := range idl.entries {
		var err error = nil
		if v.Value[v.Selection] == "" {
			err = fmt.Errorf("cannot be empty")
		} else {
			err = v.Validator(v.Value[v.Selection])
		}
		if err != nil {
			v.errorLabel.SetText(ERROR_BAD)
			errStr = err.Error()
			field = v.Label[v.Selection]
		} else {
			v.errorLabel.SetText(ERROR_GOOD)
		}
	}
	if errStr != "" {
		idl.okButton.Disable()
		idl.infoLabel.SetText(fmt.Sprintf("Error in '%s:'. Field %s.", field, errStr))
	} else {
		idl.okButton.Enable()
		idl.infoLabel.SetText(idl.info)
	}
}
