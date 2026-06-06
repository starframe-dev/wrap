package warp

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Scrollable wraps a Panel with scroll support.
// When content exceeds the allocated height, the user can scroll via mouse wheel.
type Scrollable struct {
	Content Panel
	Offset  int // scroll offset in lines
}

// NewScrollable creates a new scrollable wrapper.
func NewScrollable(content Panel) *Scrollable {
	return &Scrollable{Content: content}
}

// View renders the visible viewport of the content.
func (s *Scrollable) View(w, h int) string {
	if s.Content == nil {
		return strings.Repeat("\n", h)
	}

	// Render full content at the requested width but unlimited height
	fullContent := s.Content.View(w, 9999)
	lines := strings.Split(fullContent, "\n")

	// Clamp offset
	maxOffset := len(lines) - h
	if maxOffset < 0 {
		maxOffset = 0
	}
	if s.Offset < 0 {
		s.Offset = 0
	}
	if s.Offset > maxOffset {
		s.Offset = maxOffset
	}

	// Take visible slice
	visible := make([]string, h)
	for i := 0; i < h; i++ {
		idx := s.Offset + i
		if idx < len(lines) {
			visible[i] = padLine(lines[idx], w)
		} else {
			visible[i] = strings.Repeat(" ", w)
		}
	}
	return strings.Join(visible, "\n")
}

// Update handles scroll messages (mouse wheel, keys).
func (s *Scrollable) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			s.Offset -= 3
			if s.Offset < 0 {
				s.Offset = 0
			}
		case tea.MouseButtonWheelDown:
			s.Offset += 3
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			s.Offset--
			if s.Offset < 0 {
				s.Offset = 0
			}
		case "down":
			s.Offset++
		case "pgup":
			s.Offset -= 10
			if s.Offset < 0 {
				s.Offset = 0
			}
		case "pgdown":
			s.Offset += 10
		}
	}

	if s.Content != nil {
		return s.Content.Update(msg)
	}
	return nil
}

func padLine(line string, w int) string {
	lw := lipgloss.Width(line)
	if lw >= w {
		// Truncate carefully — find byte boundary
		for i := range line {
			if lipgloss.Width(line[:i]) > w {
				return line[:i-1]
			}
		}
		return line
	}
	return line + strings.Repeat(" ", w-lw)
}
