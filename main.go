package main

import (
	"fmt"

	"context"
	"log"

	"tinyrdm/backend/services"
	"tinyrdm/backend/types"

	"github.com/jroimartin/gocui"
)

var connectionList any

func main() {
	connSvc := services.Connection()
	browserSvc := services.Browser()
	ctx := context.Background()
	connSvc.Start(ctx)
	browserSvc.Start(ctx)
	connectionList = connSvc.ListConnection()
	// fmt.Println(r)

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("connection_list", maxX/100, maxY/100, maxX/10*2, maxY/2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Connection List"
		v.Editable = true
		v.Wrap = true
		v.Autoscroll = true
		v.Frame = true
		v.FgColor = gocui.ColorGreen
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
		v.Highlight = true
		response, ok := connectionList.(types.JSResp).Data.(types.Connections)

		if !ok {
			return fmt.Errorf("connectionList is not of type Response")
		}

		for _, conn := range response {
			fmt.Fprintln(v, conn.Name)
		}

	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
