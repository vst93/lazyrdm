package service

import (
	"fmt"
	"strings"

	"github.com/awesome-gocui/gocui"
)

// PageComponentInput 是一个内联文本输入弹窗，替代 vim 外部编辑器用于短文本输入。
type PageComponentInput struct {
	name        string
	maskName    string
	title       string
	label       string
	returnView  string
	initialText string
	resultText  string
	callbackOk  func(result string)
	callbackNo  func()
	maskInput   bool // 是否密码遮罩
	inputView   *gocui.View
}

func NewPageComponentInput(title string, label string, initialText string, maskInput bool, callbackOk func(result string), callbackNo func()) *PageComponentInput {
	returnView := ""
	if currentView := GlobalApp.Gui.CurrentView(); currentView != nil && currentView.Name() != "page_input" {
		returnView = currentView.Name()
	}
	ret := &PageComponentInput{
		name:        "page_input",
		maskName:    "page_input_mask",
		title:       title,
		label:       label,
		returnView:  returnView,
		initialText: initialText,
		resultText:  initialText,
		maskInput:   maskInput,
		callbackOk:  callbackOk,
		callbackNo:  callbackNo,
	}
	ret.Layout()
	return ret
}

func (c *PageComponentInput) Layout() *PageComponentInput {
	GlobalApp.Gui.Cursor = true

	// mask — 用蓝色背景模拟遮罩
	maskView, _ := SetViewSafe(c.maskName, 0, 0, GlobalApp.maxX-1, GlobalApp.maxY-1, 0)
	maskView.Editable = false
	maskView.Frame = false
	maskView.Wrap = false
	maskView.Clear()
	if _, err := GlobalApp.Gui.SetViewOnTop(c.maskName); err == nil {
		GlobalApp.Gui.SetCurrentView(c.maskName)
	}

	// ── 计算弹窗尺寸 ──
	labelWidth := DisplayWidth(c.label)
	inputMinWidth := 40
	bodyWidth := labelWidth + 4
	if bodyWidth < inputMinWidth {
		bodyWidth = inputMinWidth
	}
	maxAllowedWidth := GlobalApp.maxX - 6
	if bodyWidth > maxAllowedWidth {
		bodyWidth = maxAllowedWidth
	}
	viewWidth := bodyWidth + 4
	if viewWidth < 46 {
		viewWidth = 46
		if viewWidth > maxAllowedWidth+4 {
			viewWidth = maxAllowedWidth + 4
		}
	}
	viewHeight := 9

	theX0 := (GlobalApp.maxX - viewWidth) / 2
	if theX0 < 1 {
		theX0 = 1
	}
	theY0 := (GlobalApp.maxY - viewHeight) / 2
	if theY0 < 1 {
		theY0 = 1
	}
	theX1 := theX0 + viewWidth - 1
	theY1 := theY0 + viewHeight - 1

	// ── 弹窗主体 ──
	v, _ := SetViewSafe(c.name, theX0, theY0, theX1, theY1, 0)
	v.Title = " " + c.title + " "
	v.Wrap = false
	v.Editable = false
	v.Frame = true
	v.FgColor = gocui.ColorWhite | gocui.AttrBold
	v.BgColor = gocui.ColorBlue
	v.FrameColor = gocui.ColorCyan
	v.Clear()

	// 顶部空行
	v.Write([]byte("\n"))
	// label 行
	labelLine := " " + c.label
	v.Write([]byte(padRightDisplayWidth(labelLine, bodyWidth+1) + "\n"))
	v.Write([]byte("\n"))

	// ── 输入框 ──
	inputViewName := c.name + "_field"
	inputX0 := theX0 + 2
	inputY0 := theY0 + 3
	inputX1 := theX1 - 2
	inputY1 := inputY0 + 1

	iv, _ := SetViewSafe(inputViewName, inputX0, inputY0, inputX1, inputY1, 0)
	iv.Title = ""
	iv.Frame = true
	iv.Wrap = false
	iv.Editable = true
	iv.BgColor = gocui.ColorWhite
	iv.FgColor = gocui.ColorBlack
	iv.FrameColor = gocui.ColorYellow
	if c.maskInput {
		iv.Editor = &EditorPassword{}
	} else {
		iv.Editor = &EditorInput{BindValString: &c.resultText}
	}
	iv.Clear()
	iv.WriteRunes([]rune(c.resultText))
	iv.SetCursor(len([]rune(c.resultText)), 0)
	c.inputView = iv

	// ── 底部提示栏 ──
	footerY := theY1 - 2
	footerView, _ := SetViewSafe(c.name+"_footer", theX0+1, footerY, theX1-1, footerY+1, 0)
	footerView.Frame = false
	footerView.Editable = false
	footerView.BgColor = gocui.ColorBlue
	footerView.FgColor = gocui.ColorYellow
	footerView.Clear()
	footerText := "[Enter] 确认    [Esc] 取消"
	footerView.Write([]byte(" " + padRightDisplayWidth(footerText, bodyWidth)))

	if _, err := GlobalApp.Gui.SetViewOnTop(c.name); err != nil {
		return c
	}
	GlobalApp.Gui.SetViewOnTop(inputViewName)
	GlobalApp.Gui.SetViewOnTop(c.name + "_footer")
	GlobalApp.Gui.SetCurrentView(inputViewName)
	c.KeyBind()
	GlobalTipComponent.AppendList(c.name, c.KeyMapTips())
	return c
}

