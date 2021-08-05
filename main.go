// Package main provides various examples of Fyne API capabilities.
package main

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"stuartdd.com/lib"

	"fyne.io/fyne/container"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

var fileData *lib.FileData
var dataRoot *lib.DataRoot
var window fyne.Window

func shouldClose() {
	count := 0
	for _, v := range lib.EditEntryList {
		if v.IsChanged() {
			fmt.Println(v)
			count++
		}
	}
	if count > 0 {
		fmt.Println("Nope!")
	} else {
		fmt.Println("That's all folks")
		window.Close()
		os.Exit(0)
	}
}

func main() {
	if len(os.Args) == 1 {
		fmt.Println("ERROR: File name not provided")
		os.Exit(1)
	}

	a := app.NewWithID("io.fyne.demo")
	a.SetIcon(theme.FyneLogo())

	fd, err := lib.NewFileData(os.Args[1])
	if err != nil {
		fmt.Printf("ERROR: Cannot load data. %s", err)
		os.Exit(1)
	}
	fileData = fd

	dr, err := lib.NewDataRoot(fileData.GetContent())
	if err != nil {
		fmt.Printf("ERROR: Cannot process data. %s", err)
		os.Exit(1)
	}
	dataRoot = dr

	window = a.NewWindow(fmt.Sprintf("Data File: %s", fileData.GetFileName()))
	window.SetCloseIntercept(shouldClose)

	go func() {
		time.Sleep(2 * time.Second)
		for _, item := range window.MainMenu().Items[0].Items {
			if item.Label == "Quit" {
				item.Action = shouldClose
			}
		}
	}()

	newItem := fyne.NewMenuItem("New", nil)
	otherItem := fyne.NewMenuItem("Other", nil)
	otherItem.ChildMenu = fyne.NewMenu("",
		fyne.NewMenuItem("Project", func() { fmt.Println("Menu New->Other->Project") }),
		fyne.NewMenuItem("Mail", func() { fmt.Println("Menu New->Other->Mail") }),
	)

	newItem.ChildMenu = fyne.NewMenu("",
		fyne.NewMenuItem("File", func() { fmt.Println("Menu New->File") }),
		fyne.NewMenuItem("Directory", func() { fmt.Println("Menu New->Directory") }),
		otherItem,
	)

	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("Documentation", func() {
			u, _ := url.Parse("https://developer.fyne.io")
			_ = a.OpenURL(u)
		}),
		fyne.NewMenuItem("Support", func() {
			u, _ := url.Parse("https://fyne.io/support/")
			_ = a.OpenURL(u)
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Sponsor", func() {
			u, _ := url.Parse("https://github.com/sponsors/fyne-io")
			_ = a.OpenURL(u)
		}))

	mainMenu := fyne.NewMainMenu(
		// a quit item will be appended to our first menu
		fyne.NewMenu("File", newItem, fyne.NewMenuItemSeparator()),
		helpMenu,
	)

	window.SetMainMenu(mainMenu)
	window.SetMaster()

	welcomePage := lib.GetWelcomePage()
	content := container.NewMax()
	content.Objects = []fyne.CanvasObject{welcomePage.View(window, welcomePage)}
	title := widget.NewLabel(welcomePage.Title)
	setPage := func(t lib.DetailPage) {
		title.SetText(t.Title)
		content.Objects = []fyne.CanvasObject{t.View(window, t)}
		content.Refresh()
	}
	rhsPage := container.NewBorder(
		container.NewVBox(title, widget.NewSeparator(), widget.NewSeparator()), nil, nil, nil, content)

	split := container.NewHSplit(makeNav(setPage), rhsPage)
	split.Offset = 0.2
	window.SetContent(split)

	window.Resize(fyne.NewSize(640, 460))

	window.ShowAndRun()
}

func makeNav(setPage func(detailPage lib.DetailPage)) fyne.CanvasObject {
	a := fyne.CurrentApp()

	tree := &widget.Tree{
		ChildUIDs: func(uid string) []string {
			id := dataRoot.GetNavIndex(uid)
			return id
		},
		IsBranch: func(uid string) bool {
			children := dataRoot.GetNavIndex(uid)
			return len(children) > 0
		},
		CreateNode: func(branch bool) fyne.CanvasObject {
			return widget.NewLabel("?")
		},
		UpdateNode: func(uid string, branch bool, obj fyne.CanvasObject) {
			t := lib.GetDetailPage(uid, dataRoot.GetDataRootMap())
			obj.(*widget.Label).SetText(t.Title)
		},
		OnSelected: func(uid string) {
			t := lib.GetDetailPage(uid, dataRoot.GetDataRootMap())
			setPage(t)
		},
	}

	themes := fyne.NewContainerWithLayout(layout.NewGridLayout(2),
		widget.NewButton("Dark", func() {
			a.Settings().SetTheme(theme.DarkTheme())
		}),
		widget.NewButton("Light", func() {
			a.Settings().SetTheme(theme.LightTheme())
		}),
	)

	return container.NewBorder(nil, themes, nil, nil, tree)
}
