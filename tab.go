package warp

import (
    "strings"

    tea "github.com/charmbracelet/bubbletea"
)

// TabPosition defines where the tab bar is rendered.
type TabPosition int

const (
    TabTop    TabPosition = iota
    TabBottom
    TabLeft
    TabRight
)

// Tab represents a single tab with its panel tree, float panes, and focus state.
type Tab struct {
    name    string
    root    *Node
    focused Panel
    floats  []*FloatPane
    parent  *TabGroup

    // Drag state
    dragging     *SplitConfig
    flexDragging *FlexConfig
    flexDragIdx  int // index of the border being dragged in flex
    lastBorders  []BorderHit
}

func newTab(name string, parent *TabGroup) *Tab {
    return &Tab{
        name:   name,
        root:   &Node{Panel: &emptyPanel{}},
        parent: parent,
    }
}

func (t *Tab) ensureRoot() {
    if t.root == nil {
        t.root = &Node{Panel: &emptyPanel{}}
    }
}

// RootPanel returns the root panel of this tab.
// Useful as the parent argument for the first Split call.
func (t *Tab) RootPanel() Panel {
    t.ensureRoot()
    return t.root.Panel
}

// SplitVertical splits the panel vertically (left/right).
// fraction is the share for the left panel (0.0–1.0).
func (t *Tab) SplitVertical(parent Panel, fraction float64, newPanel Panel) {
    t.ensureRoot()
    node := t.root.findNode(parent)
    if node == nil {
        return
    }
    oldPanel := node.Panel
    node.Panel = nil
    node.Split = &SplitConfig{
        Direction: Vertical,
        Fraction:  clampFraction(fraction),
        First:     &Node{Panel: oldPanel},
        Second:    &Node{Panel: newPanel},
    }
}

// SplitHorizontal splits the panel horizontally (top/bottom).
// fraction is the share for the top panel (0.0–1.0).
func (t *Tab) SplitHorizontal(parent Panel, fraction float64, newPanel Panel) {
    t.ensureRoot()
    node := t.root.findNode(parent)
    if node == nil {
        return
    }
    oldPanel := node.Panel
    node.Panel = nil
    node.Split = &SplitConfig{
        Direction: Horizontal,
        Fraction:  clampFraction(fraction),
        First:     &Node{Panel: oldPanel},
        Second:    &Node{Panel: newPanel},
    }
}

// FlexItemSpec describes a panel and its flex-grow weight.
type FlexItemSpec struct {
    Panel Panel
    Grow  int
}

// FlexRow replaces the parent panel with a horizontal flex layout.
func (t *Tab) FlexRow(parent Panel, items []FlexItemSpec) {
    t.ensureRoot()
    node := t.root.findNode(parent)
    if node == nil {
        return
    }
    if len(items) == 0 {
        return
    }
    node.Panel = nil
    node.Split = nil
    flexItems := make([]*FlexItem, len(items))
    for i, spec := range items {
        grow := spec.Grow
        if grow < 0 {
            grow = 0
        }
        flexItems[i] = &FlexItem{
            Node: &Node{Panel: spec.Panel},
            Grow: grow,
        }
    }
    node.Flex = &FlexConfig{
        Direction: Horizontal,
        Items:     flexItems,
    }
}

// FlexColumn replaces the parent panel with a vertical flex layout.
func (t *Tab) FlexColumn(parent Panel, items []FlexItemSpec) {
    t.ensureRoot()
    node := t.root.findNode(parent)
    if node == nil {
        return
    }
    if len(items) == 0 {
        return
    }
    node.Panel = nil
    node.Split = nil
    flexItems := make([]*FlexItem, len(items))
    for i, spec := range items {
        grow := spec.Grow
        if grow < 0 {
            grow = 0
        }
        flexItems[i] = &FlexItem{
            Node: &Node{Panel: spec.Panel},
            Grow: grow,
        }
    }
    node.Flex = &FlexConfig{
        Direction: Vertical,
        Items:     flexItems,
    }
}

