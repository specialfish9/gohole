package query

import (
	"fmt"
	"unicode"
)

const (
	charNo = 26 + 2 // letters + "." + "-"
	dotNo  = 26
	dashNo = 27
)

type TrieNode struct {
	children [charNo]*TrieNode
}

func (n *TrieNode) Add(s string) (bool, error) {
	return n.add(s + "\x00") // Append null character to mark the end of the string.
}

func (n *TrieNode) add(s string) (bool, error) {
	if s == "" {
		// Added.
		return true, nil
	}

	index, err := getIndex(rune(s[0]))
	if err != nil {
		return false, fmt.Errorf("trie: add: %w", err)
	}

	if n.children[index] == nil {
		n.children[index] = new(TrieNode)
	}

	return n.children[index].Add(s[1:])
}

func (n *TrieNode) Contains(s string) (bool, error) {
	return n.contains(s + "\x00")
}

func (n *TrieNode) contains(s string) (bool, error) {
	if s == "" {
		return true, nil
	}

	index, err := getIndex(rune(s[0]))
	if err != nil {
		return false, fmt.Errorf("trie: add: %w", err)
	}

	if n.children[index] == nil {
		return false, nil
	}

	return n.children[index].Contains(s[1:])
}

func getIndex(r rune) (int, error) {
	if !unicode.IsLetter(r) {
		switch r {
		case '.':
			return dotNo, nil
		case '-':
			return dashNo, nil
		default:
			return -1, fmt.Errorf("invalid rune '%s'", r)
		}
	}

	upper := unicode.ToUpper(r)
	return int(upper - 'A'), nil
}
