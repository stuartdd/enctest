package lib

import (
	"net/url"

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

/*
map[:[UserA UserB]
UserA:[UserApwHints UserAnotes]
UserAnotes:[]
UserApwHints:[UserApwHintsGMailA UserApwHintsPrincipalityA]
UserB:[UserBpwHints UserBnotes]
UserBnotes:[]
UserBpwHints:[UserBpwHintsGMailB UserBpwHintsPrincipalityB]]
*/
// TutorialIndex  defines how our tutorials should be laid out in the index tree
var NavIndex = map[string][]string{
	"":      {"userA", "userB"},
	"userA": {"UserApwHints", "UserAnotes"},
	"userB": {"UserBpwHints", "UserBnotes"},
}

var DetailPages = map[string]DetailPage{
	"userA":        {"User A", "View User A Details.", welcomeScreen},
	"userB":        {"User B", "View User B Details.", welcomeScreen},
	"UserApwHints": {"PW Hints", "Hints for user", welcomeScreen},
	"UserAnotes":   {"Notes", "Notes for user.", welcomeScreen},
	"UserBpwHints": {"PW Hints", "Hints for user", welcomeScreen},
	"UserBnotes":   {"Notes", "Notes for user.", welcomeScreen},
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