// Float makes a panel floating above the layout.
func (t *Tab) Float(panel Panel, x, y, width, height int) {
    fp := &FloatPane{
        Panel:  panel,
        X:      x,
        Y:      y,
        Width:  width,
        Height: height,
        Title:  "Float",
    }
    t.floats = append(t.floats, fp)
}

// CloseFloat removes a floating pane.
func (t *Tab) CloseFloat(fp *FloatPane) {
    for i, f := range t.floats {
        if f == fp {
            t.floats = append(t.floats[:i], t.floats[i+1:]...)
            return
        }
    }
}

// Focus returns the currently focused panel.
func (t *Tab) Focus() Panel {
    return t.focused
}

// SetFocus sets the focused panel and returns a command from its Update.
func (t *Tab) SetFocus(panel Panel) tea.Cmd {
    t.focused = panel
    return nil
}

// renderContent renders the panel tree at the given dimensions.
func (t *Tab) renderContent(w, h int) string {
    t.lastBorders = findBorders(t.root, 0, 0, w, h)
    lines := renderNode(t.root, w, h)

    // Render floats on top
    for _, fp := range t.floats {
        overlayFloat(lines, fp, w, h)
    }

    return strings.Join(lines, "\n")
}

// handleMouse processes mouse events for this tab.
// offsetX, offsetY account for tab bar position.
// cw, ch are the content area dimensions.
func (t *Tab) handleMouse(msg tea.MouseMsg, offsetX, offsetY, cw, ch int) tea.Cmd {
    mx := msg.X - offsetX
    my := msg.Y - offsetY

    // Check float panes first (top z-order)
    hitFloat := false
    for i := len(t.floats) - 1; i >= 0; i-- {
        fp := t.floats[i]
        cmd := fp.handleMouse(msg, mx, my)
        if fp.CloseRequested {
            t.CloseFloat(fp)
            return nil
        }
        if cmd != nil {
            hitFloat = true
            // Bring to top
            if msg.Action == tea.MouseActionPress {
                t.floats = append(t.floats[:i], t.floats[i+1:]...)
                t.floats = append(t.floats, fp)
                t.focused = fp.Panel
            }
            return cmd
        }
        // Check if click is inside this float (for outside-click detection)
        if mx >= fp.X && mx < fp.X+fp.Width && my >= fp.Y && my < fp.Y+fp.Height {
            hitFloat = true
        }
    }

    // Close floats that want auto-close on outside click
    if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft && !hitFloat {
        for i := len(t.floats) - 1; i >= 0; i-- {
            if t.floats[i].CloseOnOutsideClick {
                t.CloseFloat(t.floats[i])
            }
        }
    }

    // Check border dragging
    switch msg.Button {
    case tea.MouseButtonLeft:
        switch msg.Action {
        case tea.MouseActionPress:
            for i, bh := range t.lastBorders {
                if t.hitBorder(bh, mx, my) {
                    if bh.Split != nil {
                        t.dragging = bh.Split
                        t.dragging.Dragging = true
                    } else if bh.Flex != nil {
                        t.flexDragging = bh.Flex
                        t.flexDragIdx = i
                        t.flexDragging.Dragging = true
                    }
                    return nil
                }
            }
            // Click on a panel — focus it, toggle collapsible
            if hit := t.panelAt(mx, my, cw, ch); hit != nil {
                t.focused = hit.Node.Panel
                // Toggle collapsible on title bar click
                if c, ok := hit.Node.Panel.(*Collapsible); ok {
                    if my == hit.Y { // Click on the first line (title bar)
                        c.Toggle()
                        t.updateFlexCollapsed(hit.Node)
                        return nil
                    }
                }
            }
        case tea.MouseActionMotion:
            if t.dragging != nil || t.flexDragging != nil {
                t.updateDrag(mx, my, cw, ch)
            }
        case tea.MouseActionRelease:
            if t.dragging != nil {
                t.dragging.Dragging = false
                t.dragging = nil
            }
            if t.flexDragging != nil {
                t.flexDragging.Dragging = false
                t.flexDragging = nil
                t.flexDragIdx = -1
            }
        }
    }

    // Forward mouse to panel under cursor (relative coordinates)
    if msg.Action == tea.MouseActionPress || msg.Action == tea.MouseActionMotion {
        if hit := t.panelAt(mx, my, cw, ch); hit != nil && hit.Node.Panel != nil {
            relMsg := tea.MouseMsg{
                X:      mx - hit.X,
                Y:      my - hit.Y,
                Action: msg.Action,
                Button: msg.Button,
                Type:   msg.Type,
                Alt:    msg.Alt,
            }
            return hit.Node.Panel.Update(relMsg)
        }
    }
    return nil
}

