package service

import (
	"fmt"

	"github.com/jroimartin/gocui"
)

type PageComponentConfirm struct {
	name        string
	title       string
	text        string
	callbackYes func()
	callbackNo  func()
}

func NewPageComponentConfirm(title string, text string, callbackYes func(), callbackNo func()) *PageComponentConfirm {
	ret := &PageComponentConfirm{
		name:        "page_confirm",
		title:       title,
		text:        text,
		callbackYes: callbackYes,
		callbackNo:  callbackNo,
	}
	ret.Layout()
	return ret
}

func (c *PageComponentConfirm) Layout() *PageComponentConfirm {
	GlobalApp.Gui.Cursor = false
	theX0 := 0
	theY0 := GlobalApp.maxY/2 - 3
	theX1 := GlobalApp.maxX - 1
	theY1 := theY0 + 5
	if GlobalApp.maxX > GlobalApp.maxY && (theY1*15/10) <= theX1 {
		theX0 = (GlobalApp.maxX - GlobalApp.maxY) / 2
		theX1 = theX0 + GlobalApp.maxY - 1
	}
	v, _ := GlobalApp.Gui.SetView(c.name, theX0, theY0, theX1, theY1)
	v.Title = c.title
	v.Wrap = true
	v.Editable = false
	v.Frame = true

	v.Clear()
	v.Write([]byte(c.text))
	GlobalApp.Gui.SetCurrentView(c.name)
	c.KeyBind()
	if GlobalApp.Gui.CurrentView().Name() == c.name {
		GlobalTipComponent.Layout(c.KeyMapTips())
	}
	return c
}

func (c *PageComponentConfirm) KeyBind() *PageComponentConfirm {
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{'y', 'Y'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.closeView()
		c.callbackYes()
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{'n', 'N'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.closeView()
		c.callbackNo()
		return nil
	})

	return c
}

func (c *PageComponentConfirm) closeView() {
	GlobalApp.Gui.DeleteView(c.name)
	GlobalApp.Gui.DeleteKeybindings(c.name)
}

func (c *PageComponentConfirm) KeyMapTips() string {
	keyMap := []KeyMapStruct{
		{"Yes", "<Y>"},
		{"No", "<N>"},
	}
	ret := ""
	for i, v := range keyMap {
		if i > 0 {
			ret += " | "
		}
		ret += fmt.Sprintf("%s: %s", v.Description, v.Key)
		i++
	}
	return ret
}
