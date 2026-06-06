package warp

// Drag-and-drop logic for split borders is implemented in tab.go:
//   - handleMouse (MouseActionPress/Motion/Release on border)
//   - updateDrag (computes new Fraction from mouse position)
//   - hitBorder (detects if mouse is on a border)
//
// Border positions are computed in render.go via findBorders().
