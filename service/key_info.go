package service

import (
	"fmt"
	"strconv"
	"tinyrdm/backend/services"
	"tinyrdm/backend/types"

	"github.com/jroimartin/gocui"
)

type LTRKeyInfoComponent struct {
	name       string
	title      string
	keyName    string
	LayoutMaxY int
	keyView    *gocui.View
	keyViewTTL *gocui.View
}

func InitKeyInfoComponent() {
	GlobalKeyInfoComponent = &LTRKeyInfoComponent{
		name:       "key_info",
		title:      "Info",
		LayoutMaxY: 0,
	}
	GlobalKeyInfoComponent.Layout()
	GlobalApp.ViewNameList = append(GlobalApp.ViewNameList, GlobalKeyInfoComponent.name)
}

func (c *LTRKeyInfoComponent) Layout() *LTRKeyInfoComponent {
	theX0, _ := GlobalDBComponent.view.Size()
	theX0 = theX0 + 2
	var err error
	var theTTL int64
	// show key info
	c.keyView, err = GlobalApp.Gui.SetView(c.name, theX0, 0, GlobalApp.maxX-25, 2)
	if err == nil || err != gocui.ErrUnknownView {
		keySummary := services.Browser().GetKeySummary(types.KeySummaryParam{
			Server: GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
			DB:     GlobalDBComponent.SelectedDB,
			Key:    c.keyName,
		})
		c.keyView.Title = " " + c.title + " "
		printString := ""
		if keySummary.Success {
			keySummaryData := keySummary.Data.(types.KeySummary)
			printString = NewTypeWord(keySummaryData.Type, "full") + " " + c.keyName
			theTTL = keySummaryData.TTL
		}
		c.keyView.Clear()
		c.keyView.Write([]byte(printString))
	}

	// show key ttl
	c.keyViewTTL, err = GlobalApp.Gui.SetView(c.name+"_ttl", GlobalApp.maxX-24, 0, GlobalApp.maxX-1, 2)
	if err == nil || err != gocui.ErrUnknownView {
		c.keyViewTTL.Clear()
		if theTTL >= 0 {
			c.keyViewTTL.Write([]byte(NewColorString("TTL: "+strconv.FormatInt(theTTL, 10)+" s"+SPACE_STRING, "black", "green", "bold")))
		} else {
			c.keyViewTTL.Write([]byte(NewColorString("TTL: "+strconv.FormatInt(theTTL, 10)+" s"+SPACE_STRING, "white", "red", "bold")))
		}
	}

	// show key detail

	if GlobalApp.Gui.CurrentView().Name() == GlobalKeyInfoComponent.name {
		// GlobalApp.Gui.SetCurrentView(GlobalKeyInfoComponent.name)
		GlobalTipComponent.Layout(c.KeyMapTip())
	}

	return c
}

func (c *LTRKeyInfoComponent) KeyMapTip() string {
	keyMap := []KeyMapStruct{
		{"Switch", "<Tab>"},
	}
	ret := ""
	for i, v := range keyMap {
		if i > 0 {
			ret += " | "
		}
		ret += fmt.Sprintf("%s: %s", v.Description, v.Key)
	}
	// return "key_info: " + ret
	return ret
}
