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
	lastTipString      string
	temporaryTipString string
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
		name: "key_map_tip",
		list: make(map[string]string, 100),
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
	if theName != "" && len(c.list) > 0 {
		for key, desc := range c.list {
			if theName == key {
				c.Layout(desc)
				break
			}
		}
	}
}

func (c *LTRTipComponent) Layout(tipString string) *LTRTipComponent {
	if tipString == c.lastTipString {
		return c
	}
	if tipString != "" {
		c.lastTipString = tipString
	}

	var err error
	if GlobalApp.maxX < 2 || GlobalApp.maxY < 2 {
		return c
	}
	c.view, err = SetViewSafe(c.name, 0, GlobalApp.maxY-2, GlobalApp.maxX, GlobalApp.maxY, 0)
	if err != nil && err != gocui.ErrUnknownView {
		PrintLn(err.Error())
		return c
	}
	c.view.Editable = false
	c.view.Frame = false
	c.view.Wrap = true
	c.view.FgColor = gocui.ColorBlue
	c.view.Clear()

	theTipString := c.lastTipString
	if c.temporaryTipString != "" {
		theTipString = c.temporaryTipString
	}
	maxWidth := GlobalApp.maxX - 2
	theTipString = c.optimizedTipForWidth(theTipString, maxWidth)
	c.view.Write([]byte(theTipString))
	return c
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
	switch tipType {
	case TipTypeWarning:
		tipString = NewColorString(tipString, "yellow")
	case TipTypeError:
		tipString = NewColorString(tipString, "red")
	case TipTypeSuccess:
		tipString = NewColorString(tipString, "green")
	}
	//获取当前的view的内容
	if c.lastTipString == tipString {
		return c
	}
	c.temporaryTipString = tipString
	// 修改展示的内容
	// c.Layout("")
	GlobalApp.Gui.Update(func(g *gocui.Gui) error {
		c.Layout("")
		return nil
	})
	// 3 s 后恢复原内容
	go func() {
		time.Sleep(time.Second * time.Duration(durationSec))
		GlobalApp.Gui.Update(func(g *gocui.Gui) error {
			c.temporaryTipString = ""
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
