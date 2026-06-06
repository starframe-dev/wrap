package warp

import (
	"encoding/base64"
	"fmt"
	"strings"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Selectable wraps a Panel with text selection support.
// Mouse drag and Shift+arrows create a selection.
// The selected region is rendered with reversed colors.
type Selectable struct {
	Content Panel

	// Selection anchor (fixed on press) and cursor (active end).
	// Both are in cell coordinates relative to the panel.
	AnchorX, AnchorY int
	CursorX, CursorY int

	HasSelection bool
	Selecting    bool // true during active mouse drag
}

// NewSelectable creates a new selectable wrapper.
func NewSelectable(content Panel) *Selectable {
	return &Selectable{Content: content}
}

// SelectedText returns the currently selected text.
func (s *Selectable) SelectedText() string {
	if !s.HasSelection || s.Content == nil {
		return ""
	}

	// Render content at a large size to get full text, then extract selection.
	// This is a simplified approach — for large content use scrollable.
	content := s.Content.View(9999, 9999)
	lines := strings.Split(content, "\n")

	sx, sy, ex, ey := s.sortedBounds()
	var parts []string
	for y := sy; y <= ey && y < len(lines); y++ {
		line := lines[y]
		lineVis := StripANSI(line)
		startX := 0
		endX := len(lineVis)
		if y == sy {
			startX = sx
		}
		if y == ey {
			endX = ex
		}
		if startX < 0 {
			startX = 0
		}
		if endX > len(lineVis) {
			endX = len(lineVis)
		}
		if startX < endX {
			parts = append(parts, extractVisRange(line, startX, endX))
		}
	}
	return strings.Join(parts, "\n")
}

// ClearSelection removes the current selection.
func (s *Selectable) ClearSelection() {
	s.HasSelection = false
	s.Selecting = false
}

// Copy returns a tea.Cmd that copies the selected text to the system
// clipboard via OSC 52. Call this when the user presses Ctrl+C.
func (s *Selectable) Copy() tea.Cmd {
	text := s.SelectedText()
	if text == "" {
		return nil
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(text))
	seq := fmt.Sprintf("\x1b]52;c;%s\x07", encoded)
	return func() tea.Msg {
		// OSC 52 works even in Bubbletea's alternate screen buffer
		fmt.Print(seq)
		return nil
	}
}

// SelectAll selects all visible content.
func (s *Selectable) SelectAll(w, h int) {
	s.AnchorX, s.AnchorY = 0, 0
	s.CursorX, s.CursorY = w-1, h-1
	s.HasSelection = true
}

// View renders the content with selection highlight.
func (s *Selectable) View(w, h int) string {
	if s.Content == nil {
		return strings.Repeat("\n", h)
	}

	content := s.Content.View(w, h)
	if !s.HasSelection {
		return content
	}

	lines := strings.Split(content, "\n")
	result := make([]string, len(lines))

	sx, sy, ex, ey := s.sortedBounds()
	for y, line := range lines {
		if y < sy || y > ey {
			result[y] = line
			continue
		}
		startX := 0
		endX := lipgloss.Width(StripANSI(line))
		if y == sy {
			startX = sx
		}
		if y == ey {
			endX = ex
		}
		result[y] = highlightRange(line, startX, endX)
	}
	return strings.Join(result, "\n")
}

// Update handles mouse and keyboard for selection.
func (s *Selectable) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonLeft:
			switch msg.Action {
			case tea.MouseActionPress:
				s.AnchorX = msg.X
				s.AnchorY = msg.Y
				s.CursorX = msg.X
				s.CursorY = msg.Y
				s.HasSelection = false
				s.Selecting = true
			case tea.MouseActionMotion:
				if s.Selecting {
					s.CursorX = msg.X
					s.CursorY = msg.Y
					s.HasSelection = true
				}
			case tea.MouseActionRelease:
				if s.Selecting {
					s.CursorX = msg.X
					s.CursorY = msg.Y
					s.Selecting = false
					if s.AnchorX == s.CursorX && s.AnchorY == s.CursorY {
						s.HasSelection = false
					}
				}
			}
		}
	case tea.KeyMsg:
		key := msg.String()
		handled := false

		// Selection keys — shift-modified arrows
		if strings.HasPrefix(key, "shift+") {
			switch key {
			case "shift+up":
				if s.CursorY > 0 {
					s.CursorY--
					s.HasSelection = true
					handled = true
				}
			case "shift+down":
				s.CursorY++
				s.HasSelection = true
				handled = true
			case "shift+left":
				if s.CursorX > 0 {
					s.CursorX--
					s.HasSelection = true
					handled = true
				}
			case "shift+right":
				s.CursorX++
				s.HasSelection = true
				handled = true
			}
		}

		if !handled {
			switch key {
			case "ctrl+a":
				s.SelectAll(9999, 9999)
				handled = true
			case "ctrl+c":
				if s.HasSelection {
					return s.Copy()
				}
			case "esc":
				if s.HasSelection {
					s.ClearSelection()
					handled = true
				}
			}
		} else {
			// Shift+arrow extends selection
			if !s.HasSelection && !s.Selecting {
				s.AnchorX = s.CursorX
				s.AnchorY = s.CursorY
				s.HasSelection = true
			}
			s.Selecting = true
			switch key {
			case "shift+left":
				s.CursorX--
				if s.CursorX < 0 {
					s.CursorX = 0
				}
				handled = true
			case "shift+right":
				s.CursorX++
				handled = true
			case "shift+up":
				s.CursorY--
				if s.CursorY < 0 {
					s.CursorY = 0
				}
				handled = true
			case "shift+down":
				s.CursorY++
				handled = true
			}
			s.Selecting = false
		}

		if handled {
			return nil
		}
	}

	if s.Content != nil {
		return s.Content.Update(msg)
	}
	return nil
}

