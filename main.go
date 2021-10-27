/*
 * Copyright (C) 2021 Stuart Davies (stuartdd)
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */
package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
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
	"github.com/stuartdd/jsonParserGo/parser"
	"stuartdd.com/gui"
	"stuartdd.com/lib"
	"stuartdd.com/pref"
	"stuartdd.com/theme2"
	"stuartdd.com/types"
)

const (
	SAVE_AS_IS              = iota
	SAVE_ENCRYPTED          = iota
	SAVE_UN_ENCRYPTED       = iota
	MAIN_THREAD_LOAD        = iota
	MAIN_THREAD_RELOAD_TREE = iota
	MAIN_THREAD_SELECT      = iota
	MAIN_THREAD_RESELECT    = iota
	MAIN_THREAD_RE_MENU     = iota

	ADD_TYPE_USER            = iota
	ADD_TYPE_HINT            = iota
	ADD_TYPE_HINT_CLONE      = iota
	ADD_TYPE_HINT_CLONE_FULL = iota
	ADD_TYPE_HINT_ITEM       = iota
	ADD_TYPE_NOTE_ITEM       = iota

	defaultScreenWidth  = 640
	defaultScreenHeight = 480

	fallbackPreferencesFile = "config.json"
	fallbackDataFile        = "data.json"

	dataFilePrefName       = "datafile"
	copyDialogTimePrefName = "dialog.copyTimeOutMS"
	linkDialogTimePrefName = "dialog.linkTimeOutMS"
	saveDialogTimePrefName = "dialog.saveTimeOutMS"
	getUrlPrefName         = "getDataUrl"
	postUrlPrefName        = "postDataUrl"
	themeVarPrefName       = "theme"
	logFileNamePrefName    = "log.fileName"
	logActivePrefName      = "log.active"
	logPrefixPrefName      = "log.prefix"
	screenWidthPrefName    = "screen.width"
	screenHeightPrefName   = "screen.height"
	screenFullPrefName     = "screen.fullScreen"
	screenSplitPrefName    = "screen.split"
	searchLastGoodPrefName = "search.lastGoodList"
	searchCasePrefName     = "search.case"
)

var (
	window                   fyne.Window
	searchWindow             *gui.SearchDataWindow
	logData                  *gui.LogData
	fileData                 *lib.FileData
	jsonData                 *lib.JsonData
	preferences              *pref.PrefData
	navTreeLHS               *widget.Tree
	saveShortcutButton       *widget.Button
	fullScreenShortcutButton *widget.Button
	editModeShortcutButton   *widget.Button
	timeStampLabel           *widget.Label
	splitContainer           *container.Split // So we can save the divider position to preferences.
	splitContainerOffset     float64          = -1
	splitContainerOffsetPref float64          = -1

	findCaseSensitive  = binding.NewBool()
	pendingSelection   = ""
	currentSelection   = ""
	shouldCloseLock    = false
	hasDataChanges     = false
	releaseTheBeast    = make(chan int, 1)
	dataIsNotLoadedYet = true
)

func abortWithUsage(message string) {
	fmt.Printf(message+"\n  Usage: %s <configfile>\n  Where: <configfile> is a json file. E.g. config.json\n", os.Args[0])
	fmt.Println("    Minimum content for this file is:\n      {\"datafile\": \"dataFile.json\"}")
	fmt.Println("    Where \"dataFile.json\" is the name of the required data file")
	fmt.Println("    This file will be updated by this application")
	os.Exit(1)
}

