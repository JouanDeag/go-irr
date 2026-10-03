package main

import (
	"reflect"
	"testing"
)

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

func TestBgpq4Args(t *testing.T) {
	oldConf := conf
	t.Cleanup(func() { conf = oldConf })

	tests := []struct {
		name        string
		addrFamily  string
		matchParent bool
		want        []string
	}{
		{"v4", "v4", true, []string{"-SRIPE", "-4", "-A", "-e", "-m 24", "-R 24", "AS123"}},
		{"v6", "v6", true, []string{"-SRIPE", "-6", "-A", "-e", "-m 48", "-R 48", "AS123"}},
		{"v4 without match parent", "v4", false, []string{"-SRIPE", "-4", "-A", "-e", "-m 24", "AS123"}},
		{"v4 blackhole", "v4-bh", true, []string{"-SRIPE", "-4", "-A", "-e", "-m 32", "-R 32", "AS123"}},
		{"v6 blackhole", "V6-BH", true, []string{"-SRIPE", "-6", "-A", "-e", "-m 128", "-R 128", "AS123"}},
		{"blackhole ignores match parent", "v4-bh", false, []string{"-SRIPE", "-4", "-A", "-e", "-m 32", "-R 32", "AS123"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf.matchParent = tt.matchParent
			got := bgpq4Args("eos", tt.addrFamily, "AS123", []string{"RIPE"})
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("bgpq4Args() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsValidAddrFamily(t *testing.T) {
	for _, family := range []string{"v4", "v6", "v4-bh", "V6-BH"} {
		if !isValidAddrFamily(family) {
			t.Errorf("isValidAddrFamily(%q) = false, want true", family)
		}
	}
	for _, family := range []string{"", "v5", "v4-foo", "4"} {
		if isValidAddrFamily(family) {
			t.Errorf("isValidAddrFamily(%q) = true, want false", family)
		}
	}
}