// sortedBounds returns selection bounds with start <= end.
func (s *Selectable) sortedBounds() (sx, sy, ex, ey int) {
	sx, sy = s.AnchorX, s.AnchorY
	ex, ey = s.CursorX, s.CursorY
	if sy > ey || (sy == ey && sx > ex) {
		sx, ex = ex, sx
		sy, ey = ey, sy
	}
	return
}

// highlightRange applies selection highlight to the visual range [startX, endX).
func highlightRange(line string, startX, endX int) string {
	if startX >= endX {
		return line
	}

	// Fast path: no ANSI in the line
	if !strings.Contains(line, "\x1b") {
		if startX <= 0 && endX >= len(line) {
			return selectionStyleANSI + line + resetStyle
		}
		if startX >= len(line) {
			return line
		}
		before := ""
		if startX > 0 {
			before = line[:startX]
		}
		selected := line[startX:]
		if endX < len(line) {
			selected = line[startX:endX]
		}
		after := ""
		if endX < len(line) {
			after = line[endX:]
		}
		return before + selectionStyleANSI + selected + resetStyle + after
	}

	// ANSI-aware path: walk through the line, tracking visual position.
	var result strings.Builder
	result.Grow(len(line) + 20)

	visPos := 0
	i := 0
	inSelection := false

	for i < len(line) {
		if line[i] == '\x1b' {
			// Copy ANSI sequence
			start := i
			if i+1 < len(line) && line[i+1] == '[' {
				i += 2
				for i < len(line) && line[i] < 0x40 {
					i++
				}
				if i < len(line) {
					i++
				}
			} else {
				i++
			}
			esc := line[start:i]
			// If we're inside selection, close it before the escape,
			// then reopen after
			if inSelection {
				result.WriteString(resetStyle)
				result.WriteString(esc)
				result.WriteString(selectionStyleANSI)
			} else {
				result.WriteString(esc)
			}
			continue
		}

		r, size := utf8.DecodeRuneInString(line[i:])
		wasInSelection := inSelection
		inSelection = visPos >= startX && visPos < endX

		if !wasInSelection && inSelection {
			result.WriteString(selectionStyleANSI)
		}
		if wasInSelection && !inSelection {
			result.WriteString(resetStyle)
		}

		result.WriteRune(r)
		i += size
		visPos++
	}

	if inSelection {
		result.WriteString(resetStyle)
	}

	return result.String()
}

// extractVisRange extracts text from visual range [startX, endX).
func extractVisRange(line string, startX, endX int) string {
	var result strings.Builder
	visPos := 0
	for i := 0; i < len(line); {
		if line[i] == '\x1b' {
			if i+1 < len(line) && line[i+1] == '[' {
				i += 2
				for i < len(line) && line[i] < 0x40 {
					i++
				}
				if i < len(line) {
					i++
				}
			} else {
				i++
			}
			continue
		}
		if visPos >= startX && visPos < endX {
			r, size := utf8.DecodeRuneInString(line[i:])
			result.WriteRune(r)
			i += size
		} else if visPos >= endX {
			break
		} else {
			_, size := utf8.DecodeRuneInString(line[i:])
			i += size
		}
		visPos++
	}
	return result.String()
}

var (
	selectionStyleANSI = "\x1b[7m"
	resetStyle         = "\x1b[0m"
)
