package yast

import "github.com/infobloxopen/themis/pdp"

func (ctx *context) unmarshalSharding(m map[interface{}]interface{}) (pdp.Shards, boundError) {
	out := pdp.NewShards()

	s, ok, err := ctx.extractMapOpt(m, yastTagSharding, yastTagSharding)
	if err != nil {
		return out, err
	}

	if !ok {
		return out, nil
	}

	for k, v := range s {
		name, min, max, servers, err := ctx.unmarshalShard(k, v)
		if err != nil {
			return out, bindError(err, yastTagSharding)
		}

		out = out.AppendShard(name, min, max, servers...)
	}

	return out, nil
}

func (ctx *context) unmarshalShard(k, v interface{}) (string, string, string, []string, boundError) {
	name, err := ctx.validateString(k, "shard name")
	if err != nil {
		return "", "", "", nil, err
	}

	m, err := ctx.validateMap(v, "shard")
	if err != nil {
		return "", "", "", nil, bindError(err, name)
	}

	rng, err := ctx.extractList(m, yastTagRange, yastTagRange)
	if err != nil {
		return "", "", "", nil, bindError(err, name)
	}

	if len(rng) != 2 {
		return "", "", "", nil, bindError(newShardRangeSizeError(rng), name)
	}

	min, err := ctx.validateString(rng[0], "lower boundary")
	if err != nil {
		return "", "", "", nil, bindError(err, name)
	}

	max, err := ctx.validateString(rng[1], "upper boundary")
	if err != nil {
		return "", "", "", nil, bindError(err, name)
	}

	if max < min {
		return "", "", "", nil, bindError(newInvalidShardRangeError(min, max), name)
	}

	items, ok, err := ctx.extractListOpt(m, yastTagServers, yastTagServers)
	if err != nil {
		return "", "", "", nil, bindError(err, name)
	}

	if !ok {
		return name, min, max, nil, nil
	}

	servers := make([]string, len(items))
	for i, v := range items {
		s, err := ctx.validateString(v, "server")
		if err != nil {
			return "", "", "", nil, bindError(bindErrorf(err, "%d", i+1), name)
		}

		servers[i] = s
	}

	return name, min, max, servers, nil
}
