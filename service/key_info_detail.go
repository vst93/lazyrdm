package service

import (
	"fmt"
	"strings"
	"tinyrdm/backend/services"
	"tinyrdm/backend/types"

	"github.com/duke-git/lancet/v2/validator"
	"github.com/jroimartin/gocui"
)

type LTRKeyInfoDetailComponent struct {
	name           string
	title          string
	LayoutMaxY     int
	view           *gocui.View
	keyValueFormat string
	viewOriginY    int
	keyValueMaxY   int
}

var keyValueFormatList = []string{"Raw", "JSON"}

func InitKeyInfoDetailComponent() {
	GlobalKeyInfoDetailComponent = &LTRKeyInfoDetailComponent{
		name:           "key_info_detail",
		title:          "Detail",
		LayoutMaxY:     0,
		keyValueFormat: "Raw",
	}
	GlobalKeyInfoDetailComponent.Layout().KeyBind()
	// c.KeyBind()
	GlobalApp.ViewNameList = append(GlobalApp.ViewNameList, GlobalKeyInfoDetailComponent.name)
}

func (c *LTRKeyInfoDetailComponent) Layout() *LTRKeyInfoDetailComponent {
	theX0, _ := GlobalDBComponent.view.Size()
	theX0 = theX0 + 2
	var err error
	// show key detail
	c.view, err = GlobalApp.gui.SetView(c.name, theX0, 3, GlobalApp.maxX-1, GlobalApp.maxY-2)
	if err == nil || err != gocui.ErrUnknownView {
		c.keyValueMaxY = 0
		c.view.Wrap = true
		c.view.Title = " " + c.title + " "
		keyDetail := services.Browser().GetKeyDetail(types.KeyDetailParam{
			Server: GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
			DB:     GlobalDBComponent.SelectedDB,
			Key:    GlobalKeyInfoComponent.keyName,
		})
		if keyDetail.Success {
			keyDetailData := keyDetail.Data.(types.KeyDetail)
			c.view.Clear()
			theVal := fmt.Sprintln(keyDetailData.Value)
			// format json data
			if c.keyValueFormat == "JSON" && validator.IsJSON(theVal) {
				theVal, _ = PrettyString(theVal)
				// c.view.Wrap = false
			}
			theValSlice := strings.Split(theVal, "\n")
			theViewX, _ := c.view.Size()
			for _, line := range theValSlice {
				lineLen := len(line)
				if lineLen > theViewX {
					c.keyValueMaxY += lineLen / theViewX
					if lineLen%theViewX > 0 {
						c.keyValueMaxY++
					}
				} else {
					c.keyValueMaxY++
				}
			}
			PrintLn(c.keyValueMaxY)
			c.view.Write([]byte(theVal))
		}
	}

	// show format select
	formatSelectView, err := GlobalApp.gui.SetView("key_value_format", GlobalApp.maxX-15, GlobalApp.maxY-4, GlobalApp.maxX-1, GlobalApp.maxY-2)
	if err == nil {
		formatSelectView.Frame = true
		formatSelectView.Clear()
		formatSelectView.Write([]byte("Format: " + c.keyValueFormat))
	}
	c.view.SetOrigin(0, c.viewOriginY)

	if GlobalApp.CurrentView == c.name {
		GlobalApp.gui.SetCurrentView(c.name)
	}

	return c
}

func (c *LTRKeyInfoDetailComponent) KeyBind() {
	GuiSetKeysbinding(GlobalApp.gui, c.name, []any{gocui.KeyCtrlF}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		nextIndex := 0
		for i, format := range keyValueFormatList {
			if format == c.keyValueFormat {
				nextIndex = i + 1
				break
			}
		}
		if nextIndex >= len(keyValueFormatList) {
			nextIndex = 0
		}
		c.keyValueFormat = keyValueFormatList[nextIndex]
		c.viewOriginY = 0
		c.Layout()
		GlobalKeyInfoDetailComponent.Layout()
		return nil
	})

	GuiSetKeysbinding(GlobalApp.gui, c.name, []any{gocui.KeyArrowUp, gocui.MouseWheelUp}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.viewOriginY--
		if c.viewOriginY < 0 {
			c.viewOriginY = 0
		}
		c.view.SetOrigin(0, c.viewOriginY)
		return nil
	})

	GuiSetKeysbinding(GlobalApp.gui, c.name, []any{gocui.KeyArrowDown, gocui.MouseWheelDown}, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		_, theViewY := c.view.Size()
		if c.viewOriginY-1 >= c.keyValueMaxY-theViewY {
			return nil
		}
		c.viewOriginY++
		c.view.SetOrigin(0, c.viewOriginY)
		return nil
	})

}
