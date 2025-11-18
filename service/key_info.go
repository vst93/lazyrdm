package service

import (
	"fmt"
	"strconv"
	"tinyrdm/backend/services"
	"tinyrdm/backend/types"

	"github.com/atotto/clipboard"
	"github.com/jroimartin/gocui"
)

type LTRKeyInfoComponent struct {
	name       string
	title      string
	keyName    string
	LayoutMaxY int
	keyView    *gocui.View
	keyViewTTL *gocui.View
}

func InitKeyInfoComponent() {
	GlobalKeyInfoComponent = &LTRKeyInfoComponent{
		name:       "key_info",
		title:      "Info",
		LayoutMaxY: 0,
	}
	GlobalKeyInfoComponent.Layout().KeyBind()
	GlobalApp.ViewNameList = append(GlobalApp.ViewNameList, GlobalKeyInfoComponent.name)
}

func (c *LTRKeyInfoComponent) Layout() *LTRKeyInfoComponent {
	theX0, _ := GlobalDBComponent.view.Size()
	theX0 = theX0 + 2
	var err error
	var theTTL int64
	// show key info
	c.keyView, err = GlobalApp.Gui.SetView(c.name, theX0, 0, GlobalApp.maxX-25, 2)
	if err == nil || err != gocui.ErrUnknownView {
		keySummary := services.Browser().GetKeySummary(types.KeySummaryParam{
			Server: GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
			DB:     GlobalDBComponent.SelectedDB,
			Key:    c.keyName,
		})
		c.keyView.Title = " " + c.title + " "
		printString := ""
		if keySummary.Success {
			keySummaryData := keySummary.Data.(types.KeySummary)
			printString = NewTypeWord(keySummaryData.Type, "full") + " " + c.keyName
			theTTL = keySummaryData.TTL
		}
		c.keyView.Clear()
		c.keyView.Write([]byte(printString))
	}

	// show key ttl
	c.keyViewTTL, err = GlobalApp.Gui.SetView(c.name+"_ttl", GlobalApp.maxX-24, 0, GlobalApp.maxX-1, 2)
	if err == nil || err != gocui.ErrUnknownView {
		c.keyViewTTL.Clear()
		if theTTL >= 0 {
			c.keyViewTTL.Write([]byte(NewColorString("TTL: "+strconv.FormatInt(theTTL, 10)+" s"+SPACE_STRING, "black", "green", "bold")))
		} else {
			c.keyViewTTL.Write([]byte(NewColorString("TTL: "+strconv.FormatInt(theTTL, 10)+" s"+SPACE_STRING, "white", "red", "bold")))
		}
	}

	// show key detail
	// if GlobalApp.Gui.CurrentView().Name() == GlobalKeyInfoComponent.name {
	// 	// GlobalApp.Gui.SetCurrentView(GlobalKeyInfoComponent.name)
	// 	GlobalTipComponent.Layout(c.KeyMapTip())
	// }
	GlobalTipComponent.AppendList(c.name, c.KeyMapTip())

	return c
}

func (c *LTRKeyInfoComponent) KeyBind() *LTRKeyInfoComponent {
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{'c'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		// copy key value
		if c.keyName == "" {
			GlobalTipComponent.LayoutTemporary("No data to copy", 2, TipTypeWarning)
			return nil
		}
		clipboard.WriteAll(c.keyName)
		GlobalTipComponent.LayoutTemporary("Copied to clipboard", 2, TipTypeSuccess)
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{'v'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		theClipboardValue, err := clipboard.ReadAll()
		if err != nil {
			GlobalTipComponent.LayoutTemporary("Clipboard is empty or not available", 3, TipTypeError)
			return nil
		}
		if theClipboardValue == c.keyName {
			GlobalTipComponent.LayoutTemporary("The value is the same as the current key", 3, TipTypeWarning)
			return nil
		}
		// 修改 key
		res := services.Browser().RenameKey(
			GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
			GlobalDBComponent.SelectedDB,
			c.keyName,
			theClipboardValue,
		)
		if !res.Success {
			GlobalTipComponent.LayoutTemporary("Failed to rename key, message: "+res.Msg, 3, TipTypeError)
			return nil
		}
		GlobalTipComponent.LayoutTemporary("Renamed successfully", 3, TipTypeSuccess)
		c.keyName = theClipboardValue
		c.Layout()

		return nil
	})

	return c
}

func (c *LTRKeyInfoComponent) KeyMapTip() string {
	keyMap := []KeyMapStruct{
		{"Switch", "<Tab>"},
		{"Copy", "<C>"},
		{"Paste", "<V>"},
	}
	ret := ""
	for i, v := range keyMap {
		if i > 0 {
			ret += " | "
		}
		ret += fmt.Sprintf("%s: %s", v.Description, v.Key)
	}
	// return "key_info: " + ret
	return ret
}
