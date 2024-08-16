# go-irr

A simple API for bgpq4, written in Go.

## Usage

```
GET /routeros/addressfamily/ASorAS-SET
GET /routeros/addressfamily/ASorAS-SET?name=myprefixlist
```

### Examples:

```
GET /arista/v4/AS208453

GET /arista/v6/AS208453:AS-SWEHOSTING

# For systems which do not permit ":" in the URI
GET /eos/v4/AS208453_AS-CUST
```

## Supported versions

```
/arista/
/eos/ # Short version without the prefix list headers
/juniper/
/bird/
/routeros6/
/routeros7/
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
