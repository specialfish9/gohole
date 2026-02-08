package filter_test

import (
	"gohole/internal/filter"
	"math/rand"
	"testing"
	"time"
)

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func generateRandomDomains(n int) []string {
	domains := make([]string, n)
	for i := range n {
		domains[i] = generateRandomString(20) + ".com"
	}
	return domains
}

func runBenchmark(b *testing.B, factory func([]string) filter.Filter, domains []string) {
	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		f := factory(domains)
		for _, d := range domains {
			_, err := f.Filter(d)
			if err != nil {
				b.Fatalf("error filtering domain: %v", err)
			}
		}
	}

	b.StopTimer()
}

func BenchmarkFilter(b *testing.B) {
	domains := generateRandomDomains(100000)

	b.Run("BasicFilter", func(b *testing.B) {
		runBenchmark(b, filter.NewBasic, domains)
	})
	b.Run("TrieFilter", func(b *testing.B) {
		runBenchmark(b, filter.NewTrie, domains)
	})
	b.Run("Trie2Filter", func(b *testing.B) {
		runBenchmark(b, filter.NewTrie2, domains)
	})
}
