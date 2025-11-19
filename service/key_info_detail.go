package service

import (
	"fmt"
	"strings"
	"tinyrdm/backend/services"
	"tinyrdm/backend/types"

	"github.com/atotto/clipboard"
	"github.com/duke-git/lancet/v2/validator"

	"github.com/awesome-gocui/gocui"
)

type LTRKeyInfoDetailComponent struct {
	name           string
	title          string
	LayoutMaxY     int
	view           *gocui.View
	keyValueFormat string
	viewOriginY    int
	keyValueMaxY   int
	CopyString     string
}

var keyValueFormatList = []string{"Raw", "JSON"}

func InitKeyInfoDetailComponent() {
	GlobalKeyInfoDetailComponent = &LTRKeyInfoDetailComponent{
		name:           "key_info_detail",
		title:          "Detail",
		LayoutMaxY:     0,
		keyValueFormat: "Raw",
	}
	GlobalKeyInfoDetailComponent.Layout().KeyBind()
	GlobalApp.ViewNameList = append(GlobalApp.ViewNameList, GlobalKeyInfoDetailComponent.name)
	GlobalTipComponent.AppendList(GlobalKeyInfoDetailComponent.name, GlobalKeyInfoDetailComponent.KeyMapTip())
}

func (c *LTRKeyInfoDetailComponent) LayoutTitle() *LTRKeyInfoDetailComponent {
	if c.view != nil && GlobalApp.Gui.CurrentView().Name() == c.name {
		c.view.Title = " [" + c.title + "] "
	} else {
		c.view.Title = " " + c.title + " "
	}
	return c
}

func (c *LTRKeyInfoDetailComponent) Layout() *LTRKeyInfoDetailComponent {
	theX0, _ := GlobalDBComponent.view.Size()
	theX0 = theX0 + 2
	var err error
	theVal := ""
	// show key detail
	c.view, err = GlobalApp.Gui.SetView(c.name, theX0, 3, GlobalApp.maxX-1, GlobalApp.maxY-2, 0)
	if err == nil || err != gocui.ErrUnknownView {
		c.keyValueMaxY = 0
		c.view.Wrap = true
		// c.view.Title = " " + c.title + " "
		if GlobalApp.Gui.CurrentView().Name() == c.name {
			c.view.Title = " [" + c.title + "] "
		} else {
			c.view.Title = " " + c.title + " "
		}
		keyDetail := services.Browser().GetKeyDetail(types.KeyDetailParam{
			Server: GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
			DB:     GlobalDBComponent.SelectedDB,
			Key:    GlobalKeyInfoComponent.keyName,
		})
		if keyDetail.Success {
			keyDetailData := keyDetail.Data.(types.KeyDetail)
			theVal = fmt.Sprintln(keyDetailData.Value)
			// format json data
			if c.keyValueFormat == "JSON" && validator.IsJSON(theVal) {
				theVal, _ = PrettyString(theVal)
				// c.view.Wrap = false
			}
			theValSlice := strings.Split(theVal, "\n")
			theViewX, _ := c.view.Size()
			for _, line := range theValSlice {
				lineLen := len(line)
				if lineLen > theViewX {
					c.keyValueMaxY += lineLen / theViewX
					if lineLen%theViewX > 0 {
						c.keyValueMaxY++
					}
				} else {
					c.keyValueMaxY++
				}
			}
			// PrintLn(c.keyValueMaxY)
		} else {
			theVal = fmt.Sprintln("")
		}
	}
	c.view.Clear()
	// theValRune = theValRune[:GlobalApp.maxX-theX0-2]
	// theVal = string(theValRune)
	// theVal = text.TrimSpace(theVal)
	c.CopyString = theVal
	// c.view.Write(DisposeMultibyteString(theVal))
	c.view.Write([]byte(theVal))

	// show format select
	formatStr := " Format: " + c.keyValueFormat + " "
	formatSelectView, err := GlobalApp.Gui.SetView("key_value_format", GlobalApp.maxX-len(formatStr)-2, GlobalApp.maxY-4, GlobalApp.maxX-1, GlobalApp.maxY-2, 0)
	if err == nil || err != gocui.ErrUnknownView {
		formatSelectView.Clear()
		formatSelectView.Write([]byte(formatStr))
	}
	formatSelectView.Frame = false
	formatSelectView.BgColor = gocui.ColorGreen

	c.view.SetOrigin(0, c.viewOriginY)

	// if GlobalApp.Gui.CurrentView().Name() == c.name {
	// 	// GlobalApp.Gui.SetCurrentView(c.name)
	// 	GlobalTipComponent.Layout(c.KeyMapTip())
	// }
	return c
}

