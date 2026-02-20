package main

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/awesome-gocui/gocui"
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
	// Height of the content for each page to bound scrolling
	pageHeights = map[Page]int{
		PageDashboard: 100,
		PageAnalytics: 150,
		PageSettings:  20,
	}
)

func main() {
	g, err := gocui.NewGui(gocui.Output256, true)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.Highlight = true
	g.Mouse = true
	g.SetManagerFunc(layout)

	setKeybindings(g)

	if err := g.MainLoop(); err != nil && !errors.Is(err, gocui.ErrQuit) {
		log.Panicln(err)
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
		v.Wrap = false // Ensure horizontal scroll works
	} else {
		v.Clear()
		v.Title = " " + pageNames[currentPage] + " "

		drawContent(v)

		// Set origin AFTER drawing content so the buffer is populated
		if err := v.SetOrigin(mainOX, mainOY); err != nil {
			// Reset if coordinates become invalid (e.g. on terminal shrink)
			mainOX, mainOY = 0, 0
			_ = v.SetOrigin(0, 0)
		}
	}

	// Sidebar Top
	if v, err := g.SetView("side_top", mainWidth+1, 0, maxX-1, 11, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.FrameColor = orange
		v.Title = " Status "
	} else {
		v.Clear()
		fmt.Fprintf(v, "\n  PAGE: %s\n  (TAB to switch)", pageNames[currentPage])
	}

	// Sidebar Bottom
	if v, err := g.SetView("side_bottom", mainWidth+1, 12, maxX-1, maxY-1, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.FrameColor = orange
		v.Title = " Technical "
	} else {
		v.Clear()
		fmt.Fprintf(v, "\n  Scroll Y: %d\n  Scroll X: %d\n  Limit:    %d", mainOY, mainOX, pageHeights[currentPage])
	}

	_, _ = g.SetCurrentView("main")
	return nil
}

func drawContent(v *gocui.View) {
	height := pageHeights[currentPage]
	for i := 0; i < height; i++ {
		switch currentPage {
		case PageDashboard:
			fmt.Fprintf(v, "[%03d] DASHBOARD FEED: Monitoring System Nodes... %s\n", i, strings.Repeat(">", i/5))
		case PageAnalytics:
			fmt.Fprintf(v, "[%03d] ANALYTICS DATA: Market Volume Index #%X\n", i, i*255)
		case PageSettings:
			fmt.Fprintf(v, "[%03d] SETTING OPTION: Parameter Config Line\n", i)
		}
	}
}

func setKeybindings(g *gocui.Gui) {
	_ = g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit)
	_ = g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, nextPage)

	_ = g.SetKeybinding("main", gocui.KeyArrowDown, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		_, vy := v.Size()
		// Only scroll if there is content below the current view
		if mainOY < pageHeights[currentPage]-(vy) {
			mainOY++
		}
		return nil
	})
	_ = g.SetKeybinding("main", gocui.KeyArrowUp, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if mainOY > 0 {
			mainOY--
		}
		return nil
	})

	_ = g.SetKeybinding("main", gocui.KeyArrowRight, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		mainOX++
		return nil
	})
	_ = g.SetKeybinding("main", gocui.KeyArrowLeft, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if mainOX > 0 {
			mainOX--
		}
		return nil
	})
}

func nextPage(g *gocui.Gui, v *gocui.View) error {
	currentPage = (currentPage + 1) % 3
	mainOX, mainOY = 0, 0 // Reset scroll when changing page
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error { return gocui.ErrQuit }
