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
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
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

	UID_POS_USER       = 0
	UID_POS_TYPE       = 1
	UID_POS_PWHINT     = 2
	UID_POS_ASSET_NAME = 2

	defaultScreenWidth  = 640
	defaultScreenHeight = 480

	fallbackPreferencesFile = "config.json"
	fallbackDataFile        = "data.json"
	uLine                   = "------------------------------------------------------------------------------------"
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
	currentSelPath     = parser.NewBarPath("")
	shouldCloseLock    = false
	hasDataChanges     = false
	releaseTheBeast    = make(chan int, 1)
	dataIsNotLoadedYet = true

	importFileFilter = []string{".csv", ".csvt"}

	dataFilePrefName          = parser.NewDotPath("file.datafile")
	backupFilePrefName        = parser.NewDotPath("file.backupfile")
	copyDialogTimePrefName    = parser.NewDotPath("dialog.copyTimeOutMS")
	linkDialogTimePrefName    = parser.NewDotPath("dialog.linkTimeOutMS")
	saveDialogTimePrefName    = parser.NewDotPath("dialog.saveTimeOutMS")
	errorDialogTimePrefName   = parser.NewDotPath("dialog.errorTimeOutMS")
	getUrlPrefName            = parser.NewDotPath("file.getDataUrl")
	postUrlPrefName           = parser.NewDotPath("file.postDataUrl")
	importPathPrefName        = parser.NewDotPath("import.path")
	importFilterPrefName      = parser.NewDotPath("import.filter")
	importCsvSkipHPrefName    = parser.NewDotPath("import.csvSkipHeader")
	importCsvDateFmtPrefName  = parser.NewDotPath("import.csvDateFormat")
	importCsvColNamesPrefName = parser.NewDotPath("import.csvColumns")
	themeVarPrefName          = parser.NewDotPath("theme")
	logFileNamePrefName       = parser.NewDotPath("log.fileName")
	logActivePrefName         = parser.NewDotPath("log.active")
	logPrefixPrefName         = parser.NewDotPath("log.prefix")
	screenWidthPrefName       = parser.NewDotPath("screen.width")
	screenHeightPrefName      = parser.NewDotPath("screen.height")
	screenFullPrefName        = parser.NewDotPath("screen.fullScreen")
	screenSplitPrefName       = parser.NewDotPath("screen.split")
	searchLastGoodPrefName    = parser.NewDotPath("search.lastGoodList")
	searchCasePrefName        = parser.NewDotPath("search.case")
)

