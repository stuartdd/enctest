// Package main provides various examples of Fyne API capabilities.
package main

import (
	"errors"
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

var (
	salt            = []byte("SQhMXVt8rQED2MxHTHxmuZLMxdJz5DQI")
	splitPrefName   = "split"
	themePrefName   = "theme"
	widthPrefName   = "width"
	heightPrefName  = "height"
	window          fyne.Window
	fileData        *lib.FileData
	dataRoot        *lib.DataRoot
	split           *container.Split
	currentUser     = ""
	fileNameArg     = ""
	shouldCloseLock = false
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println("ERROR: File name not provided")
		os.Exit(1)
	}
	fileNameArg = os.Args[1]

	a := app.NewWithID("stuartdd.enctest")
	setThemeById(a.Preferences().StringWithFallback(themePrefName, "dark"))
	a.SetIcon(theme.FyneLogo())

	fd, err := lib.NewFileData("")
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

	window = a.NewWindow(fmt.Sprintf("Data File: %s not loaded yet", fileNameArg))
	window.SetCloseIntercept(shouldClose)

	go func() {
		time.Sleep(2 * time.Second)
		for _, item := range window.MainMenu().Items[0].Items {
			if item.Label == "Quit" {
				item.Action = shouldClose
			}
		}
	}()

	saveItem := fyne.NewMenuItem("Save", func() { commitAndSaveData(false) })
	saveEncItem := fyne.NewMenuItem("Save Encrypted", func() { commitAndSaveData(true) })

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
		fyne.NewMenu("File", saveItem, saveEncItem, settingsItem, fyne.NewMenuItemSeparator()),
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

	navTree := makeNavTree(setPage)

	split = container.NewHSplit(container.NewBorder(nil, makeThemeButtons(), nil, nil, navTree), rhsPage)
	split.Offset = a.Preferences().FloatWithFallback(splitPrefName, 0.2)

	go func() {
		time.Sleep(1 * time.Second)
		loadData(fileNameArg, func() {
			dr, err := lib.NewDataRoot(fileData.GetContent())
			if err != nil {
				fmt.Printf("ERROR: Cannot process data. %s", err)
				os.Exit(1)
			}
			dataRoot = dr
			navTree = makeNavTree(setPage)
			split = container.NewHSplit(container.NewBorder(nil, makeThemeButtons(), nil, nil, navTree), rhsPage)
			window.SetContent(split)
			navTree.Select(dataRoot.GetRootUid())
		}, func(s string, e error) {
			fmt.Printf("ERROR: %s. %s", s, e.Error())
			os.Exit(1)
		})
	}()

	window.SetContent(split)
	navTree.Select(dataRoot.GetRootUid())
	window.Resize(fyne.NewSize(float32(a.Preferences().FloatWithFallback(widthPrefName, 640)), float32(a.Preferences().FloatWithFallback(heightPrefName, 460))))
	window.ShowAndRun()
}

func loadData(fileName string, success func(), fail func(string, error)) {
	fd, err := lib.NewFileData(fileName)
	if err != nil {
		fail(fmt.Sprintf("Failed to load data file %s", fileName), err)
	}
	fileData = fd
	if !fileData.IsRawJson() {
		pass := widget.NewPasswordEntry()
		dialog.ShowCustomConfirm("Enter the password to DECRYPT the file", "Confirm", "Cancel", widget.NewForm(
			widget.NewFormItem("Password", pass),
		), func(ok bool) {
			if ok && pass.Text != "" {
				fmt.Println("HERE")
				err := fileData.DecryptContents([]byte(pass.Text), salt)
				fmt.Println("THERE")
				if err != nil {
					fail(fmt.Sprintf("Failed to decrypt data file %s", fileName), err)
				} else {
					success()
				}
			} else {
				fail(fmt.Sprintf("Failed to decrypt data file %s", fileName), errors.New("password was not provided"))
			}
		}, window)
	} else {
		success()
	}
}

