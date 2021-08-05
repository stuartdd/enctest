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
	return DetailPage{"id", welcomeTitle, welcomeIntro, welcomeScreen, nil}
}

func GetDetailPage(id string, dataRootMap *map[string]interface{}) DetailPage {
	nodes := strings.Split(id, ".")
	switch len(nodes) {
	case 1:
		return DetailPage{id, id, "User details", welcomeScreen, dataRootMap}
	case 2:
		if nodes[1] == idPwDetails {
			return DetailPage{id, "PW Hints", "Password Hints overview.", welcomeScreen, dataRootMap}
		}
		if nodes[1] == idNotes {
			return DetailPage{id, "Notes", "Notes overview.", notesScreen, dataRootMap}
		}
		return DetailPage{id, "Unknown", "Not PW Hints or Notes page.", welcomeScreen, dataRootMap}
	case 3:
		if nodes[1] == idPwDetails {
			return DetailPage{id, nodes[2], "Hints page.", hintsScreen, dataRootMap}
		}
		if nodes[1] == idNotes {
			return DetailPage{id, nodes[2], "Notes page.", notesScreen, dataRootMap}
		}
		return DetailPage{id, "Unknown", "Not PW Hints or Notes page.", welcomeScreen, dataRootMap}
	}
	return DetailPage{id, id, "Root page.", welcomeScreen, dataRootMap}
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
			e = NewDetailEdit(idd, k, fmt.Sprintf("%s", v), noteChangedFunction, unDoFunction)
			EditEntryList[idd] = e
		}
		cObj = append(cObj, container.NewHBox(widget.NewLabel(" "), e.UnDo, e.Lab))
		cObj = append(cObj, widget.NewSeparator())
		cObj = append(cObj, e.Wid)
	}
	return container.NewVBox(cObj...)
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

func parseURL(urlStr string) *url.URL {
	link, err := url.Parse(urlStr)
	if err != nil {
		fyne.LogError("Could not parse URL", err)
	}

	return link
}
