package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"tinyrdm/backend/services"

	"github.com/awesome-gocui/gocui"
)

type PageComponentServerInfo struct {
	name       string
	title      string
	text       string
	returnView string
	originY    int
	view       *gocui.View
}

var GlobalServerInfoPageComponent *PageComponentServerInfo

func OpenServerInfoPage() {
	if GlobalApp == nil || GlobalApp.Gui == nil || GlobalConnectionComponent == nil {
		return
	}

	currentView := GlobalApp.Gui.CurrentView()
	if currentView == nil {
		return
	}
	if currentView.Name() == "page_confirm" || currentView.Name() == "page_help" {
		return
	}
	if currentView.Name() == "page_server_info" {
		return
	}

	connectionName := strings.TrimSpace(GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name)
	if connectionName == "" {
		GlobalTipComponent.LayoutTemporary("No active connection", 2, TipTypeWarning)
		return
	}

	serverInfo := services.Browser().ServerInfo(connectionName)
	infoText := buildServerInfoText(connectionName, serverInfo.Success, serverInfo.Data, serverInfo.Msg)

	component := &PageComponentServerInfo{
		name:       "page_server_info",
		title:      "Server Info",
		text:       infoText,
		returnView: currentView.Name(),
		originY:    0,
	}
	GlobalServerInfoPageComponent = component
	component.Layout().KeyBind()
}

func buildServerInfoText(connectionName string, success bool, data any, errMsg string) string {
	builder := strings.Builder{}
	// builder.WriteString("Server Information\n")
	// builder.WriteString("==================\n\n")
	builder.WriteString(fmt.Sprintf("Connection : %s\n", connectionName))
	builder.WriteString(fmt.Sprintf("Fetched At : %s\n", time.Now().Format("2006-01-02 15:04:05")))
	builder.WriteString(fmt.Sprintf("Status     : %s\n", map[bool]string{true: "Success", false: "Failed"}[success]))

	builder.WriteString("\nPayload\n")
	builder.WriteString("-------\n")

	if !success {
		if strings.TrimSpace(errMsg) == "" {
			errMsg = "Unknown error"
		}
		builder.WriteString("Failed to load server info.\n")
		builder.WriteString("Reason: " + errMsg + "\n")
		return builder.String()
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		builder.WriteString(fmt.Sprintf("Failed to encode payload: %v\n", err))
		builder.WriteString(fmt.Sprintf("Raw value: %v\n", data))
		return builder.String()
	}

	prettyJSON, err := PrettyString(string(jsonBytes))
	if err != nil {
		builder.WriteString(string(jsonBytes) + "\n")
		return builder.String()
	}
	builder.WriteString(prettyJSON + "\n")
	builder.WriteString("\nNavigation: ↑/↓/j/k scroll | ←/→/h/l page | Esc/q/i close\n")
	return builder.String()
}

func (c *PageComponentServerInfo) Layout() *PageComponentServerInfo {
	if GlobalApp == nil || GlobalApp.Gui == nil {
		return c
	}

	GlobalApp.Gui.Cursor = false
	x0 := 2
	y0 := 1
	x1 := GlobalApp.maxX - 3
	y1 := GlobalApp.maxY - 3
	if x1 <= x0 {
		x0 = 0
		x1 = GlobalApp.maxX - 1
	}
	if y1 <= y0 {
		y0 = 0
		y1 = GlobalApp.maxY - 1
	}

	v, err := SetViewSafe(c.name, x0, y0, x1, y1, 2)
	if err != nil && err != gocui.ErrUnknownView {
		return c
	}
	v.Title = " [INFO] " + c.title + " "
	v.Subtitle = " Read-only | Scroll: Wheel/Arrows | Close: Esc/q/?/i "
	v.Wrap = false
	v.Editable = false
	v.Frame = true
	v.FrameColor = themeFrameDialog
	v.FrameRunes = frameDouble
	v.TitleColor = gocui.ColorWhite
	v.Clear()
	v.Write([]byte(c.text))
	v.SetOrigin(0, c.originY)
	GlobalApp.Gui.SetCurrentView(c.name)
	c.view = v
	return c
}

func (c *PageComponentServerInfo) KeyBind() *PageComponentServerInfo {
	GlobalApp.Gui.DeleteKeybindings(c.name)

	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.KeyArrowDown, 'j'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.scroll(1)
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.MouseWheelDown}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.scroll(3)
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.KeyArrowUp, 'k'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.scroll(-1)
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.MouseWheelUp}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.scroll(-3)
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.KeyArrowRight, 'l'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.scroll(10)
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{gocui.KeyArrowLeft, 'h'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.scroll(-10)
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{'q', '?', 'i', gocui.KeyEsc}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.closeView()
		return nil
	})
	return c
}

func (c *PageComponentServerInfo) scroll(delta int) {
	if c.view == nil {
		return
	}
	next := c.originY + delta
	if next < 0 {
		next = 0
	}
	lines := c.view.BufferLines()
	_, viewHeight := c.view.Size()
	maxOrigin := len(lines) - viewHeight
	if maxOrigin < 0 {
		maxOrigin = 0
	}
	if next > maxOrigin {
		next = maxOrigin
	}
	c.originY = next
	c.view.SetOrigin(0, c.originY)
}

func (c *PageComponentServerInfo) closeView() {
	GlobalApp.Gui.DeleteView(c.name)
	GlobalApp.Gui.DeleteKeybindings(c.name)
	GlobalServerInfoPageComponent = nil

	if c.returnView != "" {
		if _, err := GlobalApp.Gui.SetCurrentView(c.returnView); err == nil {
			GlobalTipComponent.LayComponentTips()
			return
		}
	}
	if _, err := GlobalApp.Gui.SetCurrentView("db_list"); err == nil {
		GlobalTipComponent.LayoutTemporary("Returned to DB list", 2, TipTypeWarning)
		GlobalTipComponent.LayComponentTips()
	}
}
