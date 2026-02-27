package ui

type Binding struct {
	Key     string
	Desc    string
	Section string
}

var GlobalBindings = []Binding{
	{"c-l", "Clear Logs", "Global"},
	{"?", "Toggle help", "Global"},
	{"q", "Quit", "Global"},
}

var PageBindings = map[string][]Binding{
	"DASHBOARD": {
		{"c-o", "Copy path", "Page"},
		{"space", "Stage", "Page"},
	},
	"ANALYTICS": {
		{"c-b", "Filter status", "Page"},
		{"r", "Refresh data", "Page"},
	},
	"SETTINGS": {
		{"s", "Save config", "Page"},
		{"i", "Import keys", "Page"},
	},
}

// GetCurrentHelp combines page-specific and global bindings
func GetCurrentHelp(pageName string) []Binding {
	return append(PageBindings[pageName], GlobalBindings...)
}
