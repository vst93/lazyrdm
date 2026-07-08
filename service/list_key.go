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
	keyTypeFilter string // "", "string", "list", "hash", "set", "zset", "stream"
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
		kwRunes := []rune(theSearchKeyword)
		if string(kwRunes[0]) != "*" {
			theSearchKeyword = "*" + theSearchKeyword
		}
		kwRunes = []rune(theSearchKeyword)
		if string(kwRunes[len(kwRunes)-1]) != "*" {
			theSearchKeyword = theSearchKeyword + "*"
		}
	}
	keysInfo := services.Browser().LoadNextKeys(
		GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
		GlobalDBComponent.SelectedDB,
		theSearchKeyword,
		c.keyTypeFilter,
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
	if GlobalDBComponent == nil || GlobalDBComponent.view == nil {
		return c
	}
	_, theDBComponentH := GlobalDBComponent.view.Size()
	var err error
	// 列表
	c.view, err = SetViewSafe(c.name, 1, theDBComponentH+4, GlobalApp.maxX*2/10, GlobalApp.maxY-2, 0)
	if err != nil && err != gocui.ErrUnknownView {
		// PrintLn(err.Error())
		return c
	}
	c.view.Editable = false
	c.view.Frame = true
	c.view.FrameRunes = frameSolid
	c.view.TitleColor = gocui.ColorCyan
	c.view.Title = " Key List "
	if GlobalDBComponent.SelectedDB < 0 {
		c.view.Subtitle = ""
	} else {
		subtitle := " " + strconv.Itoa(len(c.keys)) + "/" + strconv.FormatInt(c.MaxKeys, 10)
		if strings.TrimSpace(c.searchKeyword) != "" {
			subtitle += " | filter:" + c.searchKeyword
		}
		if strings.TrimSpace(c.keyTypeFilter) != "" {
			subtitle += " | type:" + c.keyTypeFilter
		}
		c.view.Subtitle = subtitle + " "
	}
	_, c.viewMaxY = c.view.Size()

	printString := ""
	// currenLine := 0
	// totalLine := 0
	rangeBegin := c.Current - c.viewMaxY/2 + 1
	if rangeBegin < 0 {
		rangeBegin = 0
	}
	rangeEnd := rangeBegin + c.viewMaxY + 2
	if rangeEnd > len(c.keys) {
		rangeEnd = len(c.keys)
		rangeBegin = rangeEnd - c.viewMaxY - 2
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
				printString += NewColorString(keyStr+""+SPACE_STRING+"\n", "black", "cyan", "bold")
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
		c.searchView, err = SetViewSafe("search_key", GlobalApp.maxX*2/10-len(searchKeywordShow)-1, GlobalApp.maxY-4, GlobalApp.maxX*2/10, GlobalApp.maxY-2, 0)
		if err != nil && err != gocui.ErrUnknownView {
			// PrintLn(err.Error())
			return c
		}
		c.searchView.Visible = true
		c.searchView.Frame = false
		c.searchView.BgColor = themeIndicatorBg
		c.searchView.Clear()
		c.searchView.Write([]byte(searchKeywordShow))
	} else if c.searchView != nil {
		c.searchView.Visible = false
	}

	// line view
	c.lineView, err = SetViewSafe("key_list_line", 0, theDBComponentH+2, 2, GlobalApp.maxY-2, 0)
	if err == nil || err != gocui.ErrUnknownView {
		// c.lineView.FrameColor = gocui.NewRGBColor(149, 165, 166)
		c.lineView.FgColor = themeLineNum
		c.lineView.Clear()
		c.lineView.Write([]byte(lineStr))
		c.lineView.SetOrigin(0, 0)
	}
	c.lineView.FrameRunes = frameHalfTL

	// reset view x0 and x1
	c.view, _ = SetViewSafe(c.name, 1+lineViewWidth, theDBComponentH+2, GlobalApp.maxX*2/10, GlobalApp.maxY-2, 0)
	c.lineView, _ = SetViewSafe("key_list_line", 0, theDBComponentH+2, 1+lineViewWidth, GlobalApp.maxY-2, 0)

	if CurrentViewName() == c.name {
		c.view.Title = " [Key List] "
		c.lineView.FrameColor = gocui.ColorGreen
		if GlobalTipComponent != nil {
			GlobalTipComponent.Layout(c.KeyMapTip())
		}
	} else {
		c.view.Title = " Key List "
		c.lineView.FrameColor = gocui.ColorDefault
	}
	return c
}

