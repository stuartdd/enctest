// Package main provides various examples of Fyne API capabilities.
package main

import (
	"fmt"
	"log"
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

const (
	SAVE_AS_IS          = iota
	SAVE_ENCRYPTED      = iota
	SAVE_UN_ENCRYPTED   = iota
	LOAD_THREAD_LOAD    = iota
	LOAD_THREAD_LOADING = iota
	LOAD_THREAD_INPW    = iota
	LOAD_THREAD_DONE    = iota
	splitPrefName       = "split"
	themePrefName       = "theme"
	widthPrefName       = "width"
	heightPrefName      = "height"
)

var (
	window             fyne.Window
	fileData           *lib.FileData
	dataRoot           *lib.DataRoot
	splitContainer     *container.Split
	currentUser        = ""
	loadThreadFileName = ""
	shouldCloseLock    = false
	loadThreadState    = 0
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println("ERROR: File name not provided")
		os.Exit(1)
	}
	loadThreadFileName = os.Args[1]
	loadThreadState = LOAD_THREAD_LOAD

	a := app.NewWithID("stuartdd.enctest")
	setThemeById(a.Preferences().StringWithFallback(themePrefName, "dark"))
	a.SetIcon(theme.FyneLogo())

	// fd, err := lib.NewFileData("")
	// if err != nil {
	// 	fmt.Printf("ERROR: Cannot load data. %s", err)
	// 	os.Exit(1)
	// }
	// fileData = fd

	// dr, err := lib.NewDataRoot(fileData.GetContent())
	// if err != nil {
	// 	fmt.Printf("ERROR: Cannot process data. %s", err)
	// 	os.Exit(1)
	// }
	// dataRoot = dr

	window = a.NewWindow(fmt.Sprintf("Data File: %s not loaded yet", loadThreadFileName))
	window.SetCloseIntercept(shouldClose)
	go func() {
		time.Sleep(2 * time.Second)
		for _, item := range window.MainMenu().Items[0].Items {
			if item.Label == "Quit" {
				item.Action = shouldClose
			}
		}
	}()

	/*
		Create the menus
	*/
	saveItem := fyne.NewMenuItem("Save", func() { commitAndSaveData(SAVE_AS_IS, true) })
	saveEncItem := fyne.NewMenuItem("Save Encrypted", func() { commitAndSaveData(SAVE_ENCRYPTED, false) })
	saveUnEncItem := fyne.NewMenuItem("Save Un-Encrypted", func() { commitAndSaveData(SAVE_UN_ENCRYPTED, false) })
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
		fyne.NewMenu("File", saveItem, saveEncItem, saveUnEncItem, settingsItem, fyne.NewMenuItemSeparator()),
		newItem,
		helpMenu,
	)
	window.SetMainMenu(mainMenu)

	window.SetMaster()

	defaultPageRHS := gui.GetWelcomePage()
	contentRHS := container.NewMax()
	//	contentRHS.Objects = []fyne.CanvasObject{defaultPageRHS.View(window, defaultPageRHS)}
	title := widget.NewLabel(defaultPageRHS.Title)

	setPageRHS := func(detailPage gui.DetailPage) {
		if detailPage.User == "" {
			currentUser = detailPage.Title
			title.SetText(detailPage.Title)
		} else {
			currentUser = detailPage.User
			title.SetText(fmt.Sprintf("%s: %s", detailPage.User, detailPage.Title))
		}
		window.SetTitle(fmt.Sprintf("Data File: [%s]. Current User: %s", fileData.GetFileName(), currentUser))
		contentRHS.Objects = []fyne.CanvasObject{detailPage.View(window, detailPage)}
		contentRHS.Refresh()
	}

	pageRHS := container.NewBorder(
		container.NewVBox(title, widget.NewSeparator(), widget.NewSeparator()), nil, nil, nil, contentRHS)

	/*
		Load thread keeps running in background
		To Trigger it:
			set loadThreadFileName = filename
			set loadThreadState = LOAD_THREAD_LOAD
	*/
	go func() {
		for {
			fmt.Println("loop")
			time.Sleep(500 * time.Millisecond)
			if loadThreadState == LOAD_THREAD_LOAD {
				loadThreadState = LOAD_THREAD_LOADING
				fd, err := lib.NewFileData(loadThreadFileName)
				if err != nil {
					fmt.Printf("Failed to load data file %s", loadThreadFileName)
					os.Exit(1)
				}
				/*
					While file is ENCRYPTED
						Get PW and decrypt
				*/
				for !fd.IsRawJson() {
					if !(loadThreadState == LOAD_THREAD_INPW) {
						loadThreadState = LOAD_THREAD_INPW
						getPasswordAndDecrypt(fd, func() {
							// SUCCESS
							loadThreadState = LOAD_THREAD_DONE
						}, func() {
							// FAIL
							loadThreadState = LOAD_THREAD_LOADING
						})
					}
					if loadThreadState != LOAD_THREAD_DONE {
						time.Sleep(1000 * time.Millisecond)
					}
				}
				/*
					Data is decrypted so process the JSON so
						update the navigation tree
						select the root element
				*/
				dr, err := lib.NewDataRoot(fd.GetContent())
				if err != nil {
					fmt.Printf("ERROR: Cannot process data. %s", err)
					os.Exit(1)
				}
				fileData = fd
				dataRoot = dr
				navTreeLHS := makeNavTree(setPageRHS)
				navTreeLHS.Select(dataRoot.GetRootUid())
				splitContainer = container.NewHSplit(container.NewBorder(nil, makeThemeButtons(), nil, nil, navTreeLHS), pageRHS)
				window.SetContent(splitContainer)
			}
		}
	}()

	window.SetContent(contentRHS)
	window.Resize(fyne.NewSize(float32(a.Preferences().FloatWithFallback(widthPrefName, 640)), float32(a.Preferences().FloatWithFallback(heightPrefName, 460))))
	window.ShowAndRun()
}

