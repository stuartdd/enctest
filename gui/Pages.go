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
	welcomeTitle              = "Welcome"
	appDesc                   = "Welcome to Valt"
	idNotes                   = "notes"
	idPwDetails               = "pwHints"
	DataPresModePrefName      = "data.presentationmode"
	DataHintIsCalledPrefName  = "data.hintIsCalled"
	DataNoteIsCalledPrefName  = "data.noteIsCalled"
	DataTransIsCalledPrefName = "data.transIsCalled"
	DataAssetIsCalledPrefName = "data.assetIsCalled"

	PATH_SEP = "|"

	ACTION_LOG           = "log"
	ACTION_COPIED        = "copied"
	ACTION_REMOVE        = "remove"
	ACTION_RENAME        = "rename"
	ACTION_CLONE         = "clone"
	ACTION_CLONE_FULL    = "clonefull"
	ACTION_LINK          = "link"
	ACTION_UPDATED       = "update"
	ACTION_ADD_NOTE      = "addnode"
	ACTION_ADD_HINT      = "addhint"
	ACTION_ADD_HINT_ITEM = "addhintitem"
)

var (
	preferedOrderReversed = []string{"notes", "positional", "post", "pre", "link", "userId"}
	EditEntryListCache    = NewEditEntryList()
)

func NewModalEntryDialog(w fyne.Window, heading, txt string, isNote bool, annotation lib.NodeAnnotationEnum, accept func(bool, string, lib.NodeAnnotationEnum)) (modal *widget.PopUp) {
	return runModalEntryPopup(w, heading, txt, false, isNote, annotation, accept)
}

func NewModalPasswordDialog(w fyne.Window, heading, txt string, accept func(bool, string, lib.NodeAnnotationEnum)) (modal *widget.PopUp) {
	return runModalEntryPopup(w, heading, txt, true, false, lib.NOTE_TYPE_SL, accept)
}

func GetWelcomePage(uid *parser.Path, preferences pref.PrefData) *DetailPage {
	return NewDetailPage(uid, welcomeTitle, "", "", welcomeScreen, welcomeControls, nil, preferences)
}

func GetDetailTypeGroupTitle(uid *parser.Path, preferences pref.PrefData) (lib.NodeAnnotationEnum, string, string) {
	switch uid.Len() {
	case 2:
		type1, group1 := lib.GetNodeAnnotationTypeAndName(uid.StringAt(1))
		if group1 == idPwDetails {
			return type1, group1, preferences.GetStringForPathWithFallback(DataHintIsCalledPrefName, "Hint")
		}
		if group1 == idNotes {
			return type1, group1, preferences.GetStringForPathWithFallback(DataNoteIsCalledPrefName, "Note")
		}
		if group1 == lib.IdAssets {
			return type1, group1, preferences.GetStringForPathWithFallback(DataAssetIsCalledPrefName, "Asset")
		}
		return type1, group1, group1
	case 3:
		type1, group1 := lib.GetNodeAnnotationTypeAndName(uid.StringAt(1))
		_, name2 := lib.GetNodeAnnotationTypeAndName(uid.StringAt(2))
		return type1, group1, name2
	default:
		return lib.NOTE_TYPE_SL, uid.String(), uid.String()
	}
}

