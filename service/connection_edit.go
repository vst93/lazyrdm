package service

import (
	"fmt"
	"strings"
	"tinyrdm/backend/services"
	"tinyrdm/backend/types"

	"github.com/jroimartin/gocui"
	"github.com/nsf/termbox-go"
)

type LTRConnectionEditComponent struct {
	name                string
	title               string
	ConnectionConfig    types.Connection
	ConnectionConfigBak types.Connection
	viewList            []string
	viewBeginX          int
	viewBeginY          int
	viewEndX            int
	viewEndY            int
	viewNowLine         int
	viewNowCurrent      string
	KeyMapTipExtend     []KeyMapStruct
}

type LTRConnectionEditComponentFormViewConfig struct {
	name         string
	title        string
	value        EditorInput
	isNewLine    bool
	viewType     string
	xBeing       int
	xEnd         int
	radioMap     []any
	disabledEdit bool
}

func InitConnectionEditComponent(con types.Connection) *LTRConnectionEditComponent {
	connectionEditComponent := &LTRConnectionEditComponent{
		name:                "connection_edit",
		title:               " Connection Edit ",
		ConnectionConfig:    con,
		ConnectionConfigBak: con,
	}
	return connectionEditComponent
}

func (c *LTRConnectionEditComponent) Layout() *LTRConnectionEditComponent {
	GlobalApp.Gui.Cursor = false
	theX0 := 0
	theY0 := 0
	theX1 := GlobalApp.maxX - 1
	theY1 := GlobalApp.maxY - 2
	if GlobalApp.maxX > GlobalApp.maxY && (theY1*15/10) <= theX1 {
		theX0 = (GlobalApp.maxX - GlobalApp.maxY) / 2
		theX1 = theX0 + GlobalApp.maxY - 1
	}
	c.viewBeginX, c.viewBeginY, c.viewEndX, c.viewEndY = theX0, theY0, theX1, theY1
	v, _ := GlobalApp.Gui.SetView(c.name, c.viewBeginX, c.viewBeginY, c.viewEndX, c.viewEndY)
	v.Title = c.title
	// v.Editable = true
	v.Wrap = true
	c.viewNowLine = 0
	c.viewList = []string{}
	c.viewBeginY = 1
	c.KeyMapTipExtend = nil

	// json, _ := json.Marshal(c.ConnectionConfig)
	// PrintLn(string(json))

	// name
	c.formView(LTRConnectionEditComponentFormViewConfig{
		name:  c.name + "_name",
		title: "Name",
		value: EditorInput{BindValString: &c.ConnectionConfig.Name},
	})
	// group
	if c.ConnectionConfigBak.Name == "" {
		c.formView(LTRConnectionEditComponentFormViewConfig{
			name:         c.name + "_group",
			title:        "Group",
			value:        EditorInput{BindValString: &c.ConnectionConfig.Group},
			isNewLine:    true,
			disabledEdit: true,
		})
	}
	// network
	c.formView(LTRConnectionEditComponentFormViewConfig{
		name:      c.name + "_network",
		title:     "Network",
		value:     EditorInput{BindValString: &c.ConnectionConfig.Network},
		isNewLine: true,
		viewType:  "radio",
		radioMap: []any{
			"tcp",
			"unix",
		},
	})

	if c.ConnectionConfig.Network == "unix" {
		GlobalApp.Gui.DeleteView(c.name + "_host")
		GlobalApp.Gui.DeleteView(c.name + "_port")
		// Sock
		c.formView(LTRConnectionEditComponentFormViewConfig{
			name:      c.name + "_sock",
			title:     "Sock",
			value:     EditorInput{BindValString: &c.ConnectionConfig.Sock},
			isNewLine: true,
		})
	} else {
		GlobalApp.Gui.DeleteView(c.name + "_sock")
		// host
		c.formView(LTRConnectionEditComponentFormViewConfig{
			name:      c.name + "_host",
			title:     "Host",
			value:     EditorInput{BindValString: &c.ConnectionConfig.Addr},
			isNewLine: true,
			xEnd:      c.viewEndX - 21,
		})
		// port
		c.formView(LTRConnectionEditComponentFormViewConfig{
			name:      c.name + "_port",
			title:     "Port",
			value:     EditorInput{BindValInt: &c.ConnectionConfig.Port},
			isNewLine: false,
			xBeing:    c.viewEndX - 20,
		})
	}
	// username
	c.formView(LTRConnectionEditComponentFormViewConfig{
		name:      c.name + "_username",
		title:     "Username",
		value:     EditorInput{BindValString: &c.ConnectionConfig.Username},
		isNewLine: true,
	})
	// password
	c.formView(LTRConnectionEditComponentFormViewConfig{
		name:      c.name + "_password",
		title:     "Password",
		value:     EditorInput{BindValString: &c.ConnectionConfig.Password},
		isNewLine: true,
	})

	// ssh
	c.formView(LTRConnectionEditComponentFormViewConfig{
		name:      c.name + "_ssh",
		title:     "Open SSH",
		value:     EditorInput{BindValBool: &c.ConnectionConfig.SSH.Enable},
		isNewLine: true,
		viewType:  "radio",
		radioMap: []any{
			false,
			true,
		},
	})
	if c.ConnectionConfig.SSH.Enable {
		// ssh host
		c.formView(LTRConnectionEditComponentFormViewConfig{
			name:      c.name + "_ssh_host",
			title:     "SSH Host",
			value:     EditorInput{BindValString: &c.ConnectionConfig.SSH.Addr},
			isNewLine: true,
			xEnd:      c.viewEndX - 21,
		})
		// ssh port
		c.formView(LTRConnectionEditComponentFormViewConfig{
			name:      c.name + "_ssh_port",
			title:     "SSH Port",
			value:     EditorInput{BindValInt: &c.ConnectionConfig.SSH.Port},
			isNewLine: false,
			xBeing:    c.viewEndX - 20,
		})
		// ssh login type
		c.formView(LTRConnectionEditComponentFormViewConfig{
			name:      c.name + "_ssh_login_type",
			title:     "SSH Login Type",
			value:     EditorInput{BindValString: &c.ConnectionConfig.SSH.LoginType},
			isNewLine: true,
			viewType:  "radio",
			radioMap: []any{
				"pwd",
				"pkfile",
			},
		})
		if c.ConnectionConfig.SSH.LoginType == "pkfile" {
			GlobalApp.Gui.DeleteView(c.name + "_ssh_username")
			GlobalApp.Gui.DeleteView(c.name + "_ssh_password")
			// ssh key file
			c.formView(LTRConnectionEditComponentFormViewConfig{
				name:      c.name + "_ssh_key_file",
				title:     "SSH Key File Path",
				value:     EditorInput{BindValString: &c.ConnectionConfig.SSH.PKFile},
				isNewLine: true,
			})
			// ssh username
			c.formView(LTRConnectionEditComponentFormViewConfig{
				name:      c.name + "_ssh_username",
				title:     "SSH Username",
				value:     EditorInput{BindValString: &c.ConnectionConfig.SSH.Username},
				isNewLine: true,
			})
			// ssh passphrase
			c.formView(LTRConnectionEditComponentFormViewConfig{
				name:      c.name + "_ssh_passphrase",
				title:     "SSH Passphrase",
				value:     EditorInput{BindValString: &c.ConnectionConfig.SSH.Passphrase},
				isNewLine: true,
			})
		} else {
			GlobalApp.Gui.DeleteView(c.name + "_ssh_key_file")
			GlobalApp.Gui.DeleteView(c.name + "_ssh_username")
			GlobalApp.Gui.DeleteView(c.name + "_ssh_passphrase")
			// ssh username
			c.formView(LTRConnectionEditComponentFormViewConfig{
				name:      c.name + "_ssh_username",
				title:     "SSH Username",
				value:     EditorInput{BindValString: &c.ConnectionConfig.SSH.Username},
				isNewLine: true,
			})
			// ssh password
			c.formView(LTRConnectionEditComponentFormViewConfig{
				name:      c.name + "_ssh_password",
				title:     "SSH Password",
				value:     EditorInput{BindValString: &c.ConnectionConfig.SSH.Password},
				isNewLine: true,
			})
		}

	} else {
		GlobalApp.Gui.DeleteView(c.name + "_ssh_host")
		GlobalApp.Gui.DeleteView(c.name + "_ssh_port")
		GlobalApp.Gui.DeleteView(c.name + "_ssh_username")
		GlobalApp.Gui.DeleteView(c.name + "_ssh_password")
		GlobalApp.Gui.DeleteView(c.name + "_ssh_key")
		GlobalApp.Gui.DeleteView(c.name + "_ssh_key_file")
		GlobalApp.Gui.DeleteView(c.name + "_ssh_login_type")
	}

	lineWitdh := (c.viewEndX - c.viewBeginX) / 4
	// btn
	c.formBtn(c.name+"_enter", "Save", 0, lineWitdh, true)
	c.formBtn(c.name+"_cancel", "Cancel", c.viewBeginX+lineWitdh, lineWitdh, false)
	c.formBtn(c.name+"_reset", "Reset", c.viewBeginX+lineWitdh*2, lineWitdh, false)
	c.formBtn(c.name+"_test", "Test", c.viewBeginX+lineWitdh*3, lineWitdh, false)

	//表单选项选中
	if c.viewNowCurrent == "" {
		c.viewNowCurrent = c.viewList[0]
	}

	v.SetCursor(0, 0)
	v.Clear()
	c.KeyBind()

	GlobalTipComponent.Layout(c.KeyMapTip())
	return c
}

