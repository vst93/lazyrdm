package service

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/awesome-gocui/gocui"
)

type LTRTipComponent struct {
	name               string
	view               *gocui.View
	temporaryName      string
	temporaryView      *gocui.View
	lastTipString      string
	temporaryTipString string
	temporaryTipType   string
	temporaryTipSeq    int64
	list               map[string]string
}

const (
	TipTypeWarning string = "warning"
	TipTypeError   string = "error"
	TipTypeSuccess string = "success"
)

type KeyMapStruct struct {
	Description string
	Key         string
}

func InitTipComponent() {
	GlobalTipComponent = &LTRTipComponent{
		name:          "key_map_tip",
		temporaryName: "operation_tip",
		list:          make(map[string]string, 100),
	}
	GlobalTipComponent.Layout("")
}

func (c *LTRTipComponent) GetLastTipString() string {
	return c.lastTipString
}

func (c *LTRTipComponent) AppendList(key string, desc string) {
	if _, ok := c.list[key]; !ok {
		c.list[key] = desc
		c.LayComponentTips()
	}
}

func (c *LTRTipComponent) LayComponentTips() {
	theName := CurrentViewName()
	if theName == "" {
		return
	}
	tipString := strings.TrimSpace(c.tipForView(theName))
	if tipString == "" {
		tipString = c.defaultGlobalTip()
	}
	c.Layout(tipString)
}

func (c *LTRTipComponent) defaultGlobalTip() string {
	return "Switch Pane: <Tab> | Switch Connection: <Ctrl+w> | Quit: <Ctrl+q> | Help: <?>"
}

func (c *LTRTipComponent) tipForView(viewName string) string {
	if viewName == "" {
		return ""
	}

	switch viewName {
	case "connection_list":
		if GlobalConnectionComponent != nil {
			return GlobalConnectionComponent.KeyMapTip()
		}
	case "db_list":
		if GlobalDBComponent != nil {
			return GlobalDBComponent.KeyMapTip()
		}
	case "key_list", "key_list_line", "search_key":
		if GlobalKeyComponent != nil {
			return GlobalKeyComponent.KeyMapTip()
		}
	case "key_info", "key_info_ttl":
		if GlobalKeyInfoComponent != nil {
			return GlobalKeyInfoComponent.KeyMapTip()
		}
	case "key_info_detail", "key_detail_line", "key_value_format":
		if GlobalKeyInfoDetailComponent != nil {
			return GlobalKeyInfoDetailComponent.KeyMapTip()
		}
	}

	if tip, ok := c.list[viewName]; ok {
		return tip
	}

	aliasMap := map[string]string{
		"key_list_line":    "key_list",
		"search_key":       "key_list",
		"key_info_ttl":     "key_info",
		"key_detail_line":  "key_info_detail",
		"key_value_format": "key_info_detail",
	}
	if alias, ok := aliasMap[viewName]; ok {
		if tip, exists := c.list[alias]; exists {
			return tip
		}
	}

	return ""
}

func (c *LTRTipComponent) Layout(tipString string) *LTRTipComponent {
	if tipString != "" {
		c.lastTipString = tipString
	}

	var err error
	if GlobalApp.maxX < 2 || GlobalApp.maxY < 2 {
		return c
	}
	footerY0 := GlobalApp.maxY - 2
	if footerY0 < 0 {
		footerY0 = 0
	}
	c.view, err = SetViewSafe(c.name, 0, footerY0, GlobalApp.maxX, GlobalApp.maxY, 0)
	if err != nil && err != gocui.ErrUnknownView {
		PrintLn(err.Error())
		return c
	}
	if GlobalApp != nil && GlobalApp.Gui != nil {
		if _, errTop := GlobalApp.Gui.SetViewOnTop(c.name); errTop != nil && errTop != gocui.ErrUnknownView {
			PrintLn(errTop.Error())
		}
	}
	c.view.Editable = false
	c.view.Frame = false
	c.view.Wrap = false
	c.view.FgColor = gocui.ColorWhite | gocui.AttrBold
	c.view.Clear()

	theTipString := c.lastTipString
	if strings.TrimSpace(theTipString) == "" {
		theTipString = c.defaultGlobalTip()
	}
	maxWidth := GlobalApp.maxX - 2
	theTipString = c.optimizedTipForWidth(theTipString, maxWidth)
	theTipString = " " + strings.TrimSpace(theTipString)
	theTipString = padRightDisplayWidth(theTipString, maxWidth)
	c.view.Write([]byte(theTipString))
	c.layoutTemporaryTip()
	return c
}