func GetDetailPage(uid *parser.Path, dataRootMap parser.NodeI, preferences pref.PrefData) *DetailPage {
	user0 := uid.StringAt(0)
	_, group1, title := GetDetailTypeGroupTitle(uid, preferences)
	switch uid.Len() {
	case 1:
		return NewDetailPage(uid, "", "", uid.String(), welcomeScreen, userControls, dataRootMap, preferences)
	case 2:
		if group1 == idPwDetails {
			return NewDetailPage(uid, user0, group1, title+"s", welcomeScreen, hintControls, dataRootMap, preferences)
		}
		if group1 == idNotes {
			return NewDetailPage(uid, user0, group1, title+"s", detailsScreen, noteDetailsControls, dataRootMap, preferences)
		}
		if group1 == lib.IdAssets {
			return NewDetailPage(uid, user0, group1, title+"s", assetScreen, assetSummaryControls, dataRootMap, preferences)
		}
		return NewDetailPage(uid, user0, group1, title, welcomeScreen, welcomeControls, dataRootMap, preferences)
	case 3:
		_, name2 := lib.GetNodeAnnotationTypeAndName(uid.StringAt(2))
		if group1 == idPwDetails {
			return NewDetailPage(uid, user0, group1, name2, detailsScreen, hintDetailsControls, dataRootMap, preferences)
		}
		if group1 == idNotes {
			return NewDetailPage(uid, user0, group1, name2, detailsScreen, noteDetailsControls, dataRootMap, preferences)
		}
		if group1 == lib.IdAssets {
			return NewDetailPage(uid, user0, group1, name2, detailsScreen, assetControls, dataRootMap, preferences)
		}
		return NewDetailPage(uid, user0, group1, name2, welcomeScreen, welcomeControls, dataRootMap, preferences)
	default:
		return NewDetailPage(uid, user0, group1, title, welcomeScreen, welcomeControls, dataRootMap, preferences)
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

func welcomeControls(_ fyne.Window, details DetailPage, actionFunc func(string, *parser.Path, string), pref *pref.PrefData, statusDisplay *StatusDisplay) fyne.CanvasObject {
	cObj := make([]fyne.CanvasObject, 0)
	cObj = append(cObj, widget.NewLabel(details.Heading))
	return container.NewHBox(cObj...)
}

func userControls(_ fyne.Window, details DetailPage, actionFunc func(string, *parser.Path, string), pref *pref.PrefData, statusDisplay *StatusDisplay) fyne.CanvasObject {
	cObj := make([]fyne.CanvasObject, 0)
	cObj = append(cObj, NewMyIconButton("", theme.DeleteIcon(), func() {
		actionFunc(ACTION_REMOVE, details.Uid, "")
	}, statusDisplay, fmt.Sprintf("Delete: - '%s'", details.Title)))

	cObj = append(cObj, NewMyIconButton("", theme2.EditIcon(), func() {
		actionFunc(ACTION_RENAME, details.Uid, "")
	}, statusDisplay, fmt.Sprintf("Rename: - '%s'", details.Title)))

	cObj = append(cObj, widget.NewLabel(details.Heading))
	return container.NewHBox(cObj...)
}

func welcomeScreen(_ fyne.Window, details DetailPage, actionFunc func(string, *parser.Path, string), pref *pref.PrefData, statusDisplay *StatusDisplay) fyne.CanvasObject {
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

func assetControls(_ fyne.Window, details DetailPage, actionFunc func(string, *parser.Path, string), pref *pref.PrefData, statusDisplay *StatusDisplay) fyne.CanvasObject {
	head := fmt.Sprintf("%s: Asset - %s", details.User, details.Title)
	cObj := make([]fyne.CanvasObject, 0)
	cObj = append(cObj, NewMyIconButton("New", theme.ContentAddIcon(), func() {
		actionFunc(ACTION_ADD_HINT, details.Uid, "")
	}, statusDisplay, fmt.Sprintf("Add new Item to Asset: %s", details.Title)))
	cObj = append(cObj, widget.NewLabel(head))
	return container.NewHBox(cObj...)
}

func assetSummaryControls(_ fyne.Window, details DetailPage, actionFunc func(string, *parser.Path, string), pref *pref.PrefData, statusDisplay *StatusDisplay) fyne.CanvasObject {
	head := fmt.Sprintf("Asset Summary for user %s", details.User)
	cObj := make([]fyne.CanvasObject, 0)
	cObj = append(cObj, NewMyIconButton("New", theme.ContentAddIcon(), func() {
		actionFunc(ACTION_ADD_HINT, details.Uid, details.Title)
	}, statusDisplay, fmt.Sprintf("Add new Asset for user %s", details.User)))
	cObj = append(cObj, widget.NewLabel(head))
	return container.NewHBox(cObj...)
}

func hintControls(_ fyne.Window, details DetailPage, actionFunc func(string, *parser.Path, string), pref *pref.PrefData, statusDisplay *StatusDisplay) fyne.CanvasObject {
	cObj := make([]fyne.CanvasObject, 0)
	cObj = append(cObj, NewMyIconButton("New", theme.ContentAddIcon(), func() {
		actionFunc(ACTION_ADD_HINT, details.Uid, "")
	}, statusDisplay, fmt.Sprintf("Add new Hint to user: %s", details.User)))
	cObj = append(cObj, widget.NewLabel(details.Heading))
	return container.NewHBox(cObj...)
}

func noteDetailsControls(_ fyne.Window, details DetailPage, actionFunc func(string, *parser.Path, string), pref *pref.PrefData, statusDisplay *StatusDisplay) fyne.CanvasObject {
	cObj := make([]fyne.CanvasObject, 0)
	cObj = append(cObj, NewMyIconButton("New", theme.ContentAddIcon(), func() {
		actionFunc(ACTION_ADD_NOTE, details.Uid, "")
	}, statusDisplay, fmt.Sprintf("Add new Note to user: %s", details.User)))
	cObj = append(cObj, widget.NewLabel(details.Heading))
	return container.NewHBox(cObj...)
}

func getTransactionalCanvasObjects(cObj []fyne.CanvasObject, accData *lib.AccountData, editEntry *EditEntry, pref *pref.PrefData) []fyne.CanvasObject {
	transAreCalled := pref.GetStringForPathWithFallback(DataTransIsCalledPrefName, "Transaction")
	txList := accData.Transactions
	refMax := 10
	for _, tx := range txList {
		if refMax < len(tx.Ref()) {
			refMax = len(tx.Ref())
		}
	}
	cObj = append(cObj, widget.NewLabel(fmt.Sprintf("%s. List of %s(s). Current balance %0.2f", accData.AccountName, transAreCalled, accData.ClosingValue)))
	hb := container.NewHBox()
	hb.Add(NewStringFieldRight("Date:", 19))
	hb.Add(NewStringFieldRight("Reference:", refMax))
	hb.Add(NewStringFieldRight("D/C", 3))
	hb.Add(NewStringFieldRight("Amount:", 10))
	hb.Add(NewStringFieldRight("Balance:", 10))
	cObj = append(cObj, hb)
	for _, tx := range txList {
		hb := container.NewHBox()
		hb.Add(NewStringFieldRight(tx.DateTime(), 19))
		hb.Add(NewStringFieldRight(tx.Ref(), refMax))
		if tx.Value() >= 0 {
			hb.Add(NewStringFieldRight("D  ", 3))
		} else {
			hb.Add(NewStringFieldRight("  C", 3))
		}
		hb.Add(NewFloatFieldRight(tx.Value(), 10))
		hb.Add(NewFloatFieldRight(tx.LineValue(), 10))
		cObj = append(cObj, hb)
	}
	return cObj
}

func assetScreen(w fyne.Window, details DetailPage, actionFunc func(string, *parser.Path, string), pref *pref.PrefData, statusDisplay *StatusDisplay) fyne.CanvasObject {
	cObj := make([]fyne.CanvasObject, 0)
	cObj = append(cObj, widget.NewSeparator())
	assetData, err := lib.FindUserAssets(details.User)
	tot := 0.0
	if err == nil {
		refMax := 10
		for _, asset := range assetData {
			if refMax < len(asset.AccountName) {
				refMax = len(asset.AccountName)
			}
		}
		for _, asset := range assetData {
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

func detailsScreen(w fyne.Window, details DetailPage, actionFunc func(string, *parser.Path, string), pref *pref.PrefData, statusDisplay *StatusDisplay) fyne.CanvasObject {
	data := details.GetObjectsForUid()
	cObj := make([]fyne.CanvasObject, 0)
	keys := listOfNonDupeInOrderKeys(data, preferedOrderReversed)
	for _, k := range keys {
		v := data.GetNodeWithName(k)
		idd := details.Uid.StringAppend(k)
		editEntry, ok := EditEntryListCache.Get(idd)
		if !ok {
			editEntry = NewEditEntry(idd, k, v.String(),
				func(newWalue string, path *parser.Path) { //onChangeFunction
					entryChangedFunction(newWalue, path)
					actionFunc(ACTION_UPDATED, parser.NewBarPath(""), "")
				},
				unDoFunction, actionFunc, statusDisplay)
			EditEntryListCache.Add(editEntry)
		}
		editEntry.RefreshData()
		na := editEntry.NodeAnnotation
		if v.GetName() == lib.IdTransactions && v.IsContainer() {
			cObj = append(cObj, widget.NewSeparator())
			accountData, err := lib.FindUserAccount(details.User, details.Title)
			if err == nil {
				cObj = getTransactionalCanvasObjects(cObj, accountData, editEntry, pref)
			}
		} else {
			clip := NewMyIconButton("", theme.ContentCopyIcon(), func() {
				w.Clipboard().SetContent(editEntry.GetCurrentText())
				actionFunc(ACTION_COPIED, editEntry.Path, editEntry.GetCurrentText())
			}, statusDisplay, fmt.Sprintf("Copy the contents of '%s' to the clipboard", k))
			flClipboard := container.New(&FixedLayout{10, 1}, clip)
			flLab := container.New(&FixedLayout{100, 1}, editEntry.Lab)
			flLink := container.New(&FixedLayout{10, 0}, editEntry.Link)
			flUnDo := container.New(&FixedLayout{10, 0}, editEntry.UnDo)
			if len(keys) < 2 {
				editEntry.Remove.Disable()
			} else {
				editEntry.Remove.Enable()
			}
			flRemove := container.New(&FixedLayout{10, 0}, editEntry.Remove)
			flRename := container.New(&FixedLayout{10, 0}, editEntry.Rename)
			dp := details.Preferences.GetBoolWithFallback(DataPresModePrefName, true)
			cObj = append(cObj, widget.NewSeparator())
			if dp {
				switch na {
				case lib.NOTE_TYPE_RT:
					cObj = append(cObj, container.NewBorder(nil, nil, container.NewHBox(flLink, flLab), nil, widget.NewRichTextFromMarkdown(editEntry.GetCurrentText())))
				case lib.NOTE_TYPE_PO:
					cObj = append(cObj, container.NewBorder(nil, nil, container.NewHBox(flLink, flLab), nil, positional(editEntry.GetCurrentText())))
				case lib.NOTE_TYPE_IM:
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
				editEntry.Rename.Enable()
				contHeight := editEntry.Lab.MinSize().Height
				if lib.NodeAnnotationsSingleLine[na] {
					we = widget.NewEntry()
				} else {
					we = widget.NewMultiLineEntry()
					if na != lib.NOTE_TYPE_PO {
						contHeight = 250
					}
				}
				we.OnChanged = func(newWalue string) {
					entryChangedFunction(newWalue, editEntry.Path)
					actionFunc(ACTION_UPDATED, editEntry.Path, "")
				}
				we.SetText(editEntry.GetCurrentText())
				editEntry.We = we
				cObj = append(cObj, container.NewBorder(nil, nil, container.NewHBox(flRemove, flRename, flLink, flLab, flUnDo), nil, container.New(NewFixedHLayout(300, contHeight), we)))
			}
		}
	}
	return container.NewScroll(container.NewVBox(cObj...))
}

func entryChangedFunction(newWalue string, path *parser.Path) {
	ee, ok := EditEntryListCache.Get(path)
	if ok {
		ee.SetNew(newWalue)
	}
}

func unDoFunction(path *parser.Path) {
	ee, ok := EditEntryListCache.Get(path)
	if ok {
		ee.RevertEdit()
	}
}

func hintDetailsControls(_ fyne.Window, details DetailPage, actionFunc func(string, *parser.Path, string), pref *pref.PrefData, statusDisplay *StatusDisplay) fyne.CanvasObject {
	cObj := make([]fyne.CanvasObject, 0)
	cObj = append(cObj, NewMyIconButton("", theme.DeleteIcon(), func() {
		actionFunc(ACTION_REMOVE, details.Uid, "")
	}, statusDisplay, fmt.Sprintf("Delete: - '%s'", details.Title)))

	cObj = append(cObj, NewMyIconButton("", theme2.EditIcon(), func() {
		actionFunc(ACTION_RENAME, details.Uid, "")
	}, statusDisplay, fmt.Sprintf("Rename: - '%s'", details.Title)))

	cObj = append(cObj, NewMyIconButton("New", theme.ContentAddIcon(), func() {
		actionFunc(ACTION_ADD_HINT_ITEM, details.Uid, "")
	}, statusDisplay, fmt.Sprintf("Add new item to: %s", details.Title)))

	cObj = append(cObj, NewMyIconButton("", theme.ContentCopyIcon(), func() {
		actionFunc(ACTION_CLONE, details.Uid, "")
	}, statusDisplay, fmt.Sprintf("Copy: - '%s' without copying the data it contains", details.Title)))

	cObj = append(cObj, NewMyIconButton("Full", theme.ContentCopyIcon(), func() {
		actionFunc(ACTION_CLONE_FULL, details.Uid, "")
	}, statusDisplay, fmt.Sprintf("Copy: - '%s' keeping the data it contains", details.Title)))

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

func runModalEntryPopup(w fyne.Window, heading, txt string, password bool, isNote bool, annotation lib.NodeAnnotationEnum, accept func(bool, string, lib.NodeAnnotationEnum)) (modal *widget.PopUp) {
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

	if isNote {
		radioGroup = widget.NewRadioGroup(lib.NodeAnnotationPrefixNames, radinGroupChanged)
		radioGroup.SetSelected(lib.NodeAnnotationPrefixNames[annotation])
		styles = container.NewCenter(container.New(layout.NewHBoxLayout()), radioGroup)
	}
	entry := &widget.Entry{Text: txt, Password: password, OnChanged: func(s string) {}, OnSubmitted: submitInternal}
	buttons := container.NewCenter(container.New(layout.NewHBoxLayout(), widget.NewButton("Cancel", func() {
		modal.Hide()
		accept(false, entry.Text, noteTypeId)
	}), widget.NewButton("OK", func() {
		modal.Hide()
		accept(true, entry.Text, noteTypeId)
	}),
	))
	if isNote {
		modal = widget.NewModalPopUp(
			container.NewVBox(
				container.NewCenter(widget.NewLabel("Select the TYPE of the new note")),
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
	return modal
}
