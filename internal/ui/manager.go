package ui

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/awesome-gocui/gocui"
	"github.com/s4mn0v/thesaurus/internal/trading"
	"github.com/s4mn0v/thesaurus/internal/ui/pages"
)

type Manager struct {
	mu              sync.RWMutex
	pages           []Page
	activeIdx       int
	mainOY          int
	logger          *UILogger
	showHelp        bool
	showNav         bool
	helpBinding     int
	showExitConfirm bool
}

func NewManager(p []Page) *Manager {
	return &Manager{pages: p, showNav: true}
}

func (m *Manager) SetLogger(l *UILogger) { m.logger = l }

func (m *Manager) Layout(g *gocui.Gui) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	maxX, maxY := g.Size()
	orange := gocui.Get256Color(208)
	activePage := m.pages[m.activeIdx]

	navH, footerH, sideW, overviewH := -1, 2, 45, 6
	if m.showNav {
		navH = 2
	}

	cY0, cY1 := navH+1, maxY-footerH
	mX1, rX0 := maxX-sideW-1, maxX-sideW
	oY1 := cY0 + overviewH

	// 1. Navigation
	if m.showNav {
		if v, err := g.SetView("navigation", 0, 0, maxX-1, navH, 0); err != nil {
			if !errors.Is(err, gocui.ErrUnknownView) {
				return err
			}
			v.Title, v.FrameColor = " Navigation ", orange
		} else {
			v.Clear()
			for i, p := range m.pages {
				if i == m.activeIdx {
					fmt.Fprintf(v, "\033[38;5;208m %s \033[0m  ", p.Name())
				} else {
					fmt.Fprintf(v, " %s   ", p.Name())
				}
			}
		}
	} else {
		g.DeleteView("navigation")
	}

	// 2. Main
	if v, err := g.SetView("main", 0, cY0, mX1, cY1, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Wrap = true
	} else {
		v.Title = " " + activePage.Name() + " "
		v.Clear()
		v.SetOrigin(0, m.mainOY)
		activePage.Layout(v)
	}

	// 3. Overview
	if v, err := g.SetView("description", rX0, cY0, maxX-1, oY1, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Title, v.FrameColor, v.Wrap = " Overview ", orange, true
	} else {
		v.Clear()
		fmt.Fprintf(v, "\033[36mMode:\033[0m %s\n\033[37m%s\033[0m", activePage.Name(), activePage.Description())
	}

	// 4. Logs (Visual Example Feature)
	if v, err := g.SetView("side_bottom", rX0, oY1+1, maxX-1, cY1, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Title, v.FrameColor, v.Autoscroll, v.Wrap = " Logs & Activity ", orange, true, true
	}

	// 5. Footer
	if v, err := g.SetView("footer", -1, cY1, maxX, maxY, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Frame = false
	} else {
		v.Clear()
		fmt.Fprint(v, "\033[33mHelp:\033[0m \033[36m?\033[0m | \033[33mNav:\033[0m \033[36mCtrl+N\033[0m")
	}

	if err := m.renderHelp(g, maxX, maxY); err != nil {
		return err
	}
	return m.renderExitConfirm(g, maxX, maxY)
}

func (m *Manager) Listen(g *gocui.Gui, dataStream <-chan *trading.TickerData) {
	go func() {
		for data := range dataStream {
			m.mu.Lock()
			for _, p := range m.pages {
				if d, ok := p.(*pages.Dashboard); ok {
					d.Ticker = data
				}
			}
			m.mu.Unlock()
			g.Update(func(g *gocui.Gui) error { return nil })
		}
	}()
}

func (m *Manager) renderHelp(g *gocui.Gui, maxX, maxY int) error {
	if !m.showHelp {
		g.DeleteView("help")
		g.DeleteView("help_desc")
		return nil
	}

	w, h := maxX*8/10, maxY*8/10
	x0, y0 := (maxX-w)/2, (maxY-h)/2

	// 1. Help List View
	v, err := g.SetView("help", x0, y0, x0+w, y0+h-3, 0)
	if err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Title, v.FrameColor = " Keybindings ", gocui.ColorCyan
		g.SetCurrentView("help")
	}

	m.updateHelpContent(v)

	// 2. Help Description View
	if vd, err := g.SetView("help_desc", x0, y0+h-2, x0+w, y0+h, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		vd.Title, vd.FrameColor = " Description ", gocui.ColorCyan
	} else {
		vd.Clear()
		activeBindings := GetCurrentHelp(m.pages[m.activeIdx].Name())
		if m.helpBinding < len(activeBindings) {
			fmt.Fprintf(vd, " %s", activeBindings[m.helpBinding].Desc)
		}
	}

	return nil
}

