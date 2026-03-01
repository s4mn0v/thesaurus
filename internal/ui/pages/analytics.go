package pages

import (
	"fmt"

	"github.com/awesome-gocui/gocui"
)

type Analytics struct{}

func (a *Analytics) Name() string         { return "ANALYTICS" }
func (a *Analytics) Description() string  { return "Historical performance data." }
func (a *Analytics) Layout(v *gocui.View) { fmt.Fprint(v, "Analytics Content") }
