package service

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/awesome-gocui/gocui"
)

type PageComponentConfirm struct {
	name        string
	maskName    string
	title       string
	text        string
	returnView  string
	callbackYes func()
	callbackNo  func()
}

func NewPageComponentConfirm(title string, text string, callbackYes func(), callbackNo func()) *PageComponentConfirm {
	returnView := ""
	if currentView := GlobalApp.Gui.CurrentView(); currentView != nil && currentView.Name() != "page_confirm" {
		returnView = currentView.Name()
	}
	ret := &PageComponentConfirm{
		name:        "page_confirm",
		maskName:    "page_confirm_mask",
		title:       title,
		text:        text,
		returnView:  returnView,
		callbackYes: callbackYes,
		callbackNo:  callbackNo,
	}
	ret.Layout()
	return ret
}

func (c *PageComponentConfirm) Layout() *PageComponentConfirm {
	GlobalApp.Gui.Cursor = false
	maskView, _ := SetViewSafe(c.maskName, 0, 0, GlobalApp.maxX-1, GlobalApp.maxY-1, 0)
	maskView.Editable = false
	maskView.Frame = false
	maskView.Wrap = false
	maskView.Clear()
	if _, err := GlobalApp.Gui.SetViewOnTop(c.maskName); err == nil {
		GlobalApp.Gui.SetCurrentView(c.maskName)
	}

	messageLines := c.normalizedMessageLines()
	viewWidth, viewHeight := c.calculateDialogSize(messageLines)
	bodyWidth := viewWidth - 4

	theX0 := (GlobalApp.maxX - viewWidth) / 2
	if theX0 < 0 {
		theX0 = 0
	}
	theY0 := (GlobalApp.maxY - viewHeight) / 2
	if theY0 < 0 {
		theY0 = 0
	}
	theX1 := theX0 + viewWidth - 1
	theY1 := theY0 + viewHeight - 1

	v, _ := SetViewSafe(c.name, theX0, theY0, theX1, theY1, 0)
	v.Title = " Confirm"
	v.Wrap = false
	v.Editable = false
	v.Frame = true
	v.Highlight = true
	v.FgColor = themeDialogFg
	v.FrameColor = themeFrameDialog
	v.TitleColor = gocui.ColorWhite

	v.Clear()
	v.Write([]byte("\n"))
	for _, line := range messageLines {
		trimmedLine := strings.TrimSpace(line)
		trimmedLine = truncateByRuneCount(trimmedLine, bodyWidth)
		v.Write([]byte(" " + padRightDisplayWidth(trimmedLine, bodyWidth) + "\n"))
	}

	footerSpacerLines := viewHeight - len(messageLines) - 7
	if footerSpacerLines < 2 {
		footerSpacerLines = 2
	}
	for i := 0; i < footerSpacerLines; i++ {
		v.Write([]byte("\n"))
	}

	v.Write([]byte(c.footerActionLine(bodyWidth)))
	if _, err := GlobalApp.Gui.SetViewOnTop(c.name); err != nil {
		return c
	}
	GlobalApp.Gui.SetCurrentView(c.name)
	c.KeyBind()
	GlobalTipComponent.AppendList(c.name, c.KeyMapTips())
	return c
}

func (c *PageComponentConfirm) normalizedMessageLines() []string {
	messageLines := strings.Split(strings.TrimSpace(c.text), "\n")
	ret := make([]string, 0, len(messageLines))
	for _, line := range messageLines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		ret = append(ret, trimmed)
	}
	if len(ret) == 0 {
		return []string{"Are you sure?"}
	}
	return ret
}

