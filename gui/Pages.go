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
package gui

import (
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/stuartdd2/JsonParser4go/parser"
	"stuartdd.com/lib"
	"stuartdd.com/pref"
	"stuartdd.com/theme2"
)

const (
	txDateColWidth  = 20
	txNumColWidth   = 12
	txColumMaxWidth = 150

	welcomeTitle = "Welcome"
	appDesc      = "Welcome to Valt"

	PATH_SEP = "|"

	ACTION_LOG                = "log"
	ACTION_COPIED             = "copied"
	ACTION_REMOVE_CLEAN       = "removeclean"
	ACTION_REMOVE             = "remove"
	ACTION_RENAME             = "rename"
	ACTION_CLONE              = "clone"
	ACTION_CLONE_FULL         = "clonefull"
	ACTION_LINK               = "link"
	ACTION_UPDATED            = "update"
	ACTION_FILTER             = "filter"
	ACTION_ADD_HINT           = "addhint"
	ACTION_ADD_ASSET          = "addasset"
	ACTION_ADD_ASSET_ITEM     = "addassetitem"
	ACTION_ADD_TRANSACTION    = "addtransaction"
	ACTION_UPDATE_TRANSACTION = "updatetransaction"
	ACTION_IMPORT_TRANSACTION = "importtransaction"
	ACTION_ADD_HINT_ITEM      = "addhintitem"
	ACTION_ERROR_DIALOG       = "errorDialog"
	ACTION_WARN_DIALOG        = "warningDialog"
)

var (
	preferedOrderReversed = []string{"notes", "positional", "post", "pre", "link", "userId"}
	EditEntryListCache    = NewEditEntryList()
	EditMode              = false
)

func NewModalEntryDialog(w fyne.Window, heading, txt string, isAnnotated bool, annotation lib.NodeAnnotationEnum, accept func(bool, string, lib.NodeAnnotationEnum)) (modal *widget.PopUp) {
	return runModalEntryPopup(w, heading, txt, false, isAnnotated, annotation, accept)
}

func NewModalPasswordDialog(w fyne.Window, heading, txt string, accept func(bool, string, lib.NodeAnnotationEnum)) (modal *widget.PopUp) {
	return runModalEntryPopup(w, heading, txt, true, false, lib.NODE_TYPE_SL, accept)
}

func GetWelcomePage(preferences pref.PrefData, log func(string)) *DetailPage {
	return NewDetailPage(parser.NewBarPath(""), welcomeTitle, "", "", welcomeScreen, welcomeControls, nil, preferences, log)
}

/*
	For a given path LENGTH return:
	2: 	The annotation type of the second path element
		The group the values of (idPwDetails, IdAssets) or second node without annotation)
		The value of the second node mapped via preferences to display format (pwHint-->Hint etc)
	3: 	The annotation type of the second path element
		The value of the second path element without annotation
		The value of the third path element without annotation
	Else The annotation type lib.NODE_TYPE_SL (this should only happen if len == 0)
		The path as a string,
		The path as a string,
*/
func GetDetailTypeGroupTitle(selectedPath *parser.Path, preferences pref.PrefData) (lib.NodeAnnotationEnum, string, string) {
	switch selectedPath.Len() {
	case 2:
		type1, group1 := lib.GetNodeAnnotationTypeAndName(selectedPath.StringAt(1))
		return type1, group1, lib.GetNameFromNameMap(group1, "")
	case 3:
		type1, group1 := lib.GetNodeAnnotationTypeAndName(selectedPath.StringAt(1))
		_, name2 := lib.GetNodeAnnotationTypeAndName(selectedPath.StringAt(2))
		return type1, group1, lib.GetNameFromNameMap(name2, "")
	default:
		return lib.NODE_TYPE_SL, selectedPath.String(), selectedPath.String()
	}
}

