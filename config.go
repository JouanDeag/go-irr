package main

import (
	"os"
	"regexp"
	"strings"
	"time"
)

type config struct {
	sources              []string
	matchParent          bool
	listen               string
	cacheTime            time.Duration
	allowCacheBypass     bool
	allowCacheClear      bool
	allowSourceOverride  bool
}

func loadConfig(cfg *config) {
	cfg.sources = parseEnv("SOURCES",
		//default - added anything and everything relevant, whittle it down with an envvar if you want less /shruggi
		[]string{"NTTCOM", "INTERNAL", "LACNIC", "RADB", "RIPE", "RIPE-NONAUTH", "ALTDB", "BELL", "LEVEL3", "APNIC", "JPIRR", "ARIN", "BBOI", "TC", "AFRINIC", "IDNIC", "RPKI", "REGISTROBR", "CANARIE"},
		//parse custom env
		func(s string) []string {
			return strings.Split(strings.ReplaceAll(s, ", ", ","), ",")
		})

	cfg.matchParent = parseBool("MATCH_PARENT", true)

	cfg.listen = fetchEnv("LISTEN", "[::]:8080")

	cfg.cacheTime = parseEnv("CACHE_TIME", time.Hour, func(s string) time.Duration {
		d, _ := time.ParseDuration(s)
		return d
	})

	cfg.allowCacheBypass = parseBool("ALLOW_CACHE_BYPASS", false)

	cfg.allowCacheClear = parseBool("ALLOW_CACHE_CLEAR", false)

	cfg.allowSourceOverride = parseBool("ALLOW_SOURCE_OVERRIDE", false)

}

// boolPattern is anchored: an unanchored alternation matches any value
// *containing* "y", so ALLOW_CACHE_CLEAR=deny used to enable cache clearing.
var boolPattern = regexp.MustCompile(`^(true|1|y|yes)$`)

func parseBool(key string, defaultValue bool) bool {
	return parseEnv(key, defaultValue, func(s string) bool {
		return boolPattern.MatchString(strings.ToLower(strings.TrimSpace(s)))
	})
}

type envParser[T any] func(string) T

func parseEnv[T any](key string, defaultValue T, parse envParser[T]) T {
	value, found := os.LookupEnv(key)
	if found {
		return parse(value)
	} else {
		return defaultValue
	}
}

func fetchEnv(key string, defaultValue string) string {
	return parseEnv(key, defaultValue, func(s string) string { return s })
}
