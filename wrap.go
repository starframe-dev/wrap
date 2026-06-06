package warp

import (
	"strings"
	"unicode"

	"github.com/charmbracelet/lipgloss"
)

// WordWrap wraps text at word boundaries so no line exceeds width.
// Words longer than width are broken mid-word.
func WordWrap(text string, width int) []string {
	if width <= 0 {
		return nil
	}
	lines := strings.Split(text, "\n")
	var result []string
	for _, line := range lines {
		result = append(result, wrapLine(line, width)...)
	}
	return result
}

// SpaceWrap wraps text at spaces so no line exceeds width.
// Unlike WordWrap, this does NOT break words — words longer than width overflow.
func SpaceWrap(text string, width int) []string {
	if width <= 0 {
		return nil
	}
	lines := strings.Split(text, "\n")
	var result []string
	for _, line := range lines {
		result = append(result, wrapAtSpaces(line, width)...)
	}
	return result
}

func wrapLine(line string, width int) []string {
	if lipgloss.Width(line) <= width {
		return []string{line}
	}

	var result []string
	start := 0
	for start < len(line) {
		sub := line[start:]
		if lipgloss.Width(sub) <= width {
			result = append(result, sub)
			break
		}
		// Walk forward to find a word boundary
		breakAt := start
		lastWordEnd := start
		for i := start; i < len(line); i++ {
			if unicode.IsSpace(rune(line[i])) {
				lastWordEnd = i
			}
			if lipgloss.Width(line[start:i+1]) > width {
				breakAt = lastWordEnd
				break
			}
		}
		if breakAt <= start {
			breakAt = start + 1
			// Find visual boundary
			for i := start; i < len(line); i++ {
				if lipgloss.Width(line[start:i+1]) > width {
					breakAt = i
					break
				}
			}
		}
		result = append(result, strings.TrimRightFunc(line[start:breakAt], unicode.IsSpace))
		start = breakAt
		// Skip whitespace at start of next line
		for start < len(line) && unicode.IsSpace(rune(line[start])) {
			start++
		}
	}
	return result
}

func wrapAtSpaces(line string, width int) []string {
	if lipgloss.Width(line) <= width {
		return []string{line}
	}

	words := strings.Fields(line)
	if len(words) == 0 {
		return []string{line}
	}

	var result []string
	var current strings.Builder
	for _, word := range words {
		wordW := lipgloss.Width(word)
		curW := lipgloss.Width(current.String())
		if current.Len() == 0 {
			current.WriteString(word)
		} else if curW+1+wordW <= width {
			current.WriteString(" ")
			current.WriteString(word)
		} else {
			result = append(result, current.String())
			current.Reset()
			current.WriteString(word)
		}
	}
	if current.Len() > 0 {
		result = append(result, current.String())
	}
	return result
}

func isWordBreak(b byte) bool {
	return unicode.IsSpace(rune(b))
}

// WrapToString joins wrapped lines with "\n".
func WrapToString(text string, width int, useSpaceWrap bool) string {
	var lines []string
	if useSpaceWrap {
		lines = SpaceWrap(text, width)
	} else {
		lines = WordWrap(text, width)
	}
	return strings.Join(lines, "\n")
}
