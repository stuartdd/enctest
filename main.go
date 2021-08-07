// Package main provides various examples of Fyne API capabilities.
package main

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/cmd/fyne_settings/settings"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"stuartdd.com/gui"
	"stuartdd.com/lib"
)

var fileData *lib.FileData
var dataRoot *lib.DataRoot
var window fyne.Window
var currentUser string = ""
var shouldCloseLock bool = false

func main() {
	if len(os.Args) == 1 {
		fmt.Println("ERROR: File name not provided")
		os.Exit(1)
	}

	a := app.NewWithID("io.fyne.demo")
	a.Storage().RootURI()
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

	saveItem := fyne.NewMenuItem("Save", func() { commitAndSaveData() })

	settingsItem := fyne.NewMenuItem("Settings", func() {
		w := a.NewWindow("Fyne Settings")
		w.SetContent(settings.NewSettings().LoadAppearanceScreen(w))
		w.Resize(fyne.NewSize(480, 480))
		w.Show()
	})

	newItem := fyne.NewMenu("New",
		fyne.NewMenuItem("User", func() { fmt.Println("Menu New->User") }),
		fyne.NewMenuItem("PW Hint", func() { fmt.Println("Menu New->PW Hint") }),
		fyne.NewMenuItem("Note", func() { fmt.Println("Menu New->Note") }),
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
		fyne.NewMenu("File", saveItem, settingsItem, fyne.NewMenuItemSeparator()),
		newItem,
		helpMenu,
	)
	window.SetMainMenu(mainMenu)
	window.SetMaster()

	welcomePage := gui.GetWelcomePage()
	content := container.NewMax()
	content.Objects = []fyne.CanvasObject{welcomePage.View(window, welcomePage)}
	title := widget.NewLabel(welcomePage.Title)
	setPage := func(t gui.DetailPage) {
		if t.User == "" {
			currentUser = t.Title
			title.SetText(t.Title)
		} else {
			currentUser = t.User
			title.SetText(fmt.Sprintf("%s: %s", t.User, t.Title))
		}
		window.SetTitle(fmt.Sprintf("Data File: [%s]. Current User: %s", fileData.GetFileName(), currentUser))
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

func makeNav(setPage func(detailPage gui.DetailPage)) fyne.CanvasObject {
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
			t := gui.GetDetailPage(uid, dataRoot.GetDataRootMap())
			obj.(*widget.Label).SetText(t.Title)
		},
		OnSelected: func(uid string) {
			t := gui.GetDetailPage(uid, dataRoot.GetDataRootMap())
			setPage(t)
		},
	}

	themes := container.New(layout.NewGridLayout(2),
		widget.NewButton("Dark", func() {
			a.Settings().SetTheme(theme.DarkTheme())
		}),
		widget.NewButton("Light", func() {
			a.Settings().SetTheme(theme.LightTheme())
		}),
	)

	return container.NewBorder(nil, themes, nil, nil, tree)
}

func commitAndSaveData() {
	count := 0
	for _, v := range gui.EditEntryList {
		if v.IsChanged() {
			if v.CommitEdit(dataRoot.GetDataRootMap()) {
				count++
			}
		}
	}
	if count == 0 {
		dialog.NewInformation("File Save", "There were no items to save!\n\nPress OK to continue", window).Show()
	} else {
		c, err := dataRoot.ToJson()
		if err != nil {
			dialog.NewInformation("Convert To Json:", fmt.Sprintf("Error Message:\n-- %s --\nFile was not saved\nPress OK to continue", err.Error()), window).Show()
			return
		}
		fileData.SetContentString(c)
		err = fileData.StoreContent()
		if err != nil {
			dialog.NewInformation("Save File Error:", fmt.Sprintf("Error Message:\n-- %s --\nFile may not be saved!\nPress OK to continue", err.Error()), window).Show()
		} else {
			dialog.NewInformation("File Saved OK:", fmt.Sprintf("%d item(s) were saved", count), window).Show()
		}
	}
}

func shouldClose() {
	if !shouldCloseLock {
		shouldCloseLock = true
		count := 0
		for _, v := range gui.EditEntryList {
			if v.IsChanged() {
				count++
			}
		}
		if count > 0 {
			fmt.Println("Nope!")
			d := dialog.NewConfirm("Save Changes", "There are unsaved changes\nDo you want to save them before closing?", saveChangesConfirm, window)
			d.Show()
		} else {
			fmt.Println("That's all folks")
			window.Close()
		}
	}
}

func saveChangesConfirm(option bool) {
	shouldCloseLock = false
	if !option {
		fmt.Println("Quit without saving changes")
		window.Close()
	}
}