func (c *LTRListKeyComponent) KeyBind() *LTRListKeyComponent {
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.KeyArrowDown, gocui.MouseWheelDown, 'j'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if len(c.keys) == 0 {
			return nil
		}
		c.Current++
		if c.Current > len(c.keys)-1 {
			c.Current = 0
		}
		v.Clear()
		c.Layout()
		return nil
	})

	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.KeyArrowUp, gocui.MouseWheelUp, 'k'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if len(c.keys) == 0 {
			return nil
		}
		c.Current--
		if c.Current < 0 {
			c.Current = len(c.keys) - 1
		}
		v.Clear()
		c.Layout()
		return nil
	})

	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.KeyEnter, gocui.KeyArrowRight, 'l'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
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
			GlobalTipComponent.LayoutTemporary("Deleted key: "+pendDeleteKey, 2, TipTypeSuccess)
		} else {
			GlobalTipComponent.LayoutTemporary("Key not found: "+pendDeleteKey, 2, TipTypeError)
		}
		GlobalKeyComponent.Layout()
	}, func() {
		GlobalTipComponent.LayoutTemporary("Delete key cancelled", 2, TipTypeWarning)
	})

	// 新增 key （当前仅支持 string 类型）
	GuiSetKeysbindingConfirm(GlobalApp.Gui, c.name, []any{'a'}, "Create a temporary string key with 10-minute TTL?", func() {
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
			GlobalTipComponent.LayoutTemporary("Created key: "+theTmpKey, 2, TipTypeSuccess)
		} else {
			GlobalTipComponent.LayoutTemporary("Failed to create temporary key", 3, TipTypeError)
		}
	}, func() {
		GlobalTipComponent.LayoutTemporary("Create key cancelled", 2, TipTypeWarning)
	})

	// 搜索
	GuiSetKeysbindingInlineInput(GlobalApp.Gui, c.name, []any{'s'}, "Search Keys", "Keyword (supports * glob)", func() string {
		return c.searchKeyword
	}, func(editorResult string) {
		c.searchKeyword = editorResult
		c.RefreshList()
	}, func() {
		GlobalTipComponent.LayoutTemporary("Search update cancelled", 2, TipTypeWarning)
	}, nil)

	// 按类型过滤 (cycle: all → string → list → hash → set → zset → stream → all)
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{'T'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		types := []string{"", "string", "list", "hash", "set", "zset", "stream"}
		currentIdx := 0
		for i, t := range types {
			if t == c.keyTypeFilter {
				currentIdx = i
				break
			}
		}
		nextIdx := (currentIdx + 1) % len(types)
		c.keyTypeFilter = types[nextIdx]
		if c.keyTypeFilter == "" {
			GlobalTipComponent.LayoutTemporary("Type filter: all types", 2, TipTypeSuccess)
		} else {
			GlobalTipComponent.LayoutTemporary("Type filter: "+c.keyTypeFilter, 2, TipTypeSuccess)
		}
		c.RefreshList()
		return nil
	})

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
		{"Select", "↑/↓/j/k"},
		{"Open Key", "<Enter>/l/→"},
		{"Search", "<s>"},
		{"Type Filter", "<T>"},
		{"Refresh", "<r>"},
		{"Add", "<a>"},
		{"Delete", "<d>"},
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
	// return "key_list: " + ret
	return ret
}
