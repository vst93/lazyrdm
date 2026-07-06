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

func (c *PageComponentInput) Layout() *PageComponentInput {
	GlobalApp.Gui.Cursor = true

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
	// 宽度：label + 输入框，留充足左右留白
	labelWidth := DisplayWidth(c.label)
	inputWidth := 50
	if labelWidth+6 > inputWidth {
		inputWidth = labelWidth + 6
	}
	maxWidth := GlobalApp.maxX - 8
	if inputWidth > maxWidth {
		inputWidth = maxWidth
	}
	if inputWidth < 40 {
		inputWidth = 40
	}

	viewWidth := inputWidth + 6 // 左右各 3 格留白
	viewHeight := 12            // 充足的垂直空间

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

	// ── 弹窗主体（不可编辑，只放 label 和 footer）──
	dialogBodyWidth := theX1 - theX0 - 1 // 内容区宽度
	v, _ := SetViewSafe(c.name, theX0, theY0, theX1, theY1, 0)
	v.Title = " " + c.title + " "
	v.Wrap = false
	v.Editable = false
	v.Frame = true
	v.FgColor = gocui.ColorWhite
	v.BgColor = gocui.ColorBlue
	v.FrameColor = gocui.ColorCyan
	v.Clear()

	// 布局（dialog 内容区 y 从 0 开始）：
	// y=0: 空
	// y=1: label
	// y=2: 空
	// y=3: 空（输入框会覆盖这里，单独 view）
	// y=4: 空
	// y=5: 空
	// y=6: 分隔线
	// y=7: footer 提示
	// y=8+: 空

	v.Write([]byte("\n"))
	labelLine := "  " + c.label
	v.Write([]byte(padRightDisplayWidth(labelLine, dialogBodyWidth) + "\n"))
	v.Write([]byte("\n"))
	// y=3 留给输入框 view 覆盖
	v.Write([]byte("\n"))
	v.Write([]byte("\n"))
	v.Write([]byte("\n"))
	// 分隔线
	sep := strings.Repeat("─", dialogBodyWidth-2)
	v.Write([]byte("  " + sep + "\n"))
	// footer
	footerText := "  [Enter] 确认    [Esc] 取消"
	v.Write([]byte(padRightDisplayWidth(footerText, dialogBodyWidth) + "\n"))

	// ── 输入框（独立 view，可编辑）──
	// gocui Size() = x1-x0-1, y1-y0-1。要有 1 行内容，y1-y0 至少 = 2
	inputViewName := c.name + "_field"
	inputX0 := theX0 + 4
	inputX1 := theX1 - 4
	// y=3 对应 dialog 内容区第 4 行
	// dialog 的 frame 占 1 行，所以绝对 y = theY0 + 1(frame) + 3(content) = theY0 + 4
	inputY0 := theY0 + 4
	inputY1 := inputY0 + 2 // y1-y0=2 → 内容 1 行

	iv, _ := SetViewSafe(inputViewName, inputX0, inputY0, inputX1, inputY1, 0)
	iv.Title = ""
	iv.Frame = false
	iv.Wrap = false
	iv.Editable = true
	iv.BgColor = gocui.ColorWhite
	iv.FgColor = gocui.ColorBlack | gocui.AttrBold
	iv.SelBgColor = gocui.ColorWhite
	iv.SelFgColor = gocui.ColorBlack
	if c.maskInput {
		iv.Editor = &EditorPassword{}
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
		submit()
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, inputViewName, []any{gocui.KeyEsc}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		cancel()
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, inputViewName, []any{gocui.KeyCtrlJ}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		submit()
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
