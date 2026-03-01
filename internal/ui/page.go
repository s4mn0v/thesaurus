package ui

import "github.com/awesome-gocui/gocui"

type Page interface {
	Name() string
	Description() string
	Layout(v *gocui.View)
}
