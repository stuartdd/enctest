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
package lib

import (
	"strings"
)

func PadLeft(s string, w int) string {
	if len(s) > w {
		return s[:w]
	}
	var sb strings.Builder
	for i := 0; i < (w - len(s)); i++ {
		sb.WriteByte(32)
		if sb.Len() >= w {
			return sb.String()
		}
	}
	sb.WriteString(s)
	return sb.String()
}

func ParseStringForLink(s string) (string, bool) {
	var sb strings.Builder
	lc := strings.ToLower(s)
	pos := strings.Index(lc, "http://")
	if pos == -1 {
		pos = strings.Index(lc, "https://")
	}
	if pos == -1 {
		return "", false
	}
	count := 0
	for i := pos; i < len(lc); i++ {
		if lc[i] <= ' ' {
			break
		} else {
			sb.WriteByte(lc[i])
			count++
		}
	}
	if count < 12 {
		return "", false
	}
	return sb.String(), true
}
