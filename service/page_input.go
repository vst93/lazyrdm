package service

import (
	"strings"

	"github.com/atotto/clipboard"
	"github.com/awesome-gocui/gocui"
	"github.com/gdamore/tcell/v2"
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
	maskInput   bool
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

// setCursorBar 切换终端光标为不闪烁方块样式（在白底输入框中最显眼）。
// 使用 steady block 而非 bar，因为竖线太细在某些终端下几乎不可见。
// 设置光标颜色为黑色，在白底输入框上对比度最高。
// 通过 tcell 的 SetCursorStyle API 设置，光标样式会在每帧 redraw 时保持。
func setCursorBar() {
	gocuiScreen.SetCursorStyle(tcell.CursorStyleSteadyBlock, tcell.ColorBlack)
}

// setCursorDefault 恢复终端默认光标样式
func setCursorDefault() {
	gocuiScreen.SetCursorStyle(tcell.CursorStyleDefault)
}

func (c *PageComponentInput) Layout() *PageComponentInput {
	GlobalApp.Gui.Cursor = true
	setCursorBar()

	// ── mask 遮罩层 ──
	maskView, _ := SetViewSafe(c.maskName, 0, 0, GlobalApp.maxX-1, GlobalApp.maxY-1, 0)
	maskView.Editable = false
	maskView.Frame = false
	maskView.Wrap = false
	maskView.Clear()
	if _, err := GlobalApp.Gui.SetViewOnTop(c.maskName); err == nil {
		GlobalApp.Gui.SetCurrentView(c.maskName)
	}

	// ── 弹窗尺寸 ──
	labelWidth := DisplayWidth(c.label)
	inputWidth := 50
	if labelWidth+8 > inputWidth {
		inputWidth = labelWidth + 8
	}
	maxWidth := GlobalApp.maxX - 10
	if inputWidth > maxWidth {
		inputWidth = maxWidth
	}
	if inputWidth < 40 {
		inputWidth = 40
	}

	viewWidth := inputWidth + 8 // 左右各 4 格留白
	// Layout (top frame + 7 content rows + bottom frame = 9):
	//   y=0: label     y=1: gap     y=2: input row
	//   y=3: padding   y=4: gap     y=5: separator    y=6: footer
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
	dialogBodyWidth := theX1 - theX0 - 1
	v, _ := SetViewSafe(c.name, theX0, theY0, theX1, theY1, 0)
	v.Title = " " + c.title + " "
	v.Wrap = false
	v.Editable = false
	v.Frame = true
	v.FgColor = themeDialogFg
	v.BgColor = themeDialogBg
	v.FrameColor = themeFrameDialog
	v.FrameRunes = frameDouble
	v.TitleColor = gocui.ColorWhite
	v.Clear()

	// Content layout (7 content rows inside frame):
	//   y=0: label text
	//   y=1: blank gap
	//   y=2: input field row (covered by frameless inputView)
	//   y=3: input view padding (covered by inputView, not visible)
	//   y=4: blank gap
	//   y=5: separator
	//   y=6: footer
	v.Write([]byte("  " + c.label + "\n"))  // y=0
	v.Write([]byte("\n"))                     // y=1 gap
	v.Write([]byte("\n"))                     // y=2 input row (covered)
	v.Write([]byte("\n"))                     // y=3 input view padding (covered)
	v.Write([]byte("\n"))                     // y=4 gap
	sepLen := dialogBodyWidth - 4
	if sepLen < 10 {
		sepLen = 10
	}
	sep := strings.Repeat("─", sepLen)
	v.Write([]byte("  " + sep + "\n"))       // y=5 separator
	footerText := "  [Enter] OK  [Esc] Cancel"
	v.Write([]byte(padRightDisplayWidth(footerText, dialogBodyWidth) + "\n")) // y=6 footer

	// ── 输入框（独立 view）──
	inputViewName := c.name + "_field"
	inputX0 := theX0 + 5
	inputX1 := theX1 - 5
	if inputX1 <= inputX0+5 {
		inputX0 = theX0 + 3
		inputX1 = theX1 - 3
	}
	// Input field overlay at content row y=2.
	// Absolute Y = theY0 + 1(frame) + 2(content offset) = theY0 + 3.
	// y1-y0=2 gives 1 internal content row (gocui frameless Size()=y1-y0-1).
	// The view physically occupies rows [theY0+3, theY0+5], covering content y=2 and y=3 (padding).
	// Separator at y=5 (theY0+6) and footer at y=6 (theY0+7) are NOT covered.
	inputY0 := theY0 + 3
	inputY1 := inputY0 + 2 // y1-y0=2 -> 1 content row

	iv, _ := SetViewSafe(inputViewName, inputX0, inputY0, inputX1, inputY1, 0)
	iv.Title = ""
	iv.Frame = false
	iv.Wrap = false
	iv.Editable = true
	iv.BgColor = themeInputBg
	iv.FgColor = themeInputFg
	iv.SelBgColor = themeInputBg
	iv.SelFgColor = themeInputFg
	if c.maskInput {
		iv.Editor = &EditorPassword{BindValString: &c.resultText}
	} else {
		iv.Editor = &EditorInput{BindValString: &c.resultText}
	}
	iv.Clear()
	iv.WriteRunes([]rune(c.resultText))
	iv.SetCursor(len([]rune(c.resultText)), 0)
	c.inputView = iv

	// ── 层级和焦点 ──
	if _, err := GlobalApp.Gui.SetViewOnTop(c.name); err != nil {
		return c
	}
	GlobalApp.Gui.SetViewOnTop(inputViewName)
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

	GuiSetKeysbinding(GlobalApp.Gui, inputViewName, []any{gocui.KeyEnter}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if isPasting() {
			return nil
		}
		submit()
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, inputViewName, []any{gocui.KeyEsc}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		cancel()
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, inputViewName, []any{gocui.KeyCtrlJ}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if isPasting() {
			return nil
		}
		submit()
		return nil
	})
	// Ctrl+U: clear input field
	GuiSetKeysbinding(GlobalApp.Gui, inputViewName, []any{gocui.KeyCtrlU}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		v.Clear()
		c.resultText = ""
		v.SetCursor(0, 0)
		return nil
	})
	// Ctrl+Y: copy input content to clipboard
	GuiSetKeysbinding(GlobalApp.Gui, inputViewName, []any{gocui.KeyCtrlY}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		text := strings.TrimRight(v.Buffer(), "\n")
		if text != "" {
			clipboard.WriteAll(text)
			GlobalTipComponent.LayoutTemporary("Copied to clipboard", 2, TipTypeSuccess)
		}
		return nil
	})

	for _, vn := range []string{c.maskName, c.name} {
		GuiSetKeysbinding(GlobalApp.Gui, vn, []any{gocui.KeyEnter}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
			submit()
			return nil
		})
		GuiSetKeysbinding(GlobalApp.Gui, vn, []any{gocui.KeyEsc}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
			cancel()
			return nil
		})
	}
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
	setCursorDefault()
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
	return "Confirm: <Enter> | Cancel: <Esc> | Clear: <Ctrl+U> | Paste: <Ctrl+V> | Copy: <Ctrl+Y>"
}
