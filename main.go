package main

import (
	"log"
	"net/http"
	"regexp"
	"sort"
	"strings"
)

var cache prefixCache
var conf config

func init() {
	cache.init()
	loadConfig(&conf)
	go cache.purgeEvery(conf.cacheTime)
}

func main() {
	http.HandleFunc("/", handle)

	http.HandleFunc("/health", handleHealthcheck)
	http.HandleFunc("/clearCache", handleCacheClear)

	// Wrap the default mux with logging middleware so all requests are logged
	handler := loggingMiddleware(http.DefaultServeMux)
	log.Fatal(http.ListenAndServe(conf.listen, handler))

}

func handleHealthcheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func handleCacheClear(w http.ResponseWriter, r *http.Request) {
	if !conf.allowCacheClear {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	cache.init()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Cache cleared"))
}

func handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	path := strings.Split(r.URL.Path, "/")
	q := r.URL.Query()

	// path has 3 segments:
	// /vendor/addressfamily/AS1234:AS-SET
	if len(path) != 4 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if cache bypass is allowed
	if !conf.allowCacheBypass && q.Get("bypassCache") != "" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	vendor := path[1]
	addrFamily := path[2]
	asnOrAsSet := strings.ReplaceAll(strings.ToUpper(path[3]), "_", ":")

	if vendor == "" || addrFamily == "" || !strings.HasPrefix((asnOrAsSet), "AS") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Make sure asnOrAsSet is correct format
	// AS\d{1,5} or AS-SET

	isASN, _ := regexp.MatchString("^AS\\d{1,6}$", asnOrAsSet)
	isAsSet, _ := regexp.MatchString("^AS[A-Z0-9:-]{1,48}$", asnOrAsSet)

	if !isASN && !isAsSet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Determine sources: per-request override takes priority over config default
	sources := conf.sources
	if sourcesParam := q.Get("sources"); sourcesParam != "" {
		if !conf.allowSourceOverride {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		requested := strings.Split(strings.ReplaceAll(sourcesParam, " ", ""), ",")
		for i, s := range requested {
			requested[i] = strings.ToUpper(s)
			if _, ok := ValidSources[requested[i]]; !ok {
				http.Error(w, "unknown source: "+requested[i], http.StatusBadRequest)
				return
			}
		}
		sources = requested
	}
	// Deduplicate and sort sources so cache key is order-independent
	seen := make(map[string]struct{}, len(sources))
	unique := sources[:0]
	for _, s := range sources {
		if _, dup := seen[s]; !dup {
			seen[s] = struct{}{}
			unique = append(unique, s)
		}
	}
	sources = unique
	sortedSources := make([]string, len(sources))
	copy(sortedSources, sources)
	sort.Strings(sortedSources)
	sourcesKey := strings.Join(sortedSources, ",")

	// Check if the prefix list is already cached
	// If it isn't, look it up using bgpq4 and cache the result
	// Optional cache bypass
	output := cache.get(vendor, addrFamily, asnOrAsSet, sourcesKey)
	bypassRaw := q.Get("bypassCache")
	bypass := bypassRaw == "1" || strings.EqualFold(bypassRaw, "true") || strings.EqualFold(bypassRaw, "yes")

	if output == "" || bypass {
		queried := strings.TrimSpace(queryBgpq4(vendor, addrFamily, asnOrAsSet, sources))

		if queried == "" {
			if vendor == "bird" {
				// bird expects an empty array rather than a failure
				output = "NN = [];\n"
			} else {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		} else {
			output = queried
			// keep a trailing newline for compatibility with existing callers
			if !strings.HasSuffix(output, "\n") {
				output += "\n"
			}
		}

		// don't overwrite the cache when the request explicitly bypassed it
		if !bypass {
			cache.set(vendor, addrFamily, asnOrAsSet, sourcesKey, output)
		}
	}

	if q.Has("name") {
		output = strings.ReplaceAll(output, "NN", q.Get("name"))
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(output))
}
