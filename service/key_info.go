package service

import (
	"fmt"
	"strconv"
	"strings"
	"tinyrdm/backend/services"
	"tinyrdm/backend/types"

	"github.com/atotto/clipboard"
	"github.com/awesome-gocui/gocui"
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
	GlobalTipComponent.AppendList(GlobalKeyInfoComponent.name, GlobalKeyInfoComponent.KeyMapTip())
}

func (c *LTRKeyInfoComponent) LayoutTitle() *LTRKeyInfoComponent {
	if c.keyView != nil && GlobalApp.Gui.CurrentView().Name() == c.name {
		c.keyView.Title = " [" + c.title + "] "
		c.keyViewTTL.FrameColor = gocui.ColorGreen
	} else {
		c.keyView.Title = " " + c.title + " "
		c.keyViewTTL.FrameColor = gocui.ColorDefault
	}
	return c
}

func (c *LTRKeyInfoComponent) Layout() *LTRKeyInfoComponent {
	theX0, _ := GlobalDBComponent.view.Size()
	theX0 = theX0 + 2
	var err error
	var theTTL int64
	// show key info
	c.keyView, err = GlobalApp.Gui.SetView(c.name, theX0, 0, GlobalApp.maxX-1, 2, 0)
	if err == nil || err != gocui.ErrUnknownView {
		keySummary := services.Browser().GetKeySummary(types.KeySummaryParam{
			Server: GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
			DB:     GlobalDBComponent.SelectedDB,
			Key:    c.keyName,
		})
		// c.keyView.Title = " " + c.title + " "
		if GlobalApp.Gui.CurrentView().Name() == c.name {
			c.keyView.Title = " [" + c.title + "] "
		} else {
			c.keyView.Title = " " + c.title + " "
		}
		printString := ""
		if keySummary.Success {
			keySummaryData := keySummary.Data.(types.KeySummary)
			printString = NewTypeWord(keySummaryData.Type, "full") + " " + c.keyName
			theTTL = keySummaryData.TTL
		}
		c.keyView.Clear()
		c.keyView.Write([]byte(printString))
	}

	theTTLStr := ""
	theTTLStrType := 0
	if theTTL >= 0 {
		if theTTL > 86400 {
			theTTLStr += fmt.Sprintf("%dDay ", theTTL/86400)
			theTTL = theTTL % 86400
		}
		theTTLStr += fmt.Sprintf("%02d:%02d:%02d", theTTL/3600, (theTTL%3600)/60, theTTL%60)
		theTTLStr = " TTL: " + theTTLStr + ""
	} else {
		theTTLStr = " TTL: " + strconv.FormatInt(theTTL, 10) + " s"
		theTTLStrType = 1
	}
	// show key ttl
	c.keyViewTTL, err = GlobalApp.Gui.SetView(c.name+"_ttl", GlobalApp.maxX-len(theTTLStr)-3, 0, GlobalApp.maxX-1, 2, 0)
	if err == nil || err != gocui.ErrUnknownView {
		c.keyViewTTL.Clear()
		if theTTLStrType == 1 {
			c.keyViewTTL.Write([]byte(NewColorString(theTTLStr+SPACE_STRING, "white", "red", "bold")))
		} else {
			c.keyViewTTL.Write([]byte(NewColorString(theTTLStr+SPACE_STRING, "black", "green", "bold")))
		}
	}
	c.keyViewTTL.FrameRunes = []rune{'─', '│', '─', '┐', '─', '┘'}
	// c.keyViewTTL.Frame = false

	return c
}