func main() {
	var prefFile string
	if len(os.Args) < 2 {
		prefFile = fallbackPreferencesFile
	} else {
		prefFile = os.Args[1]
	}
	p, err := pref.NewPrefData(prefFile)
	if err != nil {
		abortWithUsage(fmt.Sprintf("Failed to load configuration file '%s'", prefFile))
	}
	preferences = p

	loadThreadFileName := p.GetStringForPathWithFallback(dataFilePrefName, fallbackDataFile)
	getDataUrl := p.GetStringForPathWithFallback(getUrlPrefName, "")
	postDataUrl := p.GetStringForPathWithFallback(postUrlPrefName, "")
	//
	// For extended command line options. Dont use logData use std out!
	//
	if len(os.Args) > 2 {
		switch os.Args[2] {
		case "create":
			createFile := oneOrTheOther(postDataUrl == "", loadThreadFileName, postDataUrl+"/"+loadThreadFileName)
			fmt.Printf("-> Create new data file '%s'\n", createFile)
			fmt.Printf("-> File is defined in config data file '%s'\n", prefFile)
			fmt.Println("-> Existing data will be overwritten!")
			fmt.Print("-> ARE YOU SURE. (Y/n)")
			reader := bufio.NewReader(os.Stdin)
			text, _ := reader.ReadString('\n')
			if !strings.HasPrefix(text, "Y") {
				fmt.Println("----> Action aborted. You need to type capitol Y to procceed.")
				os.Exit(0)
			}
			data := lib.CreateEmptyJsonData()
			var err error
			if postDataUrl != "" {
				_, err = parser.PostJsonBytes(fmt.Sprintf("%s/%s", postDataUrl, loadThreadFileName), data)
			} else {
				err = ioutil.WriteFile(loadThreadFileName, data, 0644)
			}
			if err != nil {
				fmt.Println(err.Error())
			} else {
				fmt.Printf("-> File %s has been created\n", createFile)
			}
		}
		os.Exit(0)
	}

	logData = gui.NewLogData(
		preferences.GetStringForPathWithFallback(logFileNamePrefName, "enctest.log"),
		preferences.GetStringForPathWithFallback(logPrefixPrefName, "INFO")+": ",
		preferences.GetBoolWithFallback(logActivePrefName, false))

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

	/*
		Create the menus
	*/
	window.SetMaster()

	wp := gui.GetWelcomePage("", *preferences)
	title := container.NewHBox()
	title.Objects = []fyne.CanvasObject{wp.CntlFunc(window, *wp, nil)}
	contentRHS := container.NewMax()
	layoutRHS := container.NewBorder(title, nil, nil, nil, contentRHS)
	buttonBar := makeButtonBar()
	searchWindow = gui.NewSearchDataWindow(closeSearchWindow, selectTreeElement)
	/*
		function called when a selection is made in the LHS tree.
		This updates the contentRHS which is the RHS page for editing data
	*/
	setPageRHSFunc := func(detailPage gui.DetailPage) {
		currentSelection = detailPage.Uid
		if searchWindow != nil {
			go searchWindow.Select(currentSelection)
		}
		log(fmt.Sprintf("Page User:'%s' Uid:'%s'", lib.GetUserFromPath(currentSelection), currentSelection))
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
			queue a valid state on the the channel to releaseTheBeast
		*/
		for { // EVER

			taskForTheBeast := <-releaseTheBeast

			switch taskForTheBeast {
			case MAIN_THREAD_LOAD:
				dataIsNotLoadedYet = true

				if logData.IsWarning() {
					timedNotification(5000, "Log file error", logData.GetErr().Error())
				}
				// Load the file and decrypt it if required
				fd, err := lib.NewFileData(loadThreadFileName, getDataUrl, postDataUrl)
				if err != nil {
					abortWithUsage(fmt.Sprintf("Failed to load data file %s. Error: %s\n", loadThreadFileName, err.Error()))
				}
				if getDataUrl != "" {
					log(fmt.Sprintf("Remote File:'%s/%s'", getDataUrl, loadThreadFileName))
				} else {
					log(fmt.Sprintf("Local File:'%s'", loadThreadFileName))
				}
				/*
					While file is ENCRYPTED
						Get PW and decrypt
						if Decryption is cancelled the application exits
				*/
				message := ""
				for fd.RequiresDecryption() {
					log(fmt.Sprintf("Requires Decryption:'%s'", loadThreadFileName))
					getPasswordAndDecrypt(fd, message, func(s string) {
						// FAIL
						message = "Error: " + strings.TrimSpace(s) + ". Please try again"
						log(fmt.Sprintf("Decryption Error:'%s'", message))
						time.Sleep(1000 * time.Millisecond)
					})
				}
				/*
					Data is decrypted so process the JSON so
						update the navigation tree
						select the root element
				*/
				dr, err := lib.NewJsonData(fd.GetContent(), dataMapUpdated)
				if err != nil {
					abortWithUsage(fmt.Sprintf("ERROR: Cannot process data in file '%s'.\n%s\n", loadThreadFileName, err))
				}
				fileData = fd
				jsonData = dr
				dataIsNotLoadedYet = false
				log(fmt.Sprintf("Data Parsed OK: File:'%s' DateTime:'%s'", loadThreadFileName, jsonData.GetTimeStampString()))
				// Follow on action to rebuild the Tree and re-display it
				futureReleaseTheBeast(0, MAIN_THREAD_RELOAD_TREE)
			case MAIN_THREAD_RELOAD_TREE:
				// Re-build the main tree view.
				// Select the root of current user if defined.
				// Init the devider (split)
				// Populate the window and we are done!
				navTreeLHS = makeNavTree(setPageRHSFunc)
				uid := jsonData.GetRootUidOrCurrentUid(currentSelection)
				log(fmt.Sprintf("Re-build nav tree. Sel:'%s' uid:'%s'", currentSelection, uid))
				selectTreeElement("MAIN_THREAD_RELOAD_TREE", uid)
				if splitContainerOffset < 0 {
					splitContainerOffset = splitContainerOffsetPref
				} else {
					splitContainerOffset = splitContainer.Offset
				}
				splitContainer = container.NewHSplit(container.NewBorder(makeSearchLHS(setPageRHSFunc), nil, nil, nil, navTreeLHS), layoutRHS)
				splitContainer.SetOffset(splitContainerOffset)
				window.SetContent(container.NewBorder(buttonBar, nil, nil, nil, splitContainer))
				futureReleaseTheBeast(0, MAIN_THREAD_RE_MENU)
			case MAIN_THREAD_RESELECT:
				log(fmt.Sprintf("Re-display RHS. Sel:'%s'", currentSelection))
				t := gui.GetDetailPage(currentSelection, jsonData.GetDataRoot(), *preferences)
				setPageRHSFunc(*t)
			case MAIN_THREAD_SELECT:
				selectTreeElement("MAIN_THREAD_SELECT", pendingSelection)
			case MAIN_THREAD_RE_MENU:
				log("Refresh menu and buttons")
				updateButtonBar()
				window.SetMainMenu(makeMenus())
			}
		}
	}()

	futureReleaseTheBeast(500, MAIN_THREAD_LOAD)
	preferences.AddChangeListener(dataPreferencesChanged, "data.")
	setFullScreen(preferences.GetBoolWithFallback(screenFullPrefName, false), false)
	window.ShowAndRun()
}

