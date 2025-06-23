package service

import (
	"github.com/jroimartin/gocui"
)

// GuiSetKeysbinding set keysbinding for a view
func GuiSetKeysbinding(g *gocui.Gui, viewname string, keys []any, mod gocui.Modifier, handler func(*gocui.Gui, *gocui.View) error) error {
	for _, key := range keys {
		if err := g.SetKeybinding(viewname, key, mod, handler); err != nil {
			return err
		}
	}
	return nil
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
type EditorInput struct{}

func (e *EditorInput) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	if mod != gocui.ModNone {
		return
	}
	switch {
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		// 删除字符
		v.EditDelete(true)
	case key == gocui.KeyArrowLeft:
		// 光标左移
		v.MoveCursor(-1, 0, false)
	case key == gocui.KeyArrowRight:
		// 光标右移
		v.MoveCursor(1, 0, false)
	case key == gocui.KeyTab:
		v.EditWrite(ch)
	default:
		if IsNormalChar(ch) {
			// 输入字符
			v.EditWrite(ch)
		}
		// keyOut = key
	}
	// return
}
