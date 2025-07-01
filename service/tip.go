package service

import "github.com/jroimartin/gocui"

type LTRTipComponent struct {
	name string
	view *gocui.View
}

type KeyMapStruct struct {
	Description string
	Key         string
}

func InitTipComponent() {
	GlobalTipComponent = &LTRTipComponent{
		name: "key_map_tip",
	}
	GlobalTipComponent.Layout("")
}

func (c *LTRTipComponent) Layout(tipString string) *LTRTipComponent {
	var err error
	c.view, err = GlobalApp.Gui.SetView(c.name, 0, GlobalApp.maxY-2, GlobalApp.maxX, GlobalApp.maxY)
	if err != nil && err != gocui.ErrUnknownView {
		PrintLn(err.Error())
		return c
	}
	c.view.Editable = false
	c.view.Frame = false
	c.view.Wrap = true
	c.view.FgColor = gocui.ColorBlue
	c.view.Clear()
	c.view.Write([]byte(tipString))
	return c
}