func abortWithUsage(message string) {
	fmt.Println(uLine)
	fmt.Println(message)
	fmt.Println(uLine)
	fmt.Printf("  Usage: %s <configfile>\n", os.Args[0])
	fmt.Println("  Where: <configfile> is a json file. E.g. config.json")
	fmt.Println("  Minimum content for this file is:\n    {\n      \"file\": {\n           \"datafile\":\"myDataFile.data\"\n      }\n    }")
	fmt.Println("    Where \"myDataFile.data\" is the name of the required data file")
	fmt.Println("    This file will be updated by this application")
	fmt.Println("  To create a NEW minimal \"myDataFile.data\" use the 'create' option.")
	fmt.Println("  For example:")
	fmt.Printf("     %s <configfile> create\n", os.Args[0])
	fmt.Printf("  This will create the file defined in the <configfile> '%s' value.\n", dataFilePrefName.String())
	fmt.Println(uLine)
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

	primaryFileName := p.GetStringWithFallback(dataFilePrefName, fallbackDataFile)
	backupFileName := p.GetStringWithFallback(backupFilePrefName, "")
	getDataUrl := p.GetStringWithFallback(getUrlPrefName, "")
	postDataUrl := p.GetStringWithFallback(postUrlPrefName, "")
	//
	// For extended command line options. Dont use logData use std out!
	//
	if len(os.Args) > 2 {
		createFile := oneOrTheOther(postDataUrl == "", primaryFileName, postDataUrl+"/"+primaryFileName)
		switch os.Args[2] {
		case "create":
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
				_, err = parser.PostJsonBytes(fmt.Sprintf("%s/%s", postDataUrl, primaryFileName), data)
			} else {
				err = ioutil.WriteFile(primaryFileName, data, 0644)
			}
			if err != nil {
				fmt.Println(err.Error())
			} else {
				fmt.Printf("-> File %s has been created\n", createFile)
			}
		default:
			fmt.Println(uLine)
			fmt.Printf("-> The line you wanted was %s %s create\n", os.Args[0], os.Args[1])
			fmt.Printf("-> This will create a NEW minimal data file called '%s' as specified in '%s'\n", createFile, os.Args[1])
			fmt.Println(uLine)
			os.Exit(1)
		}
		os.Exit(0)
	}

	logData = gui.NewLogData(
		preferences.GetStringWithFallback(logFileNamePrefName, "enctest.log"),
		preferences.GetStringWithFallback(logPrefixPrefName, "INFO")+": ",
		preferences.GetBoolWithFallback(logActivePrefName, false))

	a := app.NewWithID("stuartdd.enctest")
	a.Settings().SetTheme(theme2.NewAppTheme(preferences.GetStringWithFallback(themeVarPrefName, "dark")))
	a.SetIcon(theme2.AppLogo())

	window = a.NewWindow(fmt.Sprintf("Data File: %s not loaded yet", primaryFileName))
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
	wp := gui.GetWelcomePage(*preferences, log)
	title := container.NewHBox()
	title.Objects = []fyne.CanvasObject{wp.CntlFunc(window, *wp, nil, preferences, statusDisplay, log)}
	contentRHS := container.NewMax()
	layoutRHS := container.NewBorder(title, container.NewWithoutLayout(), nil, nil, contentRHS)
	buttonBar := makeButtonBar()
	searchWindow = gui.NewSearchDataWindow(selectTreeElement)
	lib.ClearUserAccountFilter()

	/*
		function called when a selection is made in the LHS tree.
		This updates the contentRHS which is the RHS page for editing data
	*/
	setPageRHSFunc := func(detailPage gui.DetailPage) {
		currentSelPath = detailPage.SelectedPath
		if searchWindow != nil {
			go searchWindow.Select(currentSelPath)
		}
		log(fmt.Sprintf("Page User:'%s' Uid:'%s'", currentSelPath.StringFirst(), currentSelPath))
		window.SetTitle(fmt.Sprintf("Data File: [%s]. Current User: %s", fileData.GetFileName(), currentSelPath.StringFirst()))
		/*
			Create the menus
		*/
		window.SetMainMenu(makeMenus())
		navTreeLHS.OpenBranch(currentSelPath.String())
		title.Objects = []fyne.CanvasObject{detailPage.CntlFunc(window, detailPage, controlActionFunction, preferences, statusDisplay, log)}
		title.Refresh()
		contentRHS.Objects = []fyne.CanvasObject{detailPage.ViewFunc(window, detailPage, controlActionFunction, preferences, statusDisplay, log)}
		contentRHS.Refresh()
	}

	/*
		Thread keeps running in background
		To Trigger it:
			set primaryFileName = filename
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
				fd, err := lib.NewFileData(primaryFileName, backupFileName, getDataUrl, postDataUrl)
				if err != nil {
					abortWithUsage(fmt.Sprintf("Failed to load data file %s. Error: %s", primaryFileName, err.Error()))
				}
				if getDataUrl != "" {
					log(fmt.Sprintf("Remote File:'%s/%s'", getDataUrl, primaryFileName))
				} else {
					log(fmt.Sprintf("Local File:'%s'", primaryFileName))
				}
				/*
					While file is ENCRYPTED
						Get PW and decrypt
						if Decryption is cancelled the application exits
				*/
				message := ""
				for fd.RequiresDecryption() {
					log(fmt.Sprintf("Requires Decryption:'%s'", primaryFileName))
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
					abortWithUsage(fmt.Sprintf("ERROR: Cannot process data in file '%s'.\n%s", primaryFileName, err))
				}
				fileData = fd
				jsonData = dr
				dataIsNotLoadedYet = false
				log(fmt.Sprintf("Data Parsed OK: File:'%s' DateTime:'%s'", primaryFileName, jsonData.GetTimeStampString()))
				// Follow on action to rebuild the Tree and re-display it
				futureReleaseTheBeast(0, MAIN_THREAD_RELOAD_TREE)
			case MAIN_THREAD_RELOAD_TREE:
				// Re-build the main tree view.
				// Select the root of current user if defined.
				// Init the devider (split)
				// Populate the window and we are done!
				navTreeLHS = makeNavTree(setPageRHSFunc)
				lib.InitUserAssetsCache(jsonData)
				selectTreeElement("MAIN_THREAD_RELOAD_TREE", currentSelPath)
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
				log(fmt.Sprintf("Re-display RHS. Sel:'%s'", currentSelPath))
				lib.InitUserAssetsCache(jsonData)
				t := gui.GetDetailPage(currentSelPath, jsonData.GetDataRoot(), *preferences, log)
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

