package service

import (
	"fmt"
	"strconv"

	"tinyrdm/backend/services"

	"github.com/jroimartin/gocui"
)

type LTRListKeyComponent struct {
	name          string
	title         string
	viewMaxY      int
	view          *gocui.View
	Current       int
	keys          []any
	MaxKeys       int64
	IsEnd         bool
	pendDeleteKey string
}

func InitKeyComponent() {
	GlobalKeyComponent = &LTRListKeyComponent{
		name:     "key_list",
		title:    "Keys",
		viewMaxY: 0,
		view:     nil,
	}
	GlobalKeyComponent.LoadKeys().Layout().KeyBind()
	GlobalApp.ViewNameList = append(GlobalApp.ViewNameList, GlobalKeyComponent.name)
}

func (c *LTRListKeyComponent) LoadKeys() *LTRListKeyComponent {
	if c.IsEnd {
		return c
	}
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
	theList := keysInfo.Data.(map[string]any)["keys"].([]any)
	if len(theList) > 0 {
		c.keys = append(c.keys, theList...)
	}
	// c.keys = keysInfo.Data.(map[string]any)["keys"].([]any)
	retEnd := keysInfo.Data.(map[string]any)["end"].(bool)
	c.IsEnd = retEnd
	c.MaxKeys = keysInfo.Data.(map[string]any)["maxKeys"].(int64)
	// PrintLn(retEnd)
	// PrintLn(c.MaxKeys)
	return c
}

func (c *LTRListKeyComponent) Layout() *LTRListKeyComponent {
	_, theDBComponentH := GlobalDBComponent.view.Size()
	var err error
	// 列表
	c.view, err = GlobalApp.Gui.SetView(c.name, 0, theDBComponentH+2, GlobalApp.maxX*2/10, GlobalApp.maxY-2)
	if err != nil && err != gocui.ErrUnknownView {
		PrintLn(err.Error())
		return c
	}
	c.view.Editable = false
	c.view.Frame = true
	if GlobalDBComponent.SelectedDB < 0 {
		c.view.Title = " Key List "
	} else {
		// c.view.Title = " [db" + strconv.Itoa(GlobalDBComponent.SelectedDB) + "]" + " [" + strconv.Itoa(len(c.keys)) + "/" + strconv.FormatInt(c.MaxKeys, 10) + "] "
		c.view.Title = " Key List [" + strconv.Itoa(len(c.keys)) + "/" + strconv.FormatInt(c.MaxKeys, 10) + "] "
	}
	_, c.viewMaxY = c.view.Size()

	printString := ""
	// currenLine := 0
	// totalLine := 0
	rangeBegin := c.Current - c.viewMaxY/2 + 1
	if rangeBegin < 0 {
		rangeBegin = 0
	}
	rangeEnd := rangeBegin + c.viewMaxY
	if rangeEnd > len(c.keys) {
		rangeEnd = len(c.keys)
		rangeBegin = rangeEnd - c.viewMaxY
		if rangeBegin < 0 {
			rangeBegin = 0
		}
	}
	if len(c.keys) > 0 {
		splitKeys := c.keys[rangeBegin:rangeEnd]
		for index, key := range splitKeys {
			index = index + rangeBegin
			// totalLine++
			keyStr := fmt.Sprintf("%s", key)
			if c.Current == index {
				// currenLine = totalLine
				printString += NewColorString(strconv.Itoa(index)+"-"+keyStr+""+SPACE_STRING+"\n", "white", "blue", "bold")
			} else {
				printString += fmt.Sprintf("%s\n", strconv.Itoa(index)+"-"+keyStr+""+SPACE_STRING)
			}
		}
		if c.Current >= (len(c.keys) - 1) {
			c.LoadKeys()
		}
	}

	c.view.Clear()
	c.view.Write([]byte(printString))
	// if GlobalApp.Gui.CurrentView().Name() == c.name {
	// 	GlobalTipComponent.Layout(c.KeyMapTip())
	// }
	GlobalTipComponent.AppendList(c.name, c.KeyMapTip())

	return c
}

