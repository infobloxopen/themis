package pep

import (
	"time"

	"github.com/allegro/bigcache"
)

func newCacheFromOptions(opts options) (*bigcache.BigCache, error) {
	if !opts.cache {
		return nil, nil
	}

	return bigcache.NewBigCache(bigcache.DefaultConfig(15 * time.Minute))
}
