package warp

import (
    "strings"
    "unicode/utf8"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

// FloatPane is a floating panel rendered on top of the main layout.
type FloatPane struct {
    Panel Panel
    X, Y  int
    Width int
    Height int
    Title string

    // State
    dragging       bool
    resizing       bool
    resizeEdge     string // "n", "s", "e", "w", "ne", "nw", "se", "sw"
    dragStartX     int
    dragStartY     int
    origX, origY   int
    origW, origH   int

    // CloseRequested is set when the user clicks the × button.
    // The owning Tab checks this after handleMouse and calls CloseFloat.
    CloseRequested bool

    // CloseOnOutsideClick closes the float when the user clicks outside it.
    CloseOnOutsideClick bool
}

const (
    floatMinWidth  = 10
    floatMinHeight = 3
    floatTitleH    = 1
)

// render renders the float pane into lines.
func (fp *FloatPane) render(w, h int) []string {
    lines := make([]string, fp.Height)

    // Top border with title and close button
    title := fp.Title
    closeBtn := floatCloseStyle.Render("×")
    // Reserve space: ╭ + closeBtn(2chars: " ×") + ╮ = 4 chars
    reserveW := 4
    maxTitleW := fp.Width - reserveW
    if maxTitleW < 0 {
        maxTitleW = 0
    }
    if lipgloss.Width(title) > maxTitleW && maxTitleW >= 3 {
        title = title[:maxTitleW-3] + "..."
    }
    dashesW := fp.Width - lipgloss.Width(title) - 4
    if dashesW < 0 {
        dashesW = 0
    }
    topBorder := "╭" + floatTitleStyle.Render(title) +
        floatBorderStyle.Render(strings.Repeat("─", dashesW)) +
        " " + closeBtn + "╮"
    lines[0] = topBorder

    // Content
    contentH := fp.Height - 2
    if contentH < 0 {
        contentH = 0
    }
    contentLines := padContent(fp.Panel.View(fp.Width-2, contentH), fp.Width-2, contentH)
    for i, cl := range contentLines {
        lines[i+1] = floatBorderStyle.Render("│") + floatBgStyle.Render(cl) + floatBorderStyle.Render("│")
    }

    // Bottom border
    bottomBorder := "╰" + floatBorderStyle.Render(strings.Repeat("─", fp.Width-2)) + "╯"
    lines[fp.Height-1] = bottomBorder

    return lines
}

// handleMouse processes mouse events for this float pane.
// mx, my are relative to the content area (not absolute screen).
// Returns tea.Cmd if the event was handled.
func (fp *FloatPane) handleMouse(msg tea.MouseMsg, mx, my int) tea.Cmd {
    // During active drag or resize, skip bounds check — mouse may move outside float
    if !fp.dragging && !fp.resizing {
        if mx < fp.X || mx >= fp.X+fp.Width || my < fp.Y || my >= fp.Y+fp.Height {
            return nil
        }
    }

    relX := mx - fp.X
    relY := my - fp.Y

    switch msg.Button {
    case tea.MouseButtonLeft:
        switch msg.Action {
        case tea.MouseActionPress:
            // Check close button (×) — area: fp.Width-3 to fp.Width-1 on title row
            if relY == 0 && relX >= fp.Width-3 && relX < fp.Width {
                fp.CloseRequested = true
                return nil
            }
            // Title bar drag (exclude corners which are for resize)
            if relY == 0 && relX > 0 && relX < fp.Width-1 {
                fp.dragging = true
                fp.dragStartX = mx
                fp.dragStartY = my
                fp.origX, fp.origY = fp.X, fp.Y
                return nil
            }
            edge := fp.hitEdge(relX, relY)
            if edge != "" {
                fp.resizing = true
                fp.resizeEdge = edge
                fp.dragStartX = mx
                fp.dragStartY = my
                fp.origX, fp.origY = fp.X, fp.Y
                fp.origW, fp.origH = fp.Width, fp.Height
                return nil
            }
            // Click inside — forward to panel
            if fp.Panel != nil && relY > 0 && relY < fp.Height-1 {
                innerMsg := tea.MouseMsg{
                    Action: msg.Action,
                    Button: msg.Button,
                    X:      relX - 1,
                    Y:      relY - 1,
                }
                return fp.Panel.Update(innerMsg)
            }

        case tea.MouseActionMotion:
            if fp.dragging {
                dx := mx - fp.dragStartX
                dy := my - fp.dragStartY
                fp.X = fp.origX + dx
                fp.Y = fp.origY + dy
                if fp.X < 0 {
                    fp.X = 0
                }
                if fp.Y < 0 {
                    fp.Y = 0
                }
            }
            if fp.resizing {
                dx := mx - fp.dragStartX
                dy := my - fp.dragStartY
                fp.applyResize(dx, dy)
            }

        case tea.MouseActionRelease:
            fp.dragging = false
            fp.resizing = false
            fp.resizeEdge = ""
        }
    }

    return nil
}

func (fp *FloatPane) hitEdge(x, y int) string {
    onTop := y == 0
    onBottom := y == fp.Height-1
    onLeft := x == 0
    onRight := x == fp.Width-1

    if onTop && onLeft {
        return "nw"
    }
    if onTop && onRight {
        return "ne"
    }
    if onBottom && onLeft {
        return "sw"
    }
    if onBottom && onRight {
        return "se"
    }
    if onTop {
        return "n"
    }
    if onBottom {
        return "s"
    }
    if onLeft {
        return "w"
    }
    if onRight {
        return "e"
    }
    return ""
}

func (fp *FloatPane) applyResize(dx, dy int) {
    switch fp.resizeEdge {
    case "n":
        fp.Y = fp.origY + dy
        fp.Height = fp.origH - dy
    case "s":
        fp.Height = fp.origH + dy
    case "w":
        fp.X = fp.origX + dx
        fp.Width = fp.origW - dx
    case "e":
        fp.Width = fp.origW + dx
    case "nw":
        fp.X = fp.origX + dx
        fp.Y = fp.origY + dy
        fp.Width = fp.origW - dx
        fp.Height = fp.origH - dy
    case "ne":
        fp.Y = fp.origY + dy
        fp.Width = fp.origW + dx
        fp.Height = fp.origH - dy
    case "sw":
        fp.X = fp.origX + dx
        fp.Width = fp.origW - dx
        fp.Height = fp.origH + dy
    case "se":
        fp.Width = fp.origW + dx
        fp.Height = fp.origH + dy
    }

    if fp.Width < floatMinWidth {
        fp.Width = floatMinWidth
    }
    if fp.Height < floatMinHeight {
        fp.Height = floatMinHeight
    }
    if fp.X < 0 {
        fp.X = 0
    }
    if fp.Y < 0 {
        fp.Y = 0
    }
}

// StripANSI removes ANSI escape sequences from a string.
func StripANSI(s string) string {
    var buf strings.Builder
    buf.Grow(len(s))
    for i := 0; i < len(s); i++ {
        if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
            // CSI sequence: \x1b[ + params (< 0x40) + final (>= 0x40)
            i += 2 // skip \x1b[
            for i < len(s) && s[i] < 0x40 {
                i++
            }
            // i points to final byte, skip it via loop increment
            continue
        }
        buf.WriteByte(s[i])
    }
    return buf.String()
}