func (c *LTRConnectionEditComponent) formView(config LTRConnectionEditComponentFormViewConfig) {
	if c.ConnectionConfig.Type == "group" && config.name != c.name+"_name" {
		// 分组模式下，仅显示name
		return
	}
	if config.viewType == "radio" {
		c.formViewRadio(config)
		return
	}
	name := config.name
	title := config.title
	isNewLine := config.isNewLine
	// viewType := config.viewType
	valueEditor := config.value
	xBeing := config.xBeing
	xEnd := config.xEnd

	if !config.disabledEdit {
		// 加入可选中项
		c.viewList = append(c.viewList, name)
	}
	// (*line)++
	if isNewLine {
		c.viewNowLine = c.viewNowLine + 1
	}
	if xBeing == 0 {
		xBeing = c.viewBeginX + 1
	}
	if xEnd == 0 {
		xEnd = c.viewEndX - 1
	}
	view, _ := GlobalApp.Gui.SetView(name, xBeing, c.viewBeginY+c.viewNowLine*3+1, xEnd, c.viewBeginY+c.viewNowLine*3+3)
	view.Clear()
	view.Title = " " + title + " "
	view.Frame = true

	view.FgColor = gocui.ColorWhite

	view.Clear()
	if !config.disabledEdit && (c.viewNowCurrent == name || (c.viewNowCurrent == "" && len(c.viewList) == 1)) {
		view.BgColor = gocui.ColorBlue
		view.Editable = true
		view.Editor = &valueEditor
		GlobalApp.Gui.SetCurrentView(name)
		GlobalApp.Gui.Cursor = true
	} else if !config.disabledEdit {
		view.BgColor = gocui.ColorBlack
	}
	theValue := ""
	if valueEditor.BindValString != nil {
		theValue = *valueEditor.BindValString
	} else if valueEditor.BindValInt != nil {
		theValue = fmt.Sprintf("%d", *valueEditor.BindValInt)
	}
	view.Write([]byte(theValue))
	GlobalApp.Gui.DeleteKeybindings(name)
}
func (c *LTRConnectionEditComponent) formViewRadio(config LTRConnectionEditComponentFormViewConfig) {
	name := config.name
	title := config.title
	isNewLine := config.isNewLine
	// viewType := config.viewType
	// valueEditor := config.value
	xBeing := config.xBeing
	xEnd := config.xEnd
	c.viewList = append(c.viewList, name)
	// (*line)++
	if isNewLine {
		c.viewNowLine = c.viewNowLine + 1
	}
	if xBeing == 0 {
		xBeing = c.viewBeginX + 1
	}
	if xEnd == 0 {
		xEnd = c.viewEndX - 1
	}

	view, _ := GlobalApp.Gui.SetView(name, xBeing, c.viewBeginY+c.viewNowLine*3+1, xEnd, c.viewBeginY+c.viewNowLine*3+3)
	view.Title = " " + title + " "
	view.Frame = true
	view.FgColor = gocui.ColorWhite
	view.Clear()
	if c.viewNowCurrent == name || (c.viewNowCurrent == "" && len(c.viewList) == 1) {
		view.BgColor = gocui.ColorBlue
		GlobalApp.Gui.SetCurrentView(name)
		GlobalApp.Gui.Cursor = false
		// 增加额外的快捷键提示
		c.KeyMapTipExtend = []KeyMapStruct{
			{"Choice", "<-/->"},
		}
	} else {
		view.BgColor = gocui.ColorBlack
		// c.KeyMapTipExtend = nil
	}

	theValueArr := []string{}
	// 循环配置选项
	for _, value := range config.radioMap {
		if value == config.getValue() {
			value = NewColorString(ToString(value), "black", "green", "bold")
		}
		theValueArr = append(theValueArr, ToString(value))
	}
	theValue := strings.Join(theValueArr, " / ")
	view.Write([]byte(theValue))
	GlobalApp.Gui.DeleteKeybindings(name)
	// 额外添加左右选择控制
	GuiSetKeysbinding(GlobalApp.Gui, name, []any{gocui.KeyArrowRight}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		theKey := 0
		for key, value := range config.radioMap {
			if value == config.getValue() {
				theKey = key + 1
				break
			}
		}
		if theKey >= len(config.radioMap) {
			theKey = 0
		}
		config.setValue(config.radioMap[theKey])

		c.Layout()
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, name, []any{gocui.KeyArrowLeft}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		theKey := 0
		for key, value := range config.radioMap {
			if value == config.getValue() {
				theKey = key - 1
				break
			}
		}
		if theKey < 0 {
			theKey = len(config.radioMap) - 1
		}
		config.setValue(config.radioMap[theKey])
		c.Layout()
		return nil
	})
}

