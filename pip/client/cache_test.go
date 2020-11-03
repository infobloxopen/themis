package client

import (
	"testing"
	"time"

	"github.com/allegro/bigcache/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewCacheFromOptions(t *testing.T) {
	c, err := newCacheFromOptions(makeOptions([]Option{WithCacheTTL(time.Minute)}))
	assert.NoError(t, err)
	assert.NotZero(t, c)

	c, err = newCacheFromOptions(defaults)
	assert.NoError(t, err)
	assert.Zero(t, c)
}

func TestAdjustCacheConfig(t *testing.T) {
	cfg := bigcache.DefaultConfig(15 * time.Minute)
	cfg = adjustCacheConfig(cfg)
	assert.Equal(t, 1024, cfg.Shards)
	assert.Equal(t, 536870, cfg.MaxEntriesInWindow)

	cfg = bigcache.DefaultConfig(15 * time.Minute)
	cfg.HardMaxCacheSize = 128
	cfg.MaxEntrySize = 10240
	cfg = adjustCacheConfig(cfg)
	assert.Equal(t, 256, cfg.Shards)
	assert.Equal(t, 16384, cfg.MaxEntriesInWindow)
}
