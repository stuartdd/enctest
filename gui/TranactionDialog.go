package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type InpuFieldData struct {
	Key         string
	Label       string
	Validator   func(string) error
	Value       string
	labelWidget *widget.Label
	entryWidget *widget.Entry
}

type InputDataWindow struct {
	inputDataPopUp *widget.PopUp
	title          string
	entries        map[string]*InpuFieldData
	cancelFunction func()
	selectFunction func(map[string]*InpuFieldData)
	longestLabel   int
	okButton       *widget.Button
	info           string
}

func newInpuFieldData(key string, label string, validator func(string) error, value string) *InpuFieldData {
	return &InpuFieldData{Key: key, Label: label, Validator: validator, Value: value}
}

func (ifd *InpuFieldData) String() string {
	return fmt.Sprintf("FieldData: Key: %s, Label: %s, Value: %s", ifd.Key, ifd.Label, ifd.Value)
}

func NewInputDataWindow(title string, cancelFunction func(), selectFunction func(map[string]*InpuFieldData)) *InputDataWindow {
	return &InputDataWindow{
		entries:        make(map[string]*InpuFieldData),
		title:          title,
		cancelFunction: cancelFunction,
		selectFunction: selectFunction,
		inputDataPopUp: nil,
		longestLabel:   10,
		okButton:       nil,
		info:           "Update the fields and press OK",
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
	vc.Add(container.NewCenter(widget.NewLabel(idl.info)))
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
	ifd.entryWidget = widget.NewEntry()
	ifd.entryWidget.SetText(ifd.Value)
	ifd.labelWidget = NewStringFieldRight(ifd.Label+":", idl.longestLabel+2)
	ifd.entryWidget.OnChanged = func(s string) {
		idl.validateAll()
	}
	c := container.NewBorder(nil, nil, ifd.labelWidget, nil, ifd.entryWidget)
	return c
}

func (idl *InputDataWindow) validateAll() {
	hasError := false
	for _, v := range idl.entries {
		err := v.Validator(v.entryWidget.Text)
		if err != nil {
			hasError = true
		}
	}
	if hasError {
		idl.okButton.Disable()
	} else {
		idl.okButton.Enable()
	}
}