func (c *LTRConnectionEditComponent) formBtn(name string, title string, xBeing int, width int, isNewLine bool) {
	c.viewList = append(c.viewList, name)
	if isNewLine {
		c.viewNowLine = c.viewNowLine + 1
	}
	if xBeing == 0 {
		xBeing = c.viewBeginX
	}
	view, _ := GlobalApp.Gui.SetView(name, xBeing, c.viewBeginY+c.viewNowLine*3+1, xBeing+width, c.viewBeginY+c.viewNowLine*3+3)
	view.Frame = false
	view.FgColor = gocui.ColorWhite
	view.Clear()
	if c.viewNowCurrent == name || (c.viewNowCurrent == "" && len(c.viewList) == 1) {
		// view.FgColor = gocui.ColorWhite
		view.BgColor = gocui.ColorBlue
		GlobalApp.Gui.SetCurrentView(name)
	} else {
		view.BgColor = gocui.Attribute(termbox.ColorDarkGray)
	}
	leftSpace := (width - len(title)) / 2
	// theTitle := " \n"
	theTitle := ""
	for i := 0; i < leftSpace; i++ {
		theTitle += " "
	}
	theTitle += title
	view.Write([]byte(theTitle))
	GlobalApp.Gui.DeleteKeybindings(name)
	// 增加额外的快捷键
	GuiSetKeysbinding(GlobalApp.Gui, name, []any{gocui.KeyArrowRight}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.keyBindTab(1)
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, name, []any{gocui.KeyArrowLeft}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.keyBindTab(-1)
		return nil
	})

}

