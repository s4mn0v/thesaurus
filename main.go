package main

import (
	"errors"
	"log"
	"time"

	"github.com/awesome-gocui/gocui"
	"github.com/s4mn0v/thesaurus/internal/trading"
	"github.com/s4mn0v/thesaurus/internal/ui"
)

func main() {
	g, err := gocui.NewGui(gocui.Output256, true)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	engine := trading.NewEngine()
	manager := ui.NewViewManager()
	logger := ui.NewUILogger(g)
	manager.SetLogger(logger)

	g.SetManagerFunc(manager.Layout)
	manager.SetKeybindings(g)

	go func() {
		// Faster ticker for UI responsiveness (50ms)
		uiTicker := time.NewTicker(50 * time.Millisecond)
		// Separate ticker for data simulation (1s)
		dataTicker := time.NewTicker(1 * time.Second)

		defer uiTicker.Stop()
		defer dataTicker.Stop()

		for {
			select {
			case <-dataTicker.C:
				data, err := engine.FetchTicker("BTCUSDT")
				if err != nil {
					logger.Log("\033[31mError:\033[0m %v", err)
					continue
				}
				manager.UpdateTicker(data)
				logger.Log("Market updated: %s", data.Price)
			case <-uiTicker.C:
				g.Update(func(g *gocui.Gui) error { return nil })
			}
		}
	}()

	if err := g.MainLoop(); err != nil && !errors.Is(err, gocui.ErrQuit) {
		log.Panicln(err)
	}
}