func makeNavTree(setPage func(detailPage gui.DetailPage)) *widget.Tree {
	return &widget.Tree{
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
}

func makeThemeButtons() fyne.CanvasObject {
	return container.New(layout.NewGridLayout(2),
		widget.NewButton("Dark", func() {
			setThemeById("dark")
		}),
		widget.NewButton("Light", func() {
			setThemeById("light")
		}),
	)
}

func commitAndSaveData(enc bool) {
	count := countChangedItems()
	if count == 0 {
		dialog.NewInformation("File Save", "There were no items to save!\n\nPress OK to continue", window).Show()
	} else {
		if enc {
			pass := widget.NewPasswordEntry()
			dialog.ShowCustomConfirm("Enter a password", "Confirm", "Cancel", widget.NewForm(
				widget.NewFormItem("Password", pass),
			), func(ok bool) {
				if ok && pass.Text != "" {
					_, err := commitChangedItems()
					if err != nil {
						dialog.NewInformation("Convert To Json:", fmt.Sprintf("Error Message:\n-- %s --\nFile was not saved\nPress OK to continue", err.Error()), window).Show()
						return
					}
					fmt.Printf("SAVE ENC %s %s\n", pass.Text, salt)
					err = fileData.StoreContentEncrypted([]byte(pass.Text), salt)
					if err != nil {
						dialog.NewInformation("Save Encrypted File Error:", fmt.Sprintf("Error Message:\n-- %s --\nFile may not be saved!\nPress OK to continue", err.Error()), window).Show()
					}
				} else {
					dialog.NewInformation("Save Encrypted File Error:", "Error Message:\n\n-- Password not provided --\n\nFile was not saved!\nPress OK to continue", window).Show()
				}
			}, window)
		} else {
			_, err := commitChangedItems()
			if err != nil {
				dialog.NewInformation("Convert To Json:", fmt.Sprintf("Error Message:\n-- %s --\nFile was not saved\nPress OK to continue", err.Error()), window).Show()
				return
			}
			err = fileData.StoreContent()
			if err != nil {
				dialog.NewInformation("Save File Error:", fmt.Sprintf("Error Message:\n-- %s --\nFile may not be saved!\nPress OK to continue", err.Error()), window).Show()
			} else {
				dialog.NewInformation("File Saved OK:", fmt.Sprintf("%d item(s) were saved", count), window).Show()
			}
		}
	}
}

func shouldClose() {
	if !shouldCloseLock {
		shouldCloseLock = true
		savePreferences()
		count := countChangedItems()
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

func countChangedItems() int {
	count := 0
	for _, v := range gui.EditEntryList {
		if v.IsChanged() {
			count++
		}
	}
	return count
}

func commitChangedItems() (int, error) {
	count := 0
	for _, v := range gui.EditEntryList {
		if v.IsChanged() {
			if v.CommitEdit(dataRoot.GetDataRootMap()) {
				count++
			}
		}
	}
	c, err := dataRoot.ToJson()
	if err != nil {
		return 0, err
	}
	fileData.SetContentString(c)
	return count, nil
}

func savePreferences() {
	p := fyne.CurrentApp().Preferences()
	p.SetFloat(splitPrefName, split.Offset)
	p.SetFloat(widthPrefName, float64(window.Canvas().Size().Width))
	p.SetFloat(heightPrefName, float64(window.Canvas().Size().Height))
}

func setThemeById(id string) {
	switch id {
	case "dark":
		fyne.CurrentApp().Settings().SetTheme(theme.DarkTheme())
	case "light":
		fyne.CurrentApp().Settings().SetTheme(theme.LightTheme())
	}
	fyne.CurrentApp().Preferences().SetString(themePrefName, id)
}

func saveChangesConfirm(option bool) {
	shouldCloseLock = false
	if !option {
		fmt.Println("Quit without saving changes")
		window.Close()
	}
}
