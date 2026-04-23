package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type darkTheme struct{}

var _ fyne.Theme = (*darkTheme)(nil)

func NewDarkTheme() fyne.Theme { return &darkTheme{} }

func (d *darkTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 0x1e, G: 0x1e, B: 0x1e, A: 0xff}
	case theme.ColorNameForeground:
		return color.NRGBA{R: 0xe6, G: 0xe6, B: 0xe6, A: 0xff}
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 0x7a, G: 0x7a, B: 0x7a, A: 0xff}
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 0x0e, G: 0x63, B: 0x9c, A: 0xff}
	case theme.ColorNameHover:
		return color.NRGBA{R: 0x2a, G: 0x2d, B: 0x2e, A: 0xff}
	case theme.ColorNameFocus:
		return color.NRGBA{R: 0x0e, G: 0x63, B: 0x9c, A: 0x66}
	case theme.ColorNameSelection:
		return color.NRGBA{R: 0x09, G: 0x4f, B: 0x82, A: 0xff}
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 0x25, G: 0x25, B: 0x26, A: 0xff}
	case theme.ColorNameInputBorder:
		return color.NRGBA{R: 0x3c, G: 0x3c, B: 0x3c, A: 0xff}
	case theme.ColorNameMenuBackground:
		return color.NRGBA{R: 0x25, G: 0x25, B: 0x26, A: 0xff}
	case theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 0x25, G: 0x25, B: 0x26, A: 0xff}
	case theme.ColorNameSeparator:
		return color.NRGBA{R: 0x3c, G: 0x3c, B: 0x3c, A: 0xff}
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 0x50, G: 0x50, B: 0x50, A: 0xff}
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 0x66}
	case theme.ColorNameButton:
		return color.NRGBA{R: 0x2d, G: 0x2d, B: 0x2d, A: 0xff}
	case theme.ColorNamePlaceHolder:
		return color.NRGBA{R: 0x88, G: 0x88, B: 0x88, A: 0xff}
	}
	return theme.DefaultTheme().Color(name, theme.VariantDark)
}

func (d *darkTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (d *darkTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (d *darkTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 4
	case theme.SizeNameInlineIcon:
		return 16
	case theme.SizeNameText:
		return 13
	case theme.SizeNameInputBorder:
		return 1
	}
	return theme.DefaultTheme().Size(name)
}

// Diff line colors
var (
	ColorAdd    = color.NRGBA{R: 0x0e, G: 0x46, B: 0x2d, A: 0xff}
	ColorDel    = color.NRGBA{R: 0x5a, G: 0x1d, B: 0x1d, A: 0xff}
	ColorHunk   = color.NRGBA{R: 0x26, G: 0x26, B: 0x3d, A: 0xff}
	ColorHeader = color.NRGBA{R: 0x2a, G: 0x2a, B: 0x2a, A: 0xff}
)
