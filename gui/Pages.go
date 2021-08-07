package gui

import (
	"fmt"
	"net/url"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var (
	welcomeIntro = "Please select an item from the tree view on the left."
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
	data := details.GetMapForUid()
	cObj := make([]fyne.CanvasObject, 0)
	for k, v := range *data {
		idd := details.Uid + "." + k
		e, ok := EditEntryList[idd]
		if !ok {
			e = NewDetailEdit(idd, k, fmt.Sprintf("%s", v), noteChangedFunction, unDoFunction, goToLinkFunction)
			EditEntryList[idd] = e
		}
		e.ParseForLink()
		cObj = append(cObj, container.NewHBox(e.Link, e.UnDo, e.Lab))
		cObj = append(cObj, widget.NewSeparator())
		cObj = append(cObj, e.Wid)
	}
	return container.NewVBox(cObj...)
}

func hintsScreen(_ fyne.Window, details DetailPage) fyne.CanvasObject {
	return container.NewCenter(container.NewVBox(
		// widget.NewLabelWithStyle("Display Password Hints", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		container.NewCenter(
			container.NewHBox(
				widget.NewLabel(fmt.Sprintf("%s", details.GetMapForUid())),
			),
		),
	))
}

func noteChangedFunction(s string, path string) {
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
		link, ok := ee.ParseForLink()
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