func GetDetailPage(selectedPath *parser.Path, dataMapRoot parser.NodeI, preferences pref.PrefData, log func(string)) *DetailPage {
	user0 := ""
	if selectedPath.Len() > 0 {
		user0 = selectedPath.StringAt(0)
	}
	_, group1, title := GetDetailTypeGroupTitle(selectedPath, preferences)
	switch selectedPath.Len() {
	case 1:
		return NewDetailPage(selectedPath, user0, "", selectedPath.String(), welcomeScreen, userControls, dataMapRoot, preferences, log)
	case 2:
		if group1 == lib.IdHints {
			return NewDetailPage(selectedPath, user0, group1, title+"s", welcomeScreen, hintControls, dataMapRoot, preferences, log)
		}
		if group1 == lib.IdAssets {
			return NewDetailPage(selectedPath, user0, group1, title+"s", assetScreen, assetSummaryControls, dataMapRoot, preferences, log)
		}
		return NewDetailPage(selectedPath, user0, group1, title, welcomeScreen, welcomeControls, dataMapRoot, preferences, log)
	case 3:
		_, name2 := lib.GetNodeAnnotationTypeAndName(selectedPath.StringAt(2))
		if group1 == lib.IdHints {
			return NewDetailPage(selectedPath, user0, group1, name2, detailsScreen, hintDetailsControls, dataMapRoot, preferences, log)
		}
		if group1 == lib.IdAssets {
			return NewDetailPage(selectedPath, user0, group1, name2, detailsScreen, assetDetailsControls, dataMapRoot, preferences, log)
		}
		return NewDetailPage(selectedPath, user0, group1, name2, welcomeScreen, welcomeControls, dataMapRoot, preferences, log)
	default:
		return NewDetailPage(selectedPath, user0, group1, title, welcomeScreen, welcomeControls, dataMapRoot, preferences, log)
	}
}

func loadImage(s string) (*canvas.Image, string) {
	if strings.TrimSpace(s) == "" {
		return nil, "Enter the location of the image. File or URL"
	}
	imageType, message := lib.CheckImageFile(s)
	switch imageType {
	case lib.IMAGE_NOT_SUPPORTED:
		return nil, fmt.Sprintf("File '%s' image type is not supported", s)
	case lib.IMAGE_NOT_FOUND:
		return nil, fmt.Sprintf("File image '%s' not found or invalid URL", s)
	case lib.IMAGE_GET_FAIL:
		return nil, message
	case lib.IMAGE_FILE_FOUND:
		image := canvas.NewImageFromFile(s)
		return image, ""
	case lib.IMAGE_URL:
		uri, err := storage.ParseURI(s)
		if err != nil {
			return nil, fmt.Sprintf("File url '%s' is invalid: %s", s, err.Error())
		}
		image := canvas.NewImageFromURI(uri)
		return image, ""
	}
	return nil, fmt.Sprintf("file image '%s' not found or invalid URL", s)
}

func positional(s string) fyne.CanvasObject {
	return NewPositional(s, 17, theme2.ColorForName(theme.ColorNameForeground), theme2.ColorForName(theme.ColorNameButton))
}

func welcomeControls(_ fyne.Window, details DetailPage, actionFunc func(string, *parser.Path, string), pref *pref.PrefData, statusDisplay *StatusDisplay, log func(string)) fyne.CanvasObject {
	cObj := make([]fyne.CanvasObject, 0)
	cObj = append(cObj, widget.NewLabel(details.Heading))
	return container.NewHBox(cObj...)
}

func userControls(_ fyne.Window, details DetailPage, actionFunc func(string, *parser.Path, string), pref *pref.PrefData, statusDisplay *StatusDisplay, log func(string)) fyne.CanvasObject {
	cObj := make([]fyne.CanvasObject, 0)
	cObj = append(cObj, NewMyIconButton("", theme.DeleteIcon(), func(a, b string) {
		actionFunc(ACTION_REMOVE, details.SelectedPath, "")
	}, "", "", statusDisplay, fmt.Sprintf("Delete: - '%s'", details.Title)))

	cObj = append(cObj, NewMyIconButton("", theme2.EditIcon(), func(a, b string) {
		actionFunc(ACTION_RENAME, details.SelectedPath, "")
	}, "", "", statusDisplay, fmt.Sprintf("Rename: - '%s'", details.Title)))

	cObj = append(cObj, widget.NewLabel(details.Heading))
	return container.NewHBox(cObj...)
}

