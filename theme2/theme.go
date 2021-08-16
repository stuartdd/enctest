package theme2

import (
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
	light bool
}

const IconNameLinkToWeb fyne.ThemeIconName = "linkToWeb"
const IconNameApplication fyne.ThemeIconName = "applicationIcon"

func LinkToWebIcon() fyne.Resource {
	return safeIconLookup(IconNameLinkToWeb)
}

func AppLogo() fyne.Resource {
	return resourceAppIconPng
}

func safeIconLookup(n fyne.ThemeIconName) fyne.Resource {
	t := fyne.CurrentApp().Settings().Theme()
	icon := t.Icon(n)
	if icon != nil {
		return icon
	}
	return t.Icon(theme.IconNameQuestion)
}

func NewAppTheme(varient string) *AppTheme {
	if strings.ToLower(varient) == "light" {
		return &AppTheme{light: true}
	}
	return &AppTheme{light: false}
}

var (
	_ fyne.Theme = (*AppTheme)(nil)

	darkPalette = map[fyne.ThemeColorName]color.Color{
		theme.ColorNameButton: color.NRGBA{R: 0xf4, G: 0x43, B: 0x36, A: 0x7f},
	}

	lightPalette = map[fyne.ThemeColorName]color.Color{
		theme.ColorNameButton: color.NRGBA{R: 0x43, G: 0xf4, B: 0x36, A: 0x7f},
	}
)

func (m AppTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if m.light {
		c, ok := lightPalette[name]
		if ok {
			return c
		}
		return theme.DefaultTheme().Color(name, theme.VariantLight)
	} else {
		c, ok := darkPalette[name]
		if ok {
			return c
		}
		return theme.DefaultTheme().Color(name, theme.VariantDark)
	}
}

func (m AppTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	if name == IconNameLinkToWeb {
		if m.light {
			return resourceLinkLightSvg
		}
		return resourceLinkDarkSvg
	}
	if name == IconNameApplication {
		return resourceAppIconPng
	}
	if name == theme.IconNameContentUndo {
		if m.light {
			return resourceRevertLightSvg
		}
		return resourceRevertDarkSvg
	}

	return theme.DefaultTheme().Icon(name)
}

func (m AppTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (m AppTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
