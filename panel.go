package warp

import tea "github.com/charmbracelet/bubbletea"

// Panel is the interface users implement to create content for warp panes.
// A Panel can be anything: terminal, text, graphics, form.
type Panel interface {
    // View renders the panel content at the given size.
    View(width, height int) string

    // Update handles Bubbletea messages (keys, mouse, etc).
    // The panel receives only messages that arrived while it was focused.
    Update(msg tea.Msg) tea.Cmd
}

// BasePanel provides a default no-op implementation of Panel.
// Embed it in your panel and override only the methods you need.
type BasePanel struct{}

func (BasePanel) View(width, height int) string {
    return ""
}

func (BasePanel) Update(msg tea.Msg) tea.Cmd {
    return nil
}
