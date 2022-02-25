package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"stuartdd.com/lib"
)

type inpuFieldData struct {
	Key       string
	Label     string
	Validator func(string, string) string
	Value     string
}

type InputDataWindow struct {
	inputDataWindow fyne.Window
	entries         map[string]*inpuFieldData
	cancelFunction  func()
	selectFunction  func(map[string]*inpuFieldData)
	longestLabel    int
}

func newInpuFieldData(key string, label string, validator func(string, string) string, value string) *inpuFieldData {
	return &inpuFieldData{Key: key, Label: label, Validator: validator, Value: value}
}

func (ifd *inpuFieldData) String() string {
	return fmt.Sprintf("FieldData: Key: %s, Label: %s, Value: %s", ifd.Key, ifd.Label, ifd.Value)
}

func NewInputDataWindow(cancelFunction func(), selectFunction func(map[string]*inpuFieldData)) *InputDataWindow {
	return &InputDataWindow{
		entries:         make(map[string]*inpuFieldData),
		cancelFunction:  cancelFunction,
		selectFunction:  selectFunction,
		inputDataWindow: nil,
		longestLabel:    10,
	}
}

func (idl *InputDataWindow) Add(key string, label string, validator func(string, string) string, value string) {
	if idl.longestLabel < len(label) {
		idl.longestLabel = len(label)
	}
	idl.entries[key] = newInpuFieldData(key, label, validator, value)
}

func (idl *InputDataWindow) Len() int {
	return len(idl.entries)
}

func (idl *InputDataWindow) Width() float32 {
	if idl.inputDataWindow != nil {
		return idl.inputDataWindow.Canvas().Size().Width
	}
	return 500
}

func (idl *InputDataWindow) Height() float32 {
	if idl.inputDataWindow != nil {
		return idl.inputDataWindow.Canvas().Size().Height
	}
	return 500
}
func (idl *InputDataWindow) Show(w, h float32) {
	idl.inputDataWindow = fyne.CurrentApp().NewWindow("Search Results")
	vc := container.NewVBox()
	hb := container.NewHBox()
	hb.Add(widget.NewButtonWithIcon("Close", theme.CancelIcon(), func() {
		idl.cancelFunc()
	}))

	hb.Add(widget.NewButtonWithIcon("OK", theme.ConfirmIcon(), func() {
		idl.confirmFunc()
	}))

	vc.Add(hb)
	for _, v := range idl.entries {
		vc.Add(idl.createRow(v))
	}
	vc.Add(hb)
	c := container.NewScroll(vc)
	idl.inputDataWindow.SetContent(c)
	idl.inputDataWindow.SetCloseIntercept(idl.cancelFunc)
	idl.inputDataWindow.Resize(fyne.NewSize(w, h))
	idl.inputDataWindow.SetFixedSize(true)
	idl.inputDataWindow.Show()
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
	if idl.inputDataWindow != nil {
		idl.inputDataWindow.Close()
		idl.inputDataWindow = nil
	}
}

func (idl *InputDataWindow) createRow(ifd *inpuFieldData) *fyne.Container {
	c := container.NewHBox()
	c.Add(widget.NewLabel(lib.PadLeft(ifd.Label, idl.longestLabel)))
	c.Add(widget.NewEntry())
	return c
}
