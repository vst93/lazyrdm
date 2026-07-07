package service

import (
	"github.com/gdamore/tcell/v2"
	_ "unsafe" // required for go:linkname
)

// gocuiScreen is linked to the unexported `screen` variable in
// github.com/awesome-gocui/gocui (tcell_driver.go:11).
//
// gocui does not expose its tcell.Screen, but we need to call
// SetCursorStyle() so the cursor style survives tcell's per-frame
// redraw. tcell's showCursor() emits the escape sequence for the
// configured cursorStyle on every Show() — so a raw \033[5 q written
// to stderr gets immediately overwritten by \033[0 q on the next frame.
//
//go:linkname gocuiScreen github.com/awesome-gocui/gocui.screen
var gocuiScreen tcell.Screen
