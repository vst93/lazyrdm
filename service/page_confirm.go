package service

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/awesome-gocui/gocui"
)

type PageComponentConfirm struct {
	name         string
	maskName     string
	yesBtnName   string
	noBtnName    string
	title        string
	text         string
	returnView   string
	callbackYes  func()
	callbackNo   func()
	mouseHoverYn bool // true=hover on Yes, false=hover on No
}

func NewPageComponentConfirm(title string, text string, callbackYes func(), callbackNo func()) *PageComponentConfirm {
	returnView := ""
	if currentView := GlobalApp.Gui.CurrentView(); currentView != nil && currentView.Name() != "page_confirm" {
		returnView = currentView.Name()
	}
	if strings.TrimSpace(title) == "" {
		title = "Confirmation"
	}
	if strings.TrimSpace(text) == "" {
		text = "Are you sure?"
	}
	ret := &PageComponentConfirm{
		name:        "page_confirm",
		maskName:    "page_confirm_mask",
		yesBtnName:  "page_confirm_yes",
		noBtnName:   "page_confirm_no",
		title:       title,
		text:        text,
		returnView:  returnView,
		callbackYes: callbackYes,
		callbackNo:  callbackNo,
		mouseHoverYn: true, // default hover on Yes
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
	v.Title = " " + c.title + " "
	v.Wrap = false
	v.Editable = false
	v.Frame = true
	v.Highlight = false
	v.FgColor = themeDialogFg
	v.FrameColor = themeFrameDialog
	v.FrameRunes = frameDouble
	v.TitleColor = gocui.ColorWhite

	v.Clear()
	v.Write([]byte("\n"))
	for _, line := range messageLines {
		trimmedLine := strings.TrimSpace(line)
		trimmedLine = truncateByDisplayWidth(trimmedLine, bodyWidth)
		v.Write([]byte(" " + padRightDisplayWidth(trimmedLine, bodyWidth) + "\n"))
	}

	// spacer between message and buttons
	footerSpacerLines := viewHeight - len(messageLines) - 6
	if footerSpacerLines < 1 {
		footerSpacerLines = 1
	}
	for i := 0; i < footerSpacerLines; i++ {
		v.Write([]byte("\n"))
	}

	// button area: two clickable buttons side by side
	// ┌─────────────┐  ┌──────────────┐
	// │  [Y] Confirm│  │  [N] Cancel  │
	// └─────────────┘  └──────────────┘
	btnLabelYes := " [Y] Confirm "
	btnLabelNo := " [N] Cancel "
	btnWYes := DisplayWidth(btnLabelYes) + 2 // +2 for frame
	btnWNo := DisplayWidth(btnLabelNo) + 2
	btnH := 3 // 1 content row + 2 frame lines
	gap := 3
	totalBtnW := btnWYes + gap + btnWNo
	btnX0Start := theX0 + (viewWidth-totalBtnW)/2
	if btnX0Start < theX0+2 {
		btnX0Start = theX0 + 2
	}
	btnY0 := theY0 + viewHeight - 4
	btnY1 := btnY0 + btnH - 1

	// Yes button
	yesX0 := btnX0Start
	yesX1 := yesX0 + btnWYes - 1
	yesV, _ := SetViewSafe(c.yesBtnName, yesX0, btnY0, yesX1, btnY1, 0)
	yesV.Frame = true
	yesV.Wrap = false
	yesV.Editable = false
	yesV.Highlight = false
	yesV.FrameRunes = frameSolid
	yesV.TitleColor = gocui.ColorWhite

	// No button
	noX0 := yesX1 + gap + 1
	noX1 := noX0 + btnWNo - 1
	if noX1 > theX1-1 {
		noX1 = theX1 - 1
	}
	noV, _ := SetViewSafe(c.noBtnName, noX0, btnY0, noX1, btnY1, 0)
	noV.Frame = true
	noV.Wrap = false
	noV.Editable = false
	noV.Highlight = false
	noV.FrameRunes = frameSolid
	noV.TitleColor = gocui.ColorWhite

	c.renderButtons()

	if _, err := GlobalApp.Gui.SetViewOnTop(c.name); err != nil {
		return c
	}
	GlobalApp.Gui.SetViewOnTop(c.yesBtnName)
	GlobalApp.Gui.SetViewOnTop(c.noBtnName)
	GlobalApp.Gui.SetCurrentView(c.name)
	c.KeyBind()
	GlobalTipComponent.AppendList(c.name, c.KeyMapTips())
	return c
}

// renderButtons draws the Yes/No buttons with highlight on the hovered one.
func (c *PageComponentConfirm) renderButtons() {
	btnLabelYes := " [Y] Confirm "
	btnLabelNo := " [N] Cancel "

	yesV, errY := GlobalApp.Gui.View(c.yesBtnName)
	if errY == nil {
		yesV.Clear()
		if c.mouseHoverYn {
			yesV.BgColor = themeSelBg
			yesV.FgColor = themeSelFg
			yesV.FrameColor = themeFrameActive
		} else {
			yesV.BgColor = gocui.ColorDefault
			yesV.FgColor = gocui.ColorDefault
			yesV.FrameColor = themeFrameDialog
		}
		yesV.Write([]byte(btnLabelYes))
	}

	noV, errN := GlobalApp.Gui.View(c.noBtnName)
	if errN == nil {
		noV.Clear()
		if !c.mouseHoverYn {
			noV.BgColor = themeSelBg
			noV.FgColor = themeSelFg
			noV.FrameColor = themeFrameActive
		} else {
			noV.BgColor = gocui.ColorDefault
			noV.FgColor = gocui.ColorDefault
			noV.FrameColor = themeFrameDialog
		}
		noV.Write([]byte(btnLabelNo))
	}
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
	btnWidth := DisplayWidth(" [Y] Confirm ") + DisplayWidth(" [N] Cancel ") + 8

	bodyWidth := maxMessageWidth
	if bodyWidth < titleWidth {
		bodyWidth = titleWidth
	}
	if bodyWidth < btnWidth {
		bodyWidth = btnWidth
	}
	if bodyWidth < 30 {
		bodyWidth = 30
	}

	viewWidth := bodyWidth + 4
	maxAllowedWidth := GlobalApp.maxX - 4
	if maxAllowedWidth < 36 {
		maxAllowedWidth = GlobalApp.maxX - 2
	}
	if viewWidth > maxAllowedWidth {
		viewWidth = maxAllowedWidth
	}
	if viewWidth < 36 {
		viewWidth = 36
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

func (c *PageComponentConfirm) KeyBind() *PageComponentConfirm {
	// keyboard: Enter = activate hovered button, Y = confirm, N/Esc = cancel
	activateHovered := func() {
		if c.mouseHoverYn {
			c.closeView()
			if c.callbackYes != nil {
				c.callbackYes()
			}
		} else {
			c.closeView()
			if c.callbackNo != nil {
				c.callbackNo()
			}
		}
	}
	confirm := func() {
		c.closeView()
		if c.callbackYes != nil {
			c.callbackYes()
		}
	}
	cancel := func() {
		c.closeView()
		if c.callbackNo != nil {
			c.callbackNo()
		}
	}

	for _, viewName := range []string{c.name, c.maskName, c.yesBtnName, c.noBtnName} {
		// Y/Enter = confirm (Enter also works as "activate hovered" below)
		GuiSetKeysbinding(GlobalApp.Gui, viewName, []any{'y', 'Y'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
			confirm()
			return nil
		})
		// N/Esc = cancel
		GuiSetKeysbinding(GlobalApp.Gui, viewName, []any{'n', 'N', gocui.KeyEsc}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
			cancel()
			return nil
		})
		// Enter = activate hovered button
		GuiSetKeysbinding(GlobalApp.Gui, viewName, []any{gocui.KeyEnter}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
			activateHovered()
			return nil
		})
		// Tab / Left / Right to toggle hover between buttons
		GuiSetKeysbinding(GlobalApp.Gui, viewName, []any{gocui.KeyTab, gocui.KeyArrowRight, gocui.KeyArrowLeft, 'h', 'l'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
			c.mouseHoverYn = !c.mouseHoverYn
			c.renderButtons()
			return nil
		})
	}

	// mouse click on Yes button
	GuiSetKeysbinding(GlobalApp.Gui, c.yesBtnName, []any{gocui.MouseLeft}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		confirm()
		return nil
	})
	// mouse click on No button
	GuiSetKeysbinding(GlobalApp.Gui, c.noBtnName, []any{gocui.MouseLeft}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		cancel()
		return nil
	})
	// click on mask = focus confirm dialog
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
	GlobalApp.Gui.DeleteView(c.yesBtnName)
	GlobalApp.Gui.DeleteKeybindings(c.yesBtnName)
	GlobalApp.Gui.DeleteView(c.noBtnName)
	GlobalApp.Gui.DeleteKeybindings(c.noBtnName)
	if c.returnView != "" {
		if _, err := GlobalApp.Gui.SetCurrentView(c.returnView); err == nil {
			GlobalApp.Gui.Update(func(g *gocui.Gui) error { return nil })
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
			GlobalApp.Gui.Update(func(g *gocui.Gui) error { return nil })
			return
		}
	}
}

func (c *PageComponentConfirm) KeyMapTips() string {
	keyMap := []KeyMapStruct{
		{"Confirm", "Y/Enter/Click"},
		{"Cancel", "N/Esc/Click"},
		{"Switch", "Tab/←/→"},
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
