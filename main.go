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
	MAIN_THREAD_RESELECT    = iota
	MAIN_THREAD_RE_MENU     = iota

	ADD_TYPE_USER      = iota
	ADD_TYPE_HINT      = iota
	ADD_TYPE_HINT_ITEM = iota
	ADD_TYPE_NOTE_ITEM = iota

	defaultScreenWidth  = 640
	defaultScreenHeight = 480

	allowedCharsInName      = " *@#$%^&*()_+=?"
	fallbackPreferencesFile = "preferences.json"

	dataFilePrefName       = "datafile"
	themeVarPrefName       = "theme"
	screenWidthPrefName    = "screen.width"
	screenHeightPrefName   = "screen.height"
	screenFullPrefName     = "screen.fullScreen"
	screenSplitPrefName    = "screen.split"
	searchLastGoodPrefName = "search.lastGoodList"
	searchCasePrefName     = "search.case"
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

	findCaseSensitive     = binding.NewBool()
	pendingSelection      = ""
	currentSelection      = ""
	shouldCloseLock       = false
	mainThreadNextState   = 0
	loadThreadFileName    = ""
	countStructureChanges = 0
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
	preferences.AddChangeListener(dataPreferencesChanged, "data.")

	loadThreadFileName = p.GetStringForPathWithFallback(dataFilePrefName, fallbackPreferencesFile)
	mainThreadNextState = MAIN_THREAD_IDLE

	a := app.NewWithID("stuartdd.enctest")
	a.Settings().SetTheme(theme2.NewAppTheme(preferences.GetStringForPathWithFallback(themeVarPrefName, "dark")))
	a.SetIcon(theme2.AppLogo())

	window = a.NewWindow(fmt.Sprintf("Data File: %s not loaded yet", loadThreadFileName))

	splitContainerOffsetPref = preferences.GetFloat64WithFallback(screenSplitPrefName, 0.2)
	splitContainerOffset = -1

	findCaseSensitive.Set(preferences.GetBoolWithFallback(searchCasePrefName, true))
	findCaseSensitive.AddListener(binding.NewDataListener(func() {
		b, err := findCaseSensitive.Get()
		if err == nil {
			preferences.PutBool(searchCasePrefName, b)
			preferences.Save()
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
	window.SetMaster()

	wp := gui.GetWelcomePage("", *preferences)
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
		window.SetTitle(fmt.Sprintf("Data File: [%s]. Current User: %s", fileData.GetFileName(), lib.GetUserFromPath(currentSelection)))
		window.SetMainMenu(makeMenus())
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
			mainThreadNextState is the action performed by the main loop
				Each action should leave the state as MAIN_THREAD_IDLE unless a follow on action is required
		*/
		ticker := time.NewTicker(100 * time.Millisecond)
		for range ticker.C {
			switch mainThreadNextState {
			case MAIN_THREAD_LOAD:
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
				mainThreadNextState = MAIN_THREAD_RELOAD_TREE
			case MAIN_THREAD_RELOAD_TREE:
				// Re-build the main tree view.
				// Select the root of current user if defined.
				// Init the devider (split)
				// Populate the window and we are done!
				navTreeLHS = makeNavTree(setPageRHSFunc)
				uid := dataRoot.GetRootUidOrCurrentUid(currentSelection)
				fmt.Printf("Refresh current:%s\n", uid)
				selectTreeElement(uid)

				if splitContainerOffset < 0 {
					splitContainerOffset = splitContainerOffsetPref
				} else {
					splitContainerOffset = splitContainer.Offset
				}
				splitContainer = container.NewHSplit(container.NewBorder(makeSearchLHS(setPageRHSFunc), nil, nil, nil, navTreeLHS), layoutRHS)
				splitContainer.SetOffset(splitContainerOffset)

				window.SetContent(splitContainer)
				mainThreadNextState = MAIN_THREAD_RE_MENU
			case MAIN_THREAD_RESELECT:
				fmt.Println("RE-SELECT")
				t := gui.GetDetailPage(currentSelection, dataRoot.GetDataRootMap(), *preferences)
				setPageRHSFunc(*t)
				mainThreadNextState = MAIN_THREAD_IDLE
			case MAIN_THREAD_SELECT:
				fmt.Printf("Select pending:%s\n", pendingSelection)
				selectTreeElement(pendingSelection)
				mainThreadNextState = MAIN_THREAD_IDLE
			case MAIN_THREAD_RE_MENU:
				window.SetMainMenu(makeMenus())
				mainThreadNextState = MAIN_THREAD_IDLE
			}
		}
	}()

	futureSetMainThread(1000, MAIN_THREAD_LOAD)
	setFullScreen(preferences.GetBoolWithFallback(screenFullPrefName, false))
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

func futureSetMainThread(ms int, status int) {
	go func() {
		time.Sleep(time.Duration(ms) * time.Millisecond)
		count := 100
		for mainThreadNextState != MAIN_THREAD_IDLE {
			time.Sleep(100 * time.Millisecond)
			count--
			if count <= 0 {
				return
			}
		}
		mainThreadNextState = status
	}()
}

func makeMenus() *fyne.MainMenu {
	hintName := preferences.GetStringForPathWithFallback(gui.DataHintIsCalledPrefName, "Hint")
	noteName := preferences.GetStringForPathWithFallback(gui.DataNoteIsCalledPrefName, "Note")
	hint := lib.GetHintFromPath(currentSelection)
	user := lib.GetUserFromPath(currentSelection)

	n1 := fyne.NewMenuItem(fmt.Sprintf("%s for '%s'", noteName, user), addNewNoteItem)
	n3 := fyne.NewMenuItem(fmt.Sprintf("%s for '%s'", hintName, user), addNewHint)
	n4 := fyne.NewMenuItem("User", addNewUser)
	var newItem *fyne.Menu
	if hint == "" {
		newItem = fyne.NewMenu("New", n1, n3, n4)
	} else {
		newItem = fyne.NewMenu("New",
			n1,
			fyne.NewMenuItem(fmt.Sprintf("%s Item for '%s'", hintName, hint), addNewHintItem),
			n3, n4)
	}

	var themeMenuItem *fyne.MenuItem
	if preferences.GetStringForPathWithFallback(themeVarPrefName, "dark") == "dark" {
		themeMenuItem = fyne.NewMenuItem("Light Theme", func() {
			setThemeById("light")
			futureSetMainThread(300, MAIN_THREAD_RESELECT)
		})
	} else {
		themeMenuItem = fyne.NewMenuItem("Dark Theme", func() {
			setThemeById("dark")
			futureSetMainThread(300, MAIN_THREAD_RESELECT)
		})
	}

	viewItem := fyne.NewMenu("View",
		fyne.NewMenuItem(oneOrTheOther(preferences.GetBoolWithFallback(screenFullPrefName, false), "View Windowed", "View Full Screen"), flipFullScreen),
		fyne.NewMenuItem(oneOrTheOther(preferences.GetBoolWithFallback(gui.DataPositionalPrefName, true), "Hide Positional Data", "Show Positional Data"), flipPositionalData),
		themeMenuItem,
	)

	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("Documentation", func() {
			u, _ := url.Parse("https://developer.fyne.io")
			_ = fyne.CurrentApp().OpenURL(u)
		}),
		fyne.NewMenuItem("Support", func() {
			u, _ := url.Parse("https://fyne.io/support/")
			_ = fyne.CurrentApp().OpenURL(u)
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Sponsor", func() {
			u, _ := url.Parse("https://github.com/sponsors/fyne-io")
			_ = fyne.CurrentApp().OpenURL(u)
		}))

	saveItem := fyne.NewMenuItem("Save", func() {
		commitAndSaveData(SAVE_AS_IS, true)
	})

	saveAsItem := fyne.NewMenuItem("Undefined", func() {})
	if fileData != nil {
		if fileData.IsEncryptedOnDisk() {
			saveAsItem = fyne.NewMenuItem("Save Un-Encrypted", func() {
				commitAndSaveData(SAVE_UN_ENCRYPTED, false)
			})
		} else {
			saveAsItem = fyne.NewMenuItem("Save Encrypted", func() {
				commitAndSaveData(SAVE_ENCRYPTED, false)
			})
		}
	}

	return fyne.NewMainMenu(
		// a quit item will be appended to our first menu
		fyne.NewMenu("File", saveItem, saveAsItem, fyne.NewMenuItemSeparator()),
		newItem,
		viewItem,
		helpMenu,
	)
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
			t := gui.GetDetailPage(uid, dataRoot.GetDataRootMap(), *preferences)
			obj.(*widget.Label).SetText(t.Title)
		},
		OnSelected: func(uid string) {
			t := gui.GetDetailPage(uid, dataRoot.GetDataRootMap(), *preferences)
			setPage(*t)
		},
	}
}

