package main

import (
	"bytes"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var cache map[string]map[string]map[string]string

func init() {
	cache = make(map[string]map[string]map[string]string)
	go purgeCache()
}

func purgeCache() {
	ticker := time.NewTicker(time.Hour)
	for {
		select {
		case <-ticker.C:
			// Purge the cache here
			cache = make(map[string]map[string]map[string]string)
		}
	}
}

func getPrefixListFromCache(c *gin.Context) {
	supportedVendors := map[string]string{
		"arista":    "e",
		"eos":       "e",
		"juniper":   "J",
		"bird":      "b",
		"routeros6": "K",
		"routeros7": "K7",
	}

	supportedAddressFamilies := map[string]string{
		"v4": "4",
		"v6": "6",
	}

	path := strings.Split(c.Request.URL.String(), "/")

	c.Header("Content-Type", "text/plain")

	// Minimum path length is 4
	// /vendor/addressfamily/AS1234:AS-SET

	if len(path) != 4 {
		c.String(404, "Not found")
		return
	}

	routerOs := supportedVendors[path[1]]
	addressFamily := supportedAddressFamilies[path[2]]
	asnOrAsSet := strings.Split(path[3], "?")[0]
	nameParam := strings.Split(path[3], "?") // Optional

	prefixListName := "NN"

	if len(nameParam) > 1 {
		prefixListName = strings.Split(nameParam[1], "=")[1]
	}

	if routerOs == "" || addressFamily == "" || !strings.HasPrefix((asnOrAsSet), "AS") {
		c.String(400, "Bad request")
		return
	}

	// Make sure asnOrAsSet is correct format
	// AS\d{1,5} or AS-SET

	isASN, _ := regexp.MatchString("^AS\\d{1,6}$", asnOrAsSet)
	isLegacyAsSet, _ := regexp.MatchString("^AS-[a-zA-Z0-9]{1,32}$", asnOrAsSet)
	isModernAsSet, _ := regexp.MatchString("^AS\\d{1,6}:AS-[a-zA-Z0-9]{1,32}$", asnOrAsSet)
	isAristaAsSet, _ := regexp.MatchString("^AS\\d{1,6}_AS-[a-zA-Z0-9]{1,32}$", asnOrAsSet)

	if !isASN && !isLegacyAsSet && !isModernAsSet && !isAristaAsSet {
		c.String(400, "Bad request")
		return
	}

	// Check if the prefix list is in the cache
	//	- If it is, return it
	//	- If it is not, call getPrefixList and store the result in the cache
	//	- Return the result

	cacheData := cache[routerOs][addressFamily][asnOrAsSet]

	if cacheData != "" {
		if prefixListName != "" {
			cacheData = strings.ReplaceAll(cacheData, "NN", prefixListName)
		}

		if path[1] == "eos" {
			lines := strings.Split(cacheData, "\n")
			cacheData = strings.Join(lines[2:], "\n")
		}

		c.String(200, cacheData)
		return
	}

	output := getPrefixList(addressFamily, routerOs, asnOrAsSet, isAristaAsSet)

	if output == "" {
		c.String(500, "Internal server error")
		return
	}

	if cache[routerOs] == nil {
		cache[routerOs] = make(map[string]map[string]string)
	}

	if cache[routerOs][addressFamily] == nil {
		cache[routerOs][addressFamily] = make(map[string]string)
	}

	cache[routerOs][addressFamily][asnOrAsSet] = output

	if prefixListName != "" {
		output = strings.ReplaceAll(output, "NN", prefixListName)
	}

	if path[1] == "eos" {
		lines := strings.Split(output, "\n")
		output = strings.Join(lines[2:], "\n")
	}

	c.String(200, output)
}

func getPrefixList(addressFamily string, routerOs string, asnOrAsSet string, isAristaAsSet bool) string {

	if isAristaAsSet {
		asnOrAsSet = strings.ReplaceAll(asnOrAsSet, "_", ":")
	}

	aggregate := "-3"

	if routerOs != "J" {
		aggregate = "-A"
	}

	maxLen := "-m 24"
	if addressFamily == "6" {
		maxLen = "-m 48"
	}

	cmd := exec.Command("bgpq4", aggregate, maxLen, "-"+addressFamily, "-"+routerOs, asnOrAsSet)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err != nil {
		return ""
	}

	return stdout.String()
}

func main() {

	router := gin.Default()
	router.NoRoute(getPrefixListFromCache)

	router.Run("[::]:8080")
}