func (c *LTRConnectionEditComponent) KeyBind() *LTRConnectionEditComponent {
	GuiSetKeysbinding(GlobalApp.Gui, c.viewList, []any{gocui.KeyTab, gocui.KeyArrowDown}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.keyBindTab(1)
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, c.viewList, []any{gocui.KeyArrowUp}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.keyBindTab(-1)
		return nil
	})
	GuiSetKeysbinding(GlobalApp.Gui, c.viewList, []any{gocui.KeyEnter}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		switch c.viewNowCurrent {
		case c.name + "_enter":
			if c.ConnectionConfig.Name == "" {
				GlobalTipComponent.LayoutTemporary("Name can not be empty", 3, TipTypeWarning)
				return nil
			}
			if c.ConnectionConfig.Type == "group" {
				var apiResult types.JSResp
				//修改分组名称
				if c.ConnectionConfigBak.Name == "" {
					// 创建分组
					apiResult = services.Connection().CreateGroup(c.ConnectionConfig.Name)
				} else {
					// 修改分组名称
					apiResult = services.Connection().RenameGroup(c.ConnectionConfigBak.Name, c.ConnectionConfig.Name)
				}
				if apiResult.Success {
					GlobalTipComponent.LayoutTemporary("Save group success", 2, TipTypeSuccess)
				} else {
					GlobalTipComponent.LayoutTemporary(apiResult.Msg, 3, TipTypeError)
				}
			} else {
				// 修改连接信息
				apiResult := services.Connection().SaveConnection(c.ConnectionConfigBak.Name, c.ConnectionConfig.ConnectionConfig)
				if apiResult.Success {
					GlobalTipComponent.LayoutTemporary("Save connection success", 2, TipTypeSuccess)
				} else {
					GlobalTipComponent.LayoutTemporary(apiResult.Msg, 3, TipTypeError)
				}
			}

			c.closeView()
			InitConnectionComponent()
			return nil
		case c.name + "_cancel":
			c.closeView()
			InitConnectionComponent()
			return nil
		case c.name + "_reset":
			c.ConnectionConfig = c.ConnectionConfigBak
			c.Layout()
		case c.name + "_test":
			// 测试连接
			if c.ConnectionConfig.Type == "group" {
				GlobalTipComponent.LayoutTemporary("Test group not support", 3, TipTypeWarning)
			} else {
				apiResult := services.Connection().TestConnection(c.ConnectionConfig.ConnectionConfig)
				if apiResult.Success {
					GlobalTipComponent.LayoutTemporary("Test connection : [success]", 2, TipTypeSuccess)
				} else {
					GlobalTipComponent.LayoutTemporary("Test connection : [error], "+apiResult.Msg, 5, TipTypeError)
				}
			}
		default:
			c.keyBindTab(1)
		}
		return nil
	})
	// 取消编辑
	GuiSetKeysbinding(GlobalApp.Gui, c.viewList, []any{gocui.KeyEsc}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.closeView()
		InitConnectionComponent()
		return nil
	})
	return c
}