func (c *LTRListKeyComponent) KeyBind() *LTRListKeyComponent {
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.KeyArrowDown, gocui.MouseWheelDown}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.Current++
		if c.Current > len(c.keys)-1 {
			c.Current = 0
		}
		v.Clear()
		c.Layout()
		return nil
	})

	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.KeyArrowUp, gocui.MouseWheelUp}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.Current--
		if c.Current < 0 {
			c.Current = len(c.keys) - 1
		}
		v.Clear()
		c.Layout()
		return nil
	})

	// GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.MouseLeft}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
	// 	GlobalApp.Gui.SetCurrentView(c.name)
	// 	c.Layout()
	// 	return nil
	// })

	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.KeyEnter, gocui.KeyArrowRight}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if GlobalKeyComponent.Current < 0 || GlobalKeyComponent.Current > len(GlobalKeyComponent.keys)-1 {
			return nil
		}
		GlobalKeyInfoComponent.keyName = fmt.Sprintf("%s", GlobalKeyComponent.keys[GlobalKeyComponent.Current])
		// PrintLn(GlobalKeyInfoComponent.keyName)
		GlobalKeyInfoComponent.Layout()
		GlobalKeyInfoDetailComponent.viewOriginY = 0
		GlobalKeyInfoDetailComponent.Layout()
		return nil
	})

	// 鼠标点击聚焦
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.MouseLeft}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		GlobalApp.ForceUpdate(c.name)
		return nil
	})

	// 删除 key
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.KeyDelete, 'd'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if GlobalKeyComponent.Current < 0 || GlobalKeyComponent.Current > len(GlobalKeyComponent.keys)-1 {
			return nil
		}
		c.pendDeleteKey = fmt.Sprintf("%s", GlobalKeyComponent.keys[GlobalKeyComponent.Current])
		GlobalTipComponent.LayoutTemporary("Do you want to delete key [ "+c.pendDeleteKey+"] ? (y/n)", 10, TipTypeWarning)
		return nil
	})

	// 删除 key - n
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{'n'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if c.pendDeleteKey == "" {
			return nil
		}
		GlobalTipComponent.LayoutTemporary("Cancel deleting key", 2, TipTypeWarning)
		return nil
	})

	// 删除 key - y
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{'y'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if c.pendDeleteKey == "" {
			return nil
		}
		deleteStaus := false
		for i, key := range GlobalKeyComponent.keys {
			if key == c.pendDeleteKey {
				services.Browser().DeleteKey(
					GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
					GlobalDBComponent.SelectedDB,
					c.pendDeleteKey,
					true,
				)
				GlobalKeyComponent.keys = append(GlobalKeyComponent.keys[:i], GlobalKeyComponent.keys[i+1:]...)
				deleteStaus = true
				break
			}
		}
		if deleteStaus {
			GlobalTipComponent.LayoutTemporary("Key [ "+c.pendDeleteKey+"] deleted", 2, TipTypeSuccess)
		} else {
			GlobalTipComponent.LayoutTemporary("Key [ "+c.pendDeleteKey+"] not found", 2, TipTypeError)
		}
		c.pendDeleteKey = ""
		c.Layout()
		return nil
	})

	return c
}

func (c *LTRListKeyComponent) KeyMapTip() string {
	keyMap := []KeyMapStruct{
		{"Switch", "<Tab>"},
		{"Select", "↑/↓"},
		{"Enter", "<Enter>/→"},
		{"Delete", "<D>"},
	}
	ret := ""
	for i, v := range keyMap {
		if i > 0 {
			ret += " | "
		}
		ret += fmt.Sprintf("%s: %s", v.Description, v.Key)
	}
	// return "key_list: " + ret
	return ret
}
