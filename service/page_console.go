package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"tinyrdm/backend/services"
	"tinyrdm/backend/types"

	"github.com/awesome-gocui/gocui"
	"github.com/redis/go-redis/v9"
)

type PageComponentConsole struct {
	name          string
	returnView    string
	view          *gocui.View
	inputView     *gocui.View
	outputView    *gocui.View
	outputLines   []string
	originY       int
	inputText     string
	history       []string
	historyCursor int // -1 = not browsing, >=0 = position in history
}

var GlobalConsoleComponent *PageComponentConsole

func OpenConsolePage() {
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
	if currentView.Name() == "page_console" {
		return
	}

	connectionName := strings.TrimSpace(GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name)
	if connectionName == "" {
		GlobalTipComponent.LayoutTemporary("No active connection", 2, TipTypeWarning)
		return
	}

	component := &PageComponentConsole{
		name:        "page_console",
		returnView:  currentView.Name(),
		outputLines: []string{},
		originY:     0,
	}
	component.appendOutput(fmt.Sprintf("Redis Console - %s (db: %d)", connectionName, GlobalDBComponent.SelectedDB))
	component.appendOutput("Type Redis commands and press Enter. Use ↑/↓ for history. Esc/q to close.")
	component.appendOutput("")

	GlobalConsoleComponent = component
	component.Layout().KeyBind()
}

func (c *PageComponentConsole) appendOutput(line string) {
	c.outputLines = append(c.outputLines, line)
}

func (c *PageComponentConsole) Layout() *PageComponentConsole {
	if GlobalApp == nil || GlobalApp.Gui == nil {
		return c
	}

	GlobalApp.Gui.Cursor = true

	x0 := 1
	y0 := 1
	x1 := GlobalApp.maxX - 2
	y1 := GlobalApp.maxY - 3
	if x1 <= x0 {
		x0 = 0
		x1 = GlobalApp.maxX - 1
	}
	if y1 <= y0 {
		y0 = 0
		y1 = GlobalApp.maxY - 1
	}

	// Output view
	outputY1 := y1 - 3
	if outputY1 <= y0+2 {
		outputY1 = y0 + 2
	}
	ov, err := SetViewSafe(c.name+"_output", x0, y0, x1, outputY1, 1)
	if err != nil && err != gocui.ErrUnknownView {
		return c
	}
	ov.Title = " Console Output "
	ov.Wrap = true
	ov.Editable = false
	ov.Frame = true
	ov.FrameRunes = frameSolid
	ov.TitleColor = gocui.ColorCyan
	ov.Clear()
	for _, line := range c.outputLines {
		fmt.Fprintln(ov, line)
	}
	ov.SetOrigin(0, c.originY)
	c.outputView = ov

	// Input view
	iv, err := SetViewSafe(c.name+"_input", x0, outputY1+1, x1, y1, 1)
	if err != nil && err != gocui.ErrUnknownView {
		return c
	}
	iv.Title = " Input (Enter to execute, ↑/↓ for history) "
	iv.Wrap = false
	iv.Editable = true
	iv.Frame = true
	iv.FrameRunes = frameDashed
	iv.TitleColor = gocui.ColorCyan
	iv.Editor = &EditorInput{BindValString: &c.inputText}
	iv.Clear()
	iv.Write([]byte(c.inputText))
	c.inputView = iv

	GlobalApp.Gui.SetCurrentView(c.name + "_input")
	if err := c.inputView.SetCursor(len([]rune(c.inputText)), 0); err == nil {
		// cursor set
	}

	GlobalTipComponent.AppendList(c.name, c.KeyMapTip())
	return c
}

