package ui

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/awesome-gocui/gocui"
	"github.com/s4mn0v/thesaurus/internal/trading"
)

type ViewManager struct {
	mu              sync.RWMutex
	currentPage     int
	pageNames       []string
	mainOY          int
	currentTicker   *trading.TickerData
	showHelp        bool
	helpBinding     int
	logger          *UILogger
	showExitConfirm bool
}

func NewViewManager() *ViewManager {
	return &ViewManager{
		pageNames: []string{"DASHBOARD", "ANALYTICS", "SETTINGS"},
	}
}

func (m *ViewManager) SetLogger(l *UILogger) {
	m.logger = l
}

func (m *ViewManager) UpdateTicker(data *trading.TickerData) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentTicker = data
}

func (m *ViewManager) Layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	sideW := 40
	mainW := maxX - sideW - 1
	orange := gocui.Get256Color(172)
	footerY := maxY - 2

	// 1. MAIN CONTENT
	if v, err := g.SetView("main", 0, 0, mainW, footerY, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Title, v.FrameColor, v.Wrap = " "+m.pageNames[m.currentPage]+" ", orange, true
	} else {
		v.Clear()
		m.drawPage(v)
		v.SetOrigin(0, m.mainOY)
		v.Title = " " + m.pageNames[m.currentPage] + " "
	}

	// 2. NAVIGATION
	if v, err := g.SetView("side_top", mainW+1, 0, maxX-1, 10, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Title, v.FrameColor = " Navigation ", orange
	} else {
		v.Clear()
		for i, name := range m.pageNames {
			if i == m.currentPage {
				fmt.Fprintf(v, " > \033[30;47m %s \033[0m\n", name)
			} else {
				fmt.Fprintf(v, "   %s \n", name)
			}
		}
	}

	// 3. LOGS & STATUS
	if v, err := g.SetView("side_bottom", mainW+1, 11, maxX-1, footerY, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Title, v.FrameColor, v.Autoscroll = " Logs & Status ", orange, true
	}

	// 4. FOOTER
	if v, err := g.SetView("footer", -1, footerY, maxX, maxY, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Frame = false
	} else {
		v.Clear()
		fmt.Fprintf(v, "\033[33mKeybindings:\033[0m \033[36m?\033[0m")
	}

	// 5. Confirm Exit
	if err := m.renderHelp(g, maxX, maxY); err != nil {
		return err
	}
	return m.renderExitConfirm(g, maxX, maxY)
	return m.renderHelp(g, maxX, maxY)
}

func (m *ViewManager) drawPage(v *gocui.View) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.currentPage == 0 {
		if m.currentTicker == nil {
			fmt.Fprintln(v, "Connecting to simulated exchange...")
			return
		}
		fmt.Fprintf(v, "\n \033[1mMARKET TICKER (SIMULATED)\033[0m\n")
		fmt.Fprintf(v, " ─────────────\n")
		fmt.Fprintf(v, " ASSET: %s\n", m.currentTicker.Symbol)
		fmt.Fprintf(v, " PRICE: \033[32m$%s\033[0m\n", m.currentTicker.Price)
	} else {
		fmt.Fprintf(v, "Page: %s content", m.pageNames[m.currentPage])
	}
}

func (m *ViewManager) getActiveBindings() []Binding {
	return GetCurrentHelp(m.pageNames[m.currentPage])
}

func (m *ViewManager) renderHelp(g *gocui.Gui, maxX, maxY int) error {
	if !m.showHelp {
		return nil
	}

	activeBindings := m.getActiveBindings()

	// Ensure index isn't out of bounds if page changed while help was open
	if m.helpBinding >= len(activeBindings) {
		m.helpBinding = 0
	}

	w, h := maxX*8/10, maxY*8/10
	x0, y0 := (maxX-w)/2, (maxY-h)/2
	x1, y1 := x0+w, y0+h

	v, err := g.SetView("help", x0, y0, x1, y1-3, 0)
	if err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Title, v.FrameColor = " Keybindings ", gocui.ColorCyan
		g.SetCurrentView("help")
	}

	var buf strings.Builder
	currentSection := ""
	for i, b := range activeBindings {
		if b.Section != currentSection {
			currentSection = b.Section
			buf.WriteString("\n \033[94m--- " + currentSection + " ---\033[0m \n")
		}
		line := fmt.Sprintf("  %-10s %-50s", b.Key, b.Desc)
		if i == m.helpBinding {
			buf.WriteString("\033[30;106;1m" + line + "\033[0m\n")
		} else {
			buf.WriteString(fmt.Sprintf("  \033[36m%-10s\033[0m %s\n", b.Key, b.Desc))
		}
	}

	v.Clear()
	v.Write([]byte(buf.String()))
	v.Subtitle = fmt.Sprintf(" %d of %d ", m.helpBinding+1, len(activeBindings))

	if vd, err := g.SetView("help_desc", x0, y1-2, x1, y1, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		vd.FrameColor, vd.Title = gocui.ColorCyan, " Description "
	} else {
		vd.Clear()
		if m.helpBinding < len(activeBindings) {
			vd.Write([]byte(" " + activeBindings[m.helpBinding].Desc))
		}
	}

	return nil
}

