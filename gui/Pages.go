package gui

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"stuartdd.com/theme2"
)

const (
	welcomeTitle = "Welcome"
	appDesc      = "Welcome to Valt"
	idNotes      = "notes"
	idPwDetails  = "pwHints"

	ACTION_REMOVE = "remove"
	ACTION_RENAME = "rename"
	ACTION_LINK   = "link"
)

var (
	preferedOrderReversed = []string{"notes", "positional", "post", "pre", "link", "userId"}
)

func NewModalEntryDialog(w fyne.Window, heading, txt string, accept func(bool, string)) (modal *widget.PopUp) {
	return runModalEntryPopup(w, heading, txt, false, accept)
}

func NewModalPasswordDialog(w fyne.Window, heading, txt string, accept func(bool, string)) (modal *widget.PopUp) {
	return runModalEntryPopup(w, heading, txt, true, accept)
}

func GetWelcomePage(id string) *DetailPage {
	return NewDetailPage(id, "", welcomeTitle, welcomeScreen, welcomeControls, nil)
}

func GetDetailPage(id string, dataRootMap *map[string]interface{}) *DetailPage {
	nodes := strings.Split(id, ".")
	switch len(nodes) {
	case 1:
		return NewDetailPage(id, id, "", welcomeScreen, welcomeControls, dataRootMap)
	case 2:
		if nodes[1] == idPwDetails {
			return NewDetailPage(id, "Hints", nodes[0], welcomeScreen, notesControls, dataRootMap)
		}
		if nodes[1] == idNotes {
			return NewDetailPage(id, "Notes", nodes[0], notesScreen, notesControls, dataRootMap)
		}
		return NewDetailPage(id, "Unknown", nodes[0], welcomeScreen, notesControls, dataRootMap)
	case 3:
		if nodes[1] == idPwDetails {
			return NewDetailPage(id, nodes[2], nodes[0], hintsScreen, hintsControls, dataRootMap)
		}
		if nodes[1] == idNotes {
			return NewDetailPage(id, nodes[2], nodes[0], notesScreen, notesControls, dataRootMap)
		}
		return NewDetailPage(id, "Unknown", "", welcomeScreen, notesControls, dataRootMap)
	}
	return NewDetailPage(id, id, "", welcomeScreen, notesControls, dataRootMap)
}

func entryChangedFunction(s string, path string) {
	ee := EditEntryList[path]
	if ee != nil {
		ee.SetNew(s)
	}
}

func welcomeControls(_ fyne.Window, details DetailPage, actionFunc func(string, string)) fyne.CanvasObject {
	cObj := make([]fyne.CanvasObject, 0)
	cObj = append(cObj, widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		actionFunc(ACTION_REMOVE, details.Uid)
	}))
	cObj = append(cObj, widget.NewLabel(details.Heading))
	return container.NewHBox(cObj...)
}

func welcomeScreen(_ fyne.Window, details DetailPage, actionFunc func(string, string)) fyne.CanvasObject {
	logo := canvas.NewImageFromFile("background.png")
	logo.FillMode = canvas.ImageFillContain
	logo.SetMinSize(fyne.NewSize(228, 167))

	return container.NewVBox(
		widget.NewSeparator(),
		container.NewCenter(container.NewVBox(
			widget.NewLabelWithStyle(appDesc, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			logo,
			container.NewCenter(
				container.NewHBox(
					widget.NewHyperlink("fyne.io", parseURL("https://fyne.io/")),
					widget.NewLabel("-"),
					widget.NewHyperlink("SDD", parseURL("https://github.com/stuartdd")),
					widget.NewLabel("-"),
					widget.NewHyperlink("go", parseURL("https://golang.org/")),
				),
			),
		)))
}

func notesControls(_ fyne.Window, details DetailPage, actionFunc func(string, string)) fyne.CanvasObject {
	cObj := make([]fyne.CanvasObject, 0)
	cObj = append(cObj, widget.NewLabel(details.Heading))
	return container.NewHBox(cObj...)
}

func notesScreen(_ fyne.Window, details DetailPage, actionFunc func(string, string)) fyne.CanvasObject {
	data := *details.GetMapForUid()
	cObj := make([]fyne.CanvasObject, 0)
	keys := listOfNonDupeInOrderKeys(data, preferedOrderReversed)
	for _, k := range keys {
		v := data[k]
		idd := details.Uid + "." + k
		e, ok := EditEntryList[idd]
		if !ok {
			e = NewEditEntry(idd, k, fmt.Sprintf("%s", v), entryChangedFunction, unDoFunction, actionFunc)
			EditEntryList[idd] = e
		}
		e.RefreshButtons()
		fcl := container.New(&FixedLayout{100, 1}, e.Lab)
		fcbl := container.New(&FixedLayout{10, 5}, e.Link)
		fcbr := container.New(&FixedLayout{10, 5}, e.UnDo)
		if len(keys) < 2 {
			e.Remove.Disable()
		} else {
			e.Remove.Enable()
		}
		fcre := container.New(&FixedLayout{10, 5}, e.Remove)
		fcna := container.New(&FixedLayout{10, 5}, e.Rename)
		cObj = append(cObj, widget.NewSeparator())
		cObj = append(cObj, container.NewBorder(nil, nil, container.NewHBox(fcre, fcna, fcbl, fcl), fcbr, e.Wid))
	}
	return container.NewVBox(cObj...)
}

func hintsControls(_ fyne.Window, details DetailPage, actionFunc func(action string, uid string)) fyne.CanvasObject {
	cObj := make([]fyne.CanvasObject, 0)
	cObj = append(cObj, widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		actionFunc(ACTION_REMOVE, details.Uid)
	}))
	cObj = append(cObj, widget.NewButtonWithIcon("", theme2.EditIcon(), func() {
		actionFunc(ACTION_RENAME, details.Uid)
	}))
	cObj = append(cObj, widget.NewLabel(details.Heading))
	return container.NewHBox(cObj...)
}

