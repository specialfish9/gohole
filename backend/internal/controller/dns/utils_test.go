package dns

import "testing"

func TestAddDefaultPort(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output string
		err    bool
	}{
		{
			name:   "IPv4 without port",
			input:  "127.0.0.1",
			output: "127.0.0.1:53",
			err:    false,
		},
		{
			name:   "IPv4 with port",
			input:  "127.0.0.1:54",
			output: "127.0.0.1:54",
			err:    false,
		},
		{
			name:   "IPv6 without port",
			input:  "::1",
			output: "[::1]:53",
			err:    false,
		},
		{
			name:   "IPv6 with port",
			input:  "[::1]:54",
			output: "[::1]:54",
			err:    false,
		},
		{
			name:  "invalid IP",
			input: "ciao!",
			err:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := addDefaultPort(test.input)
			if err != nil && !test.err {
				t.Errorf("expected no error, got %v", err)
			}
			if test.err && err == nil {
				t.Errorf("expected error, got nil")
			}
			if result != test.output {
				t.Errorf("expected %s, got %s", test.output, result)
			}
		})
	}
}
