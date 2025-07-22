package service

import (
	"time"

	"github.com/jroimartin/gocui"
)

type LTRTipComponent struct {
	name               string
	view               *gocui.View
	lastTipString      string
	temporaryTipString string
}

const (
	TipTypeWarning string = "warning"
	TipTypeError   string = "error"
	TipTypeSuccess string = "success"
)

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
	if tipString == c.lastTipString {
		return c
	}
	if tipString != "" {
		c.lastTipString = tipString
	}

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

	theTipString := c.lastTipString
	if c.temporaryTipString != "" {
		theTipString = c.temporaryTipString
	}
	c.view.Write([]byte(theTipString))
	return c
}

func (c *LTRTipComponent) LayoutTemporary(tipString string, durationSec int, tipType string) *LTRTipComponent {
	switch tipType {
	case TipTypeWarning:
		tipString = NewColorString(tipString, "yellow")
	case TipTypeError:
		tipString = NewColorString(tipString, "red")
	case TipTypeSuccess:
		tipString = NewColorString(tipString, "green")
	}
	//获取当前的view的内容
	if c.lastTipString == tipString {
		return c
	}
	c.temporaryTipString = tipString
	// 修改展示的内容
	// c.Layout("")
	GlobalApp.Gui.Update(func(g *gocui.Gui) error {
		c.Layout("")
		return nil
	})
	// 3 s 后恢复原内容
	go func() {
		time.Sleep(time.Second * time.Duration(durationSec))
		GlobalApp.Gui.Update(func(g *gocui.Gui) error {
			c.temporaryTipString = ""
			c.Layout("")
			return nil
		})
	}()
	return c
}