// handleKeys forwards key messages to the focused panel.
func (t *Tab) handleKeys(msg tea.KeyMsg) tea.Cmd {
    if t.focused != nil {
        return t.focused.Update(msg)
    }
    return nil
}

// broadcastMsg sends a message to all panels in this tab (tree + floats).
func (t *Tab) broadcastMsg(msg tea.Msg) []tea.Cmd {
    var cmds []tea.Cmd
    cmds = append(cmds, t.broadcastNode(t.root, msg)...)
    for _, fp := range t.floats {
        if fp.Panel != nil {
            if cmd := fp.Panel.Update(msg); cmd != nil {
                cmds = append(cmds, cmd)
            }
        }
    }
    return cmds
}

func (t *Tab) broadcastNode(node *Node, msg tea.Msg) []tea.Cmd {
    if node == nil {
        return nil
    }
    if node.IsLeaf() && node.Panel != nil {
        if cmd := node.Panel.Update(msg); cmd != nil {
            return []tea.Cmd{cmd}
        }
        return nil
    }
    var cmds []tea.Cmd
    if node.Split != nil {
        cmds = append(cmds, t.broadcastNode(node.Split.First, msg)...)
        cmds = append(cmds, t.broadcastNode(node.Split.Second, msg)...)
    }
    if node.Flex != nil {
        for _, item := range node.Flex.Items {
            cmds = append(cmds, t.broadcastNode(item.Node, msg)...)
        }
    }
    return cmds
}

func (t *Tab) hitBorder(bh BorderHit, mx, my int) bool {
    switch bh.Direction {
    case Vertical:
        return mx == bh.X && my >= bh.Y && my < bh.Y+bh.Length
    case Horizontal:
        return my == bh.Y && mx >= bh.X && mx < bh.X+bh.Length
    }
    return false
}

// panelHit describes a panel and its bounds.
type panelHit struct {
    Node *Node
    X, Y int
    W, H int
}

func (t *Tab) panelAt(mx, my, cw, ch int) *panelHit {
    return t.panelAtNode(t.root, 0, 0, cw, ch, mx, my)
}

func (t *Tab) panelAtNode(node *Node, x, y, w, h int, mx, my int) *panelHit {
    if node == nil {
        return nil
    }
    if node.IsLeaf() {
        if mx >= x && mx < x+w && my >= y && my < y+h {
            return &panelHit{Node: node, X: x, Y: y, W: w, H: h}
        }
        return nil
    }

    if node.Split != nil {
        switch node.Split.Direction {
        case Vertical:
            borderW := 1
            availW := w - borderW
            firstW := int(float64(availW) * node.Split.Fraction)
            if firstW < MinPanelSize {
                firstW = MinPanelSize
            }
            if mx < x+firstW {
                return t.panelAtNode(node.Split.First, x, y, firstW, h, mx, my)
            }
            if mx >= x+firstW+borderW {
                return t.panelAtNode(node.Split.Second, x+firstW+borderW, y, w-firstW-borderW, h, mx, my)
            }
            return nil // Border area
        case Horizontal:
            borderH := 1
            availH := h - borderH
            firstH := int(float64(availH) * node.Split.Fraction)
            if firstH < MinPanelSize {
                firstH = MinPanelSize
            }
            if my < y+firstH {
                return t.panelAtNode(node.Split.First, x, y, w, firstH, mx, my)
            }
            if my >= y+firstH+borderH {
                return t.panelAtNode(node.Split.Second, x, y+firstH+borderH, w, h-firstH-borderH, mx, my)
            }
            return nil
        }
    }

    if node.Flex != nil {
        return t.panelAtFlex(node.Flex, x, y, w, h, mx, my)
    }

    return nil
}

