package warp

import (
    "strings"

    "github.com/charmbracelet/lipgloss"
)

// renderNode renders a node tree into lines of the given dimensions.
func renderNode(node *Node, w, h int) []string {
    if node == nil {
        return makeEmptyLines(w, h)
    }
    if node.IsLeaf() {
        content := node.Panel.View(w, h)
        return padContent(content, w, h)
    }

    if node.Split != nil {
        switch node.Split.Direction {
        case Vertical:
            return renderVerticalSplit(node.Split, w, h)
        case Horizontal:
            return renderHorizontalSplit(node.Split, w, h)
        }
    }

    if node.Flex != nil {
        return renderFlex(node.Flex, w, h)
    }

    return makeEmptyLines(w, h)
}

func renderVerticalSplit(split *SplitConfig, w, h int) []string {
    borderW := 1
    availW := w - borderW
    firstW := int(float64(availW) * split.Fraction)
    if firstW < MinPanelSize {
        firstW = MinPanelSize
    }
    secondW := availW - firstW
    if secondW < MinPanelSize {
        secondW = MinPanelSize
        firstW = availW - secondW
    }

    firstLines := renderNode(split.First, firstW, h)
    secondLines := renderNode(split.Second, secondW, h)

    borderChar := borderStyle.Render("│")
    if split.Dragging {
        borderChar = borderDragStyle.Render("│")
    }

    result := make([]string, h)
    for y := 0; y < h; y++ {
        left := ""
        right := ""
        if y < len(firstLines) {
            left = firstLines[y]
        }
        if y < len(secondLines) {
            right = secondLines[y]
        }
        result[y] = left + borderChar + right
    }
    return result
}

func renderHorizontalSplit(split *SplitConfig, w, h int) []string {
    borderH := 1
    availH := h - borderH
    firstH := int(float64(availH) * split.Fraction)
    if firstH < MinPanelSize {
        firstH = MinPanelSize
    }
    secondH := availH - firstH
    if secondH < MinPanelSize {
        secondH = MinPanelSize
        firstH = availH - secondH
    }

    firstLines := renderNode(split.First, w, firstH)
    secondLines := renderNode(split.Second, w, secondH)

    borderLine := borderStyle.Render(strings.Repeat("─", w))
    if split.Dragging {
        borderLine = borderDragStyle.Render(strings.Repeat("─", w))
    }

    result := make([]string, 0, h)
    result = append(result, firstLines...)
    result = append(result, borderLine)
    result = append(result, secondLines...)
    return result
}

// renderFlex renders a flex layout with weighted children.
func renderFlex(flex *FlexConfig, w, h int) []string {
    if len(flex.Items) == 0 {
        return makeEmptyLines(w, h)
    }

    borderSize := 1
    numBorders := len(flex.Items) - 1

    switch flex.Direction {
    case Horizontal: // Row
        availW := w - numBorders*borderSize
        sizes := computeFlexSizes(availW, flex.Items)
        return renderFlexRow(flex, w, h, sizes)
    case Vertical: // Column
        availH := h - numBorders*borderSize
        sizes := computeFlexSizes(availH, flex.Items)
        return renderFlexColumn(flex, w, h, sizes)
    }

    return makeEmptyLines(w, h)
}

