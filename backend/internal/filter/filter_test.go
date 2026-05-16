package filter_test

import (
	"gohole/internal/filter"
	"testing"
)

var testDomains = []string{
	"example.com",
	"sub.domain.com",
	"test.with.dot.org.",
}

func TestFilters(t *testing.T) {
	t.Run("Basic", func(t *testing.T) {
		testFilter(t, filter.NewBasic(testDomains))
	})
	t.Run("trie", func(t *testing.T) {
		testFilter(t, filter.NewTrie(testDomains))
	})
	t.Run("trie2", func(t *testing.T) {
		testFilter(t, filter.NewTrie2(testDomains))
	})
}

func testFilter(t *testing.T, f filter.Filter) {
	assert(t, f.Size() == len(testDomains), "unexpected filter size")
	for _, domain := range testDomains {
		blocked, err := f.Filter(domain)
		assert(t, err == nil, "unexpected error filtering domain: "+domain)
		assert(t, blocked, "domain should be blocked: "+domain)
	}

	blocked, err := f.Filter("allowed.com")
	assert(t, err == nil, "unexpected error filtering domain: allowed.com")
	assert(t, !blocked, "domain should be allowed: allowed.com")
}

func assert(t *testing.T, condition bool, message string) {
	if !condition {
		t.Fatal(message)
	}
}
