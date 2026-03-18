package service

import (
	"fmt"
	"strings"

	"github.com/awesome-gocui/gocui"
)

type PageComponentConfirm struct {
	name        string
	title       string
	text        string
	returnView  string
	callbackYes func()
	callbackNo  func()
}

func NewPageComponentConfirm(title string, text string, callbackYes func(), callbackNo func()) *PageComponentConfirm {
	returnView := ""
	if currentView := GlobalApp.Gui.CurrentView(); currentView != nil && currentView.Name() != "page_confirm" {
		returnView = currentView.Name()
	}
	ret := &PageComponentConfirm{
		name:        "page_confirm",
		title:       title,
		text:        text,
		returnView:  returnView,
		callbackYes: callbackYes,
		callbackNo:  callbackNo,
	}
	ret.Layout()
	return ret
}

func (c *PageComponentConfirm) Layout() *PageComponentConfirm {
	GlobalApp.Gui.Cursor = false
	theX0 := 0
	textLineCount := len(strings.Split(c.text, "\n"))
	viewHeight := textLineCount + 5
	if viewHeight < 8 {
		viewHeight = 8
	}
	if viewHeight > GlobalApp.maxY-2 {
		viewHeight = GlobalApp.maxY - 2
	}
	theY0 := (GlobalApp.maxY - viewHeight) / 2
	if theY0 < 0 {
		theY0 = 0
	}
	theX1 := GlobalApp.maxX - 1
	theY1 := theY0 + viewHeight
	if GlobalApp.maxX > GlobalApp.maxY && (theY1*15/10) <= theX1 {
		theX0 = (GlobalApp.maxX - GlobalApp.maxY) / 2
		theX1 = theX0 + GlobalApp.maxY - 1
	}
	v, _ := SetViewSafe(c.name, theX0, theY0, theX1, theY1, 0)
	v.Title = " " + c.title + " "
	v.Wrap = true
	v.Editable = false
	v.Frame = true

	v.Clear()
	v.Write([]byte(c.text + "\n\nPress y to confirm, n to cancel."))
	GlobalApp.Gui.SetCurrentView(c.name)
	c.KeyBind()
	GlobalTipComponent.AppendList(c.name, c.KeyMapTips())
	return c
}

func (c *PageComponentConfirm) KeyBind() *PageComponentConfirm {
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{'y'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.closeView()
		c.callbackYes()
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{'n'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.closeView()
		c.callbackNo()
		return nil
	})

	return c
}

func (c *PageComponentConfirm) closeView() {
	GlobalApp.Gui.DeleteView(c.name)
	GlobalApp.Gui.DeleteKeybindings(c.name)
	if c.returnView != "" {
		if _, err := GlobalApp.Gui.SetCurrentView(c.returnView); err == nil {
			return
		}
	}
	views := GlobalApp.Gui.Views()
	if len(views) == 0 {
		return
	}
	for _, view := range views {
		if view != nil {
			GlobalApp.Gui.SetCurrentView(view.Name())
			return
		}
	}
}

func (c *PageComponentConfirm) KeyMapTips() string {
	keyMap := []KeyMapStruct{
		{"Yes", "<y>"},
		{"No", "<n>"},
	}
	ret := ""
	for i, v := range keyMap {
		if i > 0 {
			ret += " | "
		}
		ret += fmt.Sprintf("%s: %s", v.Description, v.Key)
	}
	return ret
}
