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

	cfg.matchParent = parseEnv("MATCH_PARENT", true, func(s string) bool {
		matched, _ := regexp.MatchString("true|1|y(es)?", strings.ToLower(s))
		return matched
	})

	cfg.listen = fetchEnv("LISTEN", "[::]:8080")

	cfg.cacheTime = parseEnv("CACHE_TIME", time.Hour, func(s string) time.Duration {
		d, _ := time.ParseDuration(s)
		return d
	})

	cfg.allowCacheBypass = parseEnv("ALLOW_CACHE_BYPASS", false, func(s string) bool {
		matched, _ := regexp.MatchString("true|1|y(es)?", strings.ToLower(s))
		return matched
	})

	cfg.allowCacheClear = parseEnv("ALLOW_CACHE_CLEAR", false, func(s string) bool {
		matched, _ := regexp.MatchString("true|1|y(es)?", strings.ToLower(s))
		return matched
	})

	cfg.allowSourceOverride = parseEnv("ALLOW_SOURCE_OVERRIDE", false, func(s string) bool {
		matched, _ := regexp.MatchString("true|1|y(es)?", strings.ToLower(s))
		return matched
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
