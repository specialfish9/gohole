package filter

import "log/slog"

type Filter interface {
	// Filter rturns true to allow the query, false to block it
	Filter(q string) (bool, error)
	// Size returns the number of entries in the filter
	Size() int
}

type TrieFilter struct {
	root *TrieNode
	size int
}

func (f *TrieFilter) Filter(q string) (bool, error) {
	if q[len(q)-1] == '.' {
		q = q[:len(q)-1]
	}
	found, err := f.root.Contains(q)
	if err != nil {
		return false, err
	}

	return !found, nil
}

func Trie(domains []string) Filter {
	root := NewTrieNode()

	addedDomains := 0

	for _, d := range domains {
		if err := root.Add(d); err != nil {
			slog.Error("cannot add domain to trie", "domain", d, "error", err)
		} else {
			addedDomains++
		}
	}

	return &TrieFilter{
		root: root,
		size: addedDomains,
	}
}

func (f *TrieFilter) Size() int {
	return f.size
}