/**
Select a tree element.
We need to open the parent branches or we will never see the selected element
*/
func selectTreeElement(desc, uid string) {
	user := lib.GetUserFromPath(uid)
	parent := lib.GetParentId(uid)
	log(fmt.Sprintf("selectTreeElement: Desc:'%s' User:'%s' Parent:'%s' Uid:'%s'", desc, user, parent, uid))
	navTreeLHS.OpenBranch(user)
	navTreeLHS.OpenBranch(parent)
	navTreeLHS.Select(uid)
	navTreeLHS.ScrollTo(uid)
}

func futureReleaseTheBeast(ms int, status int) {
	if ms < 1 {
		releaseTheBeast <- status
	} else {
		go func() {
			//
			// Wait for the required time before sending the request
			//
			time.Sleep(time.Duration(ms) * time.Millisecond)
			//
			// Cannot do anything until MAIN_THREAD_LOAD is run and dataIsLoaded = true
			// But we cannot give in so we wait. But not forever!
			//
			if status != MAIN_THREAD_LOAD && dataIsNotLoadedYet {
				count := 0
				for dataIsNotLoadedYet {
					time.Sleep(100 * time.Millisecond)
					count++
					if count > 10 {
						return
					}
				}
			}
			//
			// Sent the request
			//
			releaseTheBeast <- status
		}()
	}
}

