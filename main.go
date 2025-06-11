package main

import (
	"log"
	"tinyrdm-tui/service"
	"tinyrdm/backend/services"

	"github.com/jroimartin/gocui"
)

func main() {

	g, err := gocui.NewGui(gocui.OutputNormal)
	g.Mouse = true
	if err != nil {
		log.Panicln(err)
	}
	defer func() {
		services.Browser().Stop()
		g.Close()
	}()

	service.NewMainApp(g)

	// g.SetManagerFunc(service.Layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}

}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
