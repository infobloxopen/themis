package policy

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/mholt/caddy"
)

var errInvalidOption = errors.New("invalid policy plugin option")

type config struct {
	endpoints    []string
	options      map[uint16][]*edns0Opt
	attrs        *attrsConfig
	debugID      string
	debugSuffix  string
	streams      int
	hotSpot      bool
	passthrough  []string
	connTimeout  time.Duration
	autoReqSize  bool
	maxReqSize   int
	autoResAttrs bool
	maxResAttrs  int
	log          bool
	cacheTTL     time.Duration
	cacheLimit   int
}

func newConfig() config {
	return config{
		options:     make(map[uint16][]*edns0Opt),
		connTimeout: -1,
		maxReqSize:  -1,
		maxResAttrs: 64,
		attrs:       newAttrsConfig(),
	}
}

func policyParse(c *caddy.Controller) (*policyPlugin, error) {
	p := newPolicyPlugin()

	for c.Next() {
		if c.Val() == "policy" {
			c.RemainingArgs()
			for c.NextBlock() {
				if err := p.conf.parseOption(c); err != nil {
					return nil, err
				}
			}
			p.attrPool = createAttrPoolFromConfig(&p.conf)
			return p, nil
		}
	}
	return nil, fmt.Errorf("Policy setup called without keyword 'policy' in Corefile")
}

func (conf *config) parseOption(c *caddy.Controller) error {
	switch c.Val() {
	case "endpoint":
		return conf.parseEndpoint(c)

	case "edns0":
		return conf.parseEDNS0(c.RemainingArgs()...)

	case "debug_query_suffix":
		return conf.parseDebugQuerySuffix(c)

	case "streams":
		return conf.parseStreams(c)

	case "validation1":
		return conf.parseAttributes(c, attrListTypeVal1)

	case "validation2":
		return conf.parseAttributes(c, attrListTypeVal2)

	case "default_decision":
		return conf.parseAttributes(c, attrListTypeDefDecision)

	case "dnstap":
		return conf.parseDnstap(c)

	case "metrics":
		return conf.parseAttributes(c, attrListTypeMetrics)

	case "debug_id":
		return conf.parseDebugID(c)

	case "passthrough":
		return conf.parsePassthrough(c)

	case "connection_timeout":
		return conf.parseConnectionTimeout(c)

	case "log":
		return conf.parseLog(c)

	case "max_request_size":
		return conf.parseMaxRequestSize(c)

	case "max_response_attributes":
		return conf.parseMaxResponseAttributes(c)

	case "cache":
		return conf.parseCache(c)
	}

	return errInvalidOption
}

func (conf *config) parseEndpoint(c *caddy.Controller) error {
	args := c.RemainingArgs()
	if len(args) <= 0 {
		return c.ArgErr()
	}

	conf.endpoints = args
	return nil
}

func (conf *config) parseEDNS0(args ...string) error {
	// Usage: edns0 <code> <name> [ <dataType> ] [ <size> <start> <end> ].
	// Valid dataTypes are hex (default), bytes, ip.
	// Valid destTypes depend on PDP (default string).
	argsLen := len(args)
	if argsLen != 2 && argsLen != 3 && argsLen != 6 {
		return fmt.Errorf("Invalid edns0 directive. Expected 2, 3 or 6 arguments but got %d", argsLen)
	}

	dataType := "hex"
	size := "0"
	start := "0"
	end := "0"

	if argsLen > 2 {
		dataType = args[2]
	}

	if argsLen == 6 && dataType == "hex" {
		size = args[3]
		start = args[4]
		end = args[5]
	}

	code, opt, err := newEdns0Opt(args[0], args[1], dataType, size, start, end)
	if err != nil {
		return fmt.Errorf("Could not add EDNS0 %s (%s): %s", args[1], args[0], err)
	}
	opt.attrInd = conf.attrs.provideIndex(opt.name)

	opts, ok := conf.options[code]
	if !ok {
		opts = []*edns0Opt{}
	}
	conf.options[code] = append(opts, opt)

	return nil
}

func (conf *config) parseAttributes(c *caddy.Controller, listType int) error {
	args := c.RemainingArgs()
	if len(args) <= 0 {
		return c.ArgErr()
	}

	if e := conf.attrs.parseAttrList(listType, args...); e != nil {
		return c.Err(e.Error())
	}
	return nil
}

func (conf *config) parseDnstap(c *caddy.Controller) error {
	if !c.NextArg() {
		return c.ArgErr()
	}
	num, err := strconv.Atoi(c.Val())
	if err != nil {
		return c.Err(err.Error())
	}
	if num <= 0 || num > maxDnstapLists {
		return c.Err("Incorrect dnstap log level")
	}
	return conf.parseAttributes(c, attrListTypeDnstap+num-1)
}