func (c *LTRKeyInfoDetailComponent) KeyBind() {
	GlobalApp.Gui.DeleteKeybindings(c.name)
	GlobalApp.Gui.DeleteKeybindings("key_value_format")
	// format switch
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{'f', 'F'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.switchKeyValueFormat()
		return nil
	})

	//copy
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{'c', 'C'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		theVal := c.CopyString
		if theVal == "" {
			GlobalTipComponent.LayoutTemporary("No data to copy", 2, TipTypeWarning)
			return nil
		}
		clipboard.WriteAll(theVal)
		GlobalTipComponent.LayoutTemporary("Copied to clipboard", 3, TipTypeSuccess)
		return nil
	})
	// scroll
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.KeyArrowUp, gocui.MouseWheelUp}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.scroll(-1)
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.KeyArrowDown, gocui.MouseWheelDown}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.scroll(1)
		return nil
	})
	// scroll page
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.KeyArrowLeft}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.scroll(-GlobalApp.maxY + 9)
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.KeyArrowRight}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.scroll(GlobalApp.maxY - 9)
		return nil
	})

	// 鼠标点击聚焦
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.MouseLeft}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		GlobalApp.ForceUpdate(c.name)
		return nil
	})

	// key_value_format
	GuiSetKeysbinding(GlobalApp.Gui, "key_value_format", []any{gocui.MouseLeft}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.switchKeyValueFormat()
		return nil
	})

	// 刷新
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{'r'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		GlobalKeyInfoComponent.Layout()
		GlobalKeyInfoDetailComponent.viewOriginY = 0
		GlobalKeyInfoDetailComponent.Layout()
		return nil
	})

	// 修改值
	GuiSetKeysbindingConfirm(GlobalApp.Gui, c.name, []any{'p'}, "Replace value with clipboard content?", func() {
		theClipboardValue, err := clipboard.ReadAll()
		if err != nil {
			GlobalTipComponent.LayoutTemporary("Clipboard is empty or not available", 3, TipTypeError)
			return
		}
		if GlobalKeyInfoComponent.keyName == "" {
			GlobalTipComponent.LayoutTemporary("No key selected", 3, TipTypeError)
			return
		}

		// 检查 key 类型
		keySummary := services.Browser().GetKeySummary(types.KeySummaryParam{
			Server: GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
			DB:     GlobalDBComponent.SelectedDB,
			Key:    GlobalKeyInfoComponent.keyName,
		})
		if !keySummary.Success {
			GlobalTipComponent.LayoutTemporary("Failed to get key summary, message: "+keySummary.Msg, 3, TipTypeError)
			return
		}
		keySummaryData := keySummary.Data.(types.KeySummary)
		if keySummaryData.Type != "string" {
			GlobalTipComponent.LayoutTemporary("Only string type can be modified at now", 3, TipTypeError)
			return
		}

		// 修改值
		res := services.Browser().SetKeyValue(
			types.SetKeyParam{
				Server:  GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
				DB:      GlobalDBComponent.SelectedDB,
				Key:     GlobalKeyInfoComponent.keyName,
				KeyType: "string",
				Value:   theClipboardValue,
				TTL:     -1,
			},
		)
		if !res.Success {
			GlobalTipComponent.LayoutTemporary("Failed to set value, message: "+res.Msg, 3, TipTypeError)
			return
		}
		GlobalTipComponent.LayoutTemporary("Set value successfully", 3, TipTypeSuccess)
		c.Layout()
	}, func() {
		GlobalTipComponent.LayoutTemporary("Cancel set value", 3, TipTypeWarning)
	})

	// 修改值
	GuiSetKeysbindingConfirmWithVIEditor(GlobalApp.Gui, c.name, []any{'e'}, "", func() string {
		return c.CopyString
	}, func(editorResult string) {
		if GlobalKeyInfoComponent.keyName == "" {
			GlobalTipComponent.LayoutTemporary("No key selected", 3, TipTypeError)
			return
		}
		// 检查 key 类型
		keySummary := services.Browser().GetKeySummary(types.KeySummaryParam{
			Server: GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
			DB:     GlobalDBComponent.SelectedDB,
			Key:    GlobalKeyInfoComponent.keyName,
		})
		if !keySummary.Success {
			GlobalTipComponent.LayoutTemporary("Failed to get key summary, message: "+keySummary.Msg, 3, TipTypeError)
			return
		}
		keySummaryData := keySummary.Data.(types.KeySummary)
		if keySummaryData.Type != "string" {
			GlobalTipComponent.LayoutTemporary("Only string type can be modified at now", 3, TipTypeError)
			return
		}

		// 修改值
		res := services.Browser().SetKeyValue(
			types.SetKeyParam{
				Server:  GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
				DB:      GlobalDBComponent.SelectedDB,
				Key:     GlobalKeyInfoComponent.keyName,
				KeyType: "string",
				Value:   editorResult,
				TTL:     -1,
			},
		)
		if !res.Success {
			GlobalTipComponent.LayoutTemporary("Failed to set value, message: "+res.Msg, 3, TipTypeError)
			return
		}
		GlobalTipComponent.LayoutTemporary("Set value successfully", 3, TipTypeSuccess)
		c.Layout()
	}, func() {
		GlobalTipComponent.LayoutTemporary("Cancel set value", 3, TipTypeWarning)
	}, false)
}