func computeFlexSizes(avail int, items []*FlexItem) []int {
    n := len(items)
    if n == 0 {
        return nil
    }
    sizes := make([]int, n)

    // Count collapsed items and non-collapsed grow
    collapsedCount := 0
    totalGrow := 0
    for _, item := range items {
        if item.Collapsed {
            collapsedCount++
        } else {
            totalGrow += item.Grow
        }
    }

    // First pass: allocate basis (collapsed = 1)
    totalBasis := 0
    for i, item := range items {
        basis := 1
        if !item.Collapsed {
            basis = item.Basis
            if basis <= 0 {
                basis = MinPanelSize
            }
        }
        sizes[i] = basis
        totalBasis += basis
    }

    remaining := avail - totalBasis
    if remaining <= 0 {
        return sizes
    }

    // Second pass: distribute remaining space by Grow weights (collapsed get nothing)
    if totalGrow == 0 {
        // Equal distribution among non-collapsed
        nonCollapsed := n - collapsedCount
        if nonCollapsed > 0 {
            perItem := remaining / nonCollapsed
            for i, item := range items {
                if !item.Collapsed {
                    sizes[i] += perItem
                }
            }
        }
        return sizes
    }

    distributed := 0
    for i, item := range items {
        if item.Collapsed {
            continue
        }
        extra := remaining * item.Grow / totalGrow
        sizes[i] += extra
        distributed += extra
    }
    // Distribute leftover pixels to the last non-collapsed item
    leftover := remaining - distributed
    if leftover > 0 {
        for i := n - 1; i >= 0; i-- {
            if !items[i].Collapsed {
                sizes[i] += leftover
                break
            }
        }
    }

    return sizes
}

func renderFlexRow(flex *FlexConfig, w, h int, sizes []int) []string {
    if len(flex.Items) == 0 {
        return makeEmptyLines(w, h)
    }

    borderChar := borderStyle.Render("│")
    if flex.Dragging {
        borderChar = borderDragStyle.Render("│")
    }

    // Render each item
    itemLines := make([][]string, len(flex.Items))
    for i, item := range flex.Items {
        itemLines[i] = renderNode(item.Node, sizes[i], h)
    }

    result := make([]string, h)
    for y := 0; y < h; y++ {
        var buf strings.Builder
        for i := range flex.Items {
            if i > 0 {
                buf.WriteString(borderChar)
            }
            line := ""
            if y < len(itemLines[i]) {
                line = itemLines[i][y]
            }
            buf.WriteString(line)
        }
        result[y] = buf.String()
    }
    return result
}

func renderFlexColumn(flex *FlexConfig, w, h int, sizes []int) []string {
    if len(flex.Items) == 0 {
        return makeEmptyLines(w, h)
    }

    borderLine := borderStyle.Render(strings.Repeat("─", w))
    if flex.Dragging {
        borderLine = borderDragStyle.Render(strings.Repeat("─", w))
    }

    // Render each item
    itemLines := make([][]string, len(flex.Items))
    for i, item := range flex.Items {
        itemLines[i] = renderNode(item.Node, w, sizes[i])
    }

    result := make([]string, 0, h)
    for i := range flex.Items {
        if i > 0 {
            result = append(result, borderLine)
        }
        result = append(result, itemLines[i]...)
    }
    return result
}

// padContent ensures content has exactly w×h dimensions.
// Truncates by visual width (not bytes) to avoid breaking UTF-8 or ANSI sequences.
func padContent(content string, w, h int) []string {
    lines := strings.Split(content, "\n")
    result := make([]string, h)

    for y := 0; y < h; y++ {
        line := ""
        if y < len(lines) {
            line = lines[y]
        }
        lineW := lipgloss.Width(line)
        if lineW < w {
            line += strings.Repeat(" ", w-lineW)
        } else if lineW > w {
            line = truncateVisual(line, w)
        }
        result[y] = line
    }
    return result
}

// truncateVisual truncates a string to at most w visual columns.
// It never breaks UTF-8 runes or ANSI escape sequences.
func truncateVisual(s string, w int) string {
    if w <= 0 {
        return ""
    }
    var buf strings.Builder
    vis := 0
    inEsc := false
    for i := 0; i < len(s); i++ {
        b := s[i]
        if inEsc {
            buf.WriteByte(b)
            if b >= 0x40 && b <= 0x7E {
                inEsc = false
            }
            continue
        }
        if b == '\x1b' {
            buf.WriteByte(b)
            inEsc = true
            continue
        }
        if vis >= w {
            break
        }
        buf.WriteByte(b)
        // UTF-8 lead byte check: if high bit is set, this is a multi-byte rune.
        // We count it as one visual column regardless of byte length.
        if b < 0x80 || (b&0xC0) != 0x80 {
            vis++
        }
    }
    return buf.String()
}