func timedError(message string) {
	timedNotification(10000, "Error", message)
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
		m, err := lib.FindNodeForUserDataPath(jsonData.GetDataRoot(), currentSelPath)
		if err != nil {
			log(fmt.Sprintf("Data for uid [%s] not found. %s", currentSelPath, err.Error()))
		}
		if m != nil {
			log(fmt.Sprintf("uid:'%s'. Json:%s", currentSelPath, m.JsonValueIndented(4)))
		} else {
			log(fmt.Sprintf("Data for uid [%s] returned null", currentSelPath))
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
	hintName := preferences.GetStringWithFallback(gui.DataHintIsCalledPrefName, "Hint")
	assetName := preferences.GetStringWithFallback(gui.DataAssetIsCalledPrefName, "Asset")

	user := currentSelPath.StringFirst()
	selTypeItem := currentSelPath.StringAt(UID_POS_PWHINT)
	selType := currentSelPath.StringAt(UID_POS_TYPE)
	newItem := fyne.NewMenu("New")
	switch selType {
	case lib.IdHints:
		newItem.Items = append(newItem.Items, fyne.NewMenuItem(fmt.Sprintf("%s Item for '%s'", hintName, user), addNewHintItem))
		if selTypeItem != "" {
			newItem.Items = append(newItem.Items, fyne.NewMenuItem(fmt.Sprintf("Clone '%s'", selTypeItem), cloneHint))
			newItem.Items = append(newItem.Items, fyne.NewMenuItem(fmt.Sprintf("Clone Full '%s'", selTypeItem), cloneHintFull))
		}
	case lib.IdAssets:
		newItem.Items = append(newItem.Items, fyne.NewMenuItem(fmt.Sprintf("%s for '%s'", assetName, user), addNewAsset))
		if selTypeItem != "" {
			newItem.Items = append(newItem.Items, fyne.NewMenuItem(fmt.Sprintf("%s Item for '%s'", assetName, selTypeItem), addNewAssetItem))
			newItem.Items = append(newItem.Items, fyne.NewMenuItem(fmt.Sprintf("%s Transaction for '%s'", assetName, selTypeItem), addTransaction))
		}
	default:
		newItem.Items = append(newItem.Items, fyne.NewMenuItem(fmt.Sprintf("%s for '%s'", hintName, user), addNewHint))
		newItem.Items = append(newItem.Items, fyne.NewMenuItem(fmt.Sprintf("%s for '%s'", assetName, user), addNewAsset))
		newItem.Items = append(newItem.Items, fyne.NewMenuItem("New User", addNewUser))
	}

	var themeMenuItem *fyne.MenuItem
	if preferences.GetStringWithFallback(themeVarPrefName, "dark") == "dark" {
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
			_, _, title := gui.GetDetailTypeGroupTitle(parser.NewBarPath(uid), *preferences)
			obj.(*widget.Label).SetText(title)
		},
		OnSelected: func(selectedPathString string) {
			log(fmt.Sprintf("On Select:'%s'", selectedPathString))
			t := gui.GetDetailPage(parser.NewBarPath(selectedPathString), jsonData.GetDataRoot(), *preferences, log)
			setPage(*t)
		},
	}
}

