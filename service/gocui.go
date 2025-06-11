package service

import "github.com/jroimartin/gocui"

// GuiSetKeysbinding set keysbinding for a view
func GuiSetKeysbinding(g *gocui.Gui, viewname string, keys []any, mod gocui.Modifier, handler func(*gocui.Gui, *gocui.View) error) error {
	for _, key := range keys {
		if err := g.SetKeybinding(viewname, key, mod, handler); err != nil {
			return err
		}
	}
	return nil
}