func (c *PageComponentConsole) KeyBind() *PageComponentConsole {
	viewNames := []string{c.name + "_input", c.name + "_output"}
	GlobalApp.Gui.DeleteKeybindings(c.name + "_input")
	GlobalApp.Gui.DeleteKeybindings(c.name + "_output")

	// Execute command
	GuiSetKeysbinding(GlobalApp.Gui, c.name+"_input", []any{gocui.KeyEnter}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		cmd := strings.TrimSpace(c.inputText)
		if cmd == "" {
			return nil
		}
		// Save to history (avoid duplicate of last entry)
		if len(c.history) == 0 || c.history[len(c.history)-1] != cmd {
			c.history = append(c.history, cmd)
		}
		c.historyCursor = -1
		c.inputText = ""
		v.Clear()
		v.SetCursor(0, 0)
		c.executeCommand(cmd)
		return nil
	})

	// Command history navigation
	GuiSetKeysbinding(GlobalApp.Gui, c.name+"_input", []any{gocui.KeyArrowUp}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if len(c.history) == 0 {
			return nil
		}
		if c.historyCursor < 0 {
			c.historyCursor = len(c.history) - 1
		} else if c.historyCursor > 0 {
			c.historyCursor--
		}
		c.inputText = c.history[c.historyCursor]
		v.Clear()
		v.Write([]byte(c.inputText))
		v.SetCursor(len([]rune(c.inputText)), 0)
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, c.name+"_input", []any{gocui.KeyArrowDown}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if c.historyCursor < 0 {
			return nil
		}
		if c.historyCursor < len(c.history)-1 {
			c.historyCursor++
			c.inputText = c.history[c.historyCursor]
		} else {
			// Past the end: clear input
			c.historyCursor = -1
			c.inputText = ""
		}
		v.Clear()
		v.Write([]byte(c.inputText))
		v.SetCursor(len([]rune(c.inputText)), 0)
		return nil
	})

	// Scroll output
	GuiSetKeysbinding(GlobalApp.Gui, c.name+"_output", []any{gocui.KeyArrowUp, 'k'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.scrollOutput(-1)
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, c.name+"_output", []any{gocui.KeyArrowDown, 'j'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.scrollOutput(1)
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, c.name+"_output", []any{gocui.MouseWheelUp}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.scrollOutput(-3)
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, c.name+"_output", []any{gocui.MouseWheelDown}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.scrollOutput(3)
		return nil
	})
	// Page up/down for faster scrolling
	GuiSetKeysbinding(GlobalApp.Gui, c.name+"_output", []any{gocui.KeyPgup, gocui.KeyCtrlB}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		_, vh := v.Size()
		c.scrollOutput(-vh)
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, c.name+"_output", []any{gocui.KeyPgdn, gocui.KeyCtrlF}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		_, vh := v.Size()
		c.scrollOutput(vh)
		return nil
	})

	// Tab to switch between input and output
	GuiSetKeysbinding(GlobalApp.Gui, c.name+"_input", []any{gocui.KeyTab}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		GlobalApp.Gui.SetCurrentView(c.name + "_output")
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, c.name+"_output", []any{gocui.KeyTab}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		GlobalApp.Gui.SetCurrentView(c.name + "_input")
		GlobalApp.Gui.Cursor = true
		return nil
	})

	// Close
	for _, vn := range viewNames {
		GuiSetKeysbinding(GlobalApp.Gui, vn, []any{'q', gocui.KeyEsc}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
			c.closeView()
			return nil
		})
	}

	return c
}

func (c *PageComponentConsole) scrollOutput(delta int) {
	c.originY += delta
	if c.originY < 0 {
		c.originY = 0
	}
	maxOrigin := len(c.outputLines) - 1
	if c.outputView != nil {
		_, vh := c.outputView.Size()
		if maxOrigin > vh {
			maxOrigin = maxOrigin - vh
		} else {
			maxOrigin = 0
		}
	}
	if c.originY > maxOrigin {
		c.originY = maxOrigin
	}
	if c.outputView != nil {
		c.outputView.SetOrigin(0, c.originY)
	}
}

func (c *PageComponentConsole) executeCommand(cmd string) {
	connectionName := GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name
	prompt := fmt.Sprintf("%s:db%d> ", connectionName, GlobalDBComponent.SelectedDB)
	c.appendOutput(prompt + cmd)

	parts := splitRedisCommand(cmd)
	if len(parts) == 0 {
		c.appendOutput("(error) empty command")
		c.refreshOutput()
		return
	}

	// Block dangerous commands
	lowerCmd := strings.ToLower(parts[0])
	if lowerCmd == "flushall" || lowerCmd == "flushdb" || lowerCmd == "shutdown" || lowerCmd == "config" {
		c.appendOutput("(error) blocked: use the app's built-in feature for this command")
		c.refreshOutput()
		return
	}

	// Show "executing..." hint
	c.appendOutput("(executing...)")
	c.refreshOutput()

	// Run command in background goroutine to avoid blocking the UI
	go func() {
		connResp := services.Connection().GetConnection(connectionName)
		if !connResp.Success || connResp.Data == nil {
			c.replaceLastOutput("(error) connection not found")
			c.refreshOutputAsync()
			return
		}
		conn, ok := connResp.Data.(*types.Connection)
		if !ok {
			c.replaceLastOutput("(error) invalid connection data")
			c.refreshOutputAsync()
			return
		}

		options := buildRedisOptions(conn.ConnectionConfig, GlobalDBComponent.SelectedDB)
		client := redis.NewClient(options)
		defer client.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		args := make([]any, len(parts))
		for i, p := range parts {
			args[i] = p
		}

		result, err := client.Do(ctx, args...).Result()
		if err != nil {
			if err == redis.Nil {
				c.replaceLastOutput("(nil)")
			} else {
				c.replaceLastOutput("(error) " + err.Error())
			}
		} else {
			c.replaceLastOutput(formatRedisResult(result))
		}
		c.appendOutput("")
		c.refreshOutputAsync()
	}()
}

func (c *PageComponentConsole) refreshOutput() {
	if c.outputView == nil {
		return
	}
	c.outputView.Clear()
	for _, line := range c.outputLines {
		fmt.Fprintln(c.outputView, line)
	}
	_, vh := c.outputView.Size()
	total := len(c.outputLines)
	if total > vh {
		c.originY = total - vh
	} else {
		c.originY = 0
	}
	c.outputView.SetOrigin(0, c.originY)
	GlobalApp.Gui.SetCurrentView(c.name + "_input")
	GlobalApp.Gui.Cursor = true
}

