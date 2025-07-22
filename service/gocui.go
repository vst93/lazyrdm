package service

import (
	"strings"

	"github.com/jroimartin/gocui"
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
		v.MoveCursor(-1, 0, false)
	case key == gocui.KeyArrowRight:
		// 光标右移
		v.MoveCursor(1, 0, false)
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
