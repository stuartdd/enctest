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
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/stuartdd/jsonParserGo/parser"
	"stuartdd.com/pref"
	"stuartdd.com/theme2"
	"stuartdd.com/types"
)

const (
	welcomeTitle             = "Welcome"
	appDesc                  = "Welcome to Valt"
	idNotes                  = "notes"
	idPwDetails              = "pwHints"
	DataPresModePrefName     = "data.presentationmode"
	DataHintIsCalledPrefName = "data.hintIsCalled"
	DataNoteIsCalledPrefName = "data.noteIsCalled"

	ACTION_LOG        = "log"
	ACTION_COPIED     = "copied"
	ACTION_REMOVE     = "remove"
	ACTION_RENAME     = "rename"
	ACTION_CLONE      = "clone"
	ACTION_CLONE_FULL = "clonefull"
	ACTION_LINK       = "link"
	ACTION_UPDATED    = "update"
)

var (
	preferedOrderReversed = []string{"notes", "positional", "post", "pre", "link", "userId"}
	EditEntryListCache    = NewEditEntryList()
)

func NewModalEntryDialog(w fyne.Window, heading, txt string, isNote bool, annotation types.NodeAnnotationEnum, accept func(bool, string, types.NodeAnnotationEnum)) (modal *widget.PopUp) {
	return runModalEntryPopup(w, heading, txt, false, isNote, annotation, accept)
}

func NewModalPasswordDialog(w fyne.Window, heading, txt string, accept func(bool, string, types.NodeAnnotationEnum)) (modal *widget.PopUp) {
	return runModalEntryPopup(w, heading, txt, true, false, types.NOTE_TYPE_SL, accept)
}

func GetWelcomePage(id string, preferences pref.PrefData) *DetailPage {
	return NewDetailPage(id, "", welcomeTitle, welcomeScreen, welcomeControls, nil, preferences)
}

func GetDetailPage(id string, dataRootMap parser.NodeI, preferences pref.PrefData) *DetailPage {
	nodes := strings.Split(id, ".")
	user := nodes[0]
	hintsAreCalled := preferences.GetStringForPathWithFallback(DataHintIsCalledPrefName, "Hint")
	notesAreCalled := preferences.GetStringForPathWithFallback(DataNoteIsCalledPrefName, "Note")
	switch len(nodes) {
	case 1:
		return NewDetailPage(id, id, "", welcomeScreen, welcomeControls, dataRootMap, preferences)
	case 2:
		if nodes[1] == idPwDetails {
			return NewDetailPage(id, hintsAreCalled+"s", user, welcomeScreen, notesControls, dataRootMap, preferences)
		}
		if nodes[1] == idNotes {
			return NewDetailPage(id, notesAreCalled+"s", user, notesScreen, notesControls, dataRootMap, preferences)
		}
		return NewDetailPage(id, "Unknown", user, welcomeScreen, notesControls, dataRootMap, preferences)
	case 3:
		if nodes[1] == idPwDetails {
			return NewDetailPage(id, nodes[2], user, notesScreen, hintsControls, dataRootMap, preferences)
		}
		if nodes[1] == idNotes {
			return NewDetailPage(id, nodes[2], user, notesScreen, notesControls, dataRootMap, preferences)
		}
		return NewDetailPage(id, "Unknown", "", welcomeScreen, notesControls, dataRootMap, preferences)
	}
	return NewDetailPage(id, id, "", welcomeScreen, notesControls, dataRootMap, preferences)
}

func positional(s string) fyne.CanvasObject {
	g1 := container.NewHBox()
	g1.Add(widget.NewSeparator())
	for i, c := range s {
		v1 := container.NewVBox()
		v1.Add(widget.NewSeparator())
		v1.Add(container.New(NewFixedWHLayout(20, 15), widget.NewLabel(fmt.Sprintf("%d", i+1))))
		v1.Add(widget.NewSeparator())
		v1.Add(container.New(NewFixedWHLayout(20, 15), widget.NewLabel(string(c))))
		v1.Add(widget.NewSeparator())
		g1.Add(v1)
		g1.Add(widget.NewSeparator())
	}
	return g1
}

func welcomeControls(_ fyne.Window, details DetailPage, actionFunc func(string, string, string)) fyne.CanvasObject {
	cObj := make([]fyne.CanvasObject, 0)
	cObj = append(cObj, widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		actionFunc(ACTION_REMOVE, details.Uid, "")
	}))
	cObj = append(cObj, widget.NewButtonWithIcon("", theme2.EditIcon(), func() {
		actionFunc(ACTION_RENAME, details.Uid, "")
	}))
	cObj = append(cObj, widget.NewLabel(details.Heading))
	return container.NewHBox(cObj...)
}

