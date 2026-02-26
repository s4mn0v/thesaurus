package ui

import (
	"fmt"
	"time"

	"github.com/awesome-gocui/gocui"
)

type UILogger struct {
	gui *gocui.Gui
}

func NewUILogger(g *gocui.Gui) *UILogger {
	return &UILogger{gui: g}
}

func (l *UILogger) Log(format string, args ...interface{}) {
	l.gui.Update(func(g *gocui.Gui) error {
		v, err := g.View("side_bottom")
		if err != nil {
			return nil
		}
		timestamp := time.Now().Format("15:04:05")
		msg := fmt.Sprintf(format, args...)
		fmt.Fprintf(v, "\033[32m[%s]\033[0m %s\n", timestamp, msg)
		return nil
	})
}
