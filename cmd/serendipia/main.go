package main

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/awesome-gocui/gocui"
	"github.com/s4mn0v/serendipia/internal/trading"
	"github.com/s4mn0v/serendipia/internal/ui"
)

var (
	currentPage   = 0
	pageNames     = []string{"DASHBOARD", "ANALYTICS", "SETTINGS"}
	mainOY        = 0
	mu            sync.RWMutex
	currentTicker *trading.TickerData
	engine        *trading.Engine
	appLogger     *ui.UILogger
	showHelp      bool
)

func main() {
	g, err := gocui.NewGui(gocui.Output256, true)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	engine = trading.NewEngine()
	appLogger = ui.NewUILogger(g)

	g.SetManagerFunc(layout)
	setKeybindings(g)

	go fetchLoop(g)

	if err := g.MainLoop(); err != nil && !errors.Is(err, gocui.ErrQuit) {
		log.Panicln(err)
	}
}

func fetchLoop(g *gocui.Gui) {
	t := time.NewTicker(3 * time.Second)
	defer t.Stop()
	for range t.C {
		data, err := engine.FetchTicker("BTCUSDT")
		mu.Lock()
		if err != nil {
			appLogger.Log("\033[31mFetch Error:\033[0m %v", err)
		} else {
			currentTicker = data
			appLogger.Log("Market updated: %s", data.Symbol)
		}
		mu.Unlock()
		g.Update(func(g *gocui.Gui) error { return nil })
	}
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	sideW := 40
	mainW := maxX - sideW - 1
	orange := gocui.Get256Color(172)
	footerY := maxY - 2

	// 1. MAIN CONTENT (Left)
	if v, err := g.SetView("main", 0, 0, mainW, footerY, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Title, v.FrameColor, v.Wrap = " "+pageNames[currentPage]+" ", orange, true
	} else {
		v.Clear()
		drawPage(v)
		v.SetOrigin(0, mainOY)
	}

	// 2. NAVIGATION (Right Top)
	// Starts at y=0, ends at y=10
	if v, err := g.SetView("side_top", mainW+1, 0, maxX-1, 10, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Title, v.FrameColor = " Navigation ", orange
	} else {
		v.Clear()
		for i, name := range pageNames {
			if i == currentPage {
				fmt.Fprintf(v, " > \033[30;47m %s \033[0m\n", name)
			} else {
				fmt.Fprintf(v, "   %s \n", name)
			}
		}
	}

	// 3. LOGS & STATUS (Right Bottom)
	// Starts at y=11 (immediately after Navigation), ends at footerY
	if v, err := g.SetView("side_bottom", mainW+1, 11, maxX-1, footerY, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Title, v.FrameColor, v.Autoscroll = " Logs & Status ", orange, true
	}

	// 4. FOOTER (Bottom Status Bar)
	// No Frame. Uses the remaining space below footerY
	if v, err := g.SetView("footer", -1, footerY, maxX, maxY, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Frame = false
	} else {
		v.Clear()
		// fmt.Fprintf(v, " \033[33mCommit:\033[0m \033[36mc\033[0m | ")
		// fmt.Fprintf(v, "\033[33mStash:\033[0m \033[36ms\033[0m | ")
		// fmt.Fprintf(v, "\033[33mReset:\033[0m \033[36mD\033[0m | ")
		fmt.Fprintf(v, "\033[33mKeybindings:\033[0m \033[36m?\033[0m")
	}

	if showHelp {
		width, height := 60, 10
		x0 := (maxX / 2) - (width / 2)
		y0 := (maxY / 2) - (height / 2)
		x1 := x0 + width
		y1 := y0 + height

		if v, err := g.SetView("help", x0, y0, x1, y1, 0); err != nil {
			if !errors.Is(err, gocui.ErrUnknownView) {
				return err
			}
			v.Title = " Help / Keybindings "
			v.FrameColor = gocui.ColorCyan
			v.Wrap = true
			fmt.Fprintln(v, "\n [TAB]  Switch Pages")
			fmt.Fprintln(v, " [UP/DN] Scroll Content")
			fmt.Fprintln(v, " [?]     Toggle this help")
			fmt.Fprintln(v, " [^C]    Quit Application")

			if _, err := g.SetCurrentView("help"); err != nil {
				return err
			}
		}
	} else {
		g.DeleteView("help")
	}

	return nil
}

func drawPage(v *gocui.View) {
	mu.RLock()
	defer mu.RUnlock()

	if currentPage == 0 { // DASHBOARD
		if currentTicker == nil {
			fmt.Fprintln(v, "Connecting to exchange...")
			return
		}

		// Clean, filtered output
		fmt.Fprintf(v, "\n \033[1mMARKET TICKER\033[0m\n")
		fmt.Fprintf(v, " ─────────────\n")
		fmt.Fprintf(v, " ASSET: %s\n", currentTicker.Symbol)
		fmt.Fprintf(v, " PRICE: \033[32m$%s\033[0m\n", currentTicker.Price)
	} else {
		fmt.Fprintf(v, "Page: %s content", pageNames[currentPage])
	}
}

func setKeybindings(g *gocui.Gui) {
	g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error { return gocui.ErrQuit })
	g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		currentPage = (currentPage + 1) % len(pageNames)
		mainOY = 0
		// appLogger.Log("Switched to %s", pageNames[currentPage])
		return nil
	})
	g.SetKeybinding("main", gocui.KeyArrowDown, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		_, vy := v.Size()
		if mainOY < len(v.BufferLines())-vy {
			mainOY++
		}
		return nil
	})
	g.SetKeybinding("main", gocui.KeyArrowUp, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if mainOY > 0 {
			mainOY--
		}
		return nil
	})

	g.SetKeybinding("", '?', gocui.ModNone, toggleHelp)

	// Also allow closing the help with 'Enter' or 'Esc' while focused on it
	g.SetKeybinding("help", gocui.KeyEsc, gocui.ModNone, toggleHelp)
}

func toggleHelp(g *gocui.Gui, v *gocui.View) error {
	showHelp = !showHelp
	if !showHelp {
		// Return focus to main view when closing
		_, err := g.SetCurrentView("main")
		return err
	}
	appLogger.Log("Opened Help menu")
	return nil
}
