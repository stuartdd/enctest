package lib

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/container"
	"fyne.io/fyne/widget"
)

var (
	welcomeIntro = "Please select an item from the tree view on the left."
	welcomeTitle = "Welcome"
	appDesc      = "Welcome to Password Manager"
	idNotes      = "notes"
	idPwDetails  = "pwHints"
)

type DetailPage struct {
	Title, Intro string
	View         func(w fyne.Window, dataMap *map[string]interface{}) fyne.CanvasObject
	Data         *map[string]interface{}
}

func NewDetailPage(title, intro string, view func(w fyne.Window, dataMap *map[string]interface{}) fyne.CanvasObject, dataMap *map[string]interface{}) *DetailPage {
	return &DetailPage{Title: title, Intro: intro, View: view, Data: dataMap}
}

func (p *DetailPage) DataAsString() string {
	if p.Data == nil {
		return "nil"
	}
	if reflect.ValueOf(*p.Data).Kind() == reflect.String {
		return fmt.Sprintf("%v", *p.Data)
	}
	var sb strings.Builder
	for k, v := range *p.Data {
		sb.WriteString(fmt.Sprintf("key %s = %s\n", k, v))
	}
	return sb.String()
}

func GetWelcomePage() DetailPage {
	return DetailPage{welcomeTitle, welcomeIntro, welcomeScreen, nil}
}

func GetDetailPage(id string, dataMap *map[string]interface{}) DetailPage {
	nodes := strings.Split(id, ".")
	switch len(nodes) {
	case 1:
		return DetailPage{id, "User details", welcomeScreen, dataMap}
	case 2:
		if nodes[1] == idPwDetails {
			return DetailPage{"PW Hints", "Password Hints overview.", welcomeScreen, dataMap}
		}
		if nodes[1] == idNotes {
			return DetailPage{"Notes", "Notes overview.", notesScreen, dataMap}
		}
		return DetailPage{"Unknown", "Not PW Hints or Notes page.", welcomeScreen, dataMap}
	case 3:
		if nodes[1] == idPwDetails {
			return DetailPage{nodes[2], "Hints page.", hintsScreen, dataMap}
		}
		if nodes[1] == idNotes {
			return DetailPage{nodes[2], "Notes page.", notesScreen, dataMap}
		}
		return DetailPage{"Unknown", "Not PW Hints or Notes page.", welcomeScreen, dataMap}
	}
	return DetailPage{id, "Root page.", welcomeScreen, dataMap}
}

func welcomeScreen(_ fyne.Window, dataMap *map[string]interface{}) fyne.CanvasObject {
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

func notesScreen(_ fyne.Window, dataMap *map[string]interface{}) fyne.CanvasObject {
	return container.NewCenter(container.NewVBox(
		// widget.NewLabelWithStyle("Display NOTES", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		container.NewCenter(
			container.NewHBox(
				widget.NewLabel("NOTES"),
			),
		),
	))
}

func hintsScreen(_ fyne.Window, dataMap *map[string]interface{}) fyne.CanvasObject {
	return container.NewCenter(container.NewVBox(
		// widget.NewLabelWithStyle("Display Password Hints", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		container.NewCenter(
			container.NewHBox(
				widget.NewLabel("HINTS"),
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
