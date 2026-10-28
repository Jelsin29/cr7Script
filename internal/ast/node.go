package ast

// GetNode returns the embedded source range for parser and diagnostic consumers.
func (n *Node) GetNode() Node {
	return *n
}