func (c *LTRTipComponent) layoutTemporaryTip() {
	if c == nil || GlobalApp == nil || GlobalApp.Gui == nil {
		return
	}
	if strings.TrimSpace(c.temporaryTipString) == "" {
		if c.temporaryView != nil {
			GlobalApp.Gui.DeleteView(c.temporaryName)
			c.temporaryView = nil
		}
		return
	}

	maxWidth := GlobalApp.maxX - 2
	if maxWidth <= 0 || GlobalApp.maxY < 2 {
		return
	}

	displayText := strings.TrimSpace(c.temporaryTipString)
	tipWidth := DisplayWidth(displayText) + 4
	if tipWidth > maxWidth {
		tipWidth = maxWidth
	}
	if tipWidth < 28 {
		tipWidth = 28
		if tipWidth > maxWidth {
			tipWidth = maxWidth
		}
	}

	x1 := GlobalApp.maxX - 1
	x0 := x1 - tipWidth
	if x0 < 0 {
		x0 = 0
	}
	y0 := 0
	y1 := 4
	if GlobalApp.maxY < 4 {
		y1 = 1
	}
	if GlobalApp.maxY >= 4 && GlobalApp.maxY < 6 {
		y1 = 2
	}
	if GlobalApp.maxY >= 6 && GlobalApp.maxY < 8 {
		y1 = 3
	}

	var err error
	c.temporaryView, err = SetViewSafe(c.temporaryName, x0, y0, x1, y1, 0)
	if err != nil && err != gocui.ErrUnknownView {
		PrintLn(err.Error())
		return
	}
	if _, errTop := GlobalApp.Gui.SetViewOnTop(c.temporaryName); errTop != nil && errTop != gocui.ErrUnknownView {
		PrintLn(errTop.Error())
	}
	c.temporaryView.Editable = false
	c.temporaryView.Frame = true
	c.temporaryView.Wrap = false
	c.temporaryView.Title = " Notice "
	c.temporaryView.FgColor = gocui.ColorWhite
	switch c.temporaryTipType {
	case TipTypeWarning:
		c.temporaryView.FgColor = gocui.ColorYellow | gocui.AttrBold
		c.temporaryView.Title = " Warning "
	case TipTypeError:
		c.temporaryView.FgColor = gocui.ColorRed | gocui.AttrBold
		c.temporaryView.Title = " Error "
	case TipTypeSuccess:
		c.temporaryView.FgColor = gocui.ColorGreen | gocui.AttrBold
		c.temporaryView.Title = " Success "
	}
	c.temporaryView.Clear()

	bodyWidth := tipWidth - 2
	theTipString := truncateByDisplayWidth(displayText, bodyWidth-1)
	theTipString = " " + strings.TrimSpace(theTipString)
	theTipString = padRightDisplayWidth(theTipString, bodyWidth)
	if y1 >= 3 {
		c.temporaryView.Write([]byte("\n"))
	}
	c.temporaryView.Write([]byte(theTipString))
	if y1 >= 4 {
		c.temporaryView.Write([]byte("\n"))
	}
}

func padRightDisplayWidth(s string, width int) string {
	if width <= 0 {
		return ""
	}
	cur := DisplayWidth(s)
	if cur >= width {
		return s
	}
	return s + strings.Repeat(" ", width-cur)
}

func (c *LTRTipComponent) optimizedTipForWidth(tipText string, maxWidth int) string {
	tipText = strings.TrimSpace(tipText)
	if maxWidth <= 0 || tipText == "" {
		return tipText
	}
	if DisplayWidth(tipText) <= maxWidth {
		return tipText
	}

	compact := tipText
	compact = strings.ReplaceAll(compact, "[Global] ", "")
	replacer := strings.NewReplacer(
		"Switch Connection", "Conn",
		"Switch Pane", "Pane",
		"Switch Field", "Field",
		"Scroll Page", "Page",
		"Read-only", "RO",
	)
	compact = replacer.Replace(compact)
	if DisplayWidth(compact) <= maxWidth {
		return compact
	}

	parts := strings.Split(compact, " | ")
	if len(parts) <= 1 {
		return truncateByDisplayWidth(compact, maxWidth)
	}

	ret := ""
	for i, p := range parts {
		candidate := p
		if i > 0 {
			candidate = ret + " | " + p
		}
		if DisplayWidth(candidate) <= maxWidth {
			ret = candidate
			continue
		}
		if ret == "" {
			return truncateByDisplayWidth(p, maxWidth)
		}
		ellipsis := " | ..."
		if DisplayWidth(ret+ellipsis) <= maxWidth {
			return ret + ellipsis
		}
		return truncateByDisplayWidth(ret, maxWidth)
	}

	return ret
}

