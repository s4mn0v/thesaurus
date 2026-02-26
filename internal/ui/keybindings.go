package ui

type Binding struct {
	Key     string
	Desc    string
	Section string
}

var GlobalBindings = []Binding{
	{"?", "Toggle help", "Global"},
	{"q", "Quit", "Global"},
}

var PageBindings = map[string][]Binding{
	"DASHBOARD": {
		{"c-o", "Copy path", "Local"},
		{"space", "Stage", "Local"},
	},
	"ANALYTICS": {
		{"c-b", "Filter status", "Local"},
		{"r", "Refresh data", "Local"},
	},
	"SETTINGS": {
		{"s", "Save config", "Local"},
		{"i", "Import keys", "Local"},
	},
}

// GetCurrentHelp combines page-specific and global bindings
func GetCurrentHelp(pageName string) []Binding {
	return append(PageBindings[pageName], GlobalBindings...)
}