// overlayFloat draws the float pane on top of existing content lines.
// Handles ANSI escape sequences correctly by tracking visual positions.
func overlayFloat(lines []string, fp *FloatPane, totalW, totalH int) {
    if fp.X >= totalW || fp.Y >= totalH {
        return
    }

    floatLines := fp.render(totalW, totalH)

    for fy, fl := range floatLines {
        screenY := fp.Y + fy
        if screenY < 0 || screenY >= len(lines) {
            continue
        }

        origLine := lines[screenY]
        floatVisual := StripANSI(fl)
        floatWidth := lipgloss.Width(floatVisual)

        // Clamp float width so it doesn't extend beyond totalW
        if fp.X+floatWidth > totalW {
            excess := fp.X + floatWidth - totalW
            floatWidth -= excess
            // Truncate fl to visual floatWidth — find byte position
            bytePos := 0
            visCount := 0
            for bytePos < len(fl) && visCount < floatWidth {
                if fl[bytePos] == '\x1b' {
                    // Skip ANSI sequence
                    bytePos++
                    if bytePos < len(fl) && fl[bytePos] == '[' {
                        bytePos++
                        for bytePos < len(fl) && fl[bytePos] < 0x40 {
                            bytePos++
                        }
                        if bytePos < len(fl) {
                            bytePos++
                        }
                    }
                    continue
                }
                _, size := utf8.DecodeRuneInString(fl[bytePos:])
                bytePos += size
                visCount++
            }
            fl = fl[:bytePos]
        }

        // Build new line: prefix (up to visual X) + styled float + suffix (after float)
        var buf strings.Builder
        buf.Grow(len(origLine) + len(fl))

        // Copy original prefix up to visual position fp.X
        visPos := 0
        i := 0
        for i < len(origLine) && visPos < fp.X {
            if origLine[i] == '\x1b' {
                // Copy entire CSI escape sequence: \x1b[ + params + final
                start := i
                if i+1 < len(origLine) && origLine[i+1] == '[' {
                    i += 2 // skip \x1b[
                    for i < len(origLine) && origLine[i] < 0x40 {
                        i++
                    }
                    if i < len(origLine) {
                        i++ // include final byte
                    }
                } else {
                    i++ // skip unknown escape
                }
                buf.WriteString(origLine[start:i])
                continue
            }
            // Regular character
            r, size := utf8.DecodeRuneInString(origLine[i:])
            buf.WriteRune(r)
            i += size
            visPos++
        }

        // Write the styled float line (already truncated)
        buf.WriteString(fl)

        // Skip original content covered by float
        visPos = fp.X
        for i < len(origLine) && visPos < fp.X+floatWidth {
            if origLine[i] == '\x1b' {
                if i+1 < len(origLine) && origLine[i+1] == '[' {
                    i += 2 // skip \x1b[
                    for i < len(origLine) && origLine[i] < 0x40 {
                        i++
                    }
                    if i < len(origLine) {
                        i++ // skip final byte
                    }
                } else {
                    i++
                }
                continue
            }
            _, size := utf8.DecodeRuneInString(origLine[i:])
            i += size
            visPos++
        }

        // Copy remaining suffix
        if i < len(origLine) {
            buf.WriteString(origLine[i:])
        }

        lines[screenY] = buf.String()
    }
}
