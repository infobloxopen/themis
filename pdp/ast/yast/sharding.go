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
		name, shard, err := ctx.unmarshalShard(k, v)
		if err != nil {
			return out, bindError(err, yastTagSharding)
		}

		out = out.AppendShard(name, shard)
	}

	return out, nil
}

func (ctx *context) unmarshalShard(k, v interface{}) (string, pdp.Shard, boundError) {
	name, err := ctx.validateString(k, "shard name")
	if err != nil {
		return "", pdp.Shard{}, err
	}

	m, err := ctx.validateMap(v, "shard")
	if err != nil {
		return "", pdp.Shard{}, bindError(err, name)
	}

	rng, ok := m[yastTagRange]
	if !ok {
		return "", pdp.Shard{}, bindError(newMissingListError(yastTagRange), name)
	}

	shard, err := ctx.unmarshalShardEntity(m, rng)
	if err != nil {
		return "", pdp.Shard{}, bindError(err, name)
	}

	return name, shard, nil
}

func (ctx *context) unmarshalShardEntity(m map[interface{}]interface{}, v interface{}) (pdp.Shard, boundError) {
	rng, err := ctx.validateList(v, yastTagRange)
	if err != nil {
		return pdp.Shard{}, err
	}

	if len(rng) != 2 {
		return pdp.Shard{}, newShardRangeSizeError(rng)
	}

	min, err := ctx.validateString(rng[0], "lower boundary")
	if err != nil {
		return pdp.Shard{}, err
	}

	max, err := ctx.validateString(rng[1], "upper boundary")
	if err != nil {
		return pdp.Shard{}, err
	}

	if max < min {
		return pdp.Shard{}, newInvalidShardRangeError(min, max)
	}

	items, ok, err := ctx.extractListOpt(m, yastTagServers, yastTagServers)
	if err != nil {
		return pdp.Shard{}, err
	}

	if !ok {
		return pdp.Shard{
			Min: min,
			Max: max,
		}, nil
	}

	servers := make([]string, len(items))
	for i, v := range items {
		s, err := ctx.validateString(v, "server")
		if err != nil {
			return pdp.Shard{}, bindErrorf(err, "%d", i+1)
		}

		servers[i] = s
	}

	return pdp.Shard{
		Min:     min,
		Max:     max,
		Servers: servers,
	}, nil
}
