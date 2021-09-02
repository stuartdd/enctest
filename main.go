// Package main provides various examples of Fyne API capabilities.
package main

import (
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"stuartdd.com/gui"
	"stuartdd.com/lib"
	"stuartdd.com/pref"
	"stuartdd.com/theme2"
)

const (
	SAVE_AS_IS              = iota
	SAVE_ENCRYPTED          = iota
	SAVE_UN_ENCRYPTED       = iota
	MAIN_THREAD_LOAD        = iota
	MAIN_THREAD_IDLE        = iota
	MAIN_THREAD_RELOAD_TREE = iota
	MAIN_THREAD_SELECT      = iota

	ADD_TYPE_USER = iota
	ADD_TYPE_HINT = iota
	ADD_TYPE_NOTE = iota

	allowedCharsInName      = " *@#$%^&*()_+=?"
	splitPrefName           = "split"
	themeVarPrefName        = "theme"
	widthPrefName           = "width"
	heightPrefName          = "height"
	lastGoodSearchPrefName  = "lastSearch"
	searchCasePrefName      = "searchCase"
	viewFuffScreenPrefName  = "fullScreen"
	fallbackPreferencesFile = "preferences.json"
	dataFilePrefName        = "datafile"
)

var (
	window                   fyne.Window
	searchWindow             fyne.Window
	fileData                 *lib.FileData
	dataRoot                 *lib.DataRoot
	preferences              *pref.PrefData
	navTreeLHS               *widget.Tree
	splitContainer           *container.Split // So we can save the divider position to preferences.
	splitContainerOffset     float64          = -1
	splitContainerOffsetPref float64          = -1

	findCaseSensitive           = binding.NewBool()
	pendingSelection            = ""
	currentSelection            = ""
	currentUser                 = ""
	shouldCloseLock             = false
	mainThreadNextState         = 0
	mainThreadNextRunMs   int64 = 0
	loadThreadFileName          = ""
	countStructureChanges       = 0
	appIsFullScreenPref         = false
	appScreenSize               = fyne.Size{Width: 640, Height: 480}
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println("ERROR: Preferences File name not provided")
		os.Exit(1)
	}
	p, err := pref.NewPrefData(os.Args[1])
	if err != nil {
		fmt.Printf("Failed to load preferences using '%s'", os.Args[1])
		os.Exit(1)
	}
	preferences = p

	loadThreadFileName = p.GetValueForPathWithFallback(dataFilePrefName, fallbackPreferencesFile)
	mainThreadNextState = MAIN_THREAD_IDLE
	mainThreadNextRunMs = 0

	a := app.NewWithID("stuartdd.enctest")
	a.Settings().SetTheme(theme2.NewAppTheme(preferences.GetValueForPathWithFallback(themeVarPrefName, "dark")))
	a.SetIcon(theme2.AppLogo())

	window = a.NewWindow(fmt.Sprintf("Data File: %s not loaded yet", loadThreadFileName))

	appIsFullScreenPref = preferences.GetBoolWithFallback(viewFuffScreenPrefName, false)
	sw := a.Preferences().FloatWithFallback(widthPrefName, float64(appScreenSize.Width))
	sh := a.Preferences().FloatWithFallback(heightPrefName, float64(appScreenSize.Height))
	appScreenSize = fyne.NewSize(float32(sw), float32(sh))

	splitContainerOffsetPref = fyne.CurrentApp().Preferences().FloatWithFallback(splitPrefName, 0.2)
	splitContainerOffset = -1

	findCaseSensitive.Set(a.Preferences().BoolWithFallback(searchCasePrefName, true))
	findCaseSensitive.AddListener(binding.NewDataListener(func() {
		b, err := findCaseSensitive.Get()
		if err == nil {
			a.Preferences().SetBool(searchCasePrefName, b)
		}
	}))

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
		fyne.NewMenuItem("User", addNewUser),
		fyne.NewMenuItem("Hint", addNewHint),
		fyne.NewMenuItem("Note", addNewNote),
	)
	viewItem := fyne.NewMenu("View",
		fyne.NewMenuItem("Full Screen", viewFullScreen),
		fyne.NewMenuItem("Positional Data", addNewHint),
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
		viewItem,
		helpMenu,
	)
	window.SetMainMenu(mainMenu)
	window.SetMaster()

	wp := gui.GetWelcomePage("")
	title := container.NewHBox()
	title.Objects = []fyne.CanvasObject{wp.CntlFunc(window, *wp, nil)}
	contentRHS := container.NewMax()
	layoutRHS := container.NewBorder(title, nil, nil, nil, contentRHS)

	/*
		function called when a selection is made in the LHS tree.
		This updates the contentRHS which is the RHS page for editing data
	*/
	setPageRHSFunc := func(detailPage gui.DetailPage) {
		currentSelection = detailPage.Uid
		currentUser = lib.GetUserFromPath(currentSelection)
		window.SetTitle(fmt.Sprintf("Data File: [%s]. Current User: %s", fileData.GetFileName(), currentUser))
		navTreeLHS.OpenBranch(currentSelection)
		title.Objects = []fyne.CanvasObject{detailPage.CntlFunc(window, detailPage, controlActionFunction)}
		title.Refresh()
		contentRHS.Objects = []fyne.CanvasObject{detailPage.ViewFunc(window, detailPage, viewActionFunction)}
		contentRHS.Refresh()
	}

	/*
		Thread keeps running in background
		To Trigger it:
			set loadThreadFileName = filename
			requestMainThreadIn(1000, MAIN_THREAD_LOAD)
	*/
	go func() {
		/*
			--- MAIN LOOP ---
			mainThreadNextRunMs is 0 for never run the main loop
			mainThreadNextRunMs is current milliseconds + an offset.
				This will then run the Main loop at that time
			mainThreadNextRunMs is always reset to 0 when the main loop starts.

			mainThreadNextState is the action performed by the main loop
				Each action should leave the state as MAIN_THREAD_IDLE unless a follow on action is required
		*/
		for {
			if mainThreadNextRunMs > 0 && time.Now().UnixMilli() > mainThreadNextRunMs {
				mainThreadNextRunMs = 0
				if mainThreadNextState == MAIN_THREAD_LOAD {

					// Load the file and decrypt it if required
					fmt.Println("Load State")
					fd, err := lib.NewFileData(loadThreadFileName)
					if err != nil {
						fmt.Printf("Failed to load data file %s\n", loadThreadFileName)
						os.Exit(1)
					}
					/*
						While file is ENCRYPTED
							Get PW and decrypt
							if Decryption is cancelled the application exits
					*/
					message := ""
					for fd.RequiresDecryption() {
						getPasswordAndDecrypt(fd, message, func(s string) {
							// FAIL
							message = "Error: " + strings.TrimSpace(s) + ". Please try again"
							time.Sleep(1000 * time.Millisecond)
						})
					}
					/*
						Data is decrypted so process the JSON so
							update the navigation tree
							select the root element
					*/
					dr, err := lib.NewDataRoot(fd.GetContent(), dataMapUpdated)
					if err != nil {
						fmt.Printf("ERROR: Cannot process data. %s\n", err)
						os.Exit(1)
					}
					fileData = fd
					dataRoot = dr

					// Follow on action to rebuild the Tree and re-display it
					requestMainThreadIn(500, MAIN_THREAD_RELOAD_TREE)
				}

				if mainThreadNextState == MAIN_THREAD_RELOAD_TREE {
					// Re-build the main tree view.
					// Select the root of currentUser if defined.
					// Init the devider (split)
					// Populate the window and we are done!
					navTreeLHS = makeNavTree(setPageRHSFunc)
					uid := dataRoot.GetRootUidOrCurrentUid(currentSelection)
					fmt.Printf("Refresh current:%s ", uid)
					selectTreeElement(uid)

					if splitContainerOffset < 0 {
						splitContainerOffset = splitContainerOffsetPref
					} else {
						splitContainerOffset = splitContainer.Offset
					}
					splitContainer = container.NewHSplit(container.NewBorder(nil, makeLHSButtonsAndSearch(setPageRHSFunc), nil, nil, navTreeLHS), layoutRHS)
					splitContainer.SetOffset(splitContainerOffset)

					window.SetContent(splitContainer)
					requestMainThreadIn(0, MAIN_THREAD_IDLE)
				}

				// Select the tree element defined in pendingSelection
				if mainThreadNextState == MAIN_THREAD_SELECT {
					fmt.Printf("Select pending:%s ", pendingSelection)
					selectTreeElement(pendingSelection)
					requestMainThreadIn(0, MAIN_THREAD_IDLE)
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	requestMainThreadIn(1000, MAIN_THREAD_LOAD)

	if appIsFullScreenPref {
		window.SetFullScreen(true)
	} else {
		window.SetFullScreen(false)
		window.Resize(appScreenSize)
	}
	window.ShowAndRun()
}

/**
Select a tree element.
We need to open the parent branches or we will never see the selected element
*/
func selectTreeElement(uid string) {
	user := lib.GetUserFromPath(uid)
	parent := lib.GetParentId(uid)
	navTreeLHS.OpenBranch(user)
	navTreeLHS.OpenBranch(parent)
	navTreeLHS.Select(uid)
}

/**
Use the navigation index (map in DataRoot) to construct the Tree.
Calls setPage with the selected page, defined by the uid (path) and Pages (gui) api
*/
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
			setPage(*t)
		},
	}
}

