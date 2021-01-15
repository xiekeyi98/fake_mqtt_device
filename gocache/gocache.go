package gocache

import (
	"time"

	"github.com/patrickmn/go-cache"
)

var c *cache.Cache

func init() {
	c = cache.New(2*time.Minute, 10*time.Minute)
}
