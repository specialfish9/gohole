package filter

type TrieNode struct {
	children map[rune]*TrieNode
}

func NewTrieNode() *TrieNode {
	return &TrieNode{
		children: make(map[rune]*TrieNode),
	}
}

func (n *TrieNode) Add(s string) error {
	return n.add(s + "\x00") // Append null character to mark the end of the string.
}

func (n *TrieNode) add(s string) error {
	if s == "" {
		// Finish adding the string
		return nil
	}

	_, ok := n.children[rune(s[0])]
	if !ok {
		// Create a new child node if it doesn't exist
		child := NewTrieNode()
		n.children[rune(s[0])] = child
	}

	child := n.children[rune(s[0])]

	return child.add(s[1:])
}

func (n *TrieNode) Contains(s string) (bool, error) {
	return n.contains(s + "\x00")
}

func (n *TrieNode) contains(s string) (bool, error) {
	if s == "" {
		return true, nil
	}

	child, ok := n.children[rune(s[0])]
	if !ok {
		return false, nil
	}

	return child.Contains(s[1:])
}