// refreshOutputAsync is the goroutine-safe version of refreshOutput
func (c *PageComponentConsole) refreshOutputAsync() {
	GlobalApp.Gui.Update(func(g *gocui.Gui) error {
		c.refreshOutput()
		return nil
	})
}

// replaceLastOutput replaces the last line in outputLines with new text
func (c *PageComponentConsole) replaceLastOutput(text string) {
	if len(c.outputLines) == 0 {
		c.appendOutput(text)
		return
	}
	c.outputLines[len(c.outputLines)-1] = text
}

func (c *PageComponentConsole) closeView() {
	GlobalApp.Gui.DeleteView(c.name + "_output")
	GlobalApp.Gui.DeleteView(c.name + "_input")
	GlobalApp.Gui.DeleteKeybindings(c.name + "_output")
	GlobalApp.Gui.DeleteKeybindings(c.name + "_input")
	GlobalApp.Gui.Cursor = false
	GlobalConsoleComponent = nil
	delete(GlobalTipComponent.list, c.name)

	if c.returnView != "" {
		if _, err := GlobalApp.Gui.SetCurrentView(c.returnView); err == nil {
			GlobalTipComponent.LayComponentTips()
			return
		}
	}
	if _, err := GlobalApp.Gui.SetCurrentView("db_list"); err == nil {
		GlobalTipComponent.LayComponentTips()
	}
}

func (c *PageComponentConsole) KeyMapTip() string {
	keyMap := []KeyMapStruct{
		{"Execute", "<Enter>"},
		{"History", "↑/↓"},
		{"Switch I/O", "<Tab>"},
		{"Scroll", "↑/↓/j/k"},
		{"Page", "PgUp/PgDn"},
		{"Close", "<Esc>/q"},
	}
	ret := ""
	for i, v := range keyMap {
		if i > 0 {
			ret += " | "
		}
		ret += fmt.Sprintf("%s: %s", v.Description, v.Key)
	}
	return ret
}

func splitRedisCommand(cmd string) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false
	quoteChar := byte(' ')
	escaped := false
	for i := 0; i < len(cmd); i++ {
		ch := cmd[i]
		if escaped {
			current.WriteByte(ch)
			escaped = false
			continue
		}
		switch {
		case ch == '\\' && inQuotes:
			escaped = true
		case ch == '"' || ch == '\'':
			if !inQuotes {
				inQuotes = true
				quoteChar = ch
			} else if ch == quoteChar {
				inQuotes = false
				quoteChar = ' '
			} else {
				current.WriteByte(ch)
			}
		case ch == ' ' && !inQuotes:
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		default:
			current.WriteByte(ch)
		}
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	return parts
}

func formatRedisResult(result any) string {
	switch val := result.(type) {
	case nil:
		return "(nil)"
	case string:
		return `"` + val + `"`
	case int64:
		return fmt.Sprintf("(integer) %d", val)
	case int:
		return fmt.Sprintf("(integer) %d", val)
	case float64:
		return fmt.Sprintf("(float) %g", val)
	case bool:
		if val {
			return "(true)"
		}
		return "(false)"
	case []any:
		if len(val) == 0 {
			return "(empty array)"
		}
		var b strings.Builder
		for i, item := range val {
			if i > 0 {
				b.WriteString("\n")
			}
			b.WriteString(fmt.Sprintf("%d) %s", i+1, formatRedisResult(item)))
		}
		return b.String()
	case []string:
		if len(val) == 0 {
			return "(empty array)"
		}
		var b strings.Builder
		for i, item := range val {
			if i > 0 {
				b.WriteString("\n")
			}
			b.WriteString(fmt.Sprintf("%d) %s", i+1, formatRedisResult(item)))
		}
		return b.String()
	case map[string]any:
		if len(val) == 0 {
			return "(empty hash)"
		}
		var b strings.Builder
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for i, k := range keys {
			if i > 0 {
				b.WriteString("\n")
			}
			b.WriteString(fmt.Sprintf("%d) %q => %s", i+1, k, formatRedisResult(val[k])))
		}
		return b.String()
	default:
		return fmt.Sprintf("%v", val)
	}
}

func buildRedisOptions(config types.ConnectionConfig, db int) *redis.Options {
	addr := config.Addr
	if addr == "" {
		addr = "127.0.0.1:6379"
	}
	if config.Port > 0 && !strings.Contains(addr, ":") {
		addr = fmt.Sprintf("%s:%d", addr, config.Port)
	}
	opts := &redis.Options{
		Addr:     addr,
		Password: config.Password,
		DB:       db,
		Network:  "tcp",
	}
	if config.Username != "" {
		opts.Username = config.Username
	}
	if config.Network == "unix" {
		opts.Network = "unix"
		opts.Addr = config.Sock
	}
	return opts
}