/**
Section below the tree with Search details and Light and Dark theme buttons
*/
func makeSearchLHS(setPage func(detailPage gui.DetailPage)) fyne.CanvasObject {
	x := preferences.GetDropDownList(searchLastGoodPrefName)
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

func dataPreferencesChanged(path *parser.Path, value, filter string) {
	futureReleaseTheBeast(100, MAIN_THREAD_RESELECT)
}

/**
Select a tree element.
We need to open the parent branches or we will never see the selected element
*/
func selectTreeElement(desc string, uid *parser.Path) {
	if uid.IsEmpty() {
		user := jsonData.GetFirstUserName()
		navTreeLHS.OpenBranch(user)
		navTreeLHS.Select(user)
		return
	}
	n, err := parser.Find(jsonData.GetUserRoot(), uid)
	if err != nil || !n.IsContainer() {
		uid = uid.PathParent()
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
		// Return the path after the DataMapRootName.
		currentSelPath = lib.GetPathAfterDataRoot(dataPath)
		log(fmt.Sprintf("dataMapUpdated OK. Desc:'%s' DataPath:'%s'. Derived currentSelPath:'%s'", desc, dataPath, currentSelPath))
		hasDataChanges = true
	} else {
		log(fmt.Sprintf("dataMapUpdated Error. Desc:'%s' DataPath:'%s', Err:'%s'", desc, dataPath, err.Error()))
	}
}

/**
This is called when a button is pressed of the RH page
*/
func controlActionFunction(action string, dataPath *parser.Path, extra string) {
	log(fmt.Sprintf("Action:%s. Path:'%s'. Extra:'%s'", action, dataPath, extra))
	switch action {
	case gui.ACTION_REMOVE_CLEAN:
		removeAction(dataPath, -1)
	case gui.ACTION_REMOVE:
		removeAction(dataPath, 1)
	case gui.ACTION_RENAME:
		renameAction(dataPath, extra)
	case gui.ACTION_LINK:
		linkAction(dataPath, extra)
	case gui.ACTION_ADD_HINT:
		addNewHint()
	case gui.ACTION_ADD_ASSET:
		addNewAsset()
	case gui.ACTION_ADD_HINT_ITEM:
		addNewHintItem()
	case gui.ACTION_ADD_ASSET_ITEM:
		addNewAssetItem()
	case gui.ACTION_UPDATE_TRANSACTION:
		updateTransactionValue(dataPath, extra)
	case gui.ACTION_IMPORT_TRANSACTION:
		importTransactions(dataPath, extra)
	case gui.ACTION_ADD_TRANSACTION:
		addTransactionValue(dataPath, extra)
	case gui.ACTION_CLONE_FULL:
		cloneHintFull()
	case gui.ACTION_CLONE:
		cloneHint()
	case gui.ACTION_FILTER:
		futureReleaseTheBeast(300, MAIN_THREAD_RESELECT)
	case gui.ACTION_UPDATED:
		futureReleaseTheBeast(100, MAIN_THREAD_RE_MENU)
	case gui.ACTION_COPIED:
		timedNotification(preferences.GetInt64WithFallback(copyDialogTimePrefName, 1500), "Copied item text to clipboard", dataPath.String())
	case gui.ACTION_ERROR_DIALOG:
		timedNotification(preferences.GetInt64WithFallback(errorDialogTimePrefName, 2000), fmt.Sprintf("Error for data at: %s", dataPath.String()), extra)
	case gui.ACTION_WARN_DIALOG:
		timedNotification(preferences.GetInt64WithFallback(errorDialogTimePrefName, 2000), "Warning", extra)
	case gui.ACTION_LOG:
		log(gui.LogCleanString(extra, 100))
	}
}

func cloneHint() {
	n := preferences.GetStringWithFallback(gui.DataHintIsCalledPrefName, "Hint")
	addNewEntity(n+" for ", n, ADD_TYPE_HINT_CLONE, false)
}

func cloneHintFull() {
	n := preferences.GetStringWithFallback(gui.DataHintIsCalledPrefName, "Hint")
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
	n := preferences.GetStringWithFallback(gui.DataHintIsCalledPrefName, "Hint")
	ch := currentSelPath.StringAt(UID_POS_USER)
	if ch == "" {
		logInformationDialog("Add New "+n, "A User needs to be selected")
	} else {
		addNewEntity(fmt.Sprintf("%s for %s", n, ch), n, ADD_TYPE_HINT, false)
	}
}

func addTransaction() {
	addTransactionValue(currentSelPath.StringAppend(lib.IdTransactions), currentSelPath.StringLast())
}

func addTransactionValue(dataPath *parser.Path, extra string) {
	d := gui.NewInputDataWindow(
		fmt.Sprintf("Add Transaction to account '%s'", extra),
		"Update the data and press OK",
		func() {}, // On Cancel
		func(m *gui.InputData) { // On OK
			dt, _ := lib.ParseDateString(m.GetString(lib.IdTxDate, lib.FormatDateTime(time.Now())))
			err := jsonData.AddTransaction(
				dataPath,
				dt,
				m.GetString(lib.IdTxRef, "ref"),
				m.GetFloat(lib.IdTxVal, 0.1),
				lib.TransactionTypeEnum(m.GetString(lib.IdTxType, string(lib.TX_TYPE_DEB))))
			if err != nil {
				timedError(err.Error())
			}
		})

	d.Add(lib.IdTxDate, "Date", lib.CurrentDateString(), func(s string) (string, error) {
		_, e := lib.ParseDateString(s)
		if e != nil {
			return "", fmt.Errorf("format:%s or %s", lib.DATE_TIME_FORMAT_TXN, lib.DATE_FORMAT_TXN)
		} else {
			return "", nil
		}
	})

	d.Add(lib.IdTxRef, "Reference", "ref", func(s string) (string, error) {
		if s == "" {
			return "", fmt.Errorf("cannot be empty")
		} else {
			return "", nil
		}
	})

	d.Add(lib.IdTxVal, "Amount", "0.1", func(s string) (string, error) {
		v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
		if v <= 0.0 {
			return "", fmt.Errorf("cannot be 0.0 or less")
		}
		if err != nil {
			return "", fmt.Errorf("is not a valid amount")
		}
		return "", nil
	})

	d.AddOptions(lib.IdTxType, lib.TX_TYPE_LIST_LABLES, string(lib.TX_TYPE_DEB), lib.TX_TYPE_LIST_OPTIONS, func(s string) (string, error) {
		return "", nil
	})

	d.Show(window)
	go d.Validate()
}

func importCSVTransactions(dataPath *parser.Path, fileName string) (int, error) {
	skipHeader := preferences.GetBoolWithFallback(importCsvSkipHPrefName, true)
	dateFmt := preferences.GetStringWithFallback(importCsvDateFmtPrefName, lib.TIME_FORMAT_CSV)
	dataMapList := preferences.GetStringListWithFallback(importCsvColNamesPrefName, nil)
	if dataMapList == nil {
		preferences.PutStringList(importCsvColNamesPrefName, lib.IMPORT_CSV_COLUM_NAMES, false)
		dataMapList = preferences.GetStringListWithFallback(importCsvColNamesPrefName, nil)
	}
	n, _ := jsonData.FindNodeForUserDataPath(dataPath)
	if n.IsContainer() {
		t := n.(parser.NodeC).GetNodeWithName(lib.IdTransactions)
		if t == nil {
			return 0, fmt.Errorf("data error: %s node missing for %s.\nPlease add an account transactions node", lib.IdTransactions, n.GetName())
		}
		return lib.ImportCsvData(t.(parser.NodeC), fileName, skipHeader, dateFmt, dataMapList)
	}
	return 0, fmt.Errorf("data error: %s node is not a container node", n.GetName())

}

func importTransactions(dataPath *parser.Path, extra string) {
	// t := preferences.GetStringForPathWithFallback(gui.DataTransIsCalledPrefName, "Transaction")
	_, err := parser.Find(jsonData.GetUserRoot(), dataPath)
	if err != nil {
		timedError(fmt.Sprintf("canot find %s", dataPath))
		return
	}
	fod := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
		if err == nil {
			if uc == nil {
				timedNotification(2000, "Warning", "No file selected")
			} else {
				p := uc.URI().Path()
				p = p[0 : len(p)-len(uc.URI().Name())]
				n := uc.URI().Name()
				if len(p) >= 2 {
					preferences.PutString(importPathPrefName, p)
				}
				if err == nil {
					count, err := importCSVTransactions(dataPath, uc.URI().Path())
					if err != nil {
						timedError(fmt.Sprintf("Failed to import CSV file %s\nError: %s", n, err))
					} else {
						s := fmt.Sprintf("%d transaction(s) imported from file %s", count, n)
						dataMapUpdated(s, dataPath, nil)
						timedNotification(5000, "Successful import", s)
					}
				} else {
					timedError(fmt.Sprintf("Failed to read file %s\nError: %s", uc.URI().Path(), err))
				}
			}
		}
	}, window)
	uri, err := storage.ListerForURI(storage.NewFileURI(preferences.GetStringWithFallback(importPathPrefName, "/")))
	if err != nil {
		uri, _ = storage.ListerForURI(storage.NewFileURI("/"))
	}
	fod.SetLocation(uri)
	fod.SetFilter(storage.NewExtensionFileFilter(preferences.GetStringListWithFallback(importFilterPrefName, importFileFilter)))
	fod.Resize(fyne.NewSize(window.Canvas().Size().Width*0.8, window.Canvas().Size().Height*0.8))
	fod.Show()
}

