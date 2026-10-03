package main

import (
	"bytes"
	"os/exec"
	"strings"
)

// ValidSources is the set of IRR databases bgpq4 knows about.
var ValidSources = map[string]struct{}{
	"AFRINIC":      {},
	"ALTDB":        {},
	"APNIC":        {},
	"ARIN":         {},
	"ARIN-NONAUTH": {},
	"BBOI":         {},
	"BELL":         {},
	"CANARIE":      {},
	"IDNIC":        {},
	"INTERNAL":     {},
	"JPIRR":        {},
	"LACNIC":       {},
	"LEVEL3":       {},
	"NTTCOM":       {},
	"RADB":         {},
	"REGISTROBR":   {},
	"RIPE":         {},
	"RIPE-NONAUTH": {},
	"RPKI":         {},
	"TC":           {},
}

var vendorShorthands = map[string]string{
	"arista":    "e",
	"eos":       "e",
	"juniper":   "J",
	"bird":      "b",
	"routeros6": "K",
	"routeros7": "K7",
	"ios-xr":    "X",
	"ext-acl":   "E",
	"cisco":     "",
	"json":      "j",
}

type addrFamilyOptions struct {
	flag      string
	maxLen    string
	blackhole bool
}

// The -bh variants allow more specifics for BH reasons.
var addrFamilies = map[string]addrFamilyOptions{
	"v4":    {flag: "4", maxLen: "24"},
	"v6":    {flag: "6", maxLen: "48"},
	"v4-bh": {flag: "4", maxLen: "32", blackhole: true},
	"v6-bh": {flag: "6", maxLen: "128", blackhole: true},
}

func isValidAddrFamily(addrFamily string) bool {
	_, ok := addrFamilies[strings.ToLower(addrFamily)]
	return ok
}

func queryBgpq4(vendorName string, addrFamily string, asnOrAsSet string, sources []string) string {
	cmd := exec.Command("bgpq4", bgpq4Args(vendorName, addrFamily, asnOrAsSet, sources)...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err != nil {
		return ""
	}

	output := stdout.String()

	if vendorName == "eos" {
		output = stripHeadersForEos(output)
	}

	return output
}

func bgpq4Args(vendorName string, addrFamily string, asnOrAsSet string, sources []string) []string {
	var args []string

	vendor := vendorShorthands[strings.ToLower(vendorName)]
	family := addrFamilies[strings.ToLower(addrFamily)]

	args = append(args, "-S"+strings.Join(sources, ","), "-"+family.flag, "-A")

	if vendor != "" {
		args = append(args, "-"+vendor)
	}

	if vendor == "J" {
		args = append(args, "-E")
		//bgpq4 needs to make a policy of route-filters instead of a prefix-list because junos prefix-lists do not support le X for aggregation
	}

	args = append(args, "-m "+family.maxLen)
	// Blackhole lists exist to accept more-specifics, so they always need -R
	if conf.matchParent || family.blackhole {
		args = append(args, "-R "+family.maxLen)
	}

	return append(args, asnOrAsSet)
}

func stripHeadersForEos(prefixList string) string {
	output := skipLines(prefixList, 2)

	if strings.Contains(output, "deny") {
		output = skipLines(output, 1)
	}

	return output
}

func skipLines(s string, n int) string {
	parts := strings.SplitN(s, "\n", n+1)
	if len(parts) <= n {
		// bgpq4 returned fewer lines than we want to strip; treat as no output
		return ""
	}
	return parts[n]
}
