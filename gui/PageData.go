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

	"fyne.io/fyne/v2"
	"github.com/stuartdd2/JsonParser4go/parser"
	"stuartdd.com/lib"
	"stuartdd.com/pref"
)

type DetailPage struct {
	Uid                         *parser.Path
	Heading, User, Group, Title string
	ViewFunc                    func(w fyne.Window, details DetailPage, actionFunc func(string, *parser.Path, string), pref *pref.PrefData, statusDisplay *StatusDisplay) fyne.CanvasObject
	CntlFunc                    func(w fyne.Window, details DetailPage, actionFunc func(string, *parser.Path, string), pref *pref.PrefData, statusDisplay *StatusDisplay) fyne.CanvasObject
	DataRootMap                 parser.NodeI
	Preferences                 pref.PrefData
}

func NewDetailPage(
	uid *parser.Path,
	user string,
	group string,
	title string,
	viewFunc func(w fyne.Window, details DetailPage, actionFunc func(string, *parser.Path, string), pref *pref.PrefData, statusDisplay *StatusDisplay) fyne.CanvasObject,
	cntlFunc func(w fyne.Window, details DetailPage, actionFunc func(string, *parser.Path, string), pref *pref.PrefData, statusDisplay *StatusDisplay) fyne.CanvasObject,
	dataRootMap parser.NodeI,
	preferences pref.PrefData) *DetailPage {

	heading := fmt.Sprintf("User:  %s", title)
	if user != "" {
		heading = fmt.Sprintf("User:  %s - %s", user, title)
	}
	return &DetailPage{Uid: uid, Heading: heading, Title: title, Group: group, User: user, ViewFunc: viewFunc, CntlFunc: cntlFunc, DataRootMap: dataRootMap, Preferences: preferences}
}

func (p *DetailPage) GetObjectsForUid() *parser.JsonObject {
	m, err := lib.GetUserDataForUid(p.DataRootMap, p.Uid)
	if err != nil {
		panic(fmt.Sprintf("DetailPage.GetMapForUid. Uid '%s' not found. %s", p.Uid, err.Error()))
	}
	if m.GetNodeType() == parser.NT_OBJECT {
		return m.(*parser.JsonObject)
	}
	panic("DetailPage.GetMapForUid must only return JsonObject types")
}
