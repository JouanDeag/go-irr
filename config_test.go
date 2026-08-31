package main

import (
	"reflect"
	"testing"
	"time"
)

func TestLoadConfigParsesEnvironment(t *testing.T) {
	t.Setenv("SOURCES", "ripe,arin")
	t.Setenv("MATCH_PARENT", "no")
	t.Setenv("LISTEN", "127.0.0.1:9000")
	t.Setenv("CACHE_TIME", "15m")
	t.Setenv("ALLOW_CACHE_BYPASS", "yes")
	t.Setenv("ALLOW_CACHE_CLEAR", "1")
	t.Setenv("ALLOW_SOURCE_OVERRIDE", "true")

	var cfg config
	loadConfig(&cfg)

	if !reflect.DeepEqual(cfg.sources, []string{"ripe", "arin"}) {
		t.Fatalf("sources = %#v, want %#v", cfg.sources, []string{"ripe", "arin"})
	}
	if cfg.matchParent {
		t.Fatal("matchParent = true, want false")
	}
	if cfg.listen != "127.0.0.1:9000" {
		t.Fatalf("listen = %q, want %q", cfg.listen, "127.0.0.1:9000")
	}
	if cfg.cacheTime != 15*time.Minute {
		t.Fatalf("cacheTime = %s, want 15m", cfg.cacheTime)
	}
	if !cfg.allowCacheBypass {
		t.Fatal("allowCacheBypass = false, want true")
	}
	if !cfg.allowCacheClear {
		t.Fatal("allowCacheClear = false, want true")
	}
	if !cfg.allowSourceOverride {
		t.Fatal("allowSourceOverride = false, want true")
	}
}

func TestParseEnvUsesDefaultWhenUnset(t *testing.T) {
	got := parseEnv("GO_IRR_TEST_UNSET_VALUE", "fallback", func(s string) string {
		return "parsed:" + s
	})
	if got != "fallback" {
		t.Fatalf("parseEnv() = %q, want fallback", got)
	}
}

func TestLoadConfigBooleansRequireExactMatch(t *testing.T) {
	// The old unanchored pattern matched any value containing "y" or "1",
	// so these both turned the flag on.
	t.Setenv("ALLOW_CACHE_CLEAR", "deny")
	t.Setenv("MATCH_PARENT", "onlyweekdays")

	var cfg config
	loadConfig(&cfg)

	if cfg.allowCacheClear {
		t.Fatal(`allowCacheClear = true for "deny", want false`)
	}
	if cfg.matchParent {
		t.Fatal(`matchParent = true for "onlyweekdays", want false`)
	}
}