/**
Section below the tree with Search details and Light and Dark theme buttons
*/
func makeLHSButtonsAndSearch(setPage func(detailPage gui.DetailPage)) fyne.CanvasObject {
	searchEntry := widget.NewEntry()
	searchEntry.SetText(fyne.CurrentApp().Preferences().StringWithFallback(lastGoodSearchPrefName, "?"))
	c2 := container.New(
		layout.NewHBoxLayout(),
		widget.NewCheckWithData("Match Case", findCaseSensitive),
		widget.NewButtonWithIcon("", theme.SearchIcon(), func() { search(searchEntry.Text) }))

	c := container.New(
		layout.NewVBoxLayout(),
		c2,
		searchEntry)
	b := container.New(layout.NewGridLayout(2),
		widget.NewButton("Dark", func() {
			setThemeById("dark")
			t := gui.GetDetailPage(currentSelection, dataRoot.GetDataRootMap())
			setPage(*t)
		}),
		widget.NewButton("Light", func() {
			setThemeById("light")
			t := gui.GetDetailPage(currentSelection, dataRoot.GetDataRootMap())
			setPage(*t)
		}),
	)
	return container.NewVBox(widget.NewSeparator(), c, b)
}

/**
Request that the main loop is started in n milliseconds with a specific state.
*/
func requestMainThreadIn(ms int64, reqState int) {
	if reqState == MAIN_THREAD_IDLE {
		mainThreadNextRunMs = 0
		mainThreadNextState = MAIN_THREAD_IDLE
	} else {
		if ms == 0 {
			mainThreadNextRunMs = 0
		} else {
			mainThreadNextRunMs = time.Now().UnixMilli() + ms
		}
		mainThreadNextState = reqState
	}

}