func (c *LTRKeyInfoComponent) KeyBind() *LTRKeyInfoComponent {
	// 鼠标点击聚焦
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.MouseLeft}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		GlobalApp.ForceUpdate(c.name)
		return nil
	})

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

	// 粘贴剪切板值到 key 名称
	GuiSetKeysbindingConfirm(GlobalApp.Gui, c.name, []any{'p'}, "Paste from clipboard to rename key?", func() {
		theClipboardValue, err := clipboard.ReadAll()
		if err != nil {
			GlobalTipComponent.LayoutTemporary("Clipboard is empty or not available", 3, TipTypeError)
			return
		}
		if theClipboardValue == c.keyName {
			GlobalTipComponent.LayoutTemporary("The value is the same as the current key", 3, TipTypeWarning)
			return
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
			return
		}
		GlobalTipComponent.LayoutTemporary("Renamed successfully", 3, TipTypeSuccess)
		theOldKeyName := c.keyName
		c.keyName = theClipboardValue
		for i, v := range GlobalKeyComponent.keys {
			if v == theOldKeyName {
				GlobalKeyComponent.keys[i] = theClipboardValue
				break
			}
		}
		GlobalKeyComponent.Layout()
		c.Layout()
	}, func() {
		GlobalTipComponent.LayoutTemporary("Cancel rename key", 3, TipTypeWarning)
	})

	// 粘贴剪切板值到 ttl
	GuiSetKeysbindingConfirm(GlobalApp.Gui, c.name, []any{'x'}, "Paste from clipboard to replace TTL (seconds)?", func() {
		theClipboardValue, err := clipboard.ReadAll()
		if err != nil {
			GlobalTipComponent.LayoutTemporary("Clipboard is empty or not available", 3, TipTypeError)
			return
		}
		//判断剪切板内容是否为数字
		theClipboardValueInt, err := strconv.ParseInt(theClipboardValue, 10, 64)
		if err != nil {
			GlobalTipComponent.LayoutTemporary("Invalid TTL value. Please enter a number", 3, TipTypeError)
			return
		}
		// 修改 ttl
		res := services.Browser().SetKeyTTL(
			GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
			GlobalDBComponent.SelectedDB,
			c.keyName,
			theClipboardValueInt,
		)
		if !res.Success {
			GlobalTipComponent.LayoutTemporary("Failed to set TTL, message: "+res.Msg, 3, TipTypeError)
			return
		}
		GlobalTipComponent.LayoutTemporary("TTL updated successfully", 3, TipTypeSuccess)
		c.Layout()
	}, func() {
		GlobalTipComponent.LayoutTemporary("TTL setting cancelled", 3, TipTypeWarning)
	})

	// 修改 key
	GuiSetKeysbindingConfirmWithVIEditor(GlobalApp.Gui, c.name, []any{'e'}, "", func() string {
		return c.keyName
	}, func(editorResult string) {
		if editorResult == c.keyName {
			GlobalTipComponent.LayoutTemporary("The value is the same as the current key", 3, TipTypeWarning)
			return
		}

		editorResult = strings.TrimSpace(editorResult)
		if editorResult == "" {
			GlobalTipComponent.LayoutTemporary("The value is empty", 3, TipTypeWarning)
			return
		}
		// 修改 key
		res := services.Browser().RenameKey(
			GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
			GlobalDBComponent.SelectedDB,
			c.keyName,
			editorResult,
		)
		if !res.Success {
			GlobalTipComponent.LayoutTemporary("Failed to rename key, message: "+res.Msg, 3, TipTypeError)
			return
		}
		GlobalTipComponent.LayoutTemporary("Renamed successfully", 3, TipTypeSuccess)
		theOldKeyName := c.keyName
		c.keyName = editorResult
		for i, v := range GlobalKeyComponent.keys {
			if v == theOldKeyName {
				GlobalKeyComponent.keys[i] = editorResult
				break
			}
		}
		GlobalKeyComponent.Layout()
		c.Layout()
	}, func() {
		GlobalTipComponent.LayoutTemporary("Cancel rename key", 3, TipTypeWarning)
	}, false)

	// 修改 ttl
	GuiSetKeysbindingConfirmWithVIEditor(GlobalApp.Gui, c.name, []any{'t'}, "Do you want to modify the TTL (seconds)?", func() string {
		// 获取 ttl
		keySummary := services.Browser().GetKeySummary(types.KeySummaryParam{
			Server: GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
			DB:     GlobalDBComponent.SelectedDB,
			Key:    c.keyName,
		})
		theTTL := "-1"
		if keySummary.Success {
			keySummaryData := keySummary.Data.(types.KeySummary)
			theTTL = fmt.Sprintf("%d", keySummaryData.TTL)
		}
		return theTTL
	}, func(editorResult string) {
		editorResult = strings.TrimSpace(editorResult)
		if editorResult == "" {
			GlobalTipComponent.LayoutTemporary("The value is empty", 3, TipTypeWarning)
			return
		}
		//判断剪切板内容是否为数字
		theClipboardValueInt, err := strconv.ParseInt(editorResult, 10, 64)
		if err != nil {
			PrintLn(err.Error())
			GlobalTipComponent.LayoutTemporary("Invalid TTL value. Please enter a number", 3, TipTypeError)
			return
		}
		// 修改 ttl
		res := services.Browser().SetKeyTTL(
			GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
			GlobalDBComponent.SelectedDB,
			c.keyName,
			theClipboardValueInt,
		)
		if !res.Success {
			GlobalTipComponent.LayoutTemporary("Failed to set TTL, message: "+res.Msg, 3, TipTypeError)
			return
		}
		GlobalTipComponent.LayoutTemporary("TTL updated successfully", 3, TipTypeSuccess)
		c.Layout()
	}, func() {
		GlobalTipComponent.LayoutTemporary("TTL setting cancelled", 3, TipTypeWarning)
	}, false)
	return c
}

func (c *LTRKeyInfoComponent) KeyMapTip() string {
	keyMap := []KeyMapStruct{
		{"Switch", "<Tab>"},
		{"Copy", "<C>"},
		{"Edit", "<E>"},
		{"Edit TTL", "<T>"},
		{"Paste", "<P>"},
		{"Paste TTL", "<X>"},
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
