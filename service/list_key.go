package service

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"tinyrdm/backend/services"
	"tinyrdm/backend/types"

	"github.com/awesome-gocui/gocui"
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
	searchKeyword string
	searchView    *gocui.View
	lineView      *gocui.View
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
	GlobalTipComponent.AppendList(GlobalKeyComponent.name, GlobalKeyComponent.KeyMapTip())
}

func (c *LTRListKeyComponent) LoadKeys() *LTRListKeyComponent {
	if c.IsEnd {
		return c
	}
	// get key list
	if GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name == "" || GlobalDBComponent.SelectedDB < 0 {
		return c
	}
	theSearchKeyword := ""
	if c.searchKeyword != "" && c.searchKeyword != "*" {
		theSearchKeyword = c.searchKeyword
		if theSearchKeyword[0:1] != "*" {
			theSearchKeyword = "*" + theSearchKeyword
		}
		// 判断theSearchKeyword 末尾是不是 *
		if theSearchKeyword[len(theSearchKeyword)-1:] != "*" {
			theSearchKeyword = theSearchKeyword + "*"
		}
		// PrintLn(theSearchKeyword)
	}
	keysInfo := services.Browser().LoadNextKeys(
		GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
		GlobalDBComponent.SelectedDB,
		theSearchKeyword,
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
	c.view, err = GlobalApp.Gui.SetView(c.name, 1, theDBComponentH+4, GlobalApp.maxX*2/10, GlobalApp.maxY-2, 0)
	if err != nil && err != gocui.ErrUnknownView {
		PrintLn(err.Error())
		return c
	}
	c.view.Editable = false
	c.view.Frame = true
	c.view.Title = " Key List "
	if GlobalDBComponent.SelectedDB < 0 {
		c.view.Subtitle = ""
	} else {
		// c.view.Title = " [db" + strconv.Itoa(GlobalDBComponent.SelectedDB) + "]" + " [" + strconv.Itoa(len(c.keys)) + "/" + strconv.FormatInt(c.MaxKeys, 10) + "] "
		// c.view.Title = " Key List [" + strconv.Itoa(len(c.keys)) + "/" + strconv.FormatInt(c.MaxKeys, 10) + "] "
		c.view.Subtitle = " " + strconv.Itoa(len(c.keys)) + "/" + strconv.FormatInt(c.MaxKeys, 10) + " "
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

	lineStr := ""
	lineViewWidth := 0
	lineViewWidthStr := "1"

	if len(c.keys) > 0 {
		splitKeys := c.keys[rangeBegin:rangeEnd]
		lineViewWidth = len(strconv.Itoa(rangeEnd + 1))
		lineViewWidthStr = strconv.Itoa(lineViewWidth)
		for index, key := range splitKeys {
			index = index + rangeBegin
			// totalLine++
			keyStr := fmt.Sprintf("%s", key)
			if c.Current == index {
				// currenLine = totalLine
				printString += NewColorString(keyStr+""+SPACE_STRING+"\n", "white", "blue", "bold")
			} else {
				printString += fmt.Sprintf("%s\n", keyStr+""+SPACE_STRING)
			}
			lineStr += fmt.Sprintf("%"+lineViewWidthStr+"d", index+1) + "\n"
		}
		if c.Current >= (len(c.keys) - 1) {
			c.LoadKeys()
		}
	}

	c.view.Clear()
	c.view.Write([]byte(printString))

	// 显示搜索关键词
	if c.searchKeyword != "" {
		searchKeywordShow := " Search: " + c.searchKeyword + " "
		c.searchView, err = GlobalApp.Gui.SetView("search_key", GlobalApp.maxX*2/10-len(searchKeywordShow)-1, GlobalApp.maxY-4, GlobalApp.maxX*2/10, GlobalApp.maxY-2, 0)
		if err != nil && err != gocui.ErrUnknownView {
			// PrintLn(err.Error())
			return c
		}
		c.searchView.Visible = true
		c.searchView.Frame = false
		c.searchView.BgColor = gocui.ColorYellow
		c.searchView.Clear()
		c.searchView.Write([]byte(searchKeywordShow))
	} else if c.searchView != nil {
		c.searchView.Visible = false
	}

	// line view
	c.lineView, err = GlobalApp.Gui.SetView("key_list_line", 0, theDBComponentH+2, 2, GlobalApp.maxY-2, 0)
	if err == nil || err != gocui.ErrUnknownView {
		// c.lineView.FrameColor = gocui.NewRGBColor(149, 165, 166)
		c.lineView.FgColor = gocui.NewRGBColor(78, 142, 166)
		c.lineView.Clear()
		c.lineView.Write([]byte(lineStr))
		c.lineView.SetOrigin(0, 0)
	}
	c.lineView.FrameRunes = []rune{'─', '│', '┌', '─', '└', '─'}

	// reset view x0 and x1
	c.view, _ = GlobalApp.Gui.SetView(c.name, 1+lineViewWidth, theDBComponentH+2, GlobalApp.maxX*2/10, GlobalApp.maxY-2, 0)
	c.lineView, _ = GlobalApp.Gui.SetView("key_list_line", 0, theDBComponentH+2, 1+lineViewWidth, GlobalApp.maxY-2, 0)

	if GlobalApp.Gui.CurrentView().Name() == c.name {
		c.view.Title = " [Key List] "
		c.lineView.FrameColor = gocui.ColorGreen
	} else {
		c.view.Title = " Key List "
		c.lineView.FrameColor = gocui.ColorDefault
	}
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

	// 刷新
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{'r'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.RefreshList()
		return nil
	})

	// 鼠标点击聚焦
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.MouseLeft}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		GlobalApp.ForceUpdate(c.name)
		return nil
	})

	// 删除 key
	GuiSetKeysbindingConfirm(GlobalApp.Gui, c.name, []any{'d'}, "Do you want to delete the key?", func() {
		if GlobalKeyComponent.Current < 0 || GlobalKeyComponent.Current > len(GlobalKeyComponent.keys)-1 {
			return
		}
		pendDeleteKey := fmt.Sprintf("%s", GlobalKeyComponent.keys[GlobalKeyComponent.Current])
		if pendDeleteKey == "" {
			return
		}
		deleteStaus := false
		for i, key := range GlobalKeyComponent.keys {
			if key == pendDeleteKey {
				services.Browser().DeleteKey(
					GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
					GlobalDBComponent.SelectedDB,
					pendDeleteKey,
					true,
				)
				GlobalKeyComponent.keys = append(GlobalKeyComponent.keys[:i], GlobalKeyComponent.keys[i+1:]...)
				deleteStaus = true
				break
			}
		}
		if deleteStaus {
			GlobalTipComponent.LayoutTemporary("Key [ "+pendDeleteKey+"] deleted", 2, TipTypeSuccess)
		} else {
			GlobalTipComponent.LayoutTemporary("Key [ "+pendDeleteKey+"] not found", 2, TipTypeError)
		}
		GlobalKeyComponent.Layout()
	}, func() {
		GlobalTipComponent.LayoutTemporary("Delete Key cancelled", 2, TipTypeWarning)
	})

	// 新增 key （当前仅支持 string 类型）
	GuiSetKeysbindingConfirm(GlobalApp.Gui, c.name, []any{'a'}, "Do you want to create a new temporary string key?", func() {
		// 新增 key
		theTmpKey := "layrdm_tmp_key:" + time.Now().Format("20060102150405")
		res := services.Browser().SetKeyValue(
			types.SetKeyParam{
				Server:  GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
				DB:      GlobalDBComponent.SelectedDB,
				Key:     theTmpKey,
				KeyType: "string",
				Value:   "null",
				TTL:     600, // ten minutes ttl
			},
		)
		if res.Success {
			// 加到 keys 最开头
			GlobalKeyComponent.keys = append([]any{theTmpKey}, GlobalKeyComponent.keys...)
			GlobalKeyComponent.Current = 0
			GlobalKeyInfoComponent.keyName = theTmpKey
			GlobalKeyInfoComponent.Layout()
			GlobalKeyInfoDetailComponent.viewOriginY = 0
			GlobalKeyInfoDetailComponent.Layout()
			c.Layout()
			GlobalTipComponent.LayoutTemporary("Key [ "+theTmpKey+" ] added", 2, TipTypeSuccess)
		} else {
			GlobalTipComponent.LayoutTemporary("Failed to create key", 3, TipTypeError)
		}
	}, func() {
		GlobalTipComponent.LayoutTemporary("Key creation cancelled", 2, TipTypeWarning)
	})

	// 搜索
	GuiSetKeysbindingConfirmWithVIEditor(GlobalApp.Gui, c.name, []any{'s'}, "Search by keyword?", func() string {
		return c.searchKeyword
	}, func(editorResult string) {
		editorResult = strings.TrimSpace(editorResult)
		c.searchKeyword = editorResult
		c.RefreshList()
	}, func() {
	}, true)

	return c
}

// 刷新列表
func (c *LTRListKeyComponent) RefreshList() *LTRListKeyComponent {
	services.Browser().OpenDatabase(GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name, GlobalDBComponent.SelectedDB)
	c.IsEnd = false
	c.keys = []any{}
	c.Current = 0
	c.LoadKeys().Layout()
	return c
}

func (c *LTRListKeyComponent) KeyMapTip() string {
	keyMap := []KeyMapStruct{
		{"Switch", "<Tab>"},
		{"Select", "↑/↓"},
		{"Enter", "<Enter>/→"},
		{"Search", "<S>"},
		{"Delete", "<D>"},
		{"Add", "<A>"},
		{"Refresh", "<R>"},
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
