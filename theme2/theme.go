package theme2

import (
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type MyTheme struct {
	light bool
}

func NewMyTheme(varient string) *MyTheme {
	if strings.ToLower(varient) == "light" {
		return &MyTheme{light: true}
	}
	return &MyTheme{light: false}
}

var (
	_ fyne.Theme = (*MyTheme)(nil)

	darkPalette = map[fyne.ThemeColorName]color.Color{
		theme.ColorNameButton: color.NRGBA{R: 0xf4, G: 0x43, B: 0x36, A: 0x7f},
	}

	lightPalette = map[fyne.ThemeColorName]color.Color{
		theme.ColorNameButton: color.NRGBA{R: 0x43, G: 0xf4, B: 0x36, A: 0x7f},
	}
)

func (m MyTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if m.light {
		c, ok := lightPalette[name]
		if ok {
			return c
		}
	} else {
		c, ok := darkPalette[name]
		if ok {
			return c
		}
	}
	return theme.DefaultTheme().Color(name, variant)
}

func (m MyTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	if name == theme.IconNameContentUndo {
		if m.light {
			return resourceRevertLightPng
		}
		return resourceRevertDarkPng
	}
	return theme.DefaultTheme().Icon(name)
}

func (m MyTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (m MyTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
