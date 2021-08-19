package gui

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var (
	welcomeTitle = "Welcome"
	appDesc      = "Welcome to Password Manager"
	idNotes      = "notes"
	idPwDetails  = "pwHints"
)

func GetWelcomePage() DetailPage {
	return DetailPage{"id", "", welcomeTitle, welcomeScreen, nil}
}

func GetDetailPage(id string, dataRootMap *map[string]interface{}) DetailPage {
	nodes := strings.Split(id, ".")
	switch len(nodes) {
	case 1:
		return DetailPage{id, id, "", welcomeScreen, dataRootMap}
	case 2:
		if nodes[1] == idPwDetails {
			return DetailPage{id, "PW Hints", nodes[0], welcomeScreen, dataRootMap}
		}
		if nodes[1] == idNotes {
			return DetailPage{id, "Notes", nodes[0], notesScreen, dataRootMap}
		}
		return DetailPage{id, "Unknown", nodes[0], welcomeScreen, dataRootMap}
	case 3:
		if nodes[1] == idPwDetails {
			return DetailPage{id, nodes[2], nodes[0], hintsScreen, dataRootMap}
		}
		if nodes[1] == idNotes {
			return DetailPage{id, nodes[2], nodes[0], notesScreen, dataRootMap}
		}
		return DetailPage{id, "Unknown", "", welcomeScreen, dataRootMap}
	}
	return DetailPage{id, id, "", welcomeScreen, dataRootMap}
}

func welcomeScreen(_ fyne.Window, details DetailPage) fyne.CanvasObject {
	logo := canvas.NewImageFromFile("background.png")
	logo.FillMode = canvas.ImageFillContain
	logo.SetMinSize(fyne.NewSize(228, 167))
	return container.NewCenter(container.NewVBox(
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
	))
}

func notesScreen(_ fyne.Window, details DetailPage) fyne.CanvasObject {
	data := *details.GetMapForUid()
	cObj := make([]fyne.CanvasObject, 0)

	keys := make([]string, 0)
	for k, _ := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := data[k]
		idd := details.Uid + "." + k
		e, ok := EditEntryList[idd]
		if !ok {
			e = NewEditEntry(idd, k, fmt.Sprintf("%s", v), textChangedFunction, unDoFunction, goToLinkFunction)
			EditEntryList[idd] = e
		}
		e.RefreshButtons()
		cObj = append(cObj, container.NewHBox(e.Link, e.UnDo, e.Lab))
		cObj = append(cObj, widget.NewSeparator())
		cObj = append(cObj, e.Wid)
	}
	return container.NewVBox(cObj...)
}

func hintsScreen(_ fyne.Window, details DetailPage) fyne.CanvasObject {
	data := *details.GetMapForUid()
	cObj := make([]fyne.CanvasObject, 0)

	keys := make([]string, 0)
	for k, _ := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := data[k]
		idd := details.Uid + "." + k
		e, ok := EditEntryList[idd]
		if !ok {
			e = NewEditEntry(idd, k, fmt.Sprintf("%s", v), textChangedFunction, unDoFunction, goToLinkFunction)
			EditEntryList[idd] = e
		}
		e.RefreshButtons()
		cObj = append(cObj, container.NewHBox(e.Link, e.UnDo, e.Lab))
		cObj = append(cObj, widget.NewSeparator())
		cObj = append(cObj, e.Wid)
	}
	return container.NewVBox(cObj...)
}

func textChangedFunction(s string, path string) {
	ee := EditEntryList[path]
	if ee != nil {
		ee.SetNew(s)
	}
}

func unDoFunction(path string) {
	ee := EditEntryList[path]
	if ee != nil {
		ee.RevertEdit()
	}
}

func goToLinkFunction(path string) {
	ee := EditEntryList[path]
	if ee != nil {
		link, ok := ee.HasLink()
		if ok {
			fmt.Printf("This is the link [%s]", link)
		}
	}
}

func parseURL(urlStr string) *url.URL {
	link, err := url.Parse(urlStr)
	if err != nil {
		fyne.LogError("Could not parse URL", err)
	}

	return link
}
