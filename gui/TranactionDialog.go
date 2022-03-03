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
	Key        string
	Label      string
	Validator  func(string) error
	Value      string
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

func newInpuFieldData(key string, label string, validator func(string) error, value string) *InpuFieldData {
	return &InpuFieldData{Key: key, Label: label, Validator: validator, Value: value, errorLabel: widget.NewLabel(ERROR_GOOD)}
}

func (ifd *InpuFieldData) String() string {
	return fmt.Sprintf("FieldData: Key: %s, Label: %s, Value: %s", ifd.Key, ifd.Label, ifd.Value)
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

func (idl *InputDataWindow) Add(key string, label string, validator func(string) error, value string) {
	if idl.longestLabel < len(label) {
		idl.longestLabel = len(label)
	}
	idl.entries[key] = newInpuFieldData(key, label, validator, value)
}

func (idl *InputDataWindow) Len() int {
	return len(idl.entries)
}

func (idl *InputDataWindow) Show(w fyne.Window) {
	vc := container.NewVBox()
	hb := container.NewHBox()
	idl.okButton = widget.NewButtonWithIcon("OK", theme.ConfirmIcon(), func() {
		idl.confirmFunc()
	})

	hb.Add(widget.NewButtonWithIcon("Close", theme.CancelIcon(), func() {
		idl.cancelFunc()
	}))
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
	e.SetText(ifd.Value)
	l := NewStringFieldRight(ifd.Label+":", idl.longestLabel+2)
	e.OnChanged = func(s string) {
		ifd.Value = strings.TrimSpace(s)
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
		if v.Value == "" {
			err = fmt.Errorf("cannot be empty")
		} else {
			err = v.Validator(v.Value)
		}
		if err != nil {
			v.errorLabel.SetText(ERROR_BAD)
			errStr = err.Error()
			field = v.Label
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