func makeEmptyLines(w, h int) []string {
    lines := make([]string, h)
    empty := strings.Repeat(" ", w)
    for i := range lines {
        lines[i] = empty
    }
    return lines
}

// BorderHit describes a draggable border at a given position.
type BorderHit struct {
    Split     *SplitConfig
    Flex      *FlexConfig
    Direction Direction
    X, Y      int // Start position of the border
    Length    int // Length of the border in cells
}

// findBorders recursively collects all border positions from the node tree.
func findBorders(node *Node, x, y, w, h int) []BorderHit {
    if node == nil || node.IsLeaf() {
        return nil
    }

    var borders []BorderHit

    if node.Split != nil {
        split := node.Split
        switch split.Direction {
        case Vertical:
            borderW := 1
            availW := w - borderW
            firstW := int(float64(availW) * split.Fraction)
            if firstW < MinPanelSize {
                firstW = MinPanelSize
            }
            secondW := availW - firstW
            if secondW < MinPanelSize {
                secondW = MinPanelSize
                firstW = availW - secondW
            }
            borderX := x + firstW
            borders = append(borders, BorderHit{
                Split:     split,
                Direction: Vertical,
                X:         borderX,
                Y:         y,
                Length:    h,
            })
            borders = append(borders, findBorders(split.First, x, y, firstW, h)...)
            borders = append(borders, findBorders(split.Second, borderX+borderW, y, secondW, h)...)

        case Horizontal:
            borderH := 1
            availH := h - borderH
            firstH := int(float64(availH) * split.Fraction)
            if firstH < MinPanelSize {
                firstH = MinPanelSize
            }
            secondH := availH - firstH
            if secondH < MinPanelSize {
                secondH = MinPanelSize
                firstH = availH - secondH
            }
            borderY := y + firstH
            borders = append(borders, BorderHit{
                Split:     split,
                Direction: Horizontal,
                X:         x,
                Y:         borderY,
                Length:    w,
            })
            borders = append(borders, findBorders(split.First, x, y, w, firstH)...)
            borders = append(borders, findBorders(split.Second, x, borderY+borderH, w, secondH)...)
        }
    }

    if node.Flex != nil {
        borders = append(borders, findFlexBorders(node.Flex, x, y, w, h)...)
    }

    return borders
}

func findFlexBorders(flex *FlexConfig, x, y, w, h int) []BorderHit {
    if len(flex.Items) == 0 {
        return nil
    }

    borderSize := 1
    numBorders := len(flex.Items) - 1
    var borders []BorderHit

    switch flex.Direction {
    case Horizontal: // Row
        availW := w - numBorders*borderSize
        sizes := computeFlexSizes(availW, flex.Items)
        cx := x
        for i := 0; i < len(flex.Items)-1; i++ {
            cx += sizes[i]
            borders = append(borders, BorderHit{
                Flex:      flex,
                Direction: Vertical,
                X:         cx,
                Y:         y,
                Length:    h,
            })
            borders = append(borders, findBorders(flex.Items[i].Node, cx-sizes[i], y, sizes[i], h)...)
            cx += borderSize
        }
        // Last item
        lastIdx := len(flex.Items) - 1
        borders = append(borders, findBorders(flex.Items[lastIdx].Node, cx-sizes[lastIdx], y, sizes[lastIdx], h)...)

    case Vertical: // Column
        availH := h - numBorders*borderSize
        sizes := computeFlexSizes(availH, flex.Items)
        cy := y
        for i := 0; i < len(flex.Items)-1; i++ {
            cy += sizes[i]
            borders = append(borders, BorderHit{
                Flex:      flex,
                Direction: Horizontal,
                X:         x,
                Y:         cy,
                Length:    w,
            })
            borders = append(borders, findBorders(flex.Items[i].Node, x, cy-sizes[i], w, sizes[i])...)
            cy += borderSize
        }
        // Last item
        lastIdx := len(flex.Items) - 1
        borders = append(borders, findBorders(flex.Items[lastIdx].Node, x, cy-sizes[lastIdx], w, sizes[lastIdx])...)
    }

    return borders
}
