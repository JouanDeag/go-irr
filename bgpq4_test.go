package main

import "testing"

func TestSkipLines(t *testing.T) {
	got := skipLines("first\nsecond\nthird\n", 2)
	want := "third\n"
	if got != want {
		t.Fatalf("skipLines() = %q, want %q", got, want)
	}
}

func TestStripHeadersForEos(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "removes first two header lines",
			in:   "header 1\nheader 2\npermit 192.0.2.0/24\n",
			want: "permit 192.0.2.0/24\n",
		},
		{
			name: "removes deny line after headers",
			in:   "header 1\nheader 2\ndeny any\npermit 192.0.2.0/24\n",
			want: "permit 192.0.2.0/24\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripHeadersForEos(tt.in)
			if got != tt.want {
				t.Fatalf("stripHeadersForEos() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSkipLinesShorterThanRequested(t *testing.T) {
	if got := skipLines("only one line\n", 2); got != "" {
		t.Fatalf("skipLines() = %q, want empty string", got)
	}
	// Used to panic with index out of range, killing the request goroutine.
	if got := stripHeadersForEos("no ip prefix-list NN\n"); got != "" {
		t.Fatalf("stripHeadersForEos() = %q, want empty string", got)
	}
}
