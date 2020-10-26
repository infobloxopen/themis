package client

import (
	"math"

	"github.com/allegro/bigcache/v2"
)

func newCacheFromOptions(opts options) (*bigcache.BigCache, error) {
	if !opts.cache {
		return nil, nil
	}

	cfg := bigcache.DefaultConfig(opts.cacheTTL)
	cfg.MaxEntrySize = int(opts.maxSize)
	cfg.HardMaxCacheSize = opts.cacheMaxSize

	return bigcache.NewBigCache(adjustCacheConfig(cfg))
}

func adjustCacheConfig(cfg bigcache.Config) bigcache.Config {
	size := cfg.HardMaxCacheSize
	if size <= 0 {
		size = 256
	}

	win := size * 1024 * 1024 / cfg.MaxEntrySize
	shrd := 1024 * win / 600000
	shrd = int(math.Pow(2, math.Round(math.Log2(float64(shrd)))))
	if shrd < 256 {
		shrd = 256
	}

	if win/shrd < 64 {
		win = 64 * shrd
	}

	cfg.Shards = shrd
	cfg.MaxEntriesInWindow = win

	return cfg
}