func (m *Manager) updateHelpContent(v *gocui.View) {
	v.Clear()
	activeBindings := GetCurrentHelp(m.pages[m.activeIdx].Name())
	var sb strings.Builder
	for i, b := range activeBindings {
		line := fmt.Sprintf("  %-10s %-50s", b.Key, b.Desc)
		if i == m.helpBinding {
			sb.WriteString("\033[30;106;1m" + line + "\033[0m\n")
		} else {
			sb.WriteString(fmt.Sprintf("  \033[36m%-10s\033[0m %s\n", b.Key, b.Desc))
		}
	}
	v.Write([]byte(sb.String()))
}

func (m *Manager) renderExitConfirm(g *gocui.Gui, maxX, maxY int) error {
	if !m.showExitConfirm {
		g.DeleteView("exit_confirm")
		return nil
	}
	w, h := 30, 2
	x0, y0 := (maxX-w)/2, (maxY-h)/2
	if v, err := g.SetView("exit_confirm", x0, y0, x0+w, y0+h, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Title, v.FrameColor = " Quit? ", gocui.ColorRed
		g.SetCurrentView("exit_confirm")
		fmt.Fprint(v, " Are you sure? (y/n)")
	}
	return nil
}

// Logic for Keybindings Selection
func (m *Manager) helpDown(g *gocui.Gui, v *gocui.View) error {
	if m.helpBinding < len(GetCurrentHelp(m.pages[m.activeIdx].Name()))-1 {
		m.helpBinding++
	}
	return nil
}

func (m *Manager) helpUp(g *gocui.Gui, v *gocui.View) error {
	if m.helpBinding > 0 {
		m.helpBinding--
	}
	return nil
}

func (m *Manager) SetKeybindings(g *gocui.Gui) {
	g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, m.nextPage)
	g.SetKeybinding("", '?', gocui.ModNone, m.ToggleHelp)
	g.SetKeybinding("", gocui.KeyCtrlN, gocui.ModNone, m.ToggleNav)
	g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, m.toggleExitConfirm)

	// Main scrolling
	g.SetKeybinding("main", gocui.KeyArrowDown, gocui.ModNone, m.scrollDown)
	g.SetKeybinding("main", gocui.KeyArrowUp, gocui.ModNone, m.scrollUp)

	// Help navigation
	g.SetKeybinding("help", gocui.KeyArrowDown, gocui.ModNone, m.helpDown)
	g.SetKeybinding("help", gocui.KeyArrowUp, gocui.ModNone, m.helpUp)
	g.SetKeybinding("help", 'q', gocui.ModNone, m.ToggleHelp)
	g.SetKeybinding("help", gocui.KeyEsc, gocui.ModNone, m.ToggleHelp)

	// Exit confirm
	g.SetKeybinding("exit_confirm", 'y', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error { return gocui.ErrQuit })
	g.SetKeybinding("exit_confirm", 'n', gocui.ModNone, m.toggleExitConfirm)
}

// Remaining helper methods
func (m *Manager) nextPage(g *gocui.Gui, v *gocui.View) error {
	m.mu.Lock()
	m.activeIdx = (m.activeIdx + 1) % len(m.pages)
	m.mainOY = 0
	m.helpBinding = 0
	m.mu.Unlock()
	return nil
}

func (m *Manager) scrollDown(g *gocui.Gui, v *gocui.View) error {
	_, vy := v.Size()
	if m.mainOY < len(v.BufferLines())-vy {
		m.mainOY++
	}
	return nil
}

func (m *Manager) scrollUp(g *gocui.Gui, v *gocui.View) error {
	if m.mainOY > 0 {
		m.mainOY--
	}
	return nil
}

func (m *Manager) ToggleHelp(g *gocui.Gui, v *gocui.View) error {
	m.showHelp = !m.showHelp
	if !m.showHelp {
		g.SetCurrentView("main")
	}
	return nil
}

func (m *Manager) ToggleNav(g *gocui.Gui, v *gocui.View) error {
	m.showNav = !m.showNav
	return nil
}

func (m *Manager) toggleExitConfirm(g *gocui.Gui, v *gocui.View) error {
	m.showExitConfirm = !m.showExitConfirm
	if !m.showExitConfirm {
		g.SetCurrentView("main")
	}
	return nil
}