/*
Dont call directly. Use:
	futureReleaseTheBeast(100, MAIN_THREAD_RE_MENU)
*/
func updateButtonBar() {
	if countChangedItems() > 0 {
		saveShortcutButton.Enable()
	} else {
		saveShortcutButton.Disable()
	}
	if preferences.GetBoolWithFallback(screenFullPrefName, false) {
		fullScreenShortcutButton.SetText("Windowed")
	} else {
		fullScreenShortcutButton.SetText("Full Screen")
	}
	if preferences.GetBoolWithFallback(gui.DataPresModePrefName, true) {
		editModeShortcutButton.SetText("Edit Data")
	} else {
		editModeShortcutButton.SetText("Present Data")
	}
	if jsonData == nil {
		timeStampLabel.SetText("  File Not loaded")
	} else {
		timeStampLabel.SetText("  Last Updated -> " + jsonData.GetTimeStampString())
	}
}

func log(l string) {
	if logData.IsLogging() {
		logData.Log(l)
	}
}

func logInformationDialog(title, message string) dialog.Dialog {
	return logInformationDialogWithPrefix("Dialog-info", title, message)
}

func timedNotification(msDelay int64, title, message string) {
	go func() {
		dia := logInformationDialogWithPrefix("Dialog-timed", title, message)
		time.Sleep(time.Duration(msDelay) * time.Millisecond)
		dia.Hide()
	}()
}

func logInformationDialogWithPrefix(prefix, title, message string) dialog.Dialog {
	if logData.IsLogging() {
		logData.Log(fmt.Sprintf("%s: Title:'%s' Message:'%s'", prefix, title, message))
	}
	dil := dialog.NewInformation(title, message, window)
	dil.Show()
	return dil
}

func logDataRequest(action string) {
	switch action {
	case "onoff":
		logData.FlipOnOff()
	case "navmap":
		log(fmt.Sprintf("NavMap: ----------------\n%s", jsonData.GetNavIndexAsString()))
	case "select":
		m, err := lib.GetUserDataForUid(jsonData.GetDataRoot(), currentSelection)
		if err != nil {
			log(fmt.Sprintf("Data for uid [%s] not found. %s", currentSelection, err.Error()))
		}
		if m != nil {
			log(fmt.Sprintf("UID:'%s'. Json:%s", currentSelection, m.JsonValueIndented(4)))
		} else {
			log(fmt.Sprintf("Data for uid [%s] returned null", currentSelection))
		}
	default:
		log(fmt.Sprintf("Log Action: '%s' unknown", action))
	}
}