func updateTransactionValue(dataPath *parser.Path, extra string) {
	t := preferences.GetStringWithFallback(gui.DataTransIsCalledPrefName, "Transaction")
	data, err := parser.Find(jsonData.GetUserRoot(), dataPath)
	if err != nil {
		logInformationDialog("Error updating transaction value", err.Error())
		return
	}
	txd, txNode, err := lib.GetTransactionDataAndNodeForKey(data.(parser.NodeC), extra)
	if err != nil {
		logInformationDialog("Error updating transaction value", err.Error())
		return
	}
	deleteIfZero := txd.TxType() != lib.TX_TYPE_IV
	d := gui.NewInputDataWindow(
		fmt.Sprintf("Update '%s'", txd.Description()),
		fmt.Sprintf("Update %s data and press OK", t),
		func() {}, // On Cancel
		func(m *gui.InputData) { // On OK
			count := 0
			vs := m.Get(lib.IdTxVal).Value
			v, _ := strconv.ParseFloat(strings.TrimSpace(vs), 64)
			if v < 0.00001 && deleteIfZero {
				dil := dialog.NewConfirm(fmt.Sprintf("REMOVE: %s", t), fmt.Sprintf("%s\n\nAre you sure?", txd.Description()), func(b bool) {
					if b {
						data.(parser.NodeC).Remove(txNode)
						dataMapUpdated("Transaction removed", dataPath.PathParent(), nil)
					}
				}, window)
				dil.Show()
			} else {
				for _, v := range m.Data() {
					count = count + lib.UpdateNodeFromTranactionData(txNode, v.Id, v.Value)
				}
				if count > 0 {
					dataMapUpdated("Transaction updated", dataPath.PathParent(), nil)
				}
			}
		})

	d.Add(lib.IdTxRef, "Reference", txd.Ref(), func(s string) (string, error) {
		if s == "" {
			return "", fmt.Errorf("cannot be empty")
		} else {
			return "", nil
		}
	})

	d.Add(lib.IdTxVal, "Amount", txd.Val(), func(s string) (string, error) {
		v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
		if err != nil {
			return "", fmt.Errorf("is not a valid amount")
		}
		if v < 0.00001 && deleteIfZero {
			return "Zero value will REMOVE this transaction", nil
		}
		if v < 0.0 {
			return "", fmt.Errorf("cannot be less than 0.0")
		}
		return "", nil
	})

	if txd.TxType() != lib.TX_TYPE_IV {
		d.AddOptions(lib.IdTxType, lib.TX_TYPE_LIST_LABLES, string(txd.TxType()), lib.TX_TYPE_LIST_OPTIONS, func(s string) (string, error) {
			return "", nil
		})
	}

	d.Show(window)
	go d.Validate()
}