func welcomeScreen(_ fyne.Window, details DetailPage, actionFunc func(string, string, string)) fyne.CanvasObject {
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

func notesControls(_ fyne.Window, details DetailPage, actionFunc func(string, string, string)) fyne.CanvasObject {
	cObj := make([]fyne.CanvasObject, 0)
	cObj = append(cObj, widget.NewLabel(details.Heading))
	return container.NewHBox(cObj...)
}

func notesScreen(w fyne.Window, details DetailPage, actionFunc func(string, string, string)) fyne.CanvasObject {
	data := details.GetObjectsForUid()
	cObj := make([]fyne.CanvasObject, 0)
	keys := listOfNonDupeInOrderKeys(data, preferedOrderReversed)
	for _, k := range keys {
		v := data.GetNodeWithName(k)
		idd := details.Uid + "." + k
		editEntry, ok := EditEntryListCache.Get(idd)
		if !ok {
			editEntry = NewEditEntry(idd, k, v.String(),
				func(newWalue string, path string) { //onChangeFunction
					entryChangedFunction(newWalue, path)
					actionFunc(ACTION_UPDATED, path, "")
				},
				unDoFunction, actionFunc)
			EditEntryListCache.Add(editEntry)
		}
		editEntry.RefreshData()

		clip := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
			w.Clipboard().SetContent(editEntry.GetCurrentText())
			actionFunc(ACTION_COPIED, editEntry.Path, editEntry.GetCurrentText())
		})
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
		na := editEntry.NodeAnnotation
		dp := details.Preferences.GetBoolWithFallback(DataPresModePrefName, true)
		cObj = append(cObj, widget.NewSeparator())
		if na == types.NOTE_TYPE_RT && dp {
			rt := widget.NewRichTextFromMarkdown(editEntry.GetCurrentText())
			cObj = append(cObj, container.NewBorder(nil, nil, container.NewHBox(flLink, flLab), nil, rt))
		} else {
			if na == types.NOTE_TYPE_PO && dp {
				cObj = append(cObj, container.NewBorder(nil, nil, container.NewHBox(flLink, flLab), nil, positional(editEntry.GetCurrentText())))
			} else {
				if dp {
					cObj = append(cObj, container.NewBorder(nil, nil, container.NewHBox(flLink, flLab, flClipboard), nil, widget.NewLabel(editEntry.GetCurrentText())))
				} else {
					var we *widget.Entry
					editEntry.Rename.Enable()
					contHeight := editEntry.Lab.MinSize().Height
					if na == types.NOTE_TYPE_SL {
						we = widget.NewEntry()
					} else {
						we = widget.NewMultiLineEntry()
						if na != types.NOTE_TYPE_PO {
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
	}
	return container.NewScroll(container.NewVBox(cObj...))
}

func entryChangedFunction(newWalue string, path string) {
	ee, ok := EditEntryListCache.Get(path)
	if ok {
		ee.SetNew(newWalue)
	}
}

func unDoFunction(path string) {
	ee, ok := EditEntryListCache.Get(path)
	if ok {
		ee.RevertEdit()
	}
}

func hintsControls(_ fyne.Window, details DetailPage, actionFunc func(string, string, string)) fyne.CanvasObject {
	cObj := make([]fyne.CanvasObject, 0)
	cObj = append(cObj, widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		actionFunc(ACTION_REMOVE, details.Uid, "")
	}))
	cObj = append(cObj, widget.NewButtonWithIcon("", theme2.EditIcon(), func() {
		actionFunc(ACTION_RENAME, details.Uid, "")
	}))
	cObj = append(cObj, widget.NewButtonWithIcon("Clone", theme.ContentCopyIcon(), func() {
		actionFunc(ACTION_CLONE_FULL, details.Uid, "")
	}))
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

func runModalEntryPopup(w fyne.Window, heading, txt string, password bool, isNote bool, annotation types.NodeAnnotationEnum, accept func(bool, string, types.NodeAnnotationEnum)) (modal *widget.PopUp) {
	var radioGroup *widget.RadioGroup
	var styles *fyne.Container
	var noteTypeId types.NodeAnnotationEnum = 0
	submitInternal := func(s string) {
		modal.Hide()
		accept(true, s, noteTypeId)
	}

	radinGroupChanged := func(s string) {
		for i := 0; i < len(types.NodeAnnotationEnums); i++ {
			if s == types.NodeAnnotationPrefixNames[i] {
				noteTypeId = types.NodeAnnotationEnums[i]
				return
			}
		}
		noteTypeId = 0
	}

	if isNote {
		radioGroup = widget.NewRadioGroup(types.NodeAnnotationPrefixNames, radinGroupChanged)
		radioGroup.SetSelected(types.NodeAnnotationPrefixNames[annotation])
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