func (c *PageComponentInput) KeyBind() *PageComponentInput {
	inputViewName := c.name + "_field"
	GlobalApp.Gui.DeleteKeybindings(c.name)
	GlobalApp.Gui.DeleteKeybindings(inputViewName)
	GlobalApp.Gui.DeleteKeybindings(c.maskName)
	GlobalApp.Gui.DeleteKeybindings(c.name + "_footer")

	submit := func() {
		result := strings.TrimSpace(c.resultText)
		c.closeView()
		if c.callbackOk != nil {
			c.callbackOk(result)
		}
	}
	cancel := func() {
		c.closeView()
		if c.callbackNo != nil {
			c.callbackNo()
		}
	}

	// Enter to submit
	GuiSetKeysbinding(GlobalApp.Gui, inputViewName, []any{gocui.KeyEnter}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		submit()
		return nil
	})
	// Esc to cancel
	GuiSetKeysbinding(GlobalApp.Gui, inputViewName, []any{gocui.KeyEsc}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		cancel()
		return nil
	})
	// Ctrl+Enter as alternative submit (Ctrl+J)
	GuiSetKeysbinding(GlobalApp.Gui, inputViewName, []any{gocui.KeyCtrlJ}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		submit()
		return nil
	})

	// mask and footer also respond to Enter/Esc
	for _, vn := range []string{c.maskName, c.name, c.name + "_footer"} {
		GuiSetKeysbinding(GlobalApp.Gui, vn, []any{gocui.KeyEnter}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
			submit()
			return nil
		})
		GuiSetKeysbinding(GlobalApp.Gui, vn, []any{gocui.KeyEsc}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
			cancel()
			return nil
		})
	}
	// Click mask to focus input
	GuiSetKeysbinding(GlobalApp.Gui, c.maskName, []any{gocui.MouseLeft}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		GlobalApp.Gui.SetCurrentView(inputViewName)
		return nil
	})

	return c
}

func (c *PageComponentInput) closeView() {
	inputViewName := c.name + "_field"
	GlobalApp.Gui.DeleteView(inputViewName)
	GlobalApp.Gui.DeleteKeybindings(inputViewName)
	GlobalApp.Gui.DeleteView(c.name + "_footer")
	GlobalApp.Gui.DeleteKeybindings(c.name + "_footer")
	GlobalApp.Gui.DeleteView(c.name)
	GlobalApp.Gui.DeleteKeybindings(c.name)
	GlobalApp.Gui.DeleteView(c.maskName)
	GlobalApp.Gui.DeleteKeybindings(c.maskName)
	GlobalApp.Gui.Cursor = false
	delete(GlobalTipComponent.list, c.name)

	if c.returnView != "" {
		if _, err := GlobalApp.Gui.SetCurrentView(c.returnView); err == nil {
			GlobalTipComponent.LayComponentTips()
			return
		}
	}
	views := GlobalApp.Gui.Views()
	for _, view := range views {
		if view != nil {
			GlobalApp.Gui.SetCurrentView(view.Name())
			GlobalTipComponent.LayComponentTips()
			return
		}
	}
}

func (c *PageComponentInput) KeyMapTips() string {
	keyMap := []KeyMapStruct{
		{"Confirm", "<Enter>"},
		{"Cancel", "<Esc>"},
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
