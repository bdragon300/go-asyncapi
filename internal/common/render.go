package common

// UIHTMLResource represents a CSS or JS resource for UI rendering
type UIHTMLResource struct {
	Location      string
	Content       string
	Embed         bool
	FileExtension string
}
