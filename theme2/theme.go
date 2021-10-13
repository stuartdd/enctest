/*
 * Copyright (C) 2018 Stuiart Davies (stuartdd)
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
package theme2

import (
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

/*
SVG theme resource not rendered correctly on button when Disabled
- **OS:** Linux Mint 20.2 Cinnamon
- **Version:** 5.11.0-25-generic
- **Go version:** go version go1.16.7 linux/amd64
- **Fyne version:** fyne.io/fyne/v2 v2.0.4
*/

type AppTheme struct {
	lightName string
	variant   fyne.ThemeVariant
}

const (
	IconNameEdit      fyne.ThemeIconName = "edit"
	IconNameLinkToWeb fyne.ThemeIconName = "linkToWeb"
	LN                string             = "Light"
	DN                string             = "Dark"
)

var (
	_ fyne.Theme = (*AppTheme)(nil)

	appIcons = map[string]fyne.Resource{
		string(theme.IconNameContentUndo) + LN: resourceRevertLightSvg,
		string(theme.IconNameContentUndo) + DN: resourceRevertDarkSvg,
		string(IconNameLinkToWeb) + LN:         resourceLinkLightSvg,
		string(IconNameLinkToWeb) + DN:         resourceLinkDarkSvg,
		string(IconNameEdit) + LN:              resourceEditLightSvg,
		string(IconNameEdit) + DN:              resourceEditDarkSvg,
	}

	colorPalette = map[string]color.Color{
		string(theme.ColorNamePrimary) + LN:         color.NRGBA{R: 0x43, G: 0xf4, B: 0x36, A: 0xff},
		string(theme.ColorNamePrimary) + DN:         color.NRGBA{R: 0xf4, G: 0x43, B: 0x36, A: 0xff},
		string(theme.ColorNameButton) + LN:          color.NRGBA{R: 0x43, G: 0xf4, B: 0x36, A: 0x7f},
		string(theme.ColorNameButton) + DN:          color.NRGBA{R: 0xf4, G: 0x43, B: 0x36, A: 0x7f},
		string(theme.ColorNameInputBackground) + LN: color.NRGBA{R: 0x36, G: 0xFF, B: 0x36, A: 0x4f},
		string(theme.ColorNameInputBackground) + DN: color.NRGBA{R: 0xff, G: 0x36, B: 0x36, A: 0x4f},
	}
)

func LinkToWebIcon() fyne.Resource {
	return safeLookupViaTheme(IconNameLinkToWeb)
}

func EditIcon() fyne.Resource {
	return safeLookupViaTheme(IconNameEdit)
}

func AppLogo() fyne.Resource {
	return resourceAppIconPng
}

func NewAppTheme(varient string) *AppTheme {
	if strings.ToLower(varient) == "light" {
		return &AppTheme{lightName: LN, variant: theme.VariantLight}
	}
	return &AppTheme{lightName: DN, variant: theme.VariantDark}
}

func (m AppTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	c, ok := colorPalette[string(name)+m.lightName]
	if ok {
		return c
	}
	c, ok = colorPalette[string(name)]
	if ok {
		return c
	}
	return theme.DefaultTheme().Color(name, m.variant)
}

func (m AppTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	i, ok := appIcons[string(name)+m.lightName]
	if ok {
		return i
	}
	i, ok = appIcons[string(name)]
	if ok {
		return i
	}
	return theme.DefaultTheme().Icon(name)
}

func (m AppTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (m AppTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameSeparatorThickness:
		return 2
	case theme.SizeNameInlineIcon:
		return 20
	case theme.SizeNamePadding:
		return 4
	case theme.SizeNameScrollBar:
		return 16
	case theme.SizeNameScrollBarSmall:
		return 3
	case theme.SizeNameText:
		return 14
	case theme.SizeNameCaptionText:
		return 11
	case theme.SizeNameInputBorder:
		return 2
	default:
		fmt.Println(name)
		return theme.DefaultTheme().Size(name)
	}
}

func safeLookupViaTheme(n fyne.ThemeIconName) fyne.Resource {
	t := fyne.CurrentApp().Settings().Theme()
	icon := t.Icon(n)
	if icon != nil {
		return icon
	}
	return t.Icon(theme.IconNameQuestion)
}
