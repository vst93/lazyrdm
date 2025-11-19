package service

import (
	"strings"
	"time"

	"github.com/awesome-gocui/gocui"
)

// GuiSetKeysbinding set keysbinding for a view
func GuiSetKeysbinding(g *gocui.Gui, viewname any, keys []any, mod gocui.Modifier, handler func(*gocui.Gui, *gocui.View) error) error {
	// 如果 viewname 是数组,断言
	viewnameArr, ok := viewname.([]string)
	if ok {
		for _, viewname := range viewnameArr {
			err2 := GuiSetKeysbinding(g, viewname, keys, mod, handler)
			if err2 != nil {
				return err2
			}
		}
		return nil
	}
	viewnameStr, ok := viewname.(string)
	if !ok {
		return nil
	}
	for _, key := range keys {
		if err := g.SetKeybinding(viewnameStr, key, mod, handler); err != nil {
			return err
		}
	}
	return nil
}

// GuiSetKeysbindingConfirm set keysbinding for a view with confirm
func GuiSetKeysbindingConfirm(g *gocui.Gui, viewname string, keys []any, tip string, handlerYes func(), handlerNo func()) {
	tip += " (y/n)"
	GuiSetKeysbinding(g, viewname, keys, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		GlobalTipComponent.LayoutTemporary(tip, 10, TipTypeWarning)
		GlobalApp.Gui.DeleteKeybinding(viewname, 'y', gocui.ModNone)
		// GlobalApp.Gui.DeleteKeybinding(viewname, gocui.KeyEnter, gocui.ModNone)
		GlobalApp.Gui.DeleteKeybinding(viewname, 'n', gocui.ModNone)
		go func() {
			GuiSetKeysbinding(GlobalApp.Gui, viewname, []any{'y'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
				GlobalApp.Gui.DeleteKeybinding(viewname, 'y', gocui.ModNone)
				// GlobalApp.Gui.DeleteKeybinding(viewname, gocui.KeyEnter, gocui.ModNone)
				GlobalApp.Gui.DeleteKeybinding(viewname, 'n', gocui.ModNone)
				handlerYes()
				return nil
			})
			GuiSetKeysbinding(GlobalApp.Gui, viewname, []any{'n'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
				GlobalApp.Gui.DeleteKeybinding(viewname, 'y', gocui.ModNone)
				// GlobalApp.Gui.DeleteKeybinding(viewname, gocui.KeyEnter, gocui.ModNone)
				GlobalApp.Gui.DeleteKeybinding(viewname, 'n', gocui.ModNone)
				handlerNo()
				return nil
			})
			time.Sleep(time.Second * 10)
			GlobalApp.Gui.DeleteKeybinding(viewname, 'y', gocui.ModNone)
			// GlobalApp.Gui.DeleteKeybinding(viewname, gocui.KeyEnter, gocui.ModNone)
			GlobalApp.Gui.DeleteKeybinding(viewname, 'n', gocui.ModNone)
		}()
		return nil
	})
}

// GuiSetKeysbindingConfirmWithVIEditor set keysbinding for a view with confirm and vi editors
func GuiSetKeysbindingConfirmWithVIEditor(g *gocui.Gui, viewname string, keys []any, tip string, handlerGetText func() string, handlerYes func(editedText string), handlerNo func(), skipConfirm bool) {
	// 展示提示语
	if tip == "" {
		tip = "Change the value?"
	}
	tip += " (y/n)"
	GuiSetKeysbinding(g, viewname, keys, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		gocui.Suspend()
		// 调用外部编辑器
		editedText, err := EditWithExternalEditor(handlerGetText())
		if err != nil {
			// 恢复 gocui
			gocui.Resume()
			return err
		}
		// 恢复 gocui
		gocui.Resume()
		// 跳过确认
		if skipConfirm {
			handlerYes(editedText)
			return nil
		}
		GlobalTipComponent.LayoutTemporary(tip, 10, TipTypeWarning)
		GlobalApp.Gui.DeleteKeybinding(viewname, 'y', gocui.ModNone)
		GlobalApp.Gui.DeleteKeybinding(viewname, 'n', gocui.ModNone)
		// GlobalApp.Gui.DeleteKeybinding(viewname, gocui.KeyEnter, gocui.ModNone)
		go func() {
			GuiSetKeysbinding(GlobalApp.Gui, viewname, []any{'y'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
				GlobalApp.Gui.DeleteKeybinding(viewname, 'y', gocui.ModNone)
				GlobalApp.Gui.DeleteKeybinding(viewname, 'n', gocui.ModNone)
				// GlobalApp.Gui.DeleteKeybinding(viewname, gocui.KeyEnter, gocui.ModNone)
				handlerYes(editedText)
				return nil
			})
			GuiSetKeysbinding(GlobalApp.Gui, viewname, []any{'n'}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
				GlobalApp.Gui.DeleteKeybinding(viewname, 'y', gocui.ModNone)
				GlobalApp.Gui.DeleteKeybinding(viewname, 'n', gocui.ModNone)
				// GlobalApp.Gui.DeleteKeybinding(viewname, gocui.KeyEnter, gocui.ModNone)
				handlerNo()
				return nil
			})
			time.Sleep(time.Second * 10)
			GlobalApp.Gui.DeleteKeybinding(viewname, 'y', gocui.ModNone)
			GlobalApp.Gui.DeleteKeybinding(viewname, 'n', gocui.ModNone)
			// GlobalApp.Gui.DeleteKeybinding(viewname, gocui.KeyEnter, gocui.ModNone)
		}()
		return nil
	})
}

// 密码编辑器，把每个字符替换为 '*'
type EditorPassword struct{}

func (e *EditorPassword) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) (keyOut gocui.Key, chOut rune) {
	switch {
	case ch != 0:
		// 替换字符为 '*'
		chOut = '*'
		v.EditWrite(chOut)
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		// 删除字符
		v.EditDelete(true)
	default:
		keyOut = key
	}
	return
}

// 输入编辑器
type EditorInput struct {
	BindValString *string
	BindValInt    *int
	BindValBool   *bool
}

func (e *EditorInput) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	if mod != gocui.ModNone {
		return
	}
	switch {
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		// 删除字符
		v.EditDelete(true)
		if e.BindValString != nil {
			*e.BindValString = v.Buffer()
		}
		if e.BindValInt != nil {
			*e.BindValInt = int(ch)
		}
	case key == gocui.KeyArrowLeft:
		// 光标左移
		v.MoveCursor(-1, 0)
	case key == gocui.KeyArrowRight:
		// 光标右移
		v.MoveCursor(1, 0)
	case key == gocui.KeyTab:
		v.EditWrite(ch)
		if e.BindValString != nil {
			*e.BindValString = v.Buffer()
		}
		if e.BindValInt != nil {
			*e.BindValInt = int(ch)
		}
	default:
		if IsNormalChar(ch) {
			// 输入字符
			v.EditWrite(ch)
			if e.BindValString != nil {
				*e.BindValString = v.Buffer()
			}
			if e.BindValInt != nil {
				*e.BindValInt = int(ch)
			}
		}
		// keyOut = key
	}
	if e.BindValString != nil {
		*e.BindValString = strings.TrimSpace(*e.BindValString)
	}
	// return
}
