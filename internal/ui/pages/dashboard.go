package pages

import (
	"fmt"

	"github.com/awesome-gocui/gocui"
	"github.com/s4mn0v/thesaurus/internal/trading"
	"github.com/s4mn0v/thesaurus/internal/ui/theme"
)

type Dashboard struct {
	Ticker *trading.TickerData // Exported field
}

func (d *Dashboard) Name() string        { return "DASHBOARD" }
func (d *Dashboard) Description() string { return "Real-time market data monitoring." }
func (d *Dashboard) Layout(v *gocui.View) {
	if d.Ticker == nil {
		fmt.Fprintf(v, "%sConnecting to exchange...%s\n", theme.Yellow, theme.Reset)
		return
	}
	fmt.Fprintf(v, "\n %sMARKET TICKER%s\n", theme.Bold, theme.Reset)
	fmt.Fprintf(v, " ASSET: %s\n", d.Ticker.Symbol)
	fmt.Fprintf(v, " PRICE: %s$%s%s\n", theme.Green, d.Ticker.Price, theme.Reset)
}
