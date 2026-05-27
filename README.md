# go-irr

A simple API for bgpq4, written in Go.

## Usage

```
GET /routeros/addressfamily/ASorAS-SET
GET /routeros/addressfamily/ASorAS-SET?name=myprefixlist
GET /routeros/addressfamily/ASorAS-SET?sources=RIPE,ARIN
```

## Config
### Configuration options

- `SOURCES`  
    What gets passed to bgpq4's `-S` field. Default:
    `NTTCOM,INTERNAL,LACNIC,RADB,RIPE,RIPE-NONAUTH,ALTDB,BELL,LEVEL3,APNIC,JPIRR,ARIN,BBOI,TC,AFRINIC,IDNIC,RPKI,REGISTROBR,CANARIE`

    Can also be overridden per-request via the `sources` query parameter (comma-separated, case-insensitive). The per-request value takes priority over this environment variable.

- `MATCH_PARENT`  
    If bgpq4 should match parent prefixes (not just the exact route object). Enabled by default.

- `LISTEN`  
    Address and port go-irr should bind to. Default: `[::]:8080`

- `CACHE_TIME`  
    How long go-irr should cache prefix results. Default: `1h`. Accepts a sequence of decimal numbers, each with an optional fraction and a unit suffix (examples: `300ms`, `1.5h`, `2h45m`). 
    
    Valid units: `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`.

- `ALLOW_CACHE_BYPASS`  
    Allow the `bypassCache` query parameter to bypass cached results. Disabled by default.

    Accepts boolean-like values (case-insensitive): `true`, `yes`, `y`, `1`

- `ALLOW_CACHE_CLEAR`  
    Allow clearing the global cache via a request to `/clearCache`. Disabled by default.

    Accepts boolean-like values (case-insensitive): `true`, `yes`, `y`, `1`

## Usage examples:

```
GET /arista/v4/AS208453

GET /arista/v6/AS208453:AS-SWEHOSTING

# For systems which do not permit ":" in the URI
GET /eos/v4/AS208453_AS-CUST

# Override IRR sources for this request only
GET /arista/v4/AS208453?sources=RIPE,ARIN
GET /bird/v6/AS208453:AS-SWEHOSTING?sources=RPKI
```

## Supported versions

```
/arista/
/eos/ # Short version without the prefix list headers
/juniper/
/bird/
/routeros6/
/routeros7/
/ios-xr/
/ext-acl/ # generate extended access-list
/cisco/
/json/
```

## Supported address families

```
/brand/v4/
/brand/v6/
```

## Hosted version

[https://irr.as208453.net/](https://irr.as208453.net/)

## Self hosting with Docker

1. Install docker
2. Clone the repo
3. Start using docker compose
4. go-irr is now reachable via `localhost:8080`

```
docker compose up -d
```
