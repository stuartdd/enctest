// Package main provides various examples of Fyne API capabilities.
package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"stuartdd.com/gui"
	"stuartdd.com/lib"
	"stuartdd.com/theme2"
)

const (
	SAVE_AS_IS          = iota
	SAVE_ENCRYPTED      = iota
	SAVE_UN_ENCRYPTED   = iota
	LOAD_THREAD_LOAD    = iota
	LOAD_THREAD_LOADING = iota
	LOAD_THREAD_INPW    = iota
	LOAD_THREAD_DONE    = iota
	LOAD_THREAD_REFRESH = iota
	allowedCharsInName  = "*~@#$%^&*()_+=><?"
	splitPrefName       = "split"
	themeVarName        = "theme"
	widthPrefName       = "width"
	heightPrefName      = "height"
)

var (
	window             fyne.Window
	fileData           *lib.FileData
	dataRoot           *lib.DataRoot
	navTreeLHS         *widget.Tree
	splitContainer     *container.Split // So we can save the divider position to preferences.
	currentSelection   = ""
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
	a.Settings().SetTheme(theme2.NewAppTheme(a.Preferences().StringWithFallback(themeVarName, "dark")))
	a.SetIcon(theme2.AppLogo())

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

	newItem := fyne.NewMenu("New",
		fyne.NewMenuItem("User", getNewUser),
		fyne.NewMenuItem("PW Hint", addNewHint),
		fyne.NewMenuItem("Note", addNewNote),
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
		fyne.NewMenu("File", saveItem, saveEncItem, saveUnEncItem, fyne.NewMenuItemSeparator()),
		newItem,
		helpMenu,
	)
	window.SetMainMenu(mainMenu)

	window.SetMaster()

	title := widget.NewLabel("")
	contentRHS := container.NewMax()
	layoutRHS := container.NewBorder(
		container.NewVBox(title, widget.NewSeparator()), nil, nil, nil, contentRHS)

	/*
		function called when a selection is made in the LHS tree.
		This updates the contentRHS which is the RHS page for editing data
	*/
	setPageRHSFunc := func(detailPage gui.DetailPage) {
		if detailPage.User == "" {
			currentUser = detailPage.Title
			title.SetText(detailPage.Title)
		} else {
			currentUser = detailPage.User
			title.SetText(fmt.Sprintf("%s: %s", detailPage.User, detailPage.Title))
		}
		navTreeLHS.OpenBranch(detailPage.Uid)
		window.SetTitle(fmt.Sprintf("Data File: [%s]. Current User: %s", fileData.GetFileName(), currentSelection))
		contentRHS.Objects = []fyne.CanvasObject{detailPage.ViewFunc(window, detailPage)}
		contentRHS.Refresh()
	}

	/*
		Load thread keeps running in background
		To Trigger it:
			set loadThreadFileName = filename
			set loadThreadState = LOAD_THREAD_LOAD
	*/
	go func() {
		for {
			time.Sleep(500 * time.Millisecond)
			if loadThreadState == LOAD_THREAD_LOAD {
				loadThreadState = LOAD_THREAD_LOADING
				fmt.Println("Load State")
				fd, err := lib.NewFileData(loadThreadFileName)
				if err != nil {
					fmt.Printf("Failed to load data file %s\n", loadThreadFileName)
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
					fmt.Printf("ERROR: Cannot process data. %s\n", err)
					os.Exit(1)
				}
				fileData = fd
				dataRoot = dr
				loadThreadState = LOAD_THREAD_REFRESH
			}
			if loadThreadState == LOAD_THREAD_REFRESH {
				uid := dataRoot.GetRootUidOrCurrent(currentSelection)
				fmt.Println(dataRoot.ToJsonTreeMap())
				fmt.Printf("Refresh State: uid %s\n", uid)

				navTreeLHS = makeNavTree(setPageRHSFunc)
				navTreeLHS.OpenBranch(currentUser)
				navTreeLHS.Select(uid)

				splitContainer = container.NewHSplit(container.NewBorder(nil, makeThemeButtons(setPageRHSFunc), nil, nil, navTreeLHS), layoutRHS)
				splitContainer.SetOffset(fyne.CurrentApp().Preferences().FloatWithFallback(splitPrefName, 0.2))
				window.SetContent(splitContainer)
				loadThreadState = LOAD_THREAD_DONE
			}
		}
	}()

	window.Resize(fyne.NewSize(float32(a.Preferences().FloatWithFallback(widthPrefName, 640)), float32(a.Preferences().FloatWithFallback(heightPrefName, 460))))
	window.ShowAndRun()
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
			currentSelection = uid
			fmt.Printf("Select %s\n", uid)
			t := gui.GetDetailPage(uid, dataRoot.GetDataRootMap())
			setPage(t)
		},
	}
}

