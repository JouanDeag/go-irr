package main

import (
	"sync"
	"time"
)

type prefixCache struct {
	mu   sync.RWMutex
	data map[string]map[string]map[string]map[string]string
}

func (c *prefixCache) init() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]map[string]map[string]map[string]string)
}

func (c *prefixCache) purgeEvery(t time.Duration) {
	ticker := time.NewTicker(t)
	for {
		<-ticker.C
		c.init()
	}
}

func (c *prefixCache) get(vendor string, addrFamily string, asnOrAsSet string, sourcesKey string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.data[vendor][addrFamily][asnOrAsSet][sourcesKey]
}

func (c *prefixCache) set(vendor string, addrFamily string, asnOrAsSet string, sourcesKey string, v string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.data[vendor] == nil {
		c.data[vendor] = make(map[string]map[string]map[string]string)
	}

	if c.data[vendor][addrFamily] == nil {
		c.data[vendor][addrFamily] = make(map[string]map[string]string)
	}

	if c.data[vendor][addrFamily][asnOrAsSet] == nil {
		c.data[vendor][addrFamily][asnOrAsSet] = make(map[string]string)
	}

	c.data[vendor][addrFamily][asnOrAsSet][sourcesKey] = v
}
