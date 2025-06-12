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
	keyDetail  *gocui.View
}

func InitKeyInfoComponent() *LTRKeyInfoComponent {
	c := &LTRKeyInfoComponent{
		name:       "key_info",
		title:      "Key Info",
		LayoutMaxY: 0,
	}
	c.Layout()
	// c.KeyBind()
	if GlobalApp.CurrentView == c.name {
		GlobalApp.gui.SetCurrentView(c.name)
	}
	return c
}

func (c *LTRKeyInfoComponent) Layout() *LTRKeyInfoComponent {
	theX0, _ := GlobalDBComponent.view.Size()
	theX0 = theX0 + 2
	var err error
	var theTTL int64
	// show key info
	c.keyView, err = GlobalApp.gui.SetView(c.name+"_key", theX0, 0, GlobalApp.maxX-25, 2)
	if err == nil || err != gocui.ErrUnknownView {
		keySummary := services.Browser().GetKeySummary(types.KeySummaryParam{
			Server: GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
			DB:     GlobalDBComponent.SelectedDB,
			Key:    c.keyName,
		})
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
	c.keyViewTTL, err = GlobalApp.gui.SetView(c.name+"_ttl", GlobalApp.maxX-24, 0, GlobalApp.maxX-1, 2)
	if err == nil || err != gocui.ErrUnknownView {
		c.keyViewTTL.Clear()
		if theTTL >= 0 {
			c.keyViewTTL.Write([]byte(NewColorString("TTL: "+strconv.FormatInt(theTTL, 10)+" s"+SPACE_STRING, "black", "green", "bold")))
		} else {
			c.keyViewTTL.Write([]byte(NewColorString("TTL: "+strconv.FormatInt(theTTL, 10)+" s"+SPACE_STRING, "white", "red", "bold")))
		}
	}

	// show key detail
	c.keyDetail, err = GlobalApp.gui.SetView(c.name+"_detail", theX0, 3, GlobalApp.maxX-1, GlobalApp.maxY-2)
	if err == nil || err != gocui.ErrUnknownView {
		c.keyDetail.Wrap = true
		keyDetail := services.Browser().GetKeyDetail(types.KeyDetailParam{
			Server: GlobalConnectionComponent.ConnectionListSelectedConnectionInfo.Name,
			DB:     GlobalDBComponent.SelectedDB,
			Key:    c.keyName,
		})
		if keyDetail.Success {
			keyDetailData := keyDetail.Data.(types.KeyDetail)
			c.keyDetail.Clear()
			c.keyDetail.Write([]byte(fmt.Sprintln(keyDetailData.Value)))
		}
	}

	return c
}