func (t *Tab) panelAtFlex(flex *FlexConfig, x, y, w, h int, mx, my int) *panelHit {
    if len(flex.Items) == 0 {
        return nil
    }

    borderSize := 1
    numBorders := len(flex.Items) - 1

    switch flex.Direction {
    case Horizontal:
        availW := w - numBorders*borderSize
        sizes := computeFlexSizes(availW, flex.Items)
        cx := x
        for i, item := range flex.Items {
            if i > 0 {
                cx += borderSize
            }
            if mx >= cx && mx < cx+sizes[i] {
                return t.panelAtNode(item.Node, cx, y, sizes[i], h, mx, my)
            }
            cx += sizes[i]
        }
    case Vertical:
        availH := h - numBorders*borderSize
        sizes := computeFlexSizes(availH, flex.Items)
        cy := y
        for i, item := range flex.Items {
            if i > 0 {
                cy += borderSize
            }
            if my >= cy && my < cy+sizes[i] {
                return t.panelAtNode(item.Node, x, cy, w, sizes[i], mx, my)
            }
            cy += sizes[i]
        }
    }
    return nil
}

// ToggleCollapsible toggles the collapsed state of a collapsible panel.
// It also updates the flex item if the panel is inside a flex layout.
func (t *Tab) ToggleCollapsible(panel Panel) {
    t.toggleCollapsibleNode(t.root, panel)
}

func (t *Tab) toggleCollapsibleNode(node *Node, panel Panel) bool {
    if node == nil {
        return false
    }
    if node.IsLeaf() && node.Panel == panel {
        if c, ok := panel.(*Collapsible); ok {
            c.Toggle()
        }
        return true
    }
    if node.Split != nil {
        if t.toggleCollapsibleNode(node.Split.First, panel) {
            return true
        }
        return t.toggleCollapsibleNode(node.Split.Second, panel)
    }
    if node.Flex != nil {
        for _, item := range node.Flex.Items {
            if item.Node.Panel == panel {
                if c, ok := panel.(*Collapsible); ok {
                    c.Toggle()
                    item.Collapsed = c.Collapsed
                }
                return true
            }
            if t.toggleCollapsibleNode(item.Node, panel) {
                return true
            }
        }
    }
    return false
}

func (t *Tab) updateDrag(mx, my, cw, ch int) {
    if t.dragging != nil {
        t.updateSplitDrag(mx, my, cw, ch)
        return
    }
    if t.flexDragging != nil {
        t.updateFlexDrag(mx, my, cw, ch)
    }
}

func (t *Tab) updateSplitDrag(mx, my, cw, ch int) {
    for _, bh := range t.lastBorders {
        if bh.Split == t.dragging {
            switch bh.Direction {
            case Vertical:
                if cw > 0 {
                    frac := float64(mx) / float64(cw)
                    t.dragging.Fraction = clampFraction(frac)
                }
            case Horizontal:
                if ch > 0 {
                    frac := float64(my) / float64(ch)
                    t.dragging.Fraction = clampFraction(frac)
                }
            }
            break
        }
    }
}