func welcomeScreen(_ fyne.Window, details DetailPage, actionFunc func(string, *parser.Path, string), pref *pref.PrefData, statusDisplay *StatusDisplay, log func(string)) fyne.CanvasObject {
	logo := canvas.NewImageFromFile("background.png")
	logo.FillMode = canvas.ImageFillContain
	logo.SetMinSize(fyne.NewSize(228, 167))

	return container.NewVBox(
		widget.NewSeparator(),
		container.NewCenter(container.NewVBox(
			widget.NewLabelWithStyle(appDesc, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			logo,
			container.NewCenter(
				container.NewHBox(
					widget.NewHyperlink("fyne.io", parseURL("https://fyne.io/")),
					widget.NewLabel("-"),
					widget.NewHyperlink("SDD", parseURL("https://github.com/stuartdd")),
					widget.NewLabel("-"),
					widget.NewHyperlink("go", parseURL("https://golang.org/")),
				),
			),
		)))
}

func assetDetailsControls(_ fyne.Window, details DetailPage, actionFunc func(string, *parser.Path, string), pref *pref.PrefData, statusDisplay *StatusDisplay, log func(string)) fyne.CanvasObject {
	n := lib.GetNameFromNameMap(lib.IdAssets, "Asset")
	head := fmt.Sprintf("%s: %s - %s", details.User, n, details.Title)
	cObj := make([]fyne.CanvasObject, 0)
	cObj = append(cObj, NewMyIconButton("", theme.DeleteIcon(), func(a, b string) {
		actionFunc(ACTION_REMOVE_CLEAN, details.SelectedPath, "")
	}, "", "", statusDisplay, fmt.Sprintf("Delete: - '%s'", details.Title)))

	cObj = append(cObj, NewMyIconButton("", theme2.EditIcon(), func(a, b string) {
		actionFunc(ACTION_RENAME, details.SelectedPath, "")
	}, "", "", statusDisplay, fmt.Sprintf("Rename: - '%s'", details.Title)))

	cObj = append(cObj, NewMyIconButton("New", theme.ContentAddIcon(), func(a, d string) {
		actionFunc(ACTION_ADD_ASSET_ITEM, details.SelectedPath, "")
	}, "", "", statusDisplay, fmt.Sprintf("Add new Item to %s: %s", n, details.Title)))

	cObj = append(cObj, widget.NewLabel(head))
	return container.NewHBox(cObj...)
}

func assetSummaryControls(_ fyne.Window, details DetailPage, actionFunc func(string, *parser.Path, string), pref *pref.PrefData, statusDisplay *StatusDisplay, log func(string)) fyne.CanvasObject {
	n := lib.GetNameFromNameMap(lib.IdAssets, "Asset")
	head := fmt.Sprintf("%s Summary for user %s", n, details.User)
	cObj := make([]fyne.CanvasObject, 0)
	cObj = append(cObj, NewMyIconButton("New", theme.ContentAddIcon(), func(a, b string) {
		actionFunc(ACTION_ADD_ASSET, details.SelectedPath, details.Title)
	}, "", "", statusDisplay, fmt.Sprintf("Add new '%s' for user %s", n, details.User)))
	cObj = append(cObj, widget.NewLabel(head))
	return container.NewHBox(cObj...)
}

func hintControls(_ fyne.Window, details DetailPage, actionFunc func(string, *parser.Path, string), pref *pref.PrefData, statusDisplay *StatusDisplay, log func(string)) fyne.CanvasObject {
	n := lib.GetNameFromNameMap(lib.IdHints, "Hint")
	cObj := make([]fyne.CanvasObject, 0)
	cObj = append(cObj, NewMyIconButton("New", theme.ContentAddIcon(), func(a, d string) {
		actionFunc(ACTION_ADD_HINT, details.SelectedPath, "")
	}, "", "", statusDisplay, fmt.Sprintf("Add new '%s' to user: %s", n, details.User)))
	cObj = append(cObj, widget.NewLabel(details.Heading))
	return container.NewHBox(cObj...)
}

func getTransactionalCanvasObjects(actionFunc func(string, *parser.Path, string), cObj []fyne.CanvasObject, accData *lib.AccountData, pref *pref.PrefData, statusDisplay *StatusDisplay, log func(string)) []fyne.CanvasObject {
	transAreCalled := lib.GetNameFromNameMap(lib.IdTxTransactions, "Transaction")
	txList := accData.Transactions
	refMax := 10
	for _, tx := range txList {
		if refMax < len(tx.Ref()) {
			refMax = len(tx.Ref())
		}
	}
	refMax++
	hbTop := container.NewHBox()
	if EditMode {
		imp := NewMyIconButton("", theme.StorageIcon(), func(a, b string) {
			actionFunc(ACTION_IMPORT_TRANSACTION, accData.Path.PathParent(), accData.AccountName)
		}, "", "", statusDisplay, "Import from CSV file")
		hbTop.Add(imp)
	}
	filter := lib.GetUserAccountFilter(accData.User, accData.AccountName)
	filterEntry := widget.NewEntry()
	filterEntry.SetText(filter)
	filterEntry.OnChanged = func(s string) {
		lib.SetUserAccountFilter(accData.User, accData.AccountName, s)
		actionFunc(ACTION_FILTER, accData.Path.PathParent(), s)
	}
	hbTop.Add(widget.NewLabel("Filter:"))
	hbTop.Add(container.New(NewFixedWLayout(150), filterEntry))
	hbTop.Add(widget.NewLabel(fmt.Sprintf(" %d %s(s). Current balance %0.2f", len(accData.Transactions), transAreCalled, accData.ClosingValue)))
	cObj = append(cObj, hbTop)
	hb := container.NewHBox()
	if EditMode {
		add := NewMyIconButton("", theme.ContentAddIcon(), func(a, d string) {
			actionFunc(ACTION_ADD_TRANSACTION, &accData.Path, accData.AccountName)
		}, "", "", statusDisplay, fmt.Sprintf("Add a '%s' to '%s'", transAreCalled, accData.AccountName))
		hb.Add(add)
	}

	ts := fyne.TextStyle{Bold: true, Italic: false, Monospace: true, Symbol: false, TabWidth: 2}

	lb := lib.NewLine(txColumMaxWidth)
	lb.Apply("Date Time:", txDateColWidth)
	lb.ApplyRev("Reference:", refMax)
	lb.ApplyRev("In:", txNumColWidth)
	lb.ApplyRev("Out:", txNumColWidth)
	lb.ApplyRev("Balance:", txNumColWidth)
	hb.Add(widget.NewLabelWithStyle(lb.String(), fyne.TextAlignLeading, ts))
	cObj = append(cObj, hb)
	for _, tx := range txList {
		s := tx.Key()
		hb := container.NewHBox()
		if EditMode {
			rename := NewMyIconButton("", theme2.EditIcon(), func(a, b string) {
				actionFunc(ACTION_UPDATE_TRANSACTION, &accData.Path, s)
			}, "", "", statusDisplay, fmt.Sprintf("Upate '%s'", tx.Ref()))
			hb.Add(rename)
		}
		switch tx.TxType() {
		case lib.TX_TYPE_IV:
			lb.Clear()
			lb.Apply("---- -- -- -- -- --", txDateColWidth)
			lb.ApplyRev(tx.Ref(), refMax)
			lb.Apply("", txNumColWidth)
			lb.Apply("", txNumColWidth)
			lb.ApplyRev(fmt.Sprintf("%9.2f", tx.LineValue()), txNumColWidth)
			hb.Add(widget.NewLabelWithStyle(lb.String(), fyne.TextAlignLeading, ts))
			cObj = append(cObj, hb)
		case lib.TX_TYPE_CRE:
			lb.Clear()
			lb.Apply(tx.DateTime(), txDateColWidth)
			lb.ApplyRev(tx.Ref(), refMax)
			lb.ApplyRev(fmt.Sprintf("%9.2f", tx.Value()), txNumColWidth)
			lb.Apply("", txNumColWidth)
			lb.ApplyRev(fmt.Sprintf("%9.2f", tx.LineValue()), txNumColWidth)
			hb.Add(widget.NewLabelWithStyle(lb.String(), fyne.TextAlignLeading, ts))
			if lb.Contains(filter) {
				cObj = append(cObj, hb)
			}
		case lib.TX_TYPE_DEB:
			lb.Clear()
			lb.Apply(tx.DateTime(), txDateColWidth)
			lb.ApplyRev(tx.Ref(), refMax)
			lb.Apply("", txNumColWidth)
			lb.ApplyRev(fmt.Sprintf("%9.2f", tx.Value()), txNumColWidth)
			lb.ApplyRev(fmt.Sprintf("%9.2f", tx.LineValue()), txNumColWidth)
			hb.Add(widget.NewLabelWithStyle(lb.String(), fyne.TextAlignLeading, ts))
			if lb.Contains(filter) {
				cObj = append(cObj, hb)
			}
		default:
			lb.Clear()
			lb.Apply("ERR", 3)
			hb.Add(widget.NewLabelWithStyle(lb.String(), fyne.TextAlignLeading, ts))
		}
	}
	return cObj
}

func assetScreen(w fyne.Window, details DetailPage, actionFunc func(string, *parser.Path, string), pref *pref.PrefData, statusDisplay *StatusDisplay, log func(string)) fyne.CanvasObject {
	cObj := make([]fyne.CanvasObject, 0)
	cObj = append(cObj, widget.NewSeparator())
	accountData, err := lib.FindAllUserAccounts(details.User)
	tot := 0.0
	if err == nil {
		refMax := 10
		for _, asset := range accountData {
			if refMax < len(asset.AccountName) {
				refMax = len(asset.AccountName)
			}
		}
		for _, asset := range accountData {
			hb := container.NewHBox()
			lt := asset.LatestTransaction()
			if lt != nil {
				hb.Add(NewStringFieldRight(lt.DateTime(), 20))
			} else {
				hb.Add(NewStringFieldRight("No trans data!", 20))
			}
			hb.Add(NewStringFieldRight(asset.AccountName, refMax))
			hb.Add(NewFloatFieldRight(asset.ClosingValue, 10))
			tot = tot + asset.ClosingValue
			cObj = append(cObj, hb)
		}
		cObj = append(cObj, widget.NewSeparator())
		hb := container.NewHBox()
		hb.Add(NewStringFieldRight("Total Value:", refMax))
		hb.Add(NewFloatFieldRight(tot, 10))
		cObj = append(cObj, hb)
	}
	return container.NewScroll(container.NewVBox(cObj...))
}

func detailsScreen(w fyne.Window, details DetailPage, actionFunc func(string, *parser.Path, string), pref *pref.PrefData, statusDisplay *StatusDisplay, log func(string)) fyne.CanvasObject {
	data := details.GetObjectsForPage()
	cObj := make([]fyne.CanvasObject, 0)
	keys := listOfNonDupeInOrderKeys(data, preferedOrderReversed)
	transPath := parser.NewPath("", ".")
	for _, k := range keys {
		v := data.GetNodeWithName(k)
		idd := details.SelectedPath.StringAppend(k)
		editEntry, ok := EditEntryListCache.Get(idd)
		if !ok {
			editEntry = NewEditEntry(v, idd, k, v.String(),
				unDoFunction, actionFunc, statusDisplay)
			EditEntryListCache.Add(editEntry)
		}
		editEntry.RefreshData()
		na := editEntry.NodeAnnotation
		if v.GetName() == lib.IdTxTransactions && v.IsContainer() {
			transPath = editEntry.Path
		} else {
			clip := NewMyIconButton("", theme.ContentCopyIcon(), func(a, b string) {
				w.Clipboard().SetContent(editEntry.GetCurrentText())
				actionFunc(ACTION_COPIED, editEntry.Path, editEntry.GetCurrentText())
			}, "", "", statusDisplay, fmt.Sprintf("Copy the contents of '%s' to the clipboard", k))
			flClipboard := container.New(&FixedLayout{10, 1}, clip)
			flLab := container.New(&FixedLayout{100, 1}, editEntry.Lab)
			flLink := container.New(&FixedLayout{10, 0}, editEntry.Link)
			flUnDo := container.New(&FixedLayout{10, 0}, editEntry.UnDo)
			if len(keys) < 2 {
				editEntry.Remove.Disable()
			} else {
				editEntry.Remove.MyEnable()
			}
			flRemove := container.New(&FixedLayout{10, 0}, editEntry.Remove)
			flRename := container.New(&FixedLayout{10, 0}, editEntry.Rename)
			cObj = append(cObj, widget.NewSeparator())
			if !EditMode {
				switch na {
				case lib.NODE_TYPE_RT:
					cObj = append(cObj, container.NewBorder(nil, nil, container.NewHBox(flLink, flLab), nil, widget.NewRichTextFromMarkdown(editEntry.GetCurrentText())))
				case lib.NODE_TYPE_PO:
					cObj = append(cObj, container.NewBorder(nil, nil, container.NewHBox(flLink, flLab), nil, positional(editEntry.GetCurrentText())))
				case lib.NODE_TYPE_IM:
					image, message := loadImage(editEntry.GetCurrentText())
					if message == "" {
						image.FillMode = canvas.ImageFillOriginal
						cObj = append(cObj, image)
					} else {
						cObj = append(cObj, container.NewBorder(nil, nil, container.NewHBox(flLink, flLab), nil, widget.NewLabel(message)))
					}
				default:
					cObj = append(cObj, container.NewBorder(nil, nil, container.NewHBox(flLink, flLab, flClipboard), nil, widget.NewLabel(editEntry.GetCurrentText())))
				}
			} else {
				var we *widget.Entry
				editEntry.Rename.MyEnable()
				contHeight := editEntry.Lab.MinSize().Height
				if lib.NodeAnnotationsSingleLine[na] {
					we = widget.NewEntry()
				} else {
					we = widget.NewMultiLineEntry()
					if na != lib.NODE_TYPE_PO {
						contHeight = 250
					}
				}
				we.SetText(editEntry.GetCurrentText())
				we.OnChanged = func(newWalue string) {
					err := entryChangedFunction(newWalue, editEntry.Path)
					if err == nil {
						actionFunc(ACTION_UPDATED, editEntry.Path, "")
					} else {
						actionFunc(ACTION_ERROR_DIALOG, editEntry.Path, err.Error())
					}
				}
				editEntry.We = we
				cObj = append(cObj, container.NewBorder(nil, nil, container.NewHBox(flRemove, flRename, flLink, flLab, flUnDo), nil, container.New(NewFixedHLayout(300, contHeight), we)))
			}
		}
	}
	if !transPath.IsEmpty() {
		cObj = append(cObj, widget.NewSeparator())
		accountData, err := lib.FindUserAccount(details.User, details.Title)
		accountData.Path = *transPath
		if err == nil {
			cObj = getTransactionalCanvasObjects(actionFunc, cObj, accountData, pref, statusDisplay, log)
		}
	}
	return container.NewScroll(container.NewVBox(cObj...))
}

func entryChangedFunction(newWalue string, path *parser.Path) error {
	ee, ok := EditEntryListCache.Get(path)
	if ok {
		if ee.NodeType == parser.NT_NUMBER {
			nv := strings.TrimSpace(newWalue)
			if nv == "" {
				return nil
			}
			_, err := strconv.ParseFloat(strings.TrimSpace(nv), 64)
			if err != nil {
				ee.RefreshData()
				return fmt.Errorf("invalid number '%s'", nv)
			}
		}
		ee.SetNew(newWalue)
	}
	return nil
}

func unDoFunction(path *parser.Path) {
	ee, ok := EditEntryListCache.Get(path)
	if ok {
		ee.RevertEdit()
	}
}

func hintDetailsControls(_ fyne.Window, details DetailPage, actionFunc func(string, *parser.Path, string), pref *pref.PrefData, statusDisplay *StatusDisplay, log func(string)) fyne.CanvasObject {
	cObj := make([]fyne.CanvasObject, 0)
	cObj = append(cObj, NewMyIconButton("", theme.DeleteIcon(), func(a, b string) {
		actionFunc(ACTION_REMOVE, details.SelectedPath, "")
	}, "", "", statusDisplay, fmt.Sprintf("Delete: - '%s'", details.Title)))

	cObj = append(cObj, NewMyIconButton("", theme2.EditIcon(), func(a, b string) {
		actionFunc(ACTION_RENAME, details.SelectedPath, "")
	}, "", "", statusDisplay, fmt.Sprintf("Rename: - '%s'", details.Title)))

	cObj = append(cObj, NewMyIconButton("New", theme.ContentAddIcon(), func(a, b string) {
		actionFunc(ACTION_ADD_HINT_ITEM, details.SelectedPath, "")
	}, "", "", statusDisplay, fmt.Sprintf("Add new item to: %s", details.Title)))

	cObj = append(cObj, NewMyIconButton("", theme.ContentCopyIcon(), func(a, b string) {
		actionFunc(ACTION_CLONE, details.SelectedPath, "")
	}, "", "", statusDisplay, fmt.Sprintf("Copy: - '%s' without copying the data it contains", details.Title)))

	cObj = append(cObj, NewMyIconButton("Full", theme.ContentCopyIcon(), func(a, b string) {
		actionFunc(ACTION_CLONE_FULL, details.SelectedPath, "")
	}, "", "", statusDisplay, fmt.Sprintf("Copy: - '%s' keeping the data it contains", details.Title)))

	cObj = append(cObj, widget.NewLabel(details.Heading))
	return container.NewHBox(cObj...)
}

func listOfNonDupeInOrderKeys(m *parser.JsonObject, ordered []string) []string {
	keys := m.GetSortedKeys()
	sort.Strings(keys)
	for _, s := range ordered {
		pos, found := contains(keys, s)
		if found && pos > 0 {
			for i := pos; i > 0; i-- {
				keys[i] = keys[i-1]
			}
			keys[0] = s
		}
	}
	return keys
}

func contains(s []string, str string) (int, bool) {
	for i, v := range s {
		if v == str {
			return i, true
		}
	}
	return 0, false
}

func parseURL(urlStr string) *url.URL {
	link, err := url.Parse(urlStr)
	if err != nil {
		fyne.LogError("Could not parse URL", err)
	}

	return link
}

func runModalEntryPopup(w fyne.Window, heading, txt string, password bool, isAnnotated bool, annotation lib.NodeAnnotationEnum, accept func(bool, string, lib.NodeAnnotationEnum)) (modal *widget.PopUp) {
	var radioGroup *widget.RadioGroup
	var styles *fyne.Container
	var noteTypeId lib.NodeAnnotationEnum = 0
	submitInternal := func(s string) {
		modal.Hide()
		accept(true, s, noteTypeId)
	}

	radinGroupChanged := func(s string) {
		for i := 0; i < len(lib.NodeAnnotationEnums); i++ {
			if s == lib.NodeAnnotationPrefixNames[i] {
				noteTypeId = lib.NodeAnnotationEnums[i]
				return
			}
		}
		noteTypeId = 0
	}

	if isAnnotated {
		radioGroup = widget.NewRadioGroup(lib.NodeAnnotationPrefixNames, radinGroupChanged)
		radioGroup.SetSelected(lib.NodeAnnotationPrefixNames[annotation])
		styles = container.NewCenter(container.New(layout.NewHBoxLayout()), radioGroup)
	}
	entry := &widget.Entry{Text: txt, Password: password, OnChanged: nil, OnSubmitted: submitInternal}
	buttons := container.NewCenter(container.New(layout.NewHBoxLayout(), widget.NewButton("Cancel", func() {
		modal.Hide()
		accept(false, entry.Text, noteTypeId)
	}), widget.NewButton("OK", func() {
		modal.Hide()
		accept(true, entry.Text, noteTypeId)
	}),
	))
	if isAnnotated {
		modal = widget.NewModalPopUp(
			container.NewVBox(
				container.NewCenter(widget.NewLabel("Select the TYPE of item from below")),
				styles,
				container.NewCenter(widget.NewLabel(heading)),
				entry,
				buttons,
			),
			w.Canvas(),
		)
	} else {
		modal = widget.NewModalPopUp(
			container.NewVBox(
				container.NewCenter(widget.NewLabel(heading)),
				entry,
				buttons,
			),
			w.Canvas(),
		)
	}
	w.Canvas().SetOnTypedKey(func(ke *fyne.KeyEvent) {
		if ke.Name == "Return" {
			modal.Hide()
			accept(true, entry.Text, noteTypeId)
		} else {
			if ke.Name == "Escape" {
				modal.Hide()
				accept(false, entry.Text, noteTypeId)
			}
		}
	})
	modal.Show()
	w.Canvas().Focus(entry)
	return modal
}
