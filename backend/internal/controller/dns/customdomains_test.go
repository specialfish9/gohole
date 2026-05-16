package dns

import (
	"net/netip"
	"testing"
)

func TestParseCustomDomains(t *testing.T) {
	testCases := []struct {
		name     string
		input    map[string]any
		expected map[string]netip.Addr
		err      bool
	}{
		{
			name:     "empty",
			input:    map[string]any{},
			expected: map[string]netip.Addr{},
			err:      false,
		},
		{
			name: "v4 and v6",
			input: map[string]any{
				"example4.com": "192.168.1.1",
				"domain4.com":  "127.0.0.1",
				"example6.com": "2345:0425:2CA1:0000:0000:0567:5673:23b5",
				"domain6.com":  "::1",
			},
			expected: map[string]netip.Addr{
				"example4.com": netip.MustParseAddr("192.168.1.1"),
				"domain4.com":  netip.MustParseAddr("127.0.0.1"),
				"example6.com": netip.MustParseAddr("2345:0425:2CA1:0000:0000:0567:5673:23b5"),
				"domain6.com":  netip.MustParseAddr("::1"),
			},
			err: false,
		},
		{
			name: "invalid",
			input: map[string]any{
				"example.com": "this is not a valid IPv4 address",
			},
			expected: nil,
			err:      true,
		},
		{
			name: "invalid",
			input: map[string]any{
				"example.com": 42,
			},
			expected: nil,
			err:      true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parseCustomDomains(tc.input)
			if tc.err && err == nil {
				t.Fatalf("parseCustomDomains(%s) expected error, got nil", tc.name)
			}
			if !tc.err && err != nil {
				t.Fatalf("parseCustomDomains(%s) expected no error, got %v", tc.name, err)
			}

			if len(result) != len(tc.expected) {
				t.Fatalf("expected %d domains, got %d", len(tc.expected), len(result))
			}
			for domain := range tc.expected {
				if _, exists := result[normalizeName(domain)]; !exists {
					t.Errorf("expected domain %q not found in result", domain)
				}
			}
		})
	}

}