func (c *LTRConnectionEditComponent) keyBindTab(mod int) {
	if len(c.viewList) > 1 {
		if mod >= 0 {
			if c.viewNowCurrent == c.viewList[len(c.viewList)-1] {
				c.viewNowCurrent = c.viewList[0]
			} else {
				for i, viewName := range c.viewList {
					if viewName == c.viewNowCurrent {
						c.viewNowCurrent = c.viewList[i+1]
						break
					}
				}
			}
		} else {
			if c.viewNowCurrent == c.viewList[0] {
				c.viewNowCurrent = c.viewList[len(c.viewList)-1]
			} else {
				for i, viewName := range c.viewList {
					if viewName == c.viewNowCurrent {
						c.viewNowCurrent = c.viewList[i-1]
						break
					}
				}
			}
		}
		c.Layout()
	}
}

func (c *LTRConnectionEditComponent) KeyMapTip() string {
	keyMap := []KeyMapStruct{
		{"Switch", "<Tab>/<Enter>"},
		{"Submit", "<Enter>"},
		{"Cancel", "<Esc>"},
	}
	keyMap = append(keyMap, c.KeyMapTipExtend...)
	ret := ""
	for i, v := range keyMap {
		if i > 0 {
			ret += " | "
		}
		ret += fmt.Sprintf("%s: %s", v.Description, v.Key)
		i++
	}
	return ret
}

func (c *LTRConnectionEditComponent) closeView() {
	for _, viewName := range c.viewList {
		GlobalApp.Gui.DeleteView(viewName)
		GlobalApp.Gui.DeleteKeybindings(viewName)
	}
	GlobalApp.Gui.DeleteView(c.name)
	GlobalApp.Gui.DeleteKeybindings(c.name)
}

func (cf *LTRConnectionEditComponentFormViewConfig) getValue() any {
	if cf.value.BindValString != nil {
		return *cf.value.BindValString
	} else if cf.value.BindValInt != nil {
		return *cf.value.BindValInt
	} else if cf.value.BindValBool != nil {
		return *cf.value.BindValBool
	} else {
		return nil
	}
}

func (cf *LTRConnectionEditComponentFormViewConfig) setValue(newValue any) {
	if cf.value.BindValString != nil {
		*cf.value.BindValString = newValue.(string)
	} else if cf.value.BindValInt != nil {
		*cf.value.BindValInt = newValue.(int)
	} else if cf.value.BindValBool != nil {
		*cf.value.BindValBool = newValue.(bool)
	} else {
		return
	}
}