func makeThemeButtons(setPage func(detailPage gui.DetailPage)) fyne.CanvasObject {
	return container.New(layout.NewGridLayout(2),
		widget.NewButton("Dark", func() {
			setThemeById("dark")
			t := gui.GetDetailPage(currentSelection, dataRoot.GetDataRootMap())
			setPage(t)
		}),
		widget.NewButton("Light", func() {
			setThemeById("light")
			t := gui.GetDetailPage(currentSelection, dataRoot.GetDataRootMap())
			setPage(t)
		}),
	)
}

func getNewUser() {
	entry := widget.NewEntry()
	dialog.ShowCustomConfirm("Enter the name of the user", "Confirm", "Cancel", widget.NewForm(
		widget.NewFormItem("Password", entry),
	), func(ok bool) {
		if ok {
			problem, ok := validateEntryAndAddKey(entry.Text, "user")
			if !ok {
				dialog.NewInformation("Add New User", "Error: "+problem, window).Show()
			}
		}
	}, window)
}

func addNewNote() {
	entry := widget.NewEntry()
	dialog.ShowCustomConfirm(fmt.Sprintf("Enter the name of the new note for user '%s'", currentUser), "Confirm", "Cancel", widget.NewForm(
		widget.NewFormItem("Note Name", entry),
	), func(ok bool) {
		if ok {
			problem, ok := validateEntryAndAddKey(entry.Text, "note")
			if !ok {
				dialog.NewInformation("Add New Note", "Error: "+problem, window).Show()
			}
		}
	}, window)
}

func addNewHint() {
	entry := widget.NewEntry()
	dialog.ShowCustomConfirm(fmt.Sprintf("Enter the name of the new hint for user '%s'", currentUser), "Confirm", "Cancel", widget.NewForm(
		widget.NewFormItem("Hint Name", entry),
	), func(ok bool) {
		if ok {
			problem, ok := validateEntryAndAddKey(entry.Text, "hint")
			if !ok {
				dialog.NewInformation("Add New Hint", "Error: "+problem, window).Show()
			}
		}
	}, window)
}

func validateEntryAndAddKey(entry, addType string) (string, bool) {
	if len(entry) == 0 {
		return "Input is undefined", false
	}
	if len(entry) < 2 {
		return "Input is too short", false
	}
	lcEntry := strings.ToLower(entry)
	for _, c := range lcEntry {
		if c < ' ' {
			return "Input must not contain control characters", false
		}
		if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (strings.ContainsRune(allowedCharsInName, c)) {
			continue
		}
		return fmt.Sprintf("Input must not contain character '%c'. Only 0..9, a..z, A..Z and %s chars are allowed", c, allowedCharsInName), false

	}
	var err error = nil
	var path string = ""
	var name string = ""
	switch addType {
	case "user":
		name, err = dataRoot.AddUser(entry)
	case "note":
		path, err = dataRoot.AddNote(currentUser, entry, "note")
	case "hint":
		path, err = dataRoot.AddApplication(currentUser, entry)
	}
	loadThreadState = LOAD_THREAD_REFRESH
	if err != nil {
		return err.Error(), false
	}
	if path != "" {
		currentSelection = path
	}
	if name != "" {
		currentUser = name
	}
	return "", true
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
			fmt.Printf("Failed to decrypt data file %s. password was not provided\n", fd.GetFileName())
			os.Exit(1)
		}
	}, window)
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
	if splitContainer != nil {
		p.SetFloat(splitPrefName, splitContainer.Offset)
	}
	p.SetFloat(widthPrefName, float64(window.Canvas().Size().Width))
	p.SetFloat(heightPrefName, float64(window.Canvas().Size().Height))
}

func setThemeById(varient string) {
	t := theme2.NewAppTheme(varient)
	fyne.CurrentApp().Settings().SetTheme(t)
	fyne.CurrentApp().Preferences().SetString(themeVarName, varient)
}

func saveChangesConfirm(option bool) {
	shouldCloseLock = false
	if !option {
		fmt.Println("Quit without saving changes")
		window.Close()
	}
}
