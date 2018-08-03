package pdp

type Shard struct {
	min     string
	max     string
	servers []string
}

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
	shards := make([]shard, len(s.shards)+1)
	copy(shards, s.shards)
	shards[len(s.shards)] = shard{
		name:    name,
		min:     min,
		max:     max,
		servers: servers,
	}

	return Shards{
		shards: shards,
	}
}

func (s Shards) RemoveShard(name string) (Shards, error) {
	if len(s.shards) > 0 {
		if len(s.shards) == 1 {
			if s.shards[0].name == name {
				return NewShards(), nil
			}
		} else {
			for i, item := range s.shards {
				if item.name == name {
					last := len(s.shards) - 1
					shards := make([]shard, last)

					if i > 0 {
						copy(shards, s.shards[:i])
					}

					if i < last {
						copy(shards[i:], s.shards[i+1:])
					}

					return Shards{
						shards: shards,
					}, nil
				}
			}
		}
	}

	return s, newMissingShardError(name)
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
