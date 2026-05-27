package main

import (
	"bytes"
	"os/exec"
	"strings"
)

// ValidSources is the set of IRR databases bgpq4 knows about.
var ValidSources = map[string]struct{}{
	"AFRINIC":     {},
	"ALTDB":       {},
	"APNIC":       {},
	"ARIN":        {},
	"ARIN-NONAUTH": {},
	"BBOI":        {},
	"BELL":        {},
	"CANARIE":     {},
	"IDNIC":       {},
	"INTERNAL":    {},
	"JPIRR":       {},
	"LACNIC":      {},
	"LEVEL3":      {},
	"NTTCOM":      {},
	"RADB":        {},
	"REGISTROBR":  {},
	"RIPE":        {},
	"RIPE-NONAUTH": {},
	"RPKI":        {},
	"TC":          {},
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

var addrFamilyShorthands = map[string]string{
	"v4": "4",
	"v6": "6",
}

func queryBgpq4(vendorName string, addrFamily string, asnOrAsSet string, sources []string) string {
	var args []string

	vendor := vendorShorthands[strings.ToLower(vendorName)]
	addrFamily = addrFamilyShorthands[strings.ToLower(addrFamily)]

	args = append(args, "-S"+strings.Join(sources, ","), "-"+addrFamily, "-A")

	if vendor != "" {
		args = append(args, "-"+vendor)
	}

	if vendor == "J" {
		args = append(args, "-E")
		//bgpq4 needs to make a policy of route-filters instead of a prefix-list
		//because junos prefix-lists do not support le X for aggregation
	}

	maxLen := "24"
	if addrFamily == "6" {
		maxLen = "48"
	}
	args = append(args, "-m "+maxLen)
	if conf.matchParent {
		args = append(args, "-R "+maxLen)
	}

	args = append(args, asnOrAsSet)
	cmd := exec.Command("bgpq4", args...)

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

func stripHeadersForEos(prefixList string) string {
	output := skipLines(prefixList, 2)

	if strings.Contains(output, "deny") {
		output = skipLines(output, 1)
	}

	return output
}

func skipLines(s string, n int) string {
	return strings.SplitN(s, "\n", n+1)[n]
}