/**
This is called when a heading button is pressed of the RH page
*/
func controlActionFunction(action string, uid string) {
	fmt.Printf("Control %s %s\n", action, uid)
	viewActionFunction(action, uid)
}

/**
This is called when a detail button is pressed of the RH page
*/
func viewActionFunction(action string, uid string) {
	switch action {
	case gui.ACTION_REMOVE:
		removeAction(uid)
	case gui.ACTION_RENAME:
		renameAction(uid)
	case gui.ACTION_LINK:
		linkAction(uid)
	}
}

/**
Called if there is a structural change in the model
*/
func dataMapUpdated(desc, user, path string, err error) {
	if err == nil {
		fmt.Printf("Updated: %s User: %s Path:%s\n", desc, user, path)
		currentSelection = path
		countStructureChanges++
	}
	requestMainThreadIn(100, MAIN_THREAD_RELOAD_TREE)
}

/**
flip the full screen flag and set the screen to full screen if required
*/
func viewFullScreen() {
	appIsFullScreenPref = !appIsFullScreenPref
	if appIsFullScreenPref {
		appScreenSize = window.Canvas().Size()
		window.SetFullScreen(true)
	} else {
		window.SetFullScreen(false)
		window.Resize(appScreenSize)
	}
}

/**
Remove a node from the main data (model) and update the tree view
dataMapUpdated id called if a change is made to the model
*/
func removeAction(uid string) {
	dialog.NewConfirm("Remove entry", fmt.Sprintf("%s\nAre you sure?", uid), func(b bool) {
		err := dataRoot.Remove(uid, 1)
		if err != nil {
			dialog.NewInformation("Remove item error", err.Error(), window).Show()
		}
	}, window).Show()
}

