package lib

import (
	"net/url"
	"strings"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/cmd/fyne_demo/data"
	"fyne.io/fyne/container"
	"fyne.io/fyne/widget"
)

type DetailPage struct {
	Title, Intro string
	View         func(w fyne.Window) fyne.CanvasObject
}

func GetDetailPages(id string) DetailPage {
	nodes := strings.Split(id, ".")
	switch len(nodes) {
	case 1:
		return DetailPage{id, "User page.", welcomeScreen}
	case 2:
		if nodes[1] == "pwHints" {
			return DetailPage{"PW Hints", "Hints overview.", welcomeScreen}
		}
		if nodes[1] == "notes" {
			return DetailPage{"Notes", "Notes overview.", welcomeScreen}
		}
		return DetailPage{"Unknown", "Not PW Hints Not Notes page.", welcomeScreen}
	case 3:
		if nodes[1] == "pwHints" {
			return DetailPage{nodes[2], "Hints page.", welcomeScreen}
		}
		if nodes[1] == "notes" {
			return DetailPage{nodes[2], "Notes page.", welcomeScreen}
		}
		return DetailPage{"Unknown", "Not PW Hints Not Notes page.", welcomeScreen}
	}
	return DetailPage{id, "Root page.", welcomeScreen}
}

func welcomeScreen(_ fyne.Window) fyne.CanvasObject {
	logo := canvas.NewImageFromResource(data.FyneScene)
	logo.FillMode = canvas.ImageFillContain
	logo.SetMinSize(fyne.NewSize(228, 167))

	return container.NewCenter(container.NewVBox(
		widget.NewLabelWithStyle("Welcome to the Fyne toolkit demo app", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		logo,
		container.NewHBox(
			widget.NewHyperlink("fyne.io", parseURL("https://fyne.io/")),
			widget.NewLabel("-"),
			widget.NewHyperlink("documentation", parseURL("https://fyne.io/develop/")),
			widget.NewLabel("-"),
			widget.NewHyperlink("sponsor", parseURL("https://github.com/sponsors/fyne-io")),
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
