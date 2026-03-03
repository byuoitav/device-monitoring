package messenger

import "testing"

func TestNormalizeHubAddress(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "host and port unchanged",
			in:   "localhost:7100",
			want: "localhost:7100",
		},
		{
			name: "ws url normalized",
			in:   "ws://localhost:7100",
			want: "localhost:7100",
		},
		{
			name: "wss url normalized",
			in:   "wss://hub.example.org:7443",
			want: "hub.example.org:7443",
		},
		{
			name: "trim whitespace and trailing slash",
			in:   "  ws://127.0.0.1:7100/ ",
			want: "127.0.0.1:7100",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeHubAddress(tc.in)
			if got != tc.want {
				t.Fatalf("normalizeHubAddress(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
