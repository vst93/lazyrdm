package service

import (
	"fmt"
	"strconv"

	"tinyrdm/backend/services"

	"github.com/jroimartin/gocui"
)

type LTRListKeyComponent struct {
	name       string
	title      string
	LayoutMaxY int
	view       *gocui.View
	Current    int
	keys       []any
	MaxKeys    int64
}

func InitKeyComponent() *LTRListKeyComponent {
	c := &LTRListKeyComponent{
		name:       "key_list",
		title:      "Key List",
		LayoutMaxY: 0,
		view:       nil,
	}
	c.Layout()
	return c
}

func (c *LTRListKeyComponent) LoadKeys() *LTRListKeyComponent {
	// get key list
	keysInfo := services.Browser().LoadNextKeys(
		GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
		GlobalDBComponent.SelectedDB,
		"",
		"",
		false,
	)
	c.keys = keysInfo.Data.(map[string]any)["keys"].([]any)
	// retEnd := keysInfo.Data.(map[string]any)["end"].(bool)
	c.MaxKeys = keysInfo.Data.(map[string]any)["maxKeys"].(int64)
	return c
}

func (c *LTRListKeyComponent) Layout() *LTRListKeyComponent {

	if len(c.keys) == 0 {
		return c
	}
	v, err := GlobalApp.gui.SetView(c.name, 0, GlobalApp.maxY*2/10+1, GlobalApp.maxX*2/10, GlobalApp.maxY-1)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return c
		}
		v.Title = c.title + " (" + strconv.FormatInt(c.MaxKeys, 10) + ")"
		v.Editable = false
		v.Frame = true
		_, c.LayoutMaxY = v.Size()

		GlobalApp.gui.SetCurrentView(c.name)
	}

	printString := ""
	currenLine := 0
	totalLine := 0
	for index, key := range c.keys {
		if c.Current == index {
			printString += NewColorString("["+key.(string)+"]"+SPACE_STRING+"\n", "white", "blue", "bold")
		} else {
			printString += fmt.Sprintf("%s\n", ""+key.(string)+""+SPACE_STRING)
		}
	}
	if currenLine > c.LayoutMaxY/2 {
		originLine := currenLine - c.LayoutMaxY/2
		if originLine < 0 {
			originLine = 0
		}
		if originLine > totalLine-c.LayoutMaxY {
			originLine = totalLine - c.LayoutMaxY
		}
		v.SetOrigin(0, originLine)
	} else {
		v.SetOrigin(0, 0)
	}
	v.Clear()
	v.Write([]byte(printString))
	c.view = v
	return c
}

func (c *LTRListKeyComponent) KeyBind() *LTRListKeyComponent {
	GuiSetKeysbinding(GlobalApp.gui, c.name, []any{gocui.KeyArrowDown}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.Current++
		if c.Current > len(c.keys)-1 {
			c.Current = 0
		}
		v.Clear()
		c.Layout()
		return nil
	})

	GuiSetKeysbinding(GlobalApp.gui, c.name, []any{gocui.KeyArrowUp}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.Current--
		if c.Current < 0 {
			c.Current = len(c.keys) - 1
		}
		v.Clear()
		c.Layout()
		return nil
	})

	return c
}
