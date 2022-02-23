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
	"github.com/stuartdd2/JsonParser4go/parser"
	"stuartdd.com/gui"
	"stuartdd.com/lib"
	"stuartdd.com/pref"
	"stuartdd.com/theme2"
)

const (
	SAVE_AS_IS = iota
	SAVE_ENCRYPTED
	SAVE_UN_ENCRYPTED

	MAIN_THREAD_LOAD = iota
	MAIN_THREAD_RELOAD_TREE
	MAIN_THREAD_RESELECT
	MAIN_THREAD_RE_MENU

	ADD_TYPE_USER = iota
	ADD_TYPE_HINT
	ADD_TYPE_ASSET
	ADD_TYPE_ASSET_ITEM
	ADD_TYPE_HINT_CLONE
	ADD_TYPE_HINT_CLONE_FULL
	ADD_TYPE_HINT_ITEM
	ADD_TYPE_NOTE_ITEM

	UID_POS_USER   = 0
	UID_POS_PWHINT = 2

	defaultScreenWidth  = 640
	defaultScreenHeight = 480

	fallbackPreferencesFile = "config.json"
	fallbackDataFile        = "data.json"

	dataFilePrefName       = "file.datafile"
	copyDialogTimePrefName = "dialog.copyTimeOutMS"
	linkDialogTimePrefName = "dialog.linkTimeOutMS"
	saveDialogTimePrefName = "dialog.saveTimeOutMS"
	getUrlPrefName         = "file.getDataUrl"
	postUrlPrefName        = "file.postDataUrl"
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
	saveShortcutButton       *gui.MyButton
	fullScreenShortcutButton *gui.MyButton
	editModeShortcutButton   *gui.MyButton
	timeStampLabel           *widget.Label
	statusDisplay            *gui.StatusDisplay
	splitContainer           *container.Split // So we can save the divider position to preferences.
	splitContainerOffset     float64          = -1
	splitContainerOffsetPref float64          = -1

	findCaseSensitive  = binding.NewBool()
	currentUid         = parser.NewBarPath("")
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

	window.SetMaster()

	statusDisplay = gui.NewStatusDisplay("Select an item from the list above", "Hint")
	wp := gui.GetWelcomePage(parser.NewBarPath(""), *preferences)
	title := container.NewHBox()
	title.Objects = []fyne.CanvasObject{wp.CntlFunc(window, *wp, nil, preferences, statusDisplay)}
	contentRHS := container.NewMax()
	layoutRHS := container.NewBorder(title, container.NewWithoutLayout(), nil, nil, contentRHS)
	buttonBar := makeButtonBar()
	searchWindow = gui.NewSearchDataWindow(closeSearchWindow, selectTreeElement)
	/*
		function called when a selection is made in the LHS tree.
		This updates the contentRHS which is the RHS page for editing data
	*/
	setPageRHSFunc := func(detailPage gui.DetailPage) {
		currentUid = detailPage.Uid
		if searchWindow != nil {
			go searchWindow.Select(currentUid)
		}
		log(fmt.Sprintf("Page User:'%s' Uid:'%s'", currentUid.StringFirst(), currentUid))
		window.SetTitle(fmt.Sprintf("Data File: [%s]. Current User: %s", fileData.GetFileName(), currentUid.StringFirst()))
		/*
			Create the menus
		*/
		window.SetMainMenu(makeMenus())
		navTreeLHS.OpenBranch(currentUid.String())
		title.Objects = []fyne.CanvasObject{detailPage.CntlFunc(window, detailPage, controlActionFunction, preferences, statusDisplay)}
		title.Refresh()
		contentRHS.Objects = []fyne.CanvasObject{detailPage.ViewFunc(window, detailPage, controlActionFunction, preferences, statusDisplay)}
		contentRHS.Refresh()
	}

	/*
		Thread keeps running in background
		To Trigger it:
			set loadThreadFileName = filename
			futureReleaseTheBeast(1000, MAIN_THREAD_LOAD)
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
				futureReleaseTheBeast(100, MAIN_THREAD_RELOAD_TREE)
			case MAIN_THREAD_RELOAD_TREE:
				// Re-build the main tree view.
				// Select the root of current user if defined.
				// Init the devider (split)
				// Populate the window and we are done!
				navTreeLHS = makeNavTree(setPageRHSFunc)
				lib.InitUserAssetsCache(jsonData.GetDataRoot())
				selectTreeElement("MAIN_THREAD_RELOAD_TREE", currentUid)
				if splitContainerOffset < 0 {
					splitContainerOffset = splitContainerOffsetPref
				} else {
					splitContainerOffset = splitContainer.Offset
				}
				splitContainer = container.NewHSplit(container.NewBorder(makeSearchLHS(setPageRHSFunc), nil, nil, nil, navTreeLHS), layoutRHS)
				splitContainer.SetOffset(splitContainerOffset)
				window.SetContent(container.NewBorder(buttonBar, statusDisplay.StatusContainer, nil, nil, splitContainer))
				futureReleaseTheBeast(0, MAIN_THREAD_RE_MENU)
			case MAIN_THREAD_RESELECT:
				log(fmt.Sprintf("Re-display RHS. Sel:'%s'", currentUid))
				lib.InitUserAssetsCache(jsonData.GetDataRoot())
				t := gui.GetDetailPage(currentUid, jsonData.GetDataRoot(), *preferences)
				setPageRHSFunc(*t)
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
	log("ShowAndRun")
	window.ShowAndRun()
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
			// MAIN_THREAD_LOAD has priority
			if status == MAIN_THREAD_LOAD {
				releaseTheBeast <- status
				return
			}
			//
			// Cannot do anything until MAIN_THREAD_LOAD is run and dataIsNotLoadedYet = false
			// But we cannot give in so we wait. But not forever!
			//
			if dataIsNotLoadedYet {
				count := 0
				for dataIsNotLoadedYet {
					time.Sleep(100 * time.Millisecond)
					count++
					if count > 100 {
						log(fmt.Sprintf("ReleaseTheBeast timed out waiting for data to load. Status:%d", status))
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
	saveShortcutButton.SetStatusMessage(fmt.Sprintf("Save changes to: %s", fileData.GetFileName()))
	if preferences.GetBoolWithFallback(screenFullPrefName, false) {
		fullScreenShortcutButton.SetText("Windowed")
		fullScreenShortcutButton.SetStatusMessage("Set display to Windowed")
	} else {
		fullScreenShortcutButton.SetText("Full Screen")
		fullScreenShortcutButton.SetStatusMessage("Set display to Full Screen")
	}
	if preferences.GetBoolWithFallback(gui.DataPresModePrefName, true) {
		editModeShortcutButton.SetText("Edit Data")
		editModeShortcutButton.SetStatusMessage("Allow user to edit the data")
	} else {
		editModeShortcutButton.SetText("Present Data")
		editModeShortcutButton.SetStatusMessage("Display the data in presentation mode")
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
		m, err := lib.GetUserDataForUid(jsonData.GetDataRoot(), currentUid)
		if err != nil {
			log(fmt.Sprintf("Data for uid [%s] not found. %s", currentUid, err.Error()))
		}
		if m != nil {
			log(fmt.Sprintf("uid:'%s'. Json:%s", currentUid, m.JsonValueIndented(4)))
		} else {
			log(fmt.Sprintf("Data for uid [%s] returned null", currentUid))
		}
	default:
		log(fmt.Sprintf("Log Action: '%s' unknown", action))
	}
}

func makeButtonBar() *fyne.Container {
	saveShortcutButton = gui.NewMyIconButton("Save", theme.DocumentSaveIcon(), func() {
		commitAndSaveData(SAVE_AS_IS, true)
	}, statusDisplay, "Save changes")
	fullScreenShortcutButton = gui.NewMyIconButton("FULL SCREEN", theme.ComputerIcon(), flipFullScreen, statusDisplay, "Display full screen or Show windowed")
	editModeShortcutButton = gui.NewMyIconButton("EDIT", theme.DocumentIcon(), flipPositionalData, statusDisplay, "Allow editing of the data")

	quit := gui.NewMyIconButton("EXIT", theme.LogoutIcon(), shouldClose, statusDisplay, "Exit the application")
	timeStampLabel = widget.NewLabel("  File Not loaded")
	return container.NewHBox(quit, saveShortcutButton, gui.Padding50, fullScreenShortcutButton, editModeShortcutButton, timeStampLabel)
}

func makeMenus() *fyne.MainMenu {
	hintName := preferences.GetStringForPathWithFallback(gui.DataHintIsCalledPrefName, "Hint")
	noteName := preferences.GetStringForPathWithFallback(gui.DataNoteIsCalledPrefName, "Note")
	assetName := preferences.GetStringForPathWithFallback(gui.DataAssetIsCalledPrefName, "Asset")
	hint := currentUid.StringAt(UID_POS_PWHINT)
	user := currentUid.StringAt(UID_POS_USER)

	n1 := fyne.NewMenuItem(fmt.Sprintf("%s for '%s'", noteName, user), addNewNoteItem)
	n3 := fyne.NewMenuItem(fmt.Sprintf("%s for '%s'", hintName, user), addNewHint)
	n4 := fyne.NewMenuItem(fmt.Sprintf("%s for '%s'", assetName, user), addNewAsset)
	n5 := fyne.NewMenuItem("User", addNewUser)
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
			n4, n5)
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
			type1, group1, title := gui.GetDetailTypeGroupTitle(parser.NewBarPath(uid), *preferences)
			log(fmt.Sprintf("On Update:'UID:%s Type:[%s] Group:%s Title:%s'", uid, lib.NodeAnnotationPrefixNames[type1], group1, title))
			obj.(*widget.Label).SetText(title)
		},
		OnSelected: func(uid string) {
			log(fmt.Sprintf("On Select:'%s'", uid))
			t := gui.GetDetailPage(parser.NewBarPath(uid), jsonData.GetDataRoot(), *preferences)
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
		gui.NewMyIconButton("", theme.SearchIcon(), func() { search(searchEntry.Text) }, statusDisplay, "Search for the given text"),
		widget.NewCheckWithData("Match Case", findCaseSensitive))
	c := container.New(
		layout.NewVBoxLayout(),
		c2,
		searchEntry)

	return container.NewVBox(widget.NewSeparator(), c)
}

func dataPreferencesChanged(path, value, filter string) {
	futureReleaseTheBeast(100, MAIN_THREAD_RESELECT)
}

/**
Select a tree element.
We need to open the parent branches or we will never see the selected element
*/
func selectTreeElement(desc string, uid *parser.Path) {
	if uid.IsEmpty() {
		uid = jsonData.GetUserPath(uid)
	} else {
		n, err := parser.Find(jsonData.GetUserRoot(), uid)
		if err != nil || !n.IsContainer() {
			uid = uid.PathParent()
		}
	}
	user := uid.StringFirst()
	parent := uid.PathParent().String()
	log(fmt.Sprintf("SelectTreeElement: Desc:'%s' User:'%s' Parent:'%s' Uid:'%s'", desc, user, parent, uid))
	navTreeLHS.OpenBranch(user)
	navTreeLHS.OpenBranch(parent)
	navTreeLHS.Select(uid.String())
	navTreeLHS.ScrollTo(uid.String())
}

/**
Called if there is a structural change in the model
*/
func dataMapUpdated(desc string, dataPath *parser.Path, err error) {
	defer func() {
		if r := recover(); r != nil {
			log(fmt.Sprintf("dataMapUpdated Recovered. Desc:'%s' DataPath:'%s', panic:'%s'", desc, dataPath, r))
		}
		futureReleaseTheBeast(100, MAIN_THREAD_RELOAD_TREE)
	}()
	if err == nil {
		currentUid = lib.GetUidPathFromDataPath(dataPath)
		log(fmt.Sprintf("dataMapUpdated OK. Desc:'%s' DataPath:'%s'. Derived currentUid:'%s'", desc, dataPath, currentUid))
		hasDataChanges = true
	} else {
		log(fmt.Sprintf("dataMapUpdated Error. Desc:'%s' DataPath:'%s', Err:'%s'", desc, dataPath, err.Error()))
	}
}

/**
This is called when a button is pressed of the RH page
*/
func controlActionFunction(action string, dataPath *parser.Path, extra string) {
	log(fmt.Sprintf("Action:%s path:'%s' extra:'%s'", action, dataPath, extra))
	switch action {
	case gui.ACTION_REMOVE:
		removeAction(dataPath)
	case gui.ACTION_RENAME:
		renameAction(dataPath, extra)
	case gui.ACTION_LINK:
		linkAction(dataPath, extra)
	case gui.ACTION_ADD_NOTE:
		addNewNoteItem()
	case gui.ACTION_ADD_HINT:
		addNewHint()
	case gui.ACTION_ADD_ASSET:
		addNewAsset()
	case gui.ACTION_ADD_HINT_ITEM:
		addNewHintItem()
	case gui.ACTION_ADD_ASSET_ITEM:
		addNewAssetItem()
	case gui.ACTION_CLONE_FULL:
		cloneHintFull()
	case gui.ACTION_CLONE:
		cloneHint()
	case gui.ACTION_UPDATED:
		futureReleaseTheBeast(100, MAIN_THREAD_RE_MENU)
	case gui.ACTION_COPIED:
		timedNotification(preferences.GetInt64WithFallback(copyDialogTimePrefName, 1500), "Copied item text to clipboard", dataPath.String())
	case gui.ACTION_ERROR_DIALOG:
		timedNotification(preferences.GetInt64WithFallback(copyDialogTimePrefName, 2000), fmt.Sprintf("Error for data at: %s", dataPath.String()), extra)
	case gui.ACTION_LOG:
		log(gui.LogCleanString(extra, 100))
	}
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
Add a user via addNewEntity
*/
func addNewUser() {
	addNewEntity("User", "User", ADD_TYPE_USER, false)
}

/**
Add a hint via addNewEntity
*/
func addNewHint() {
	n := preferences.GetStringForPathWithFallback(gui.DataHintIsCalledPrefName, "Hint")
	ch := currentUid.StringAt(UID_POS_USER)
	if ch == "" {
		logInformationDialog("Add New "+n, "A User needs to be selected")
	} else {
		addNewEntity(fmt.Sprintf("%s for %s", n, ch), n, ADD_TYPE_HINT, false)
	}
}

/**
Add a asset via addNewEntity
*/
func addNewAsset() {
	n := preferences.GetStringForPathWithFallback(gui.DataAssetIsCalledPrefName, "Asset")
	ch := currentUid.StringAt(UID_POS_USER)
	if ch == "" {
		logInformationDialog("Add New "+n, "A User needs to be selected")
	} else {
		addNewEntity(fmt.Sprintf("%s for %s", n, ch), n, ADD_TYPE_ASSET, false)
	}
}

/**
Selecting the menu to add an item to a hint
*/
func addNewAssetItem() {
	n := preferences.GetStringForPathWithFallback(gui.DataAssetIsCalledPrefName, "Asset")
	ch := currentUid.StringAt(UID_POS_USER)
	if ch == "" {
		logInformationDialog("Add New Atem to "+n, "A User needs to be selected")
	} else {
		addNewEntity(fmt.Sprintf("%s Item for %s", n, ch), n, ADD_TYPE_ASSET_ITEM, true)
	}
}

/**
Selecting the menu to add an item to a hint
*/
func addNewHintItem() {
	n := preferences.GetStringForPathWithFallback(gui.DataNoteIsCalledPrefName, "Hint")
	ch := currentUid.StringAt(UID_POS_USER)
	if ch == "" {
		logInformationDialog("Add New Item to "+n, "A User needs to be selected")
	} else {
		addNewEntity(fmt.Sprintf("%s Item for %s", n, ch), n, ADD_TYPE_HINT_ITEM, true)
	}
}

/**
Selecting the menu to add an item to the notes
*/
func addNewNoteItem() {
	n := preferences.GetStringForPathWithFallback(gui.DataNoteIsCalledPrefName, "Note")
	ch := currentUid.StringAt(UID_POS_USER)
	if ch == "" {
		logInformationDialog("Add New Item to "+n, "A User needs to be selected")
	} else {
		addNewEntity(fmt.Sprintf("%s Item for %s", n, ch), n, ADD_TYPE_NOTE_ITEM, true)
	}
}

/**
Add an entity to the model.
Delegate to DataRoot for the logic. Call back on dataMapUpdated function if a change is made
*/
func addNewEntity(head string, name string, addType int, isNote bool) {
	cu := currentUid.PathAt(UID_POS_USER)
	gui.NewModalEntryDialog(window, "Enter the name of the new "+head, "", isNote, lib.NOTE_TYPE_SL, func(accept bool, newName string, nt lib.NodeAnnotationEnum) {
		if accept {
			entityName, err := lib.ProcessEntityName(newName, nt)
			if err == nil {
				switch addType {
				case ADD_TYPE_USER:
					err = jsonData.AddUser(entityName)
				case ADD_TYPE_NOTE_ITEM:
					err = jsonData.AddNoteItem(cu, entityName)
				case ADD_TYPE_HINT:
					err = jsonData.AddHint(cu, entityName)
				case ADD_TYPE_ASSET:
					err = jsonData.AddAsset(cu, entityName)
				case ADD_TYPE_ASSET_ITEM:
					err = jsonData.AddSubItem(currentUid, entityName, "asset")
				case ADD_TYPE_HINT_CLONE:
					err = jsonData.CloneHint(currentUid, entityName, false)
				case ADD_TYPE_HINT_CLONE_FULL:
					err = jsonData.CloneHint(currentUid, entityName, true)
				case ADD_TYPE_HINT_ITEM:
					err = jsonData.AddSubItem(currentUid, entityName, "hint")
				}
			}
			if err != nil {
				logInformationDialog("Add New "+name, "Error: "+err.Error())
			}
		}
	})
}

/**
Remove a node from the main data (model) and update the tree view
dataMapUpdated id called if a change is made to the model
*/
func removeAction(dataPath *parser.Path) {
	log(fmt.Sprintf("removeAction Uid:'%s'", dataPath))
	_, removeName := lib.GetNodeAnnotationTypeAndName(dataPath.StringLast())
	dialog.NewConfirm("Remove entry", fmt.Sprintf("'%s'\nAre you sure?", removeName), func(ok bool) {
		if ok {
			err := jsonData.Remove(dataPath, 1)
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
func renameAction(dataPath *parser.Path, extra string) {
	log(fmt.Sprintf("renameAction dataPath:'%s' Extra:'%s'", dataPath, extra))
	m, _ := jsonData.GetUserDataForUid(dataPath)
	if m != nil {
		at, fromName := lib.GetNodeAnnotationTypeAndName(dataPath.StringLast())
		toName := ""
		isNote := false
		if jsonData.IsStringNode(m) {
			toName = fromName
			isNote = true
		}
		gui.NewModalEntryDialog(window, fmt.Sprintf("Rename entry '%s' ", fromName), toName, isNote, at, func(accept bool, toName string, nt lib.NodeAnnotationEnum) {
			if accept {
				s, err := lib.ProcessEntityName(toName, nt)
				if err != nil {
					logInformationDialog("Name validation error", err.Error())
				} else {
					err := jsonData.Rename(dataPath, s)
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
func linkAction(uid *parser.Path, urlStr string) {
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
	jsonData.Search(func(path *parser.Path, desc string) {
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
Get the password to decrypt the loaded data contained in FileData
*/
func getPasswordAndDecrypt(fd *lib.FileData, message string, fail func(string)) {
	if message == "" {
		message = "Enter the password to DECRYPT the file"
	}
	running := true
	gui.NewModalPasswordDialog(window, message, "", func(ok bool, value string, nt lib.NodeAnnotationEnum) {
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
			gui.NewModalPasswordDialog(window, "Enter the password to DECRYPT the file", "", func(ok bool, value string, nt lib.NodeAnnotationEnum) {
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