func (c *PageComponentConfirm) calculateDialogSize(messageLines []string) (int, int) {
	maxMessageWidth := 0
	for _, line := range messageLines {
		lineWidth := DisplayWidth(line)
		if lineWidth > maxMessageWidth {
			maxMessageWidth = lineWidth
		}
	}
	titleWidth := DisplayWidth(strings.TrimSpace(c.title)) + 6
	leftAction := "[Enter/Y] Confirm"
	rightAction := "[Esc/N] Cancel"
	actionWidth := DisplayWidth(leftAction) + DisplayWidth(rightAction) + 4

	bodyWidth := maxMessageWidth
	if bodyWidth < titleWidth {
		bodyWidth = titleWidth
	}
	if bodyWidth < actionWidth {
		bodyWidth = actionWidth
	}
	if bodyWidth < 28 {
		bodyWidth = 28
	}

	viewWidth := bodyWidth + 4
	maxAllowedWidth := GlobalApp.maxX - 4
	if maxAllowedWidth < 34 {
		maxAllowedWidth = GlobalApp.maxX - 2
	}
	if viewWidth > maxAllowedWidth {
		viewWidth = maxAllowedWidth
	}
	if viewWidth < 34 {
		viewWidth = 34
		if viewWidth > maxAllowedWidth {
			viewWidth = maxAllowedWidth
		}
	}

	viewHeight := len(messageLines) + 8
	if viewHeight < 11 {
		viewHeight = 11
	}
	maxAllowedHeight := GlobalApp.maxY - 2
	if viewHeight > maxAllowedHeight {
		viewHeight = maxAllowedHeight
	}

	return viewWidth, viewHeight
}

func (c *PageComponentConfirm) footerActionLine(bodyWidth int) string {
	leftText := "[Enter/Y] Confirm"
	rightText := "[Esc/N] Cancel"
	joined := leftText + strings.Repeat(" ", 6) + rightText
	if DisplayWidth(joined) > bodyWidth {
		joined = leftText + "  " + rightText
	}
	return " " + padRightDisplayWidth(joined, bodyWidth)
}

func (c *PageComponentConfirm) KeyBind() *PageComponentConfirm {
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{'y', 'Y', gocui.KeyEnter}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.closeView()
		if c.callbackYes != nil {
			c.callbackYes()
		}
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, c.name, []any{'n', 'N', gocui.KeyEsc}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.closeView()
		if c.callbackNo != nil {
			c.callbackNo()
		}
		return nil
	})

	GuiSetKeysbinding(GlobalApp.Gui, c.maskName, []any{'y', 'Y', gocui.KeyEnter}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.closeView()
		if c.callbackYes != nil {
			c.callbackYes()
		}
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, c.maskName, []any{'n', 'N', gocui.KeyEsc}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.closeView()
		if c.callbackNo != nil {
			c.callbackNo()
		}
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, c.maskName, []any{gocui.MouseLeft}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		_, _ = g.SetCurrentView(c.name)
		return nil
	})

	return c
}

func (c *PageComponentConfirm) closeView() {
	GlobalApp.Gui.DeleteView(c.name)
	GlobalApp.Gui.DeleteKeybindings(c.name)
	GlobalApp.Gui.DeleteView(c.maskName)
	GlobalApp.Gui.DeleteKeybindings(c.maskName)
	if c.returnView != "" {
		if _, err := GlobalApp.Gui.SetCurrentView(c.returnView); err == nil {
			return
		}
	}
	views := GlobalApp.Gui.Views()
	if len(views) == 0 {
		return
	}
	for _, view := range views {
		if view != nil {
			GlobalApp.Gui.SetCurrentView(view.Name())
			return
		}
	}
}

func (c *PageComponentConfirm) KeyMapTips() string {
	keyMap := []KeyMapStruct{
		{"Confirm", "<y>/<Enter>"},
		{"Cancel", "<n>/<Esc>"},
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

func truncateByRuneCount(input string, maxRuneCount int) string {
	if maxRuneCount <= 0 {
		return ""
	}
	if utf8.RuneCountInString(input) <= maxRuneCount {
		return input
	}
	if maxRuneCount <= 3 {
		return strings.Repeat(".", maxRuneCount)
	}

	maxContentRune := maxRuneCount - 3
	b := strings.Builder{}
	runeCount := 0
	for _, r := range input {
		if runeCount >= maxContentRune {
			break
		}
		b.WriteRune(r)
		runeCount++
	}
	b.WriteString("...")
	return b.String()
}

func centerByDisplayWidth(input string, width int) string {
	if width <= 0 {
		return ""
	}
	text := truncateByRuneCount(strings.TrimSpace(input), width)
	textWidth := DisplayWidth(text)
	if textWidth >= width {
		return " " + text
	}
	leftPadding := (width - textWidth) / 2
	if leftPadding < 0 {
		leftPadding = 0
	}
	ret := " " + strings.Repeat(" ", leftPadding) + text
	return padRightDisplayWidth(ret, width+1)
}