/**
Rename a node from the main data (model) and update the tree view
dataMapUpdated id called if a change is made to the model
*/
func renameAction(uid string) {
	m, _ := dataRoot.GetDataForUid(uid)
	if m != nil {
		fromName := lib.GetLastId(uid)
		gui.NewModalEntryDialog(window, fmt.Sprintf("Rename entry '%s' ", fromName), "", func(accept bool, s string) {
			if accept {
				err := validateEntityName(s)
				if err != nil {
					dialog.NewInformation("Name validation error", err.Error(), window).Show()
				} else {
					if fromName == s {
						dialog.NewInformation("Rename item error", "Rename to the same name", window).Show()
					} else {
						err := dataRoot.Rename(uid, s)
						if err != nil {
							dialog.NewInformation("Rename item error", err.Error(), window).Show()
						}
					}
				}
			}
		})
	}
}

/**
Activate a link in a browser if it is contained in a note of hint
*/
func linkAction(s string) {
	if s != "" {
		s, err := url.Parse(s)
		if err != nil {
			dialog.NewInformation("Link is invalid", err.Error(), window).Show()
		} else {
			err = fyne.CurrentApp().OpenURL(s)
			if err != nil {
				dialog.NewInformation("ink could not be opened", err.Error(), window).Show()
			}
		}
	}
}

/*
The search button has been pressed
*/
func search(s string) {
	matchCase, _ := findCaseSensitive.Get()

	// Do the search. The map contains the returned search entries
	// This ensures no duplicates are displayed
	// The map key is the human readable results e.g. 'User [Hint] app: noteName'
	// The values are paths within the model! user.pwHints.app.noteName
	mapPaths := make(map[string]string)
	dataRoot.Search(func(path, desc string) {
		mapPaths[desc] = path
	}, s, matchCase)

	// Fine all the keys and sort them
	paths := make([]string, 0)
	for k := range mapPaths {
		paths = append(paths, k)
	}
	sort.Strings(paths)

	// Use the sorted keys to populate the result window
	if len(paths) > 0 {
		fyne.CurrentApp().Preferences().SetString(lastGoodSearchPrefName, s)
		list := widget.NewList(
			func() int { return len(paths) },
			func() fyne.CanvasObject {
				return widget.NewLabel("")
			},
			func(lii widget.ListItemID, co fyne.CanvasObject) {
				co.(*widget.Label).SetText(paths[lii])
			},
		)
		// When an item is selected, start the main thread to select it in the main tree
		// pendingSelection is the path to the selected item
		list.OnSelected = func(id widget.ListItemID) {
			pendingSelection = mapPaths[paths[id]]
			requestMainThreadIn(100, MAIN_THREAD_SELECT)
		}
		go showSearchResultsWindow(window.Canvas().Size().Width/2, window.Canvas().Size().Height/2, list)
	} else {
		dialog.NewInformation("Search results", fmt.Sprintf("Nothing found for search '%s'", s), window).Show()
	}
}

/**
Display the search results in a NON modal window
*/
func showSearchResultsWindow(w float32, h float32, list *widget.List) {
	if searchWindow != nil {
		searchWindow.Close()
		searchWindow = nil
	}
	c := container.NewScroll(list)
	searchWindow = fyne.CurrentApp().NewWindow("Search List")
	searchWindow.SetContent(c)
	searchWindow.Resize(fyne.NewSize(w, h))
	searchWindow.SetFixedSize(true)
	searchWindow.Show()
}

/**
Selecting the menu to add a user
*/
func addNewUser() {
	addNewEntity("User", "User", ADD_TYPE_USER)
}

/**
Selecting the menu to add a hint
*/
func addNewHint() {
	addNewEntity("Hint for "+currentUser, "Hint", ADD_TYPE_HINT)
}

/**
Selecting the menu to add a note
*/
func addNewNote() {
	addNewEntity("Note for "+currentUser, "Note", ADD_TYPE_NOTE)
}

/**
Add an entity to the model.
Delegate to DataRoot for the logic. Call back on dataMapUpdated function if a change is made
*/
func addNewEntity(head string, name string, addType int) {
	gui.NewModalEntryDialog(window, "Enter the name of the new "+head, "", func(accept bool, s string) {
		if accept {
			err := validateEntityName(s)
			if err == nil {
				switch addType {
				case ADD_TYPE_USER:
					err = dataRoot.AddUser(s)
				case ADD_TYPE_NOTE:
					err = dataRoot.AddNote(currentUser, s)
				case ADD_TYPE_HINT:
					err = dataRoot.AddHint(currentUser, s)
				}
			}
			if err != nil {
				dialog.NewInformation("Add New "+name, "Error: "+err.Error(), window).Show()
			}
		}
	})
}