/**
Section below the tree with Search details and Light and Dark theme buttons
*/
func makeSearchLHS(setPage func(detailPage gui.DetailPage)) fyne.CanvasObject {
	x := preferences.GetStringList(searchLastGoodPrefName)
	searchEntry := widget.NewSelectEntry(x)
	searchEntry.SetText(x[0])
	c2 := container.New(
		layout.NewHBoxLayout(),
		widget.NewLabel("Find:"),
		widget.NewButtonWithIcon("", theme.SearchIcon(), func() { search(searchEntry.Text) }),
		widget.NewCheckWithData("Match Case", findCaseSensitive))
	c := container.New(
		layout.NewVBoxLayout(),
		c2,
		searchEntry)

	return container.NewVBox(widget.NewSeparator(), c)
}

/**
This is called when a heading button is pressed of the RH page
*/
func controlActionFunction(action string, uid string) {
	fmt.Printf("Control %s %s\n", action, uid)
	viewActionFunction(action, uid)
}

func dataPreferencesChanged(path, value, filter string) {
	futureSetMainThread(100, MAIN_THREAD_RESELECT)
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
	futureSetMainThread(100, MAIN_THREAD_RELOAD_TREE)
}

func flipPositionalData() {
	p := preferences.GetBoolWithFallback(gui.DataPositionalPrefName, true)
	preferences.PutBool(gui.DataPositionalPrefName, !p)
	futureSetMainThread(500, MAIN_THREAD_RE_MENU)
}