func (c *LTRKeyInfoDetailComponent) KeyMapTip() string {
	keyMap := []KeyMapStruct{
		{"Switch", "<Tab>"},
		{"Switch Format", "<F>"},
		{"Edit", "<E>"},
		{"Copy", "<C>"},
		{"Paste", "<P>"},
		{"Scroll", "↑/↓"},
		{"Scroll Page", "←/→"},
		{"Refresh", "<R>"},
	}
	ret := ""
	for i, v := range keyMap {
		if i > 0 {
			ret += " | "
		}
		ret += fmt.Sprintf("%s: %s", v.Description, v.Key)
		i++
	}
	// return "key_detail: " + ret
	return ret
}

func (c *LTRKeyInfoDetailComponent) scroll(n int) {
	c.viewOriginY += n
	if c.viewOriginY < 0 {
		c.viewOriginY = 0
	}
	_, theViewY := c.view.Size()
	if c.keyValueMaxY-theViewY <= c.viewOriginY {
		c.viewOriginY = c.keyValueMaxY - theViewY
	}

	c.view.SetOrigin(0, c.viewOriginY)
}

func (c *LTRKeyInfoDetailComponent) switchKeyValueFormat() {
	nextIndex := 0
	for i, format := range keyValueFormatList {
		if format == c.keyValueFormat {
			nextIndex = i + 1
			break
		}
	}
	if nextIndex >= len(keyValueFormatList) {
		nextIndex = 0
	}
	c.keyValueFormat = keyValueFormatList[nextIndex]
	c.viewOriginY = 0
	c.Layout()
	// GlobalKeyInfoDetailComponent.Layout()
}
