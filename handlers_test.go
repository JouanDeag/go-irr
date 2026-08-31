package main

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func resetHandlerState(t *testing.T) {
	t.Helper()

	oldConf := conf
	oldCache := cache

	conf = config{
		sources:             []string{"RIPE", "ARIN", "RIPE"},
		matchParent:         true,
		listen:              "127.0.0.1:0",
		cacheTime:           time.Hour,
		allowCacheBypass:    false,
		allowCacheClear:     false,
		allowSourceOverride: false,
	}
	cache.init()

	t.Cleanup(func() {
		conf = oldConf
		cache = oldCache
	})
}

func TestHandleHealthcheck(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	handleHealthcheck(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if rr.Body.String() != "OK" {
		t.Fatalf("body = %q, want OK", rr.Body.String())
	}
}

func TestHandleCacheClear(t *testing.T) {
	resetHandlerState(t)
	cache.set("arista", "v4", "AS123", "ARIN,RIPE", "cached")

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/clearCache", nil)
	handleCacheClear(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status with clear disabled = %d, want %d", rr.Code, http.StatusForbidden)
	}
	if got := cache.get("arista", "v4", "AS123", "ARIN,RIPE"); got != "cached" {
		t.Fatalf("cache after forbidden clear = %q, want cached", got)
	}

	conf.allowCacheClear = true
	rr = httptest.NewRecorder()
	handleCacheClear(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status with clear enabled = %d, want %d", rr.Code, http.StatusOK)
	}
	if rr.Body.String() != "Cache cleared" {
		t.Fatalf("body = %q, want Cache cleared", rr.Body.String())
	}
	if got := cache.get("arista", "v4", "AS123", "ARIN,RIPE"); got != "" {
		t.Fatalf("cache after clear = %q, want empty string", got)
	}
}

func TestHandleServesCachedOutputAndRenamesList(t *testing.T) {
	resetHandlerState(t)
	cache.set("arista", "v4", "AS123", "ARIN,RIPE", "NN permit 192.0.2.0/24\n")

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/arista/v4/as123?name=CUSTOM", nil)
	handle(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if got, want := rr.Header().Get("Content-Type"), "text/plain"; got != want {
		t.Fatalf("content type = %q, want %q", got, want)
	}
	if got, want := rr.Body.String(), "CUSTOM permit 192.0.2.0/24\n"; got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
}

func TestHandleSourceOverrideUsesOrderIndependentCacheKey(t *testing.T) {
	resetHandlerState(t)
	conf.allowSourceOverride = true
	cache.set("json", "v6", "AS208453:AS-SWEHOSTING", "ARIN,RIPE", "NN = [];\n")

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/json/v6/AS208453_AS-SWEHOSTING?sources=ripe,arin&name=PREFIXES", nil)
	handle(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if got, want := rr.Body.String(), "PREFIXES = [];\n"; got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
}

func TestHandleRejectsInvalidRequests(t *testing.T) {
	tests := []struct {
		name string
		path string
		want int
		conf func()
	}{
		{
			name: "wrong path segment count",
			path: "/arista/v4",
			want: http.StatusBadRequest,
		},
		{
			name: "non AS identifier",
			path: "/arista/v4/FOO",
			want: http.StatusBadRequest,
		},
		{
			name: "cache bypass forbidden",
			path: "/arista/v4/AS123?bypassCache=1",
			want: http.StatusForbidden,
		},
		{
			name: "source override forbidden",
			path: "/arista/v4/AS123?sources=RIPE",
			want: http.StatusForbidden,
		},
		{
			name: "unknown source rejected",
			path: "/arista/v4/AS123?sources=RIPE,UNKNOWN",
			want: http.StatusBadRequest,
			conf: func() {
				conf.allowSourceOverride = true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetHandlerState(t)
			if tt.conf != nil {
				tt.conf()
			}

			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			handle(rr, req)

			if rr.Code != tt.want {
				t.Fatalf("status = %d, want %d", rr.Code, tt.want)
			}
		})
	}
}

func TestHandleDoesNotMutateConfiguredSources(t *testing.T) {
	resetHandlerState(t)
	conf.sources = []string{"RIPE", "ARIN", "RIPE", "NTTCOM"}
	cache.set("arista", "v4", "AS123", "ARIN,NTTCOM,RIPE", "NN permit 192.0.2.0/24\n")

	rr := httptest.NewRecorder()
	handle(rr, httptest.NewRequest(http.MethodGet, "/arista/v4/AS123", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	// Dedup used to write into conf.sources' backing array, corrupting shared
	// config and racing between concurrent requests.
	want := []string{"RIPE", "ARIN", "RIPE", "NTTCOM"}
	if !reflect.DeepEqual(conf.sources, want) {
		t.Fatalf("conf.sources = %#v, want %#v", conf.sources, want)
	}
}
