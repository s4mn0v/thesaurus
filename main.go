package main

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/awesome-gocui/gocui"
	v2 "github.com/s4mn0v/bitget/pkg/client/v2"
)

type Page int

const (
	PageDashboard Page = iota
	PageAnalytics
	PageSettings
)

var (
	currentPage    = PageDashboard
	pageNames      = []string{"DASHBOARD", "ANALYTICS", "SETTINGS"}
	mainOX, mainOY = 0, 0
	pageHeights    = map[Page]int{
		PageDashboard: 100,
		PageAnalytics: 150,
		PageSettings:  20,
	}

	// State management
	mu         sync.RWMutex
	tickerInfo string
	market     *v2.MixMarketClient
)

func main() {
	g, err := gocui.NewGui(gocui.Output256, true)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	// Initialize SDK Client
	market = new(v2.MixMarketClient).Init()

	g.Highlight = true
	g.Mouse = true
	g.SetManagerFunc(layout)

	setKeybindings(g)

	// Start background fetcher
	go fetchTickerLoop(g)

	if err := g.MainLoop(); err != nil && !errors.Is(err, gocui.ErrQuit) {
		log.Panicln(err)
	}
}

// fetchTickerLoop polls the API concurrently.
func fetchTickerLoop(g *gocui.Gui) {
	params := map[string]string{
		"symbol":      "BTCUSDT",
		"productType": "USDT-FUTURES",
	}

	t := time.NewTicker(2 * time.Second)
	defer t.Stop()

	for range t.C {
		resp, err := market.Ticker(params)

		mu.Lock()
		if err != nil {
			tickerInfo = fmt.Sprintf("Error: %v", err)
		} else {
			tickerInfo = resp
		}
		mu.Unlock()

		// Signal UI refresh
		g.Update(func(g *gocui.Gui) error { return nil })
	}
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	sidebarW := 35
	mainWidth := maxX - sidebarW - 1
	orange := gocui.Get256Color(172)

	if v, err := g.SetView("main", 0, 0, mainWidth, maxY-1, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.FrameColor = orange
		v.Wrap = true
	} else {
		v.Clear()
		v.Title = " " + pageNames[currentPage] + " "
		drawContent(v)
		_ = v.SetOrigin(mainOX, mainOY)
	}

	// Navigation View
	if v, err := g.SetView("side_top", mainWidth+1, 0, maxX-1, 11, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.FrameColor = orange
		v.Title = " Navigation "
	} else {
		v.Clear()
		fmt.Fprintln(v, "\n PAGES:")
		for i, name := range pageNames {
			if Page(i) == currentPage {
				fmt.Fprintf(v, "  \033[30;47m %-12s \033[0m\n", name)
			} else {
				fmt.Fprintf(v, "    %-12s \n", name)
			}
		}
		fmt.Fprintf(v, "\n (TAB to switch)")
	}

	// Technical View
	if v, err := g.SetView("side_bottom", mainWidth+1, 12, maxX-1, maxY-1, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.FrameColor = orange
		v.Title = " Market Status "
	} else {
		v.Clear()
		mu.RLock()
		fmt.Fprintf(v, "\n  Symbol: BTCUSDT\n  Data size: %d bytes", len(tickerInfo))
		mu.RUnlock()
	}

	_, _ = g.SetCurrentView("main")
	return nil
}

func drawContent(v *gocui.View) {
	switch currentPage {
	case PageDashboard:
		mu.RLock()
		data := tickerInfo
		mu.RUnlock()
		if data == "" {
			fmt.Fprintln(v, "Loading market data...")
		} else {
			fmt.Fprintln(v, data)
		}
	case PageAnalytics:
		for i := range pageHeights[PageAnalytics] {
			fmt.Fprintf(v, "[%03d] ANALYTICS DATA: Index #%X\n", i, i*255)
		}
	case PageSettings:
		fmt.Fprintln(v, "Configuration parameters...")
	}
}

func setKeybindings(g *gocui.Gui) {
	_ = g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit)
	_ = g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, nextPage)
	_ = g.SetKeybinding("main", gocui.KeyArrowDown, gocui.ModNone, scrollDown)
	_ = g.SetKeybinding("main", gocui.KeyArrowUp, gocui.ModNone, scrollUp)
}

func scrollDown(g *gocui.Gui, v *gocui.View) error {
	_, vy := v.Size()

	lines := len(v.BufferLines())

	if mainOY < lines-vy {
		mainOY++
	}
	return nil
}

func scrollUp(g *gocui.Gui, v *gocui.View) error {
	if mainOY > 0 {
		mainOY--
	}
	return nil
}

func nextPage(g *gocui.Gui, v *gocui.View) error {
	currentPage = Page((int(currentPage) + 1) % len(pageNames))
	mainOX, mainOY = 0, 0
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error { return gocui.ErrQuit }