/**
flip the full screen flag and set the screen to full screen if required
*/
func flipFullScreen() {
	setFullScreen(!window.FullScreen())
}

func setFullScreen(fullScreen bool) {
	if fullScreen {
		// Ensure that 0.0 is never stored.
		if window.Canvas().Size().Width > 1 && window.Canvas().Size().Height > 1 {
			preferences.PutFloat32(screenWidthPrefName, window.Canvas().Size().Width)
			preferences.PutFloat32(screenHeightPrefName, window.Canvas().Size().Height)
		}
		window.Resize(fyne.NewSize(defaultScreenWidth, defaultScreenHeight))
		window.SetFullScreen(true)
	} else {
		window.SetFullScreen(false)
		sw := preferences.GetFloat32WithFallback(screenWidthPrefName, defaultScreenWidth)
		sh := preferences.GetFloat32WithFallback(screenHeightPrefName, defaultScreenHeight)
		window.Resize(fyne.NewSize(sw, sh))
	}
	preferences.PutBool(screenFullPrefName, fullScreen)
	futureSetMainThread(500, MAIN_THREAD_RE_MENU)
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
Activate a link in a browser if it is contained in a note or hint
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
		preferences.PutStringList(searchLastGoodPrefName, s, 10)
		defer futureSetMainThread(200, MAIN_THREAD_RELOAD_TREE)
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
			futureSetMainThread(100, MAIN_THREAD_SELECT)
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
	n := preferences.GetStringForPathWithFallback(gui.DataHintIsCalledPrefName, "Hint")
	addNewEntity(n+" for ", n, ADD_TYPE_HINT)
}

/**
Selecting the menu to add an item to a hint
*/
func addNewHintItem() {
	n := preferences.GetStringForPathWithFallback(gui.DataNoteIsCalledPrefName, "Hint")
	ch := lib.GetHintFromPath(currentSelection)
	if ch == "" {
		dialog.NewInformation("Add New "+n, fmt.Sprintf("A %s needs to be selected", n), window).Show()
	} else {
		addNewEntity(fmt.Sprintf("%s Item for %s", n, ch), n, ADD_TYPE_HINT_ITEM)
	}
}

/**
Selecting the menu to add an item to the notes
*/
func addNewNoteItem() {
	n := preferences.GetStringForPathWithFallback(gui.DataNoteIsCalledPrefName, "Note")
	addNewEntity(n+" Item for ", n, ADD_TYPE_NOTE_ITEM)
}

/**
Add an entity to the model.
Delegate to DataRoot for the logic. Call back on dataMapUpdated function if a change is made
*/
func addNewEntity(head string, name string, addType int) {
	cu := lib.GetUserFromPath(currentSelection)
	gui.NewModalEntryDialog(window, "Enter the name of the new "+head, "", func(accept bool, s string) {
		if accept {
			err := validateEntityName(s)
			if err == nil {
				switch addType {
				case ADD_TYPE_USER:
					err = dataRoot.AddUser(s)
				case ADD_TYPE_NOTE_ITEM:
					err = dataRoot.AddNoteItem(cu, s)
				case ADD_TYPE_HINT:
					err = dataRoot.AddHint(cu, s)
				case ADD_TYPE_HINT_ITEM:
					err = dataRoot.AddHintItem(cu, lib.GetHintFromPath(currentSelection), s)
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

func callbackAfterSave() {
	defer futureSetMainThread(500, MAIN_THREAD_RE_MENU)
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
						err = fileData.StoreContentEncrypted([]byte(value), callbackAfterSave)
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
				err = fileData.StoreContentAsIs(callbackAfterSave)
			} else {
				err = fileData.StoreContentUnEncrypted(callbackAfterSave)
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

	if splitContainer != nil {
		preferences.PutFloat64(screenSplitPrefName, splitContainer.Offset)
	}

	if !window.FullScreen() {
		// Ensure that 0.0 is never stored.
		if window.Canvas().Size().Width > 1 && window.Canvas().Size().Height > 1 {
			preferences.PutFloat32(screenWidthPrefName, window.Canvas().Size().Width)
			preferences.PutFloat32(screenHeightPrefName, window.Canvas().Size().Height)
		}
	}

	preferences.Save()
}

func setThemeById(varient string) {
	t := theme2.NewAppTheme(varient)
	fyne.CurrentApp().Settings().SetTheme(t)
	preferences.PutString(themeVarPrefName, varient)
}

func saveChangesConfirm(option bool) {
	shouldCloseLock = false
	if !option {
		fmt.Println("Quit without saving changes")
		window.Close()
	}
}

func oneOrTheOther(one bool, s1, s2 string) string {
	if one {
		return s1
	}
	return s2
}
