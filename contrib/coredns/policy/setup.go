package policy

import (
	"errors"
	"fmt"
	"time"

	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/middleware"

	"github.com/mholt/caddy"
)

const (
	defaultTimeoutSecs = 5
)

func init() {
	caddy.RegisterPlugin("policy", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	mw, err := policyParse(c)

	if err != nil {
		return middleware.Error("policy", err)
	}

	c.OnStartup(func() error {
		err := mw.Connect()
		if err != nil {
			return middleware.Error("policy", err)
		}
		return nil
	})

	c.OnShutdown(func() error {
		return nil
	})

	dnsserver.GetConfig(c).AddMiddleware(func(next middleware.Handler) middleware.Handler {
		mw.Next = next
		return mw
	})

	return nil
}

func policyParse(c *caddy.Controller) (*PolicyMiddleware, error) {
	mw := &PolicyMiddleware{Timeout: time.Duration(defaultTimeoutSecs) * time.Second}

	for c.Next() {
		if c.Val() == "policy" {
			c.RemainingArgs()
			//mw.Zones = c.RemainingArgs()
			//if len(mw.Zones) == 0 {
			//	mw.Zones = make([]string, len(c.ServerBlockKeys))
			//	copy(mw.Zones, c.ServerBlockKeys)
			//}
			//middleware.Zones(mw.Zones).Normalize()
			for c.NextBlock() {
				switch c.Val() {
				case "endpoint":
					args := c.RemainingArgs()
					if len(args) > 0 {
						mw.Endpoint = args[0]
						continue
					}
					return nil, c.ArgErr()
				case "timeout":
					args := c.RemainingArgs()
					if len(args) > 0 {
						d, err := time.ParseDuration(args[0])
						if err != nil {
							return nil, fmt.Errorf("Invalid timeout duration'%s': %v. Valid examples are: 5s, 30s, 2m, etc.", args[0], err)
						}
						mw.Timeout = d
						continue
					}
					return nil, c.ArgErr()
				case "edns0":
					args := c.RemainingArgs()
					if len(args) < 2 || (len(args) != 4 && len(args) != 2) {
						return nil, fmt.Errorf("Invalid edns0 directive. Usage: edns0 <code> <name> [ <dataType> <destType> ]. Valid dataTypes are hex (default), bytes, ip. Valid destTypes depend on PDP (default string).")
					}
					dataType := "hex"
					destType := "string"
					if len(args) == 4 {
						dataType = args[2]
						destType = args[3]
					}

					err := mw.AddEDNS0Map(args[0], args[1], dataType, destType)
					if err != nil {
						return nil, fmt.Errorf("Could not add EDNS0 map for %s: %s", args[0], err)
					}
				}
			}
			return mw, nil
		}
	}
	return nil, errors.New("Policy setup called without keyword 'policy' in Corefile")
}
