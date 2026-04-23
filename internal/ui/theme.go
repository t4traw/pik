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

// Diff line & gutter colors — VS Code-ish palette.
// Base tints are intentionally subtle so syntax-highlighted text stays readable;
// emphasis colors are vivid so word-level diffs pop against the base tint.
var (
	ColorAdd     = color.NRGBA{R: 0x14, G: 0x30, B: 0x1e, A: 0xff} // subtle green tint
	ColorDel     = color.NRGBA{R: 0x3a, G: 0x1a, B: 0x1a, A: 0xff} // subtle red tint
	ColorAddEmph = color.NRGBA{R: 0x28, G: 0x86, B: 0x36, A: 0xff} // vivid green (word diff)
	ColorDelEmph = color.NRGBA{R: 0xb0, G: 0x3a, B: 0x3a, A: 0xff} // vivid red (word diff)

	ColorAddStripe = color.NRGBA{R: 0x3f, G: 0xa8, B: 0x4a, A: 0xff} // left-edge stripe for added lines
	ColorDelStripe = color.NRGBA{R: 0xd0, G: 0x40, B: 0x40, A: 0xff} // left-edge stripe for removed lines

	ColorHunk   = color.NRGBA{R: 0x26, G: 0x26, B: 0x3d, A: 0xff}
	ColorHeader = color.NRGBA{R: 0x2a, G: 0x2a, B: 0x2a, A: 0xff}

	ColorGutterBG = color.NRGBA{R: 0x1a, G: 0x1a, B: 0x1a, A: 0xff}
	ColorGutterFG = color.NRGBA{R: 0x6a, G: 0x6a, B: 0x6a, A: 0xff}

	ColorDefaultFG = color.NRGBA{R: 0xd4, G: 0xd4, B: 0xd4, A: 0xff}
)

// Syntax highlight palette (VS Code Dark+ style).
var (
	ColorSynKeyword  = color.NRGBA{R: 0x56, G: 0x9c, B: 0xd6, A: 0xff} // blue
	ColorSynType     = color.NRGBA{R: 0x4e, G: 0xc9, B: 0xb0, A: 0xff} // teal
	ColorSynString   = color.NRGBA{R: 0xce, G: 0x91, B: 0x78, A: 0xff} // orange-brown
	ColorSynNumber   = color.NRGBA{R: 0xb5, G: 0xce, B: 0xa8, A: 0xff} // light green
	ColorSynComment  = color.NRGBA{R: 0x6a, G: 0x99, B: 0x55, A: 0xff} // green
	ColorSynFunc     = color.NRGBA{R: 0xdc, G: 0xdc, B: 0xaa, A: 0xff} // pale yellow
	ColorSynVar      = color.NRGBA{R: 0x9c, G: 0xdc, B: 0xfe, A: 0xff} // light blue
	ColorSynOp       = color.NRGBA{R: 0xd4, G: 0xd4, B: 0xd4, A: 0xff} // default fg
	ColorSynPunct    = color.NRGBA{R: 0xd4, G: 0xd4, B: 0xd4, A: 0xff}
	ColorSynConst    = color.NRGBA{R: 0x4f, G: 0xc1, B: 0xff, A: 0xff} // brighter blue
	ColorSynBuiltin  = color.NRGBA{R: 0xdc, G: 0xdc, B: 0xaa, A: 0xff}
	ColorSynTag      = color.NRGBA{R: 0x56, G: 0x9c, B: 0xd6, A: 0xff}
	ColorSynAttr     = color.NRGBA{R: 0x9c, G: 0xdc, B: 0xfe, A: 0xff}
	ColorSynEscape   = color.NRGBA{R: 0xd7, G: 0xba, B: 0x7d, A: 0xff}
	ColorSynError    = color.NRGBA{R: 0xf4, G: 0x47, B: 0x47, A: 0xff}
	ColorSynNamespc  = color.NRGBA{R: 0x4e, G: 0xc9, B: 0xb0, A: 0xff}
	ColorSynOther    = color.NRGBA{R: 0xd4, G: 0xd4, B: 0xd4, A: 0xff}
)
