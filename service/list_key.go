package service

import (
	"fmt"
	"strconv"

	"tinyrdm/backend/services"
	"tinyrdm/backend/types"

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
	// GlobalApp.ViewNameList = append(GlobalApp.ViewNameList, c.name)
	c.LoadKeys().Layout().KeyBind()
	if GlobalApp.CurrentView == c.name {
		GlobalApp.gui.SetCurrentView(c.name)
	}
	return c
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

	if c.view == nil {
		v, err := GlobalApp.gui.SetView(c.name, 0, GlobalApp.maxY*2/10+1, GlobalApp.maxX*2/10, GlobalApp.maxY-1)
		if err != nil && err != gocui.ErrUnknownView {
			PrintLn(err.Error())
			return c
		}
		v.Editable = false
		v.Frame = true
		if GlobalDBComponent.SelectedDB < 0 {
			v.Title = " [not selected db] "
		} else {
			v.Title = " [db" + strconv.Itoa(GlobalDBComponent.SelectedDB) + "]" + " [count:" + strconv.FormatInt(c.MaxKeys, 10) + "] "
		}
		c.view = v
	}
	_, c.LayoutMaxY = c.view.Size()

	printString := ""
	currenLine := 0
	totalLine := 0
	if len(c.keys) > 0 {
		for index, key := range c.keys {
			if c.Current == index {
				// get key info
				keyType := services.Browser().GetKeyType(
					types.KeySummaryParam{
						Server: GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
						DB:     GlobalDBComponent.SelectedDB,
						Key:    key.(string),
					},
				)
				theKeyTypeStr := ""
				if keyType.Success {
					keyTypeData := keyType.Data.(types.KeySummary)
					theKeyTypeStr = keyTypeData.Type
				}
				printString += NewTypeWord(theKeyTypeStr) + NewColorString(" "+key.(string)+""+SPACE_STRING+"\n", "white", "blue", "bold")
			} else {
				printString += fmt.Sprintf("%s\n", ""+key.(string)+""+SPACE_STRING)
			}
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
		c.view.SetOrigin(0, originLine)
	} else {
		c.view.SetOrigin(0, 0)
	}
	c.view.Clear()
	c.view.Write([]byte(printString))
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
