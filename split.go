package warp

// Direction specifies the split orientation.
type Direction int

const (
    // Vertical splits left/right (side by side).
    Vertical Direction = iota
    // Horizontal splits top/bottom (stacked).
    Horizontal
)

// MinPanelSize is the minimum size in characters for any panel.
const MinPanelSize = 3

// SplitConfig is an internal node in the panel tree.
// It splits its area between two children.
type SplitConfig struct {
    Direction Direction
    Fraction  float64 // Share of First child (0.0–1.0)
    First     *Node
    Second    *Node
    Dragging  bool // true during drag-and-drop
}

// FlexItem is a single item inside a FlexConfig.
type FlexItem struct {
    Node      *Node
    Grow      int  // flex-grow weight
    Shrink    int  // flex-shrink (unused for now)
    Basis     int  // flex-basis (min size), 0 = auto
    Collapsed bool // when true, the item takes only 1 line/char
}

// FlexConfig lays out its children in a row or column with weighted sizes.
type FlexConfig struct {
    Direction Direction
    Items     []*FlexItem
    Dragging  bool
}

// Node is a node in the panel tree — either a terminal (Panel) or internal (Split/Flex).
type Node struct {
    Panel Panel
    Split *SplitConfig
    Flex  *FlexConfig
}

// IsLeaf returns true if this node contains a Panel (not a Split or Flex).
func (n *Node) IsLeaf() bool {
    return n.Panel != nil
}

// findNode locates the node containing the given panel in the tree.
// Returns nil if not found.
func (n *Node) findNode(panel Panel) *Node {
    if n == nil {
        return nil
    }
    if n.Panel == panel {
        return n
    }
    if n.Split != nil {
        if found := n.Split.First.findNode(panel); found != nil {
            return found
        }
        if found := n.Split.Second.findNode(panel); found != nil {
            return found
        }
    }
    if n.Flex != nil {
        for _, item := range n.Flex.Items {
            if found := item.Node.findNode(panel); found != nil {
                return found
            }
        }
    }
    return nil
}

// replaceNode replaces oldNode with newNode in the tree.
// Returns true if replacement was successful.
func (n *Node) replaceNode(old, new *Node) bool {
    if n == nil {
        return false
    }
    if n.Split != nil {
        if n.Split.First == old {
            n.Split.First = new
            return true
        }
        if n.Split.Second == old {
            n.Split.Second = new
            return true
        }
        return n.Split.First.replaceNode(old, new) || n.Split.Second.replaceNode(old, new)
    }
    if n.Flex != nil {
        for _, item := range n.Flex.Items {
            if item.Node == old {
                item.Node = new
                return true
            }
            if item.Node.replaceNode(old, new) {
                return true
            }
        }
    }
    return false
}

// collectLeafNodes collects all leaf (Panel) nodes in order.
func (n *Node) collectLeafNodes() []*Node {
    if n == nil {
        return nil
    }
    if n.IsLeaf() {
        return []*Node{n}
    }
    var result []*Node
    if n.Split != nil {
        result = append(result, n.Split.First.collectLeafNodes()...)
        result = append(result, n.Split.Second.collectLeafNodes()...)
    }
    if n.Flex != nil {
        for _, item := range n.Flex.Items {
            result = append(result, item.Node.collectLeafNodes()...)
        }
    }
    return result
}
