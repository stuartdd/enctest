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
package types

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type NodeAnnotationEnum int

const (
	NOTE_TYPE_SL NodeAnnotationEnum = iota
	NOTE_TYPE_ML
	NOTE_TYPE_RT
	NOTE_TYPE_PO
)

type MyButton struct {
	widget.Button
	onTapped func()
}

func NewMyButton(text string, f func()) *MyButton {
	button := &MyButton{onTapped: f}
	button.ExtendBaseWidget(button)
	button.SetText(text)
	return button
}

func (t *MyButton) Tapped(_ *fyne.PointEvent) {
	t.onTapped()
}

var (
	nodeAnnotationPrefix      = []string{"", "!ml", "!rt", "!po"}
	NodeAnnotationPrefixNames = []string{"Single Line", "Multi Line", "Rich Text", "Positional"}
	NodeAnnotationEnums       = []NodeAnnotationEnum{NOTE_TYPE_SL, NOTE_TYPE_ML, NOTE_TYPE_RT, NOTE_TYPE_PO}
)

func IndexOfAnnotation(annotation string) int {
	for i, v := range nodeAnnotationPrefix {
		if v == annotation {
			return i
		}
	}
	return 0
}

func GetNodeAnnotationTypeAndName(combinedName string) (NodeAnnotationEnum, string) {
	pos := strings.IndexRune(combinedName, '!')
	if pos < 0 {
		return NOTE_TYPE_SL, combinedName
	}
	aStr := combinedName[pos:]
	indx := IndexOfAnnotation(aStr)
	if indx == 0 {
		return NOTE_TYPE_SL, combinedName
	}
	return NodeAnnotationEnums[indx], combinedName[:pos]
}

func GetNodeAnnotationNameWithPrefix(nae NodeAnnotationEnum, name string) string {
	return name + nodeAnnotationPrefix[nae]
}

func GetNodeAnnotationPrefixName(nae NodeAnnotationEnum) string {
	return nodeAnnotationPrefix[nae]
}