func makeButtonBar() *fyne.Container {
	saveShortcutButton = widget.NewButton("Save", func() {
		commitAndSaveData(SAVE_AS_IS, true)
	})
	fullScreenShortcutButton = widget.NewButton("FULL SCREEN", flipFullScreen)
	editModeShortcutButton = widget.NewButton("EDIT", flipPositionalData)

	quit := widget.NewButton("EXIT", shouldClose)
	timeStampLabel = widget.NewLabel("  File Not loaded")
	return container.NewHBox(quit, saveShortcutButton, gui.Padding50, fullScreenShortcutButton, editModeShortcutButton, timeStampLabel)
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
			n3,
			fyne.NewMenuItem(fmt.Sprintf("Clone '%s'", hint), cloneHint),
			fyne.NewMenuItem(fmt.Sprintf("Clone Full '%s'", hint), cloneHintFull),
			n4)
	}

	var themeMenuItem *fyne.MenuItem
	if preferences.GetStringForPathWithFallback(themeVarPrefName, "dark") == "dark" {
		themeMenuItem = fyne.NewMenuItem("Light Theme", func() {
			setThemeById("light")
			futureReleaseTheBeast(300, MAIN_THREAD_RESELECT)
		})
	} else {
		themeMenuItem = fyne.NewMenuItem("Dark Theme", func() {
			setThemeById("dark")
			futureReleaseTheBeast(300, MAIN_THREAD_RESELECT)
		})
	}

	viewItem := fyne.NewMenu("View",
		fyne.NewMenuItem(oneOrTheOther(preferences.GetBoolWithFallback(screenFullPrefName, false), "View Windowed", "View Full Screen"), flipFullScreen),
		fyne.NewMenuItem(oneOrTheOther(preferences.GetBoolWithFallback(gui.DataPresModePrefName, true), "Edit Data", "Present Data"), flipPositionalData),
		themeMenuItem,
	)

	m := make([]*fyne.MenuItem, 0)
	m = append(m, fyne.NewMenuItem("Support", func() {
		u, _ := url.Parse("https://fyne.io/support/")
		_ = fyne.CurrentApp().OpenURL(u)
	},
	))
	m = append(m, fyne.NewMenuItem("Documentation", func() {
		u, _ := url.Parse("https://developer.fyne.io")
		_ = fyne.CurrentApp().OpenURL(u)
	}))

	if logData.IsReady() {
		m = append(m,
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem(oneOrTheOther(logData.IsLogging(), "Log Stop", "Log Start"), func() {
				logDataRequest("onoff")
				futureReleaseTheBeast(100, MAIN_THREAD_RE_MENU)
			}))
		if logData.IsLogging() {
			m = append(m,
				fyne.NewMenuItem("Log Selection", func() {
					logDataRequest("select")
				}),
				fyne.NewMenuItem("Log Nav Map", func() {
					logDataRequest("navmap")
				}),
			)
		}
	}

	helpMenu := fyne.NewMenu("Help", m...)

	saveItem := fyne.NewMenuItem("Save", func() {
		commitAndSaveData(SAVE_AS_IS, true)
	})

	saveAsItem := fyne.NewMenuItem("Undefined", func() {})
	if fileData != nil {
		if fileData.IsEncrypted() {
			saveAsItem = fyne.NewMenuItem("Save Un-Encrypted", func() {
				commitAndSaveData(SAVE_UN_ENCRYPTED, false)
			})
		} else {
			saveAsItem = fyne.NewMenuItem("Save Encrypted", func() {
				commitAndSaveData(SAVE_ENCRYPTED, false)
			})
		}
	}

	mainMenu := fyne.NewMainMenu(
		// a quit item will be appended to our first menu
		fyne.NewMenu("File", saveItem, saveAsItem, fyne.NewMenuItemSeparator()),
		newItem,
		viewItem,
		helpMenu,
	)

	return mainMenu
}

