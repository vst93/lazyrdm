package main

import (
	"tinyrdm-tui/service"

	"log"

	"github.com/jroimartin/gocui"
)

var ConnectionList any

func main() {
	// fmt.Println(r)

	g, err := gocui.NewGui(gocui.OutputNormal)
	g.Mouse = true
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	service.InitConnectionComponent(g).KeyBind().Layout()

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
