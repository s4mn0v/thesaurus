package main

import (
	"errors"
	"log"
	"time"

	"github.com/awesome-gocui/gocui"
	"github.com/s4mn0v/thesaurus/internal/trading"
	"github.com/s4mn0v/thesaurus/internal/ui"
	"github.com/s4mn0v/thesaurus/internal/ui/pages"
)

func main() {
	g, err := gocui.NewGui(gocui.Output256, true)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	engine := trading.NewEngine()

	appPages := []ui.Page{
		&pages.Dashboard{},
		&pages.Analytics{},
	}

	manager := ui.NewManager(appPages)
	manager.SetLogger(ui.NewUILogger(g))

	g.SetManagerFunc(manager.Layout)
	manager.SetKeybindings(g)

	dataStream := engine.StreamTicker("BTCUSDT", 1*time.Second)
	manager.Listen(g, dataStream)

	if err := g.MainLoop(); err != nil && !errors.Is(err, gocui.ErrQuit) {
		log.Panicln(err)
	}
}
