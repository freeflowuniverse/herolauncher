package book

// MarkdownFile represents a markdown file in the book
type MarkdownFile struct {
	Name  string
	Path  string
	Title string
	Dir   string
}

// Anchor represents a heading in a markdown file
type Anchor struct {
	ID    string
	Text  string
	Level int
}

// SidebarItem represents an item in a sidebar section
type SidebarItem struct {
	Title    string        `json:"Title"`
	Href     string        `json:"Href"`
	External bool          `json:"External"`
	Children []SidebarItem `json:"Children,omitempty"`
	IsDir    bool          `json:"IsDir,omitempty"`
}

// SidebarSection represents a section in the sidebar
type SidebarSection struct {
	Title string        `json:"Title"`
	Items []SidebarItem `json:"Items"`
}

// Configuration represents the book configuration
type Configuration struct {
	Sidebar []SidebarSection `json:"Sidebar"`
	Title   string           `json:"Title,omitempty"`
	BaseURL string           `json:"BaseURL,omitempty"`
}
