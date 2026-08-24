package main

import "testing"

func TestPrefixCacheSetAndGet(t *testing.T) {
	var c prefixCache
	c.init()

	if got := c.get("arista", "v4", "AS123", "ARIN,RIPE"); got != "" {
		t.Fatalf("empty cache returned %q, want empty string", got)
	}

	c.set("arista", "v4", "AS123", "ARIN,RIPE", "prefix-list output")
	c.set("arista", "v6", "AS123", "ARIN,RIPE", "ipv6 output")
	c.set("bird", "v4", "AS123", "ARIN,RIPE", "bird output")

	tests := []struct {
		name       string
		vendor     string
		addrFamily string
		asnOrAsSet string
		sourcesKey string
		want       string
	}{
		{
			name:       "exact key returns cached value",
			vendor:     "arista",
			addrFamily: "v4",
			asnOrAsSet: "AS123",
			sourcesKey: "ARIN,RIPE",
			want:       "prefix-list output",
		},
		{
			name:       "address family is isolated",
			vendor:     "arista",
			addrFamily: "v6",
			asnOrAsSet: "AS123",
			sourcesKey: "ARIN,RIPE",
			want:       "ipv6 output",
		},
		{
			name:       "vendor is isolated",
			vendor:     "bird",
			addrFamily: "v4",
			asnOrAsSet: "AS123",
			sourcesKey: "ARIN,RIPE",
			want:       "bird output",
		},
		{
			name:       "missing source key returns empty string",
			vendor:     "arista",
			addrFamily: "v4",
			asnOrAsSet: "AS123",
			sourcesKey: "RIPE",
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.get(tt.vendor, tt.addrFamily, tt.asnOrAsSet, tt.sourcesKey)
			if got != tt.want {
				t.Fatalf("cache get = %q, want %q", got, tt.want)
			}
		})
	}

	c.init()
	if got := c.get("arista", "v4", "AS123", "ARIN,RIPE"); got != "" {
		t.Fatalf("cache get after init = %q, want empty string", got)
	}
}
