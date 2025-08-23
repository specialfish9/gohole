package query

import "log"

type Filter interface {
	// Returns true to allow the query, false to block it
	Filter(q string) (bool, error)
}

type TrieFilter struct {
	root *TrieNode
}

func (f *TrieFilter) Filter(q string) (bool, error) {
	found, err := f.root.Contains(q)
	if err != nil {
		return false, err
	}

	return !found, nil
}

func Trie(domains []string) Filter {
	root := new(TrieNode)

	for _, d := range domains {
		log.Printf("INFO adding %s\n", d)
		root.Add(d)
	}

	return &TrieFilter{
		root: root,
	}
}