func hintsScreen(_ fyne.Window, details DetailPage, actionFunc func(action string, uid string)) fyne.CanvasObject {
	data := *details.GetMapForUid()
	cObj := make([]fyne.CanvasObject, 0)
	keys := listOfNonDupeInOrderKeys(data, preferedOrderReversed)
	for _, k := range keys {
		v := data[k]
		idd := details.Uid + "." + k
		e, ok := EditEntryList[idd]
		if !ok {
			e = NewEditEntry(idd, k, fmt.Sprintf("%s", v), entryChangedFunction, unDoFunction, actionFunc)
			EditEntryList[idd] = e
		}
		e.RefreshButtons()
		fcl := container.New(&FixedLayout{100, 1}, e.Lab)
		fcbl := container.New(&FixedLayout{10, 5}, e.Link)
		fcbr := container.New(&FixedLayout{10, 5}, e.UnDo)
		if len(keys) < 2 {
			e.Remove.Disable()
		} else {
			e.Remove.Enable()
		}
		fcre := container.New(&FixedLayout{10, 5}, e.Remove)
		fcna := container.New(&FixedLayout{10, 5}, e.Rename)
		cObj = append(cObj, widget.NewSeparator())
		cObj = append(cObj, container.NewBorder(nil, nil, container.NewHBox(fcre, fcna, fcbl, fcl), fcbr, e.Wid))

	}
	return container.NewVBox(cObj...)
}

func listOfNonDupeInOrderKeys(m map[string]interface{}, ordered []string) []string {
	keys := make([]string, 0)
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, s := range ordered {
		pos, found := contains(keys, s)
		if found && pos > 0 {
			for i := pos; i > 0; i-- {
				keys[i] = keys[i-1]
			}
			keys[0] = s
		}
	}
	return keys
}

func contains(s []string, str string) (int, bool) {
	for i, v := range s {
		if v == str {
			return i, true
		}
	}
	return 0, false
}

func unDoFunction(path string) {
	ee := EditEntryList[path]
	if ee != nil {
		ee.RevertEdit()
	}
}

func parseURL(urlStr string) *url.URL {
	link, err := url.Parse(urlStr)
	if err != nil {
		fyne.LogError("Could not parse URL", err)
	}

	return link
}

func runModalEntryPopup(w fyne.Window, heading, txt string, password bool, accept func(bool, string)) (modal *widget.PopUp) {
	submitInternal := func(s string) {
		modal.Hide()
		accept(true, s)
	}
	entry := &widget.Entry{Text: txt, Password: password, OnChanged: func(s string) {}, OnSubmitted: submitInternal}
	modal = widget.NewModalPopUp(
		container.NewVBox(
			widget.NewLabel("   "+heading+"   "),
			entry,
			container.NewCenter(container.New(layout.NewHBoxLayout(), widget.NewButton("Cancel", func() {
				modal.Hide()
				accept(false, entry.Text)
			}), widget.NewButton("OK", func() {
				modal.Hide()
				accept(true, entry.Text)
			}),
			))),
		w.Canvas(),
	)
	w.Canvas().SetOnTypedKey(func(ke *fyne.KeyEvent) {
		if ke.Name == "Return" {
			modal.Hide()
			accept(true, entry.Text)
		} else {
			if ke.Name == "Escape" {
				modal.Hide()
				accept(false, entry.Text)
			}
		}
	})
	modal.Show()
	return modal
}