/**
Use the navigation index (map in DataRoot) to construct the Tree.
Calls setPage with the selected page, defined by the uid (path) and Pages (gui) api
*/
func makeNavTree(setPage func(detailPage gui.DetailPage)) *widget.Tree {
	return &widget.Tree{
		ChildUIDs: func(uid string) []string {
			id := jsonData.GetNavIndex(uid)
			return id
		},
		IsBranch: func(uid string) bool {
			children := jsonData.GetNavIndex(uid)
			return len(children) > 0
		},
		CreateNode: func(branch bool) fyne.CanvasObject {
			return widget.NewLabel("?")
		},
		UpdateNode: func(uid string, branch bool, obj fyne.CanvasObject) {
			t := gui.GetDetailPage(uid, jsonData.GetDataRoot(), *preferences)
			obj.(*widget.Label).SetText(t.Title)
		},
		OnSelected: func(uid string) {
			t := gui.GetDetailPage(uid, jsonData.GetDataRoot(), *preferences)
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
func controlActionFunction(action string, uid string, extra string) {
	viewActionFunction(action, uid, extra)
}

func dataPreferencesChanged(path, value, filter string) {
	futureReleaseTheBeast(100, MAIN_THREAD_RESELECT)
}

/**
This is called when a detail button is pressed of the RH page
*/
func viewActionFunction(action string, uid string, extra string) {
	switch action {
	case gui.ACTION_LOG:
		log(gui.LogCleanString(extra, 100))
	case gui.ACTION_REMOVE:
		removeAction(uid)
	case gui.ACTION_RENAME:
		renameAction(uid, extra)
	case gui.ACTION_LINK:
		linkAction(uid, extra)
	case gui.ACTION_UPDATED:
		futureReleaseTheBeast(100, MAIN_THREAD_RE_MENU)
	case gui.ACTION_COPIED:
		timedNotification(preferences.GetInt64WithFallback(copyDialogTimePrefName, 2500), "Copied item text to clipboard", uid)
	}
}

/**
Called if there is a structural change in the model
*/
func dataMapUpdated(desc, user, path string, err error) {
	if err == nil {
		log(fmt.Sprintf("dataMapUpdated Desc:'%s' User:'%s' Path:'%s'", desc, user, path))
		pp := lib.GetParentId(path)
		if jsonData.GetNavIndex(pp) == nil {
			path = pp
		}
		currentSelection = path
		hasDataChanges = true
	} else {
		log(fmt.Sprintf("dataMapUpdated Desc:'%s' User:'%s' Path:'%s', Err:'%s'", desc, user, path, err.Error()))
	}
	futureReleaseTheBeast(100, MAIN_THREAD_RELOAD_TREE)
}

func flipPositionalData() {
	p := preferences.GetBoolWithFallback(gui.DataPresModePrefName, true)
	preferences.PutBool(gui.DataPresModePrefName, !p)
	futureReleaseTheBeast(500, MAIN_THREAD_RE_MENU)
}

/**
flip the full screen flag and set the screen to full screen if required
*/
func flipFullScreen() {
	setFullScreen(!window.FullScreen(), true)
}

func setFullScreen(fullScreen, refreshMenu bool) {
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
	if refreshMenu {
		futureReleaseTheBeast(500, MAIN_THREAD_RE_MENU)
	}
}

/**
Remove a node from the main data (model) and update the tree view
dataMapUpdated id called if a change is made to the model
*/
func removeAction(uid string) {
	log(fmt.Sprintf("removeAction Uid:'%s'", uid))
	_, removeName := types.GetNodeAnnotationTypeAndName(lib.GetLastId(uid))
	dialog.NewConfirm("Remove entry", fmt.Sprintf("'%s'\nAre you sure?", removeName), func(ok bool) {
		if ok {
			err := jsonData.Remove(uid, 1)
			if err != nil {
				logInformationDialog("Remove item error", err.Error())
			}
		}
	}, window).Show()
}

/**
Rename a node from the main data (model) and update the tree view
dataMapUpdated id called if a change is made to the model
*/
func renameAction(uid string, extra string) {
	log(fmt.Sprintf("renameAction Uid:'%s'  Extra:'%s'", uid, extra))
	m, _ := jsonData.GetUserDataForUid(uid)
	if m != nil {
		at, fromName := types.GetNodeAnnotationTypeAndName(lib.GetLastId(uid))
		toName := ""
		isNote := false
		if jsonData.IsStringNode(m) {
			toName = fromName
			isNote = true
		}
		gui.NewModalEntryDialog(window, fmt.Sprintf("Rename entry '%s' ", fromName), toName, isNote, at, func(accept bool, toName string, nt types.NodeAnnotationEnum) {
			if accept {
				s, err := lib.ProcessEntityName(toName, nt)
				if err != nil {
					logInformationDialog("Name validation error", err.Error())
				} else {
					err := jsonData.Rename(uid, s)
					if err != nil {
						logInformationDialog("Rename item error", err.Error())
					}
				}
			}
		})
	}
}

/**
Activate a link in a browser if it is contained in a note or hint
*/
func linkAction(uid, urlStr string) {
	log(fmt.Sprintf("linkAction Uid:'%s' Url:%s", uid, urlStr))
	if urlStr != "" {
		s, err := url.Parse(urlStr)
		if err != nil {
			logInformationDialog("Error: Link failed to parse", err.Error())
		} else {
			err = fyne.CurrentApp().OpenURL(s)
			if err != nil {
				logInformationDialog("Error: Link could not be opened", err.Error())
			} else {
				timedNotification(preferences.GetInt64WithFallback(linkDialogTimePrefName, 2500), "Open Link (URL)", urlStr)
			}
		}
	} else {
		logInformationDialog("Error: Link could not be opened", "An empty link was provided")
	}
}

func closeSearchWindow() {
	if searchWindow != nil {
		searchWindow.Close()
	}
}

/*
The search button has been pressed
*/
func search(s string) {
	if s == "" {
		return
	}
	matchCase, _ := findCaseSensitive.Get()

	// Do the search. The map contains the returned search entries
	// This ensures no duplicates are displayed
	// The map key is the human readable results e.g. 'User [Hint] app: noteName'
	// The values are paths within the model! user.pwHints.app.noteName
	searchWindow.Reset()
	jsonData.Search(func(path, desc string) {
		searchWindow.Add(desc, path)
	}, s, matchCase)

	// Use the sorted keys to populate the result window
	if searchWindow.Len() > 0 {
		preferences.PutStringList(searchLastGoodPrefName, s, 10)
		defer futureReleaseTheBeast(200, MAIN_THREAD_RELOAD_TREE)
		searchWindow.Show(500, 350, s)
	} else {
		logInformationDialog("Search results", fmt.Sprintf("Nothing found for search '%s'", s))
	}
}

/**
Selecting the menu to add a user
*/
func addNewUser() {
	addNewEntity("User", "User", ADD_TYPE_USER, false)
}

/**
Selecting the menu to add a hint
*/
func addNewHint() {
	n := preferences.GetStringForPathWithFallback(gui.DataHintIsCalledPrefName, "Hint")
	addNewEntity(n+" for ", n, ADD_TYPE_HINT, false)
}

func cloneHint() {
	n := preferences.GetStringForPathWithFallback(gui.DataHintIsCalledPrefName, "Hint")
	addNewEntity(n+" for ", n, ADD_TYPE_HINT_CLONE, false)
}

func cloneHintFull() {
	n := preferences.GetStringForPathWithFallback(gui.DataHintIsCalledPrefName, "Hint")
	addNewEntity(n+" for ", n, ADD_TYPE_HINT_CLONE_FULL, false)
}

/**
Selecting the menu to add an item to a hint
*/
func addNewHintItem() {
	n := preferences.GetStringForPathWithFallback(gui.DataNoteIsCalledPrefName, "Hint")
	ch := lib.GetHintFromPath(currentSelection)
	if ch == "" {
		logInformationDialog("Add New "+n, fmt.Sprintf("A %s needs to be selected", n))
	} else {
		addNewEntity(fmt.Sprintf("%s Item for %s", n, ch), n, ADD_TYPE_HINT_ITEM, true)
	}
}

/**
Selecting the menu to add an item to the notes
*/
func addNewNoteItem() {
	n := preferences.GetStringForPathWithFallback(gui.DataNoteIsCalledPrefName, "Note")
	addNewEntity(n+" Item for ", n, ADD_TYPE_NOTE_ITEM, true)
}

/**
Add an entity to the model.
Delegate to DataRoot for the logic. Call back on dataMapUpdated function if a change is made
*/
func addNewEntity(head string, name string, addType int, isNote bool) {
	cu := lib.GetUserFromPath(currentSelection)
	gui.NewModalEntryDialog(window, "Enter the name of the new "+head, "", isNote, types.NOTE_TYPE_SL, func(accept bool, newName string, nt types.NodeAnnotationEnum) {
		if accept {
			s, err := lib.ProcessEntityName(newName, nt)
			if err == nil {
				switch addType {
				case ADD_TYPE_USER:
					err = jsonData.AddUser(s)
				case ADD_TYPE_NOTE_ITEM:
					err = jsonData.AddNoteItem(cu, s)
				case ADD_TYPE_HINT:
					err = jsonData.AddHint(cu, s)
				case ADD_TYPE_HINT_CLONE:
					err = jsonData.CloneHint(cu, currentSelection, s, false)
				case ADD_TYPE_HINT_CLONE_FULL:
					err = jsonData.CloneHint(cu, currentSelection, s, true)
				case ADD_TYPE_HINT_ITEM:
					err = jsonData.AddHintItem(cu, currentSelection, s)
				}
			}
			if err != nil {
				logInformationDialog("Add New "+name, "Error: "+err.Error())
			}
		}
	})
}

/**
Get the password to decrypt the loaded data contained in FileData
*/
func getPasswordAndDecrypt(fd *lib.FileData, message string, fail func(string)) {
	if message == "" {
		message = "Enter the password to DECRYPT the file"
	}
	running := true
	gui.NewModalPasswordDialog(window, message, "", func(ok bool, value string, nt types.NodeAnnotationEnum) {
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
			abortWithUsage(fmt.Sprintf("Failed to decrypt data file '%s'\nPassword was not provided\n", fd.GetFileName()))
		}
	})
	// This method must not end until OK or Cancel are pressed
	for running {
		time.Sleep(500 * time.Millisecond)
	}
}

func callbackAfterSave() {
	timedNotification(preferences.GetInt64WithFallback(saveDialogTimePrefName, 3000), "Saved", fileData.GetFileName())
	futureReleaseTheBeast(500, MAIN_THREAD_RE_MENU)
}

/**
Commit changes to the model and save it to file.
If we are saving encrypted then capture a password and encrypt the file
Otherwise save it as it was loaded!
commitChangedItems writes any un-doable entries to the model. Nothing structural!
Once we have done all that we must update the button bar to disable the save button
*/
func commitAndSaveData(enc int, mustBeChanged bool) {
	count := countChangedItems()
	if count == 0 && mustBeChanged {
		logInformationDialog("File Save", "There were no items to save!\n\nPress OK to continue")
	} else {
		if enc == SAVE_ENCRYPTED {
			gui.NewModalPasswordDialog(window, "Enter the password to DECRYPT the file", "", func(ok bool, value string, nt types.NodeAnnotationEnum) {
				if ok {
					if value != "" {
						_, err := commitChangedItems()
						if err != nil {
							logInformationDialog("Convert To Json:", fmt.Sprintf("Error Message:\n-- %s --\nFile was not saved\nPress OK to continue", err.Error()))
							return
						}
						err = fileData.StoreContentEncrypted([]byte(value), callbackAfterSave)
						if err != nil {
							logInformationDialog("Save Encrypted File Error:", fmt.Sprintf("Error Message:\n-- %s --\nFile may not be saved!\nPress OK to continue", err.Error()))
						} else {
							hasDataChanges = false
						}
					} else {
						logInformationDialog("Save Encrypted File Error:", "Error Message:\n\n-- Password not provided --\n\nFile was not saved!\nPress OK to continue")
					}
				}
			})
		} else {
			_, err := commitChangedItems()
			if err != nil {
				logInformationDialog("Convert To Json:", fmt.Sprintf("Error Message:\n-- %s --\nFile was not saved\nPress OK to continue", err.Error()))
				return
			}
			if enc == SAVE_AS_IS {
				err = fileData.StoreContentAsIs(callbackAfterSave)
			} else {
				err = fileData.StoreContentUnEncrypted(callbackAfterSave)
			}
			if err != nil {
				logInformationDialog("Save File Error:", fmt.Sprintf("Error Message:\n-- %s --\nFile may not be saved!\nPress OK to continue", err.Error()))
			} else {
				hasDataChanges = false
			}
		}
	}
}

func shouldClose() {
	if !shouldCloseLock {
		shouldCloseLock = true // shouldCloseLock is cleared in the saveChangesDialogAction.
		savePreferences()
		if searchWindow != nil {
			searchWindow.Close()
		}
		count := countChangedItems()
		if count > 0 {
			d := dialog.NewConfirm("Close Warning", "There are unsaved changes\nDo you want to save them before closing?", saveChangesDialogAction, window)
			d.Show()
		} else {
			shouldCloseLock = false
			logData.WaitAndClose()
			window.Close()
		}
	}
}

func countChangedItems() int {
	count := gui.EditEntryListCache.Count()
	if hasDataChanges {
		count++
	}
	return count
}

func commitChangedItems() (int, error) {
	count := gui.EditEntryListCache.Commit(jsonData.GetDataRoot())
	jsonData.SetDateTime()
	c := jsonData.ToJson()
	fileData.SetContent([]byte(c))
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
	if logData.GetErr() == nil {
		preferences.PutBool(logActivePrefName, logData.IsLogging())
	}
	preferences.Save()
}

func setThemeById(varient string) {
	t := theme2.NewAppTheme(varient)
	fyne.CurrentApp().Settings().SetTheme(t)
	preferences.PutString(themeVarPrefName, varient)
}

func saveChangesDialogAction(option bool) {
	shouldCloseLock = false
	if !option {
		log("Quit without saving changes")
		logData.WaitAndClose()
		window.Close()
	}
}

func oneOrTheOther(one bool, s1, s2 string) string {
	if one {
		return s1
	}
	return s2
}