/**
Validate the names of entities. These result in JSON entity names so require
some restrictions.
*/
func validateEntityName(entry string) error {
	if len(entry) == 0 {
		return fmt.Errorf("input is undefined")
	}
	if len(entry) < 2 {
		return fmt.Errorf("input '%s' is too short. Must be longer that 1 char", entry)
	}
	lcEntry := strings.ToLower(entry)
	for _, c := range lcEntry {
		if c < ' ' {
			return fmt.Errorf("input must not contain control characters")
		}
		if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (strings.ContainsRune(allowedCharsInName, c)) {
			continue
		}
		return fmt.Errorf("input must not contain character '%c'. Only '0..9', 'a..z', 'A..Z' and '%s' chars are allowed", c, allowedCharsInName)
	}
	return nil
}

/**
Get the password to decrypt the loaded data contained in FileData
*/
func getPasswordAndDecrypt(fd *lib.FileData, message string, fail func(string)) {
	if message == "" {
		message = "Enter the password to DECRYPT the file"
	}
	running := true
	gui.NewModalPasswordDialog(window, message, "", func(ok bool, value string) {
		if ok {
			if value == "" {
				fail("password is empty")
			} else {
				err := fd.DecryptContents([]byte(value))
				if err != nil {
					// Encryption failed so sanitise the error message and pass it back so we can try again
					s := err.Error()
					p := strings.Index(s, ":")
					if p > 0 {
						fail(s[p+1:])
					} else {
						fail(s)
					}
				}
			}
			running = false
		} else {
			fmt.Printf("Failed to decrypt data file %s. password was not provided\n", fd.GetFileName())
			os.Exit(1)
		}
	})
	// This method must not end until OK or Cancel are pressed
	for running {
		time.Sleep(500 * time.Millisecond)
	}
}

/**
Commit changes to the model and save it to file.
If we are saving encrypted then capture a password and encrypt the file
Otherwise save it as it was loaded!
commitChangedItems writes any un-doable entries to the model. Nothing structural!
*/
func commitAndSaveData(enc int, mustBeChanged bool) {
	count := countChangedItems()
	if count == 0 && mustBeChanged {
		dialog.NewInformation("File Save", "There were no items to save!\n\nPress OK to continue", window).Show()
	} else {
		if enc == SAVE_ENCRYPTED {
			gui.NewModalPasswordDialog(window, "Enter the password to DECRYPT the file", "", func(ok bool, value string) {
				if ok {
					if value != "" {
						_, err := commitChangedItems()
						if err != nil {
							dialog.NewInformation("Convert To Json:", fmt.Sprintf("Error Message:\n-- %s --\nFile was not saved\nPress OK to continue", err.Error()), window).Show()
							return
						}
						err = fileData.StoreContentEncrypted([]byte(value))
						if err != nil {
							dialog.NewInformation("Save Encrypted File Error:", fmt.Sprintf("Error Message:\n-- %s --\nFile may not be saved!\nPress OK to continue", err.Error()), window).Show()
						} else {
							countStructureChanges = 0
						}
					} else {
						dialog.NewInformation("Save Encrypted File Error:", "Error Message:\n\n-- Password not provided --\n\nFile was not saved!\nPress OK to continue", window).Show()
					}
				}
			})
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
				countStructureChanges = 0
			}
		}
	}
}

func shouldClose() {
	savePreferences()
	if !shouldCloseLock {
		shouldCloseLock = true
		if searchWindow != nil {
			searchWindow.Close()
		}
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
	count := countStructureChanges
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

	if !window.FullScreen() {
		p.SetFloat(widthPrefName, float64(window.Canvas().Size().Width))
		p.SetFloat(heightPrefName, float64(window.Canvas().Size().Height))
	}
}

func setThemeById(varient string) {
	t := theme2.NewAppTheme(varient)
	fyne.CurrentApp().Settings().SetTheme(t)
	preferences.PutRootString(themeVarPrefName, varient)
	preferences.Save()
}

func saveChangesConfirm(option bool) {
	shouldCloseLock = false
	if !option {
		fmt.Println("Quit without saving changes")
		window.Close()
	}
}