/**
Add a asset via addNewEntity
*/
func addNewAsset() {
	n := preferences.GetStringWithFallback(gui.DataAssetIsCalledPrefName, "Asset")
	ch := currentSelPath.StringAt(UID_POS_USER)
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
	n := preferences.GetStringWithFallback(gui.DataAssetIsCalledPrefName, "Asset")
	ch := currentSelPath.StringAt(UID_POS_USER)
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
	n := preferences.GetStringWithFallback(gui.DataHintIsCalledPrefName, "Hint")
	ch := currentSelPath.StringAt(UID_POS_USER)
	if ch == "" {
		logInformationDialog("Add New Item to "+n, "A User needs to be selected")
	} else {
		addNewEntity(fmt.Sprintf("%s Item for %s", n, ch), n, ADD_TYPE_HINT_ITEM, true)
	}
}

/**
Add an entity to the model.
Delegate to DataRoot for the logic. Call back on dataMapUpdated function if a change is made
*/
func addNewEntity(head string, name string, addType int, isAnnotated bool) {
	cu := currentSelPath.PathAt(UID_POS_USER)
	gui.NewModalEntryDialog(window, "Enter the name of the new "+head, "", isAnnotated, lib.NODE_TYPE_SL, func(accept bool, newName string, nt lib.NodeAnnotationEnum) {
		if accept {
			entityName, err := lib.ProcessEntityName(newName, nt)
			if err == nil {
				switch addType {
				case ADD_TYPE_USER:
					err = jsonData.AddUser(entityName)
				case ADD_TYPE_HINT:
					err = jsonData.AddHint(cu, entityName)
				case ADD_TYPE_ASSET:
					err = jsonData.AddAsset(cu, entityName)
				case ADD_TYPE_ASSET_ITEM:
					err = jsonData.AddSubItem(currentSelPath, entityName, "asset")
				case ADD_TYPE_HINT_CLONE:
					err = jsonData.CloneHint(currentSelPath, entityName, false)
				case ADD_TYPE_HINT_CLONE_FULL:
					err = jsonData.CloneHint(currentSelPath, entityName, true)
				case ADD_TYPE_HINT_ITEM:
					err = jsonData.AddSubItem(currentSelPath, entityName, "hint")
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
func removeAction(dataPath *parser.Path, min int) {
	log(fmt.Sprintf("removeAction Uid:'%s'", dataPath))
	_, removeName := lib.GetNodeAnnotationTypeAndName(dataPath.StringLast())
	dialog.NewConfirm("Remove entry", fmt.Sprintf("'%s'\nAre you sure?", removeName), func(ok bool) {
		if ok {
			err := jsonData.Remove(dataPath, min)
			if err != nil {
				logInformationDialog("Remove item error", err.Error())
			}
		}
	}, window).Show()
}

/**
Rename a node from the main data (model) and update the tree view
dataMapUpdated is called if a change is made to the model
*/
func renameAction(dataPath *parser.Path, extra string) {
	log(fmt.Sprintf("renameAction dataPath:'%s' Extra:'%s'", dataPath, extra))
	m, _ := jsonData.FindNodeForUserDataPath(dataPath)
	if m != nil {
		at, fromName := lib.GetNodeAnnotationTypeAndName(dataPath.StringLast())
		toName := ""
		isAnnotated := false
		if m.GetNodeType() == parser.NT_STRING {
			toName = fromName
			isAnnotated = true
		}
		gui.NewModalEntryDialog(window, fmt.Sprintf("Rename entry '%s' ", fromName), toName, isAnnotated, at, func(accept bool, toName string, nt lib.NodeAnnotationEnum) {
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

/*
The search button has been pressed
*/
func search(searchFor string) {
	if searchFor == "" {
		return
	}
	matchCase, _ := findCaseSensitive.Get()
	lib.ClearUserAccountFilter()

	// Do the search. The map contains the returned search entries
	// This ensures no duplicates are displayed
	// The map key is the human readable results e.g. 'User [Hint] app: noteName'
	// The values are paths within the model! user.pwHints.app.noteName
	if searchWindow != nil {
		searchWindow.Close()
	}
	searchWindow = gui.NewSearchDataWindow(selectTreeElement)
	searchWindow.Reset()
	jsonData.Search(func(trail *parser.Trail) {
		user := searchStringNodeName(trail.GetNodeAt(0))
		kind := searchStringNodeName(trail.GetNodeAt(1))
		switch kind {
		case lib.IdAssets:
			s := searchStringTrailFromTo(trail, 2, 4)
			p := trail.GetPath(0, 3, "|")
			n := preferences.GetStringWithFallback(gui.DataAssetIsCalledPrefName, "Asset")
			t3 := trail.GetNodeAt(3)
			if t3 != nil {
				if searchStringNodeName(t3) == lib.IdTransactions {
					lib.SetUserAccountFilter(user, searchStringNodeName(trail.GetNodeAt(2)), searchFor)
					t5 := trail.GetLast()
					if t5 != nil {
						searchWindow.Add(fmt.Sprintf("%s %s [ %s ] In Transaction [ %s ]", user, n, s, t5.String()), p)
					} else {
						searchWindow.Add(fmt.Sprintf("%s %s [ %s ] In Transaction Data", user, n, s), p)
					}
				} else {
					searchWindow.Add(fmt.Sprintf("%s %s [ %s ] In Field [ %s ]", user, n, s, searchStringNodeName(t3)), p)
				}
			}
		default:
			s := searchStringTrailFromTo(trail, 2)
			p := trail.GetPath(0, 3, "|")
			n := preferences.GetStringWithFallback(gui.DataHintIsCalledPrefName, "Hint")
			if trail.Len() > 3 {
				searchWindow.Add(fmt.Sprintf("%s %s [ %s ] In Field [ %s ]", user, n, s, searchStringNodeName(trail.GetNodeAt(3))), p)
			} else {
				searchWindow.Add(fmt.Sprintf("%s %s [ %s ]", user, n, s), p)
			}
		}
	}, searchFor, !matchCase)

	// Use the sorted keys to populate the result window
	if searchWindow.Len() > 0 {
		preferences.AddToDropDownList(searchLastGoodPrefName, searchFor, 10)
		searchWindow.Show(800, 350, searchFor)
	} else {
		logInformationDialog("Search results", fmt.Sprintf("Nothing found for search '%s'", searchFor))
	}
	defer futureReleaseTheBeast(200, MAIN_THREAD_RELOAD_TREE)
}

func searchStringTrailFromTo(trail *parser.Trail, ind ...int) string {
	var st strings.Builder
	for _, v := range ind {
		if v >= 0 && v < trail.Len() {
			n := searchStringNodeName(trail.GetNodeAt(uint(v)))
			if n != "" {
				st.WriteString(n)
				st.WriteRune('.')
			}
		}
	}
	return strings.Trim(st.String(), ".")
}

func searchStringNodeName(node parser.NodeI) string {
	if node == nil {
		return "nil"
	}
	_, n := lib.GetNodeAnnotationTypeAndName(node.GetName())
	return n
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
			abortWithUsage(fmt.Sprintf("Failed to decrypt data file '%s'\nPassword was not provided", fd.GetFileName()))
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
