package service

import (
	"fmt"
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
}

var keyValueFormatList = []string{"Raw", "JSON"}

func InitKeyInfoDetailComponent() {
	GlobalKeyInfoDetailComponent = &LTRKeyInfoDetailComponent{
		name:           "key_info_detail",
		title:          "Key Value",
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
		c.view.Wrap = true
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
			}
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
		c.Layout()
		GlobalKeyInfoDetailComponent.Layout()
		return nil
	})
}