func truncateByDisplayWidth(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if DisplayWidth(s) <= maxWidth {
		return s
	}
	if maxWidth <= 3 {
		return strings.Repeat(".", maxWidth)
	}
	limit := maxWidth - 3
	ret := ""
	current := 0
	for _, r := range s {
		charWidth := DisplayWidth(string(r))
		if current+charWidth > limit {
			break
		}
		ret += string(r)
		current += charWidth
	}
	return ret + "..."
}

func (c *LTRTipComponent) LayoutTemporary(tipString string, durationSec int, tipType string) *LTRTipComponent {
	tipString = strings.TrimSpace(tipString)
	//获取当前的view的内容
	if c.lastTipString == tipString {
		return c
	}
	c.temporaryTipSeq++
	temporaryTipSeq := c.temporaryTipSeq
	c.temporaryTipString = tipString
	c.temporaryTipType = tipType
	GlobalApp.Gui.Update(func(g *gocui.Gui) error {
		c.Layout("")
		return nil
	})
	go func() {
		time.Sleep(time.Second * time.Duration(durationSec))
		GlobalApp.Gui.Update(func(g *gocui.Gui) error {
			if temporaryTipSeq != c.temporaryTipSeq {
				return nil
			}
			c.temporaryTipString = ""
			c.temporaryTipType = ""
			c.Layout("")
			return nil
		})
	}()
	return c
}

func (c *LTRTipComponent) BuildHelpText(currentViewName string) string {
	globalKeyMap := []KeyMapStruct{
		{"Quit", "<Ctrl+q>"},
		{"Switch Pane", "<Tab>"},
		{"Switch Connection", "<Ctrl+w>"},
		{"Help", "<?>"},
	}

	builder := strings.Builder{}
	builder.WriteString("LazyRDM Keyboard Shortcuts\n")
	builder.WriteString("==========================\n\n")
	builder.WriteString("Global\n")
	builder.WriteString("------\n")
	builder.WriteString(c.formatKeyMapList(globalKeyMap))

	currentTip := ""
	if currentViewName != "" {
		if tip, ok := c.list[currentViewName]; ok {
			currentTip = strings.TrimSpace(tip)
		}
	}
	if currentTip == "" {
		lastTip := strings.TrimSpace(c.lastTipString)
		if strings.Contains(lastTip, ":") {
			currentTip = lastTip
		}
	}

	if currentViewName != "" {
		builder.WriteString("\nCurrent View\n")
		builder.WriteString("------------\n")
		builder.WriteString(fmt.Sprintf("%s\n", currentViewName))
		if currentTip != "" {
			builder.WriteString(c.formatKeyMapList(c.parseTipString(currentTip)))
		} else {
			builder.WriteString("- No shortcut tips available for this view.\n")
		}
	}

	if len(c.list) > 0 {
		builder.WriteString("\nAll Views\n")
		builder.WriteString("---------\n")
		viewNames := make([]string, 0, len(c.list))
		for viewName := range c.list {
			viewNames = append(viewNames, viewName)
		}
		sort.Strings(viewNames)
		for _, viewName := range viewNames {
			builder.WriteString(fmt.Sprintf("\n%s\n", viewName))
			builder.WriteString(c.formatKeyMapList(c.parseTipString(c.list[viewName])))
		}
	}

	builder.WriteString("\nHelp Navigation\n")
	builder.WriteString("---------------\n")
	builder.WriteString("- Scroll: ↑/↓/j/k or mouse wheel\n")
	builder.WriteString("- Page: ←/→/h/l\n")
	builder.WriteString("- Close help: <Esc>/q/?\n")
	return builder.String()
}

func (c *LTRTipComponent) parseTipString(tipText string) []KeyMapStruct {
	ret := make([]KeyMapStruct, 0)
	parts := strings.Split(tipText, "|")
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item == "" {
			continue
		}
		kv := strings.SplitN(item, ":", 2)
		if len(kv) == 2 {
			description := strings.TrimSpace(kv[0])
			key := strings.TrimSpace(kv[1])
			if description != "" || key != "" {
				ret = append(ret, KeyMapStruct{Description: description, Key: key})
			}
		} else {
			ret = append(ret, KeyMapStruct{Description: item, Key: ""})
		}
	}
	return ret
}

func (c *LTRTipComponent) formatKeyMapList(keyMap []KeyMapStruct) string {
	if len(keyMap) == 0 {
		return "- No shortcuts configured.\n"
	}
	builder := strings.Builder{}
	for _, item := range keyMap {
		if item.Key == "" {
			builder.WriteString(fmt.Sprintf("- %s\n", item.Description))
		} else {
			builder.WriteString(fmt.Sprintf("- %-16s %s\n", item.Description+":", item.Key))
		}
	}
	return builder.String()
}