func getPasswordAndDecrypt(fd *lib.FileData, success func(), fail func()) {
	pass := widget.NewPasswordEntry()
	dialog.ShowCustomConfirm("Enter the password to DECRYPT the file", "Confirm", "Exit Application", widget.NewForm(
		widget.NewFormItem("Password", pass),
	), func(ok bool) {
		if ok {
			if pass.Text == "" {
				fail()
			} else {
				err := fd.DecryptContents([]byte(pass.Text))
				if err != nil {
					fail()
				} else {
					success()
				}
			}
		} else {
			fmt.Printf("Failed to decrypt data file %s. password was not provided", fd.GetFileName())
			os.Exit(1)
		}
	}, window)
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

func KeyDown(k *fyne.KeyEvent) {
	switch k.Name {
	case fyne.KeyReturn:
		log.Fatal("keyUP", fmt.Sprint("%h", k.Name))
	}
	// s.Field.KeyUp(k)
	// widget.Refresh(s)
}

func commitAndSaveData(enc int, mustBeChanged bool) {
	count := countChangedItems()
	if count == 0 && mustBeChanged {
		dialog.NewInformation("File Save", "There were no items to save!\n\nPress OK to continue", window).Show()
	} else {
		if enc == SAVE_ENCRYPTED {
			pass := widget.NewPasswordEntry()
			dialog.ShowCustomConfirm("Enter the password to ENCRYPT the file", "Confirm", "Cancel", widget.NewForm(
				widget.NewFormItem("Password", pass),
			), func(ok bool) {
				if ok && pass.Text != "" {
					_, err := commitChangedItems()
					if err != nil {
						dialog.NewInformation("Convert To Json:", fmt.Sprintf("Error Message:\n-- %s --\nFile was not saved\nPress OK to continue", err.Error()), window).Show()
						return
					}
					err = fileData.StoreContentEncrypted([]byte(pass.Text))
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
			if enc == SAVE_AS_IS {
				err = fileData.StoreContentAsIs()
			} else {
				err = fileData.StoreContentUnEncrypted()
			}
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
	p.SetFloat(splitPrefName, splitContainer.Offset)
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