func (conf *config) parseStreams(c *caddy.Controller) error {
	args := c.RemainingArgs()
	if len(args) < 1 || len(args) > 2 {
		return c.ArgErr()
	}

	streams, err := strconv.ParseInt(args[0], 10, 32)
	if err != nil {
		return fmt.Errorf("Could not parse number of streams: %s", err)
	}
	if streams < 1 {
		return fmt.Errorf("Expected at least one stream got %d", streams)
	}

	conf.streams = int(streams)

	if len(args) > 1 {
		switch strings.ToLower(args[1]) {
		default:
			return fmt.Errorf("Expected round-robin or hot-spot balancing but got %s", args[1])

		case "round-robin":
			conf.hotSpot = false

		case "hot-spot":
			conf.hotSpot = true
		}
	} else {
		conf.hotSpot = false
	}

	return nil
}

func (conf *config) parseDebugQuerySuffix(c *caddy.Controller) error {
	args := c.RemainingArgs()
	if len(args) != 1 {
		return c.ArgErr()
	}

	conf.debugSuffix = args[0]
	return nil
}

func (conf *config) parseDebugID(c *caddy.Controller) error {
	args := c.RemainingArgs()
	if len(args) != 1 {
		return c.ArgErr()
	}

	conf.debugID = args[0]
	return nil
}

func (conf *config) parsePassthrough(c *caddy.Controller) error {
	args := c.RemainingArgs()
	if len(args) <= 0 {
		return c.ArgErr()
	}

	conf.passthrough = args
	return nil
}

func (conf *config) parseConnectionTimeout(c *caddy.Controller) error {
	args := c.RemainingArgs()
	if len(args) != 1 {
		return c.ArgErr()
	}

	if strings.ToLower(args[0]) == "no" {
		conf.connTimeout = -1
	} else {
		timeout, err := time.ParseDuration(args[0])
		if err != nil {
			return fmt.Errorf("Could not parse timeout: %s", err)
		}

		conf.connTimeout = timeout
	}

	return nil
}

func (conf *config) parseLog(c *caddy.Controller) error {
	args := c.RemainingArgs()
	if len(args) != 0 {
		return c.ArgErr()
	}

	conf.log = true
	return nil
}

func (conf *config) parseMaxRequestSize(c *caddy.Controller) error {
	args := c.RemainingArgs()
	if len(args) < 1 || len(args) > 2 {
		return c.ArgErr()
	}

	s := ""
	if strings.ToLower(args[0]) == "auto" {
		conf.autoReqSize = true
		if len(args) > 1 {
			s = args[1]
		}
	} else {
		s = args[0]
	}

	if len(s) > 0 {
		size, err := strconv.ParseUint(s, 10, 0)
		if err != nil {
			return fmt.Errorf("Could not parse PDP request size limit: %s", err)
		}

		if size > math.MaxInt32 {
			return fmt.Errorf("Size limit %d (> %d) for PDP request is too high", size, math.MaxInt32)
		}

		conf.maxReqSize = int(size)
	}

	return nil
}

func (conf *config) parseMaxResponseAttributes(c *caddy.Controller) error {
	args := c.RemainingArgs()
	if len(args) != 1 {
		return c.ArgErr()
	}

	if strings.ToLower(args[0]) == "auto" {
		conf.autoResAttrs = true
		return nil
	}

	n, err := strconv.ParseUint(args[0], 10, 0)
	if err != nil {
		return fmt.Errorf("Could not parse PDP response attributes limit: %s", err)
	}

	if n > math.MaxInt32 {
		return fmt.Errorf("Attributes limit %d (> %d) for PDP response is too high", n, math.MaxInt32)
	}

	conf.maxResAttrs = int(n)
	return nil
}

func (conf *config) parseCache(c *caddy.Controller) error {
	args := c.RemainingArgs()
	if len(args) > 2 {
		return c.ArgErr()
	}

	if len(args) > 0 {
		ttl, err := time.ParseDuration(args[0])
		if err != nil {
			return fmt.Errorf("Could not parse decision cache TTL: %s", err)
		}

		if ttl <= 0 {
			return fmt.Errorf("Can't set decision cache TTL to %s", ttl)
		}

		conf.cacheTTL = ttl
	} else {
		conf.cacheTTL = 10 * time.Minute
	}

	if len(args) > 1 {
		n, err := strconv.ParseUint(args[1], 10, 0)
		if err != nil {
			return fmt.Errorf("Could not parse decision cache limit: %s", err)
		}

		if n > math.MaxInt32 {
			return fmt.Errorf("Cache limit %d (> %d) is too high", n, math.MaxInt32)
		}

		conf.cacheLimit = int(n)
	}

	return nil
}
