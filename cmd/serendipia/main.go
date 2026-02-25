package main

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/awesome-gocui/gocui"
	"github.com/gookit/color"
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
	helpBinding   = 0
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
		w, h := 80, 20
		x0, y0 := (maxX/2)-(w/2), (maxY/2)-(h/2)
		x1, y1 := x0+w, y0+h

		if v, err := g.SetView("help", x0, y0, x1, y1-4, 0); err != nil {
			if !errors.Is(err, gocui.ErrUnknownView) {
				return err
			}
			v.Title, v.FrameColor = " Keybindings ", gocui.ColorCyan
			g.SetCurrentView("help")
		} else {
			v.Clear()

			// Define Styles
			selectedStyle := color.Style{color.FgBlack, color.BgHiCyan, color.OpBold}
			keyStyle := color.FgCyan
			sectionStyle := color.FgLightBlue

			currentSection := ""
			for i, b := range ui.HelpMenu {
				// Section Header handling
				if b.Section != currentSection {
					currentSection = b.Section
					fmt.Fprint(v, sectionStyle.Render(fmt.Sprintf("\n --- %s --- \n", currentSection)))
				}

				// Content Formatting
				// We pad the string to ensure the background color forms a full bar
				line := fmt.Sprintf("  %-10s %-50s", b.Key, b.Desc)

				if i == helpBinding {
					// Apply background color for the selected item
					fmt.Fprintln(v, selectedStyle.Render(line))
				} else {
					// Normal rendering
					fmt.Fprintf(v, "%s %s\n", keyStyle.Render(fmt.Sprintf("  %-10s", b.Key)), b.Desc)
				}
			}
			v.Subtitle = fmt.Sprintf(" %d of %d ", helpBinding+1, len(ui.HelpMenu))
		}

		if v, err := g.SetView("help_desc", x0, y1-3, x1, y1, 0); err != nil {
			if !errors.Is(err, gocui.ErrUnknownView) {
				return err
			}
			v.FrameColor, v.Title = gocui.ColorCyan, " Description "
		} else {
			v.Clear()
			if helpBinding < len(ui.HelpMenu) {
				fmt.Fprintf(v, " %s", ui.HelpMenu[helpBinding].Desc)
			}
		}
		g.SetCurrentView("help")
	} else {
		g.DeleteView("help")
		g.DeleteView("help_desc")
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

	// Navigation inside Help
	g.SetKeybinding("help", gocui.KeyArrowDown, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if helpBinding < len(ui.HelpMenu)-1 {
			helpBinding++
		}
		return nil
	})
	g.SetKeybinding("help", gocui.KeyArrowUp, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if helpBinding > 0 {
			helpBinding--
		}
		return nil
	})
	g.SetKeybinding("help", gocui.KeyEnter, gocui.ModNone, toggleHelp)
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
