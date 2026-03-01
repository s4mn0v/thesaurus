package pages

import (
	"fmt"

	"github.com/awesome-gocui/gocui"
	"github.com/s4mn0v/thesaurus/internal/ui/theme"
)

type Automate struct{}

func (p *Automate) Name() string { return "AUTOMATE" }

func (p *Automate) Layout(v *gocui.View) {
	fmt.Fprintf(v, "Hi world this is: AUTOMATE", theme.Bold, theme.Reset)
}
