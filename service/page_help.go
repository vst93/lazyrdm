package service

import (
	"github.com/awesome-gocui/gocui"
)

type PageComponentHelp struct {
	name       string
	title      string
	text       string
	returnView string
	originY    int
	view       *gocui.View
}

var GlobalHelpPageComponent *PageComponentHelp

func OpenHelpPage() {
	if GlobalApp == nil || GlobalApp.Gui == nil || GlobalTipComponent == nil {
		return
	}

	currentView := GlobalApp.Gui.CurrentView()
	if currentView == nil {
		return
	}
	if currentView.Name() == "page_help" {
		return
	}
	if currentView.Name() == "page_confirm" {
		return
	}

	component := &PageComponentHelp{
		name:       "page_help",
		title:      "Help",
		text:       GlobalTipComponent.BuildHelpText(currentView.Name()),
		returnView: currentView.Name(),
		originY:    0,
	}
	GlobalHelpPageComponent = component
	component.Layout().KeyBind()
}

func (c *PageComponentHelp) Layout() *PageComponentHelp {
	if GlobalApp == nil || GlobalApp.Gui == nil {
		return c
	}

	GlobalApp.Gui.Cursor = false
	x0 := 2
	y0 := 1
	x1 := GlobalApp.maxX - 3
	y1 := GlobalApp.maxY - 3
	if x1 <= x0 {
		x0 = 0
		x1 = GlobalApp.maxX - 1
	}
	if y1 <= y0 {
		y0 = 0
		y1 = GlobalApp.maxY - 1
	}

	v, err := SetViewSafe(c.name, x0, y0, x1, y1, 1)
	if err != nil && err != gocui.ErrUnknownView {
		return c
	}
	v.Title = " [HELP] " + c.title + " "
	v.Subtitle = " Read-only | Scroll: Wheel/Arrows | Close: Esc/q/? "
	v.Wrap = false
	v.Editable = false
	v.Frame = true
	v.FrameColor = themeFrameDialog
	v.Clear()
	v.Write([]byte(c.text))
	v.SetOrigin(0, c.originY)
	GlobalApp.Gui.SetCurrentView(c.name)
	c.view = v
	return c
}

func (c *PageComponentHelp) KeyBind() *PageComponentHelp {
	GlobalApp.Gui.DeleteKeybindings(c.name)

	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.KeyArrowDown, 'j'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.scroll(1)
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.MouseWheelDown}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.scroll(3)
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.KeyArrowUp, 'k'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.scroll(-1)
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.MouseWheelUp}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.scroll(-3)
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.KeyArrowRight, 'l'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.scroll(8)
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.KeyArrowLeft, 'h'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.scroll(-8)
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{'q', '?', gocui.KeyEsc}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.closeView()
		return nil
	})
	return c
}

func (c *PageComponentHelp) scroll(delta int) {
	if c.view == nil {
		return
	}
	next := c.originY + delta
	if next < 0 {
		next = 0
	}
	lines := c.view.BufferLines()
	_, viewHeight := c.view.Size()
	maxOrigin := len(lines) - viewHeight
	if maxOrigin < 0 {
		maxOrigin = 0
	}
	if next > maxOrigin {
		next = maxOrigin
	}
	c.originY = next
	c.view.SetOrigin(0, c.originY)
}

func (c *PageComponentHelp) closeView() {
	GlobalApp.Gui.DeleteView(c.name)
	GlobalApp.Gui.DeleteKeybindings(c.name)
	GlobalHelpPageComponent = nil

	if c.returnView != "" {
		if _, err := GlobalApp.Gui.SetCurrentView(c.returnView); err == nil {
			GlobalTipComponent.LayComponentTips()
			return
		}
	}

	if _, err := GlobalApp.Gui.SetCurrentView("connection_list"); err == nil {
		GlobalTipComponent.LayoutTemporary("Returned to connection list", 2, TipTypeWarning)
		GlobalTipComponent.LayComponentTips()
	}
}
