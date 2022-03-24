package gui

import (
	"fmt"
	"strconv"
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
	Labels     []string
	Options    []string
	Value      string
	Validator  func(string) (string, error)
	errorLabel *widget.Label
}

type InputData struct {
	entries []*InpuFieldData
}

type InputDataWindow struct {
	inputDataPopUp *widget.PopUp
	title          string
	info           string
	entries        *InputData
	cancelFunction func()
	selectFunction func(*InputData)
	longestLabel   int
	okButton       *widget.Button
	infoLabel      *widget.Label
}

func newInpuData() *InputData {
	return &InputData{entries: make([]*InpuFieldData, 0)}
}

func (inD *InputData) add(id string, labels []string, value string, options []string, validator func(string) (string, error)) {
	inD.entries = append(inD.entries, newInpuFieldData(id, labels, value, options, validator))
}

func (inD *InputData) Data() []*InpuFieldData {
	return inD.entries
}

func (inD *InputData) Get(name string) *InpuFieldData {
	for _, idf := range inD.entries {
		if idf.Id == name {
			return idf
		}
	}
	return nil
}

func (inD *InputData) GetString(name string, def string) string {
	d := inD.Get(name)
	if d == nil {
		return def
	}
	return d.Value
}

func (inD *InputData) GetFloat(name string, def float64) float64 {
	d := inD.Get(name)
	if d == nil {
		return def
	}
	v, err := strconv.ParseFloat(d.Value, 64)
	if err != nil {
		return 0.0
	}
	return v
}

func newInpuFieldData(id string, labels []string, value string, options []string, validator func(string) (string, error)) *InpuFieldData {
	return &InpuFieldData{
		Id:         id,
		Labels:     labels,
		Options:    options,
		Value:      strings.TrimSpace(value),
		Validator:  validator,
		errorLabel: widget.NewLabel(ERROR_GOOD)}
}

func (ifd *InpuFieldData) String() string {
	return fmt.Sprintf("FieldData: ID: %s, Label: %s, Value: %s", ifd.Id, ifd.Labels[0], ifd.Value)
}

func NewInputDataWindow(title string, info string, cancelFunction func(), selectFunction func(*InputData)) *InputDataWindow {
	return &InputDataWindow{
		entries:        newInpuData(),
		title:          title,
		info:           info,
		cancelFunction: cancelFunction,
		selectFunction: selectFunction,
		inputDataPopUp: nil,
		longestLabel:   10,
		okButton:       nil,
		infoLabel:      widget.NewLabel(info),
	}
}

func (idl *InputDataWindow) AddOptions(id string, labels []string, value string, options []string, validator func(string) (string, error)) {
	idl.entries.add(id, labels, value, options, validator)
}

func (idl *InputDataWindow) Add(id string, label string, value string, validator func(string) (string, error)) {
	if idl.longestLabel < len(label) {
		idl.longestLabel = len(label)
	}
	idl.entries.add(id, append(make([]string, 0), label), value, make([]string, 0), validator)
}

func (idl *InputDataWindow) Validate() {
	idl.validateAll()
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
	for _, v := range idl.entries.Data() {
		if len(v.Options) == 0 {
			vc.Add(idl.createRow(v))
		} else {
			vc.Add(idl.createOption(v))
		}
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

func (idl *InputDataWindow) createOption(ifd *InpuFieldData) *fyne.Container {
	e := widget.NewRadioGroup(ifd.Labels, func(s string) {
		ifd.Value = ifd.Options[findIndexOf(ifd.Labels, s)]
		idl.validateAll()
	})
	e.SetSelected(ifd.Labels[findIndexOf(ifd.Options, ifd.Value)])
	c := container.NewBorder(nil, nil, NewStringFieldRight("", idl.longestLabel+2), nil, e)
	return c
}

func (idl *InputDataWindow) createRow(ifd *InpuFieldData) *fyne.Container {
	e := widget.NewEntry()
	e.SetText(strings.TrimSpace(ifd.Value))
	l := NewStringFieldRight(ifd.Labels[0]+":", idl.longestLabel+2)
	e.OnChanged = func(s string) {
		ifd.Value = strings.TrimSpace(s)
		idl.validateAll()
	}
	c := container.NewBorder(nil, nil, l, ifd.errorLabel, e)
	return c
}

func (idl *InputDataWindow) validateAll() {
	errStr := ""
	infoStr := ""
	field := ""
	for _, v := range idl.entries.Data() {
		var err error = nil
		s, err := v.Validator(v.Value)
		if err != nil {
			v.errorLabel.SetText(ERROR_BAD)
			errStr = err.Error()
			field = v.Labels[0]
		} else {
			infoStr = s
			v.errorLabel.SetText(ERROR_GOOD)
		}
	}
	if errStr != "" {
		idl.okButton.Disable()
		idl.infoLabel.SetText(fmt.Sprintf("Error in '%s:'. Field %s.", field, errStr))
	} else {
		idl.okButton.Enable()
		if infoStr != "" {
			idl.infoLabel.SetText(idl.info)
		} else {
			idl.infoLabel.SetText(infoStr)
		}
	}
}

func findIndexOf(haystack []string, needle string) int {
	for i, v := range haystack {
		if v == needle {
			return i
		}
	}
	return 0
}