func (t *Tab) updateFlexDrag(mx, my, cw, ch int) {
    if t.flexDragIdx < 0 || t.flexDragIdx >= len(t.flexDragging.Items)-1 {
        return
    }

    switch t.flexDragging.Direction {
    case Horizontal:
        if cw <= 0 {
            return
        }
        // Total available width for flex items
        borderSize := 1
        numBorders := len(t.flexDragging.Items) - 1
        availW := cw - numBorders*borderSize
        if availW <= 0 {
            return
        }

        // Mouse position relative to flex start
        // Find cumulative size up to dragIdx
        sizes := computeFlexSizes(availW, t.flexDragging.Items)
        cum := 0
        for i := 0; i <= t.flexDragIdx; i++ {
            cum += sizes[i]
        }
        cum += t.flexDragIdx * borderSize

        // New relative position
        rel := mx - cum + sizes[t.flexDragIdx]
        if rel < MinPanelSize {
            rel = MinPanelSize
        }
        if rel > availW-MinPanelSize {
            rel = availW - MinPanelSize
        }

        // Adjust grow weights proportionally
        leftGrow := t.flexDragging.Items[t.flexDragIdx].Grow
        rightGrow := t.flexDragging.Items[t.flexDragIdx+1].Grow
        if leftGrow+rightGrow == 0 {
            leftGrow = 1
            rightGrow = 1
        }
        ratio := float64(rel) / float64(availW)
        totalGrow := leftGrow + rightGrow
        t.flexDragging.Items[t.flexDragIdx].Grow = int(ratio * float64(totalGrow))
        if t.flexDragging.Items[t.flexDragIdx].Grow < 1 {
            t.flexDragging.Items[t.flexDragIdx].Grow = 1
        }
        t.flexDragging.Items[t.flexDragIdx+1].Grow = totalGrow - t.flexDragging.Items[t.flexDragIdx].Grow

    case Vertical:
        if ch <= 0 {
            return
        }
        borderSize := 1
        numBorders := len(t.flexDragging.Items) - 1
        availH := ch - numBorders*borderSize
        if availH <= 0 {
            return
        }

        sizes := computeFlexSizes(availH, t.flexDragging.Items)
        cum := 0
        for i := 0; i <= t.flexDragIdx; i++ {
            cum += sizes[i]
        }
        cum += t.flexDragIdx * borderSize

        rel := my - cum + sizes[t.flexDragIdx]
        if rel < MinPanelSize {
            rel = MinPanelSize
        }
        if rel > availH-MinPanelSize {
            rel = availH - MinPanelSize
        }

        leftGrow := t.flexDragging.Items[t.flexDragIdx].Grow
        rightGrow := t.flexDragging.Items[t.flexDragIdx+1].Grow
        if leftGrow+rightGrow == 0 {
            leftGrow = 1
            rightGrow = 1
        }
        ratio := float64(rel) / float64(availH)
        totalGrow := leftGrow + rightGrow
        t.flexDragging.Items[t.flexDragIdx].Grow = int(ratio * float64(totalGrow))
        if t.flexDragging.Items[t.flexDragIdx].Grow < 1 {
            t.flexDragging.Items[t.flexDragIdx].Grow = 1
        }
        t.flexDragging.Items[t.flexDragIdx+1].Grow = totalGrow - t.flexDragging.Items[t.flexDragIdx].Grow
    }
}

// updateFlexCollapsed updates the FlexItem.Collapsed flag for a collapsible panel.
func (t *Tab) updateFlexCollapsed(target *Node) {
    t.updateFlexCollapsedNode(t.root, target)
}

func (t *Tab) updateFlexCollapsedNode(node *Node, target *Node) bool {
    if node == nil {
        return false
    }
    if node.Flex != nil {
        for _, item := range node.Flex.Items {
            if item.Node == target {
                if c, ok := target.Panel.(*Collapsible); ok {
                    item.Collapsed = c.Collapsed
                }
                return true
            }
            if t.updateFlexCollapsedNode(item.Node, target) {
                return true
            }
        }
    }
    if node.Split != nil {
        if t.updateFlexCollapsedNode(node.Split.First, target) {
            return true
        }
        return t.updateFlexCollapsedNode(node.Split.Second, target)
    }
    return false
}

func clampFraction(f float64) float64 {
    if f < 0.1 {
        return 0.1
    }
    if f > 0.9 {
        return 0.9
    }
    return f
}

// emptyPanel is used as a placeholder when a tab has no user panels yet.
type emptyPanel struct{}

func (emptyPanel) View(width, height int) string {
    return ""
}

func (emptyPanel) Update(msg tea.Msg) tea.Cmd {
    return nil
}
