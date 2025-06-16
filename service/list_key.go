package service

import (
	"fmt"
	"strconv"

	"tinyrdm/backend/services"

	"github.com/jroimartin/gocui"
)

type LTRListKeyComponent struct {
	name        string
	title       string
	viewMaxY    int
	view        *gocui.View
	viewOriginY int
	Current     int
	keys        []any
	MaxKeys     int64
}

func InitKeyComponent() {
	GlobalKeyComponent = &LTRListKeyComponent{
		name:     "key_list",
		title:    "Key List",
		viewMaxY: 0,
		view:     nil,
	}
	GlobalKeyComponent.LoadKeys().Layout().KeyBind()
	GlobalApp.ViewNameList = append(GlobalApp.ViewNameList, GlobalKeyComponent.name)
}

func (c *LTRListKeyComponent) LoadKeys() *LTRListKeyComponent {
	// get key list
	if GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name == "" || GlobalDBComponent.SelectedDB < 0 {
		return c
	}
	keysInfo := services.Browser().LoadNextKeys(
		GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
		GlobalDBComponent.SelectedDB,
		"",
		"",
		false,
	)
	if !keysInfo.Success {
		return c
	}
	c.keys = keysInfo.Data.(map[string]any)["keys"].([]any)
	// retEnd := keysInfo.Data.(map[string]any)["end"].(bool)
	c.MaxKeys = keysInfo.Data.(map[string]any)["maxKeys"].(int64)
	return c
}

func (c *LTRListKeyComponent) Layout() *LTRListKeyComponent {
	_, theDBComponentH := GlobalDBComponent.view.Size()
	var err error
	c.view, err = GlobalApp.gui.SetView(c.name, 0, theDBComponentH+2, GlobalApp.maxX*2/10, GlobalApp.maxY-2)
	if err != nil && err != gocui.ErrUnknownView {
		PrintLn(err.Error())
		return c
	}
	c.view.Editable = false
	c.view.Frame = true
	if GlobalDBComponent.SelectedDB < 0 {
		c.view.Title = " [not selected db] "
	} else {
		c.view.Title = " [db" + strconv.Itoa(GlobalDBComponent.SelectedDB) + "]" + " [" + strconv.Itoa(len(c.keys)) + "/" + strconv.FormatInt(c.MaxKeys, 10) + "] "
	}
	_, c.viewMaxY = c.view.Size()

	printString := ""
	currenLine := 0
	totalLine := 0
	if len(c.keys) > 0 {
		for index, key := range c.keys {
			totalLine++
			// keyStr, ok := key.(string)
			// if !ok {
			// 	continue
			// }
			keyStr := fmt.Sprintf("%s", key)
			if c.Current == index {
				currenLine = totalLine
				// get key info
				// keyType := services.Browser().GetKeyType(
				// 	types.KeySummaryParam{
				// 		Server: GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
				// 		DB:     GlobalDBComponent.SelectedDB,
				// 		Key:    key.(string),
				// 	},
				// )
				// theKeyTypeStr := ""
				// if keyType.Success {
				// 	keyTypeData := keyType.Data.(types.KeySummary)
				// 	theKeyTypeStr = keyTypeData.Type
				// }
				// printString += NewTypeWord(theKeyTypeStr) + NewColorString(" "+key.(string)+""+SPACE_STRING+"\n", "white", "blue", "bold")
				printString += NewColorString(strconv.Itoa(totalLine)+"-"+keyStr+""+SPACE_STRING+"\n", "white", "blue", "bold")
			} else {
				printString += fmt.Sprintf("%s\n", strconv.Itoa(totalLine)+"-"+keyStr+""+SPACE_STRING)
			}
		}
	}
	if c.viewOriginY < 0 {
		c.viewOriginY = 0
	}
	if currenLine > c.viewMaxY/2 {
		c.viewOriginY = currenLine - c.viewMaxY/2
	} else {
		c.viewOriginY = 0
	}
	if c.viewOriginY+c.viewMaxY > len(c.keys) {
		c.viewOriginY = len(c.keys) - c.viewMaxY
	}
	c.view.SetOrigin(0, c.viewOriginY)
	c.view.Clear()
	c.view.Write([]byte(printString))
	if GlobalApp.CurrentView == c.name {
		GlobalApp.gui.SetCurrentView(c.name)
	}
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

	GuiSetKeysbinding(GlobalApp.gui, c.name, []any{gocui.MouseWheelDown}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if c.viewOriginY+1 > len(c.keys)-c.viewMaxY {
			// c.viewOriginY = c.viewMaxY
			return nil
		}
		c.viewOriginY++
		c.Current = c.viewOriginY + c.viewMaxY/2 - 1
		c.view.SetOrigin(0, c.viewOriginY)
		return nil
	})
	GuiSetKeysbinding(GlobalApp.gui, c.name, []any{gocui.MouseWheelUp}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.viewOriginY--
		if c.viewOriginY < 0 {
			c.viewOriginY = 0
		}
		c.Current = c.viewOriginY + c.viewMaxY/2 - 1
		c.view.SetOrigin(0, c.viewOriginY)
		return nil
	})

	GuiSetKeysbinding(GlobalApp.gui, c.name, []any{gocui.MouseLeft}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		GlobalApp.CurrentView = c.name
		c.Layout()
		return nil
	})

	GuiSetKeysbinding(GlobalApp.gui, c.name, []any{gocui.KeyEnter, gocui.KeyArrowRight}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		// get key info
		// PrintLn(GlobalKeyComponent.Current)
		// PrintLn(GlobalKeyComponent.keys)
		GlobalKeyInfoComponent.keyName = fmt.Sprintf("%s", GlobalKeyComponent.keys[GlobalKeyComponent.Current])
		// PrintLn(GlobalKeyInfoComponent.keyName)
		GlobalKeyInfoComponent.Layout()
		GlobalKeyInfoDetailComponent.viewOriginY = 0
		GlobalKeyInfoDetailComponent.Layout()
		return nil
	})

	return c
}
