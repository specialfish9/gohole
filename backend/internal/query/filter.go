package query

type Filter interface {
	// Returns true to allow the query, false to block it
	Filter(q string) (bool, error)
}

type TrieFilter struct {
	root *TrieNode
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
	root := new(TrieNode)

	for _, d := range domains {
		root.Add(d)
	}

	return &TrieFilter{
		root: root,
	}
}
