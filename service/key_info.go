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
	if c.keyView != nil && CurrentViewName() == c.name {
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
	c.keyView, err = SetViewSafe(c.name, theX0, 0, GlobalApp.maxX-1, 2, 0)
	if err != nil && err != gocui.ErrUnknownView {
		return c
	}
	c.keyView.TitleColor = gocui.ColorCyan
	c.keyView.FrameRunes = frameSolid
	keySummary := services.Browser().GetKeySummary(types.KeySummaryParam{
		Server: GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
		DB:     GlobalDBComponent.SelectedDB,
		Key:    c.keyName,
	})
	if CurrentViewName() == c.name {
		c.keyView.Title = " [" + c.title + "] "
	} else {
		c.keyView.Title = " " + c.title + " "
	}
	printString := ""
	if keySummary.Success {
		keySummaryData := keySummary.Data.(types.KeySummary)
		typeWord := NewTypeWord(keySummaryData.Type, "full")
		printString = typeWord + " " + c.keyName
		sizeStr := ""
		if keySummaryData.Length > 0 {
			sizeStr = fmt.Sprintf("len=%d", keySummaryData.Length)
		}
		if keySummaryData.Size > 0 {
			if sizeStr != "" {
				sizeStr += ", "
			}
			sizeStr += fmt.Sprintf("size=%d", keySummaryData.Size)
		}
		if sizeStr != "" {
			printString += "  [" + sizeStr + "]"
		}
		theTTL = keySummaryData.TTL
	}
	c.keyView.Clear()
	c.keyView.Write([]byte(printString))

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
	c.keyViewTTL, err = SetViewSafe(c.name+"_ttl", GlobalApp.maxX-len(theTTLStr)-3, 0, GlobalApp.maxX-1, 2, 0)
	if err == nil || err != gocui.ErrUnknownView {
		c.keyViewTTL.Clear()
		if theTTLStrType == 1 {
			c.keyViewTTL.Write([]byte(NewColorString(theTTLStr+SPACE_STRING, "white", "red", "bold")))
		} else {
			c.keyViewTTL.Write([]byte(NewColorString(theTTLStr+SPACE_STRING, "black", "green", "bold")))
		}
	}
	c.keyViewTTL.FrameRunes = frameHalfTR
	// c.keyViewTTL.Frame = false
	if CurrentViewName() == c.name && GlobalTipComponent != nil {
		GlobalTipComponent.Layout(c.KeyMapTip())
	}

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
			GlobalTipComponent.LayoutTemporary("No key selected to copy", 2, TipTypeWarning)
			return nil
		}
		clipboard.WriteAll(c.keyName)
		GlobalTipComponent.LayoutTemporary("Copied key name to clipboard", 2, TipTypeSuccess)
		return nil
	})

	// 粘贴剪切板值到 key 名称
	GuiSetKeysbindingConfirm(GlobalApp.Gui, c.name, []any{'p'}, "Rename key using clipboard content?", func() {
		theClipboardValue, err := clipboard.ReadAll()
		if err != nil {
			GlobalTipComponent.LayoutTemporary("Clipboard is empty or unavailable", 3, TipTypeError)
			return
		}
		if theClipboardValue == c.keyName {
			GlobalTipComponent.LayoutTemporary("Clipboard value matches current key name", 3, TipTypeWarning)
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
			GlobalTipComponent.LayoutTemporary("Failed to rename key: "+res.Msg, 3, TipTypeError)
			return
		}
		GlobalTipComponent.LayoutTemporary("Key renamed successfully", 3, TipTypeSuccess)
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
		GlobalTipComponent.LayoutTemporary("Rename key cancelled", 3, TipTypeWarning)
	})

	// 粘贴剪切板值到 ttl
	GuiSetKeysbindingConfirm(GlobalApp.Gui, c.name, []any{'x'}, "Replace TTL (seconds) using clipboard content?", func() {
		theClipboardValue, err := clipboard.ReadAll()
		if err != nil {
			GlobalTipComponent.LayoutTemporary("Clipboard is empty or unavailable", 3, TipTypeError)
			return
		}
		//判断剪切板内容是否为数字
		theClipboardValueInt, err := strconv.ParseInt(theClipboardValue, 10, 64)
		if err != nil {
			GlobalTipComponent.LayoutTemporary("Invalid TTL: please enter an integer in seconds", 3, TipTypeError)
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
			GlobalTipComponent.LayoutTemporary("Failed to update TTL: "+res.Msg, 3, TipTypeError)
			return
		}
		GlobalTipComponent.LayoutTemporary("TTL updated", 3, TipTypeSuccess)
		c.Layout()
	}, func() {
		GlobalTipComponent.LayoutTemporary("TTL update cancelled", 3, TipTypeWarning)
	})

	// 修改 key
	GuiSetKeysbindingInlineInput(GlobalApp.Gui, c.name, []any{'e'}, "Rename Key", "New key name", func() string {
		return c.keyName
	}, func(editorResult string) {
		if editorResult == c.keyName {
			GlobalTipComponent.LayoutTemporary("New key name matches current key name", 3, TipTypeWarning)
			return
		}

		if editorResult == "" {
			GlobalTipComponent.LayoutTemporary("Key name cannot be empty", 3, TipTypeWarning)
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
			GlobalTipComponent.LayoutTemporary("Failed to rename key: "+res.Msg, 3, TipTypeError)
			return
		}
		GlobalTipComponent.LayoutTemporary("Key renamed successfully", 3, TipTypeSuccess)
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
		GlobalTipComponent.LayoutTemporary("Rename key cancelled", 3, TipTypeWarning)
	}, func() bool {
		if strings.TrimSpace(c.keyName) == "" {
			GlobalTipComponent.LayoutTemporary("No key selected", 3, TipTypeWarning)
			return false
		}
		return true
	})

	// 修改 ttl
	GuiSetKeysbindingInlineInput(GlobalApp.Gui, c.name, []any{'t'}, "Edit TTL", "TTL (seconds, -1 = no expire)", func() string {
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
		if editorResult == "" {
			GlobalTipComponent.LayoutTemporary("TTL cannot be empty", 3, TipTypeWarning)
			return
		}
		//判断内容是否为数字
		theClipboardValueInt, err := strconv.ParseInt(editorResult, 10, 64)
		if err != nil {
			PrintLn(err.Error())
			GlobalTipComponent.LayoutTemporary("Invalid TTL: please enter an integer in seconds", 3, TipTypeError)
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
			GlobalTipComponent.LayoutTemporary("Failed to update TTL: "+res.Msg, 3, TipTypeError)
			return
		}
		GlobalTipComponent.LayoutTemporary("TTL updated", 3, TipTypeSuccess)
		c.Layout()
	}, func() {
		GlobalTipComponent.LayoutTemporary("TTL update cancelled", 3, TipTypeWarning)
	}, nil)

	// 刷新
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{'r'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		GlobalKeyInfoComponent.Layout()
		GlobalKeyInfoDetailComponent.viewOriginY = 0
		GlobalKeyInfoDetailComponent.Layout()
		return nil
	})

	// 删除 key
	GuiSetKeysbindingConfirm(GlobalApp.Gui, c.name, []any{'d'}, "Delete this key permanently?", func() {
		if strings.TrimSpace(c.keyName) == "" {
			GlobalTipComponent.LayoutTemporary("No key selected", 2, TipTypeWarning)
			return
		}
		res := services.Browser().DeleteKey(
			GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
			GlobalDBComponent.SelectedDB,
			c.keyName,
			true,
		)
		if res.Success {
			// Remove from key list
			for i, k := range GlobalKeyComponent.keys {
				if fmt.Sprintf("%s", k) == c.keyName {
					GlobalKeyComponent.keys = append(GlobalKeyComponent.keys[:i], GlobalKeyComponent.keys[i+1:]...)
					break
				}
			}
			GlobalKeyInfoComponent.keyName = ""
			GlobalTipComponent.LayoutTemporary("Deleted key", 2, TipTypeSuccess)
			GlobalKeyComponent.Layout()
			GlobalKeyInfoComponent.Layout()
			GlobalKeyInfoDetailComponent.Layout()
		} else {
			GlobalTipComponent.LayoutTemporary("Delete failed: "+res.Msg, 4, TipTypeError)
		}
	}, func() {
		GlobalTipComponent.LayoutTemporary("Delete key cancelled", 2, TipTypeWarning)
	})

	return c
}

func (c *LTRKeyInfoComponent) KeyMapTip() string {
	keyMap := []KeyMapStruct{
		{"Rename", "<e>"},
		{"Edit TTL", "<t>"},
		{"Delete Key", "<d>"},
		{"Copy", "<c>"},
		{"Paste Rename", "<p>"},
		{"Paste TTL", "<x>"},
		{"Refresh", "<r>"},
		{"Pane", "<Tab>"},
		{"Conn/Quit/Help", "<Ctrl+w>/<Ctrl+q>/<?>"},
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