func (m *ViewManager) SetKeybindings(g *gocui.Gui) {
	g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error { return gocui.ErrQuit })
	g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, m.nextPage)
	g.SetKeybinding("main", gocui.KeyArrowDown, gocui.ModNone, m.scrollDown)
	g.SetKeybinding("main", gocui.KeyArrowUp, gocui.ModNone, m.scrollUp)
	g.SetKeybinding("", '?', gocui.ModNone, m.ToggleHelp)

	g.SetKeybinding("help", gocui.KeyArrowDown, gocui.ModNone, m.helpDown)
	g.SetKeybinding("help", gocui.KeyArrowUp, gocui.ModNone, m.helpUp)
	g.SetKeybinding("help", gocui.KeyEnter, gocui.ModNone, m.ToggleHelp)
	g.SetKeybinding("help", gocui.KeyEsc, gocui.ModNone, m.ToggleHelp)
	g.SetKeybinding("help", 'q', gocui.ModNone, m.ToggleHelp)

	// Logs

	g.SetKeybinding("", gocui.KeyCtrlL, gocui.ModNone, m.clearLogs)

	// Confirm Exit

	g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, m.toggleExitConfirm)

	g.SetKeybinding("exit_confirm", 'y', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error { return gocui.ErrQuit })
	g.SetKeybinding("exit_confirm", 'n', gocui.ModNone, m.toggleExitConfirm)
	g.SetKeybinding("exit_confirm", gocui.KeyEsc, gocui.ModNone, m.toggleExitConfirm)
}

func (m *ViewManager) ToggleHelp(g *gocui.Gui, v *gocui.View) error {
	m.showHelp = !m.showHelp
	if !m.showHelp {
		g.DeleteView("help")
		g.DeleteView("help_desc")
		_, err := g.SetCurrentView("main")
		return err
	}
	// m.logger.Log("Opened Help menu")
	return nil
}

func (m *ViewManager) nextPage(g *gocui.Gui, v *gocui.View) error {
	m.currentPage = (m.currentPage + 1) % len(m.pageNames)
	m.mainOY = 0
	m.helpBinding = 0 // Reset help cursor when switching pages
	return nil
}

func (m *ViewManager) scrollDown(g *gocui.Gui, v *gocui.View) error {
	_, vy := v.Size()
	if m.mainOY < len(v.BufferLines())-vy {
		m.mainOY++
	}
	return nil
}

func (m *ViewManager) scrollUp(g *gocui.Gui, v *gocui.View) error {
	if m.mainOY > 0 {
		m.mainOY--
	}
	return nil
}

func (m *ViewManager) helpDown(g *gocui.Gui, v *gocui.View) error {
	if m.helpBinding < len(m.getActiveBindings())-1 {
		m.helpBinding++
	}
	return nil
}

func (m *ViewManager) helpUp(g *gocui.Gui, v *gocui.View) error {
	if m.helpBinding > 0 {
		m.helpBinding--
	}
	return nil
}

func (m *ViewManager) clearLogs(g *gocui.Gui, v *gocui.View) error {
	vLogs, err := g.View("side_bottom")
	if err != nil {
		return nil
	}

	vLogs.Clear()
	return vLogs.SetOrigin(0, 0)
}

func (m *ViewManager) renderExitConfirm(g *gocui.Gui, maxX, maxY int) error {
	if !m.showExitConfirm {
		g.DeleteView("exit_confirm")
		return nil
	}

	w, h := 30, 3
	x0, y0 := (maxX-w)/2, (maxY-h)/2
	x1, y1 := x0+w, y0+h

	if v, err := g.SetView("exit_confirm", x0, y0, x1, y1, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Title = " Quit? "
		v.FrameColor = gocui.ColorRed
		fmt.Fprint(v, " Are you sure? (y/n)")
		g.SetCurrentView("exit_confirm")
	}
	return nil
}

func (m *ViewManager) toggleExitConfirm(g *gocui.Gui, v *gocui.View) error {
	m.showExitConfirm = !m.showExitConfirm
	if !m.showExitConfirm {
		g.DeleteView("exit_confirm")
		_, err := g.SetCurrentView("main")
		return err
	}
	return nil
}
