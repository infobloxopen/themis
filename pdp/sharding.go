package pdp

type shard struct {
	name    string
	min     string
	max     string
	servers []string
}

type Shards struct {
	shards []shard
}

func NewShards() Shards {
	return Shards{
		shards: []shard{},
	}
}

func (s Shards) appendShards(shards ...Shards) Shards {
	totalShards := s.shards
	for _, item := range shards {
		totalShards = append(totalShards, item.shards...)
	}

	return Shards{
		shards: totalShards,
	}
}

func (s Shards) AppendShard(name, min, max string, servers ...string) Shards {
	return Shards{
		shards: append(s.shards, shard{
			name:    name,
			min:     min,
			max:     max,
			servers: servers,
		}),
	}
}

func (s Shards) Map() map[string][]string {
	out := make(map[string][]string, len(s.shards))
	for _, shard := range s.shards {
		out[shard.name] = shard.servers
	}

	return out
}

func (s Shards) get(id string) (string, bool) {
	for _, shard := range s.shards {
		if id >= shard.min && id <= shard.max {
			if len(shard.servers) > 0 {
				return shard.name, true
			}

			return shard.name, false
		}
	}

	return "", false
}
