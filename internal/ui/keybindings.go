package ui

type Binding struct {
	Key     string
	Desc    string
	Section string // "Local", "Global", etc.
}

var HelpMenu = []Binding{
	{"c-o", "Copy path to clipboard", "Local"},
	{"space", "Stage", "Local"},
	{"c-b", "Filter files by status", "Local"},
	{"y", "Copy to clipboard...", "Local"},
	{"?", "Toggle help", "Global"},
	{"q", "Quit", "Global"},
}
