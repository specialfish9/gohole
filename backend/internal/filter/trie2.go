package filter

import "github.com/dghubble/trie"

type Trie2Filter struct {
	size int
	trie *trie.RuneTrie
}

func NewTrie2(domains []string) Filter {
	t := trie.NewRuneTrie()
	for _, d := range domains {
		t.Put(d, struct{}{})
	}

	return &Trie2Filter{
		size: len(domains),
		trie: t,
	}
}

func (f *Trie2Filter) Filter(q string) (bool, error) {
	return f.trie.Get(q) != nil, nil
}

func (f *Trie2Filter) Size() int {
	return f.size
}
