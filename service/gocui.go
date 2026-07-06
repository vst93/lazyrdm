package service

import (
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/awesome-gocui/gocui"
	"golang.org/x/term"
)

func confirmTitleFromTip(defaultTitle string, tip string) string {
	text := strings.ToLower(strings.TrimSpace(tip))
	if text == "" {
		return defaultTitle
	}
	switch {
	case strings.Contains(text, "delete"):
		return "Delete Confirmation"
	case strings.Contains(text, "rename"):
		return "Rename Confirmation"
	case strings.Contains(text, "ttl"):
		return "TTL Confirmation"
	case strings.Contains(text, "replace"):
		return "Replace Confirmation"
	case strings.Contains(text, "import"):
		return "Import Confirmation"
	case strings.Contains(text, "export"):
		return "Export Confirmation"
	case strings.Contains(text, "apply") || strings.Contains(text, "update"):
		return "Apply Changes"
	default:
		return defaultTitle
	}
}

func IsConfirmModalActive(g *gocui.Gui) bool {
	if g == nil {
		return false
	}
	if currentView := g.CurrentView(); currentView != nil && currentView.Name() == "page_confirm" {
		return true
	}
	if _, err := g.View("page_confirm"); err == nil {
		return true
	}
	return false
}

func activeOverlayViewName(g *gocui.Gui) string {
	if g == nil {
		return ""
	}
	if _, err := g.View("page_confirm"); err == nil {
		return "page_confirm"
	}
	if _, err := g.View("page_help"); err == nil {
		return "page_help"
	}
	if _, err := g.View("page_server_info"); err == nil {
		return "page_server_info"
	}
	if _, err := g.View("page_console_output"); err == nil {
		return "page_console_output"
	}
	return ""
}

func canHandleOverlayViewBinding(bindingView string, overlayView string) bool {
	if bindingView == overlayView {
		return true
	}
	if overlayView == "page_confirm" && bindingView == "page_confirm_mask" {
		return true
	}
	return false
}

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
		wrappedHandler := func(g *gocui.Gui, v *gocui.View) error {
			overlayView := activeOverlayViewName(g)
			if overlayView != "" && !canHandleOverlayViewBinding(viewnameStr, overlayView) {
				if _, err := g.SetCurrentView(overlayView); err != nil {
					return nil
				}
				return nil
			}
			return handler(g, v)
		}
		if err := g.SetKeybinding(viewnameStr, key, mod, wrappedHandler); err != nil {
			return err
		}
	}
	return nil
}

// GuiSetKeysbindingConfirm set keysbinding for a view with confirm
func GuiSetKeysbindingConfirm(g *gocui.Gui, viewname string, keys []any, tip string, handlerYes func(), handlerNo func()) {
	GuiSetKeysbinding(g, viewname, keys, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if tip == "" {
			tip = "Are you sure you want to continue?"
		}
		NewPageComponentConfirm(confirmTitleFromTip("Action Confirmation", tip), tip, func() {
			if handlerYes != nil {
				handlerYes()
			}
		}, func() {
			if handlerNo != nil {
				handlerNo()
			}
		})
		return nil
	})
}

// GuiSetKeysbindingConfirmWithVIEditor set keysbinding for a view with confirm and vi editors
func GuiSetKeysbindingConfirmWithVIEditor(g *gocui.Gui, viewname string, keys []any, tip string, handlerGetText func() string, handlerYes func(editedText string), handlerNo func(), skipConfirm bool, canProceed func() bool) {
	// 展示提示语
	if tip == "" {
		tip = "Apply this change?"
	}
	GuiSetKeysbinding(g, viewname, keys, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if canProceed != nil && !canProceed() {
			return nil
		}
		initialText := handlerGetText()
		currentView := g.CurrentView()
		// 保存原始终端状态
		var oldState *term.State
		if runtime.GOOS != "windows" {
			if state, err := term.MakeRaw(int(os.Stdin.Fd())); err == nil {
				oldState = state
				defer term.Restore(int(os.Stdin.Fd()), oldState)
			}
		}
		// 禁用鼠标输入
		disableMouseInput()
		gocui.Suspend()
		defer func() {
			// 完全重置终端
			resetTerminalCompletely()
			gocui.Resume()
			// 重新启用鼠标和重绘界面
			if g != nil {
				g.Update(func(g *gocui.Gui) error {
					// 延迟重新启用鼠标
					if runtime.GOOS != "windows" {
						go func() {
							time.Sleep(100 * time.Millisecond)
							enableMouseInput()
						}()
					}
					if IsConfirmModalActive(g) {
						_, _ = g.SetCurrentView("page_confirm")
						return nil
					}
					// 恢复当前视图
					if currentView != nil {
						if view, err := g.View(currentView.Name()); err == nil {
							g.SetCurrentView(view.Name())
						}
					}
					return nil
				})
			}
		}()

		editedText, err := EditWithExternalEditor(initialText)
		if err != nil {
			return err
		}

		// gocui.Suspend()
		// // 恢复 gocui
		// defer gocui.Resume()
		// // 调用外部编辑器
		// editedText, err := RunEditorInSubprocess(handlerGetText())
		// if err != nil {
		// 	// 恢复 gocui
		// 	// gocui.Resume()
		// 	return err
		// }

		// 跳过确认
		if skipConfirm {
			if handlerYes != nil {
				handlerYes(editedText)
			}
			return nil
		}
		NewPageComponentConfirm(confirmTitleFromTip("Apply Changes", tip), tip, func() {
			if handlerYes != nil {
				handlerYes(editedText)
			}
		}, func() {
			if handlerNo != nil {
				handlerNo()
			}
		})
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
