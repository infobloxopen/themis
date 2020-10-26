package pep

import (
	"testing"
	"time"

	"github.com/allegro/bigcache/v2"
)

func TestAdjustCacheConfig(t *testing.T) {
	cfg := bigcache.DefaultConfig(15 * time.Minute)
	cfg = adjustCacheConfig(cfg)
	if cfg.Shards != 1024 || cfg.MaxEntriesInWindow != 536870 {
		t.Errorf("Expected %d shards and %d entries in window but got %d and %d",
			1024, 536870, cfg.Shards, cfg.MaxEntriesInWindow)
	}

	cfg = bigcache.DefaultConfig(15 * time.Minute)
	cfg.HardMaxCacheSize = 128
	cfg.MaxEntrySize = 10240
	cfg = adjustCacheConfig(cfg)
	if cfg.Shards != 256 || cfg.MaxEntriesInWindow != 16384 {
		t.Errorf("Expected %d shards and %d entries in window but got %d and %d",
			1024, 536870, cfg.Shards, cfg.MaxEntriesInWindow)
	}
}
