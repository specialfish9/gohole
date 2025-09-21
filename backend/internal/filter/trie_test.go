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

func TestTrie(t *testing.T) {
	f := filter.Trie(testDomains)

	assert(t, f.Size() == len(testDomains), "unexpected filter size")
	for _, domain := range testDomains {
		allowed, err := f.Filter(domain)
		assert(t, err == nil, "unexpected error filtering domain: "+domain)
		assert(t, !allowed, "domain should be blocked: "+domain)
	}

	allowed, err := f.Filter("allowed.com")
	assert(t, err == nil, "unexpected error filtering domain: allowed.com")
	assert(t, allowed, "domain should be allowed: allowed.com")
}

func assert(t *testing.T, condition bool, message string) {
	if !condition {
		t.Fatal(message)
	}
}
