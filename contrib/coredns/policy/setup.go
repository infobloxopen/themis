package policy

import (
	"fmt"

	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/middleware"

	"github.com/mholt/caddy"
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

	c.OnRestart(func() error {
		mw.Close()
		return nil
	})

	c.OnFinalShutdown(func() error {
		mw.Close()
		return nil
	})

	dnsserver.GetConfig(c).AddMiddleware(func(next middleware.Handler) middleware.Handler {
		mw.Next = next
		return mw
	})

	return nil
}

func policyParse(c *caddy.Controller) (*PolicyMiddleware, error) {
	mw := &PolicyMiddleware{Trace: dnsserver.GetConfig(c).Handler("trace")}

	for c.Next() {
		if c.Val() == "policy" {
			c.RemainingArgs()
			for c.NextBlock() {
				switch c.Val() {
				case "endpoint":
					args := c.RemainingArgs()
					if len(args) > 0 {
						mw.Endpoints = args
						continue
					}
					return nil, c.ArgErr()
				case "edns0":
					args := c.RemainingArgs()
					// Usage: edns0 <code> <name> [ <dataType> <destType> ] [ <stringOffset> <stringSize> ].
					// Valid dataTypes are hex (default), bytes, ip.
					// Valid destTypes depend on PDP (default string).
					argsLen := len(args)
					if argsLen == 2 || argsLen == 4 || argsLen == 6 {
						dataType := "hex"
						destType := "string"
						stringOffset := "0"
						stringSize := "0"
						if argsLen > 2 {
							dataType = args[2]
							destType = args[3]
						}
						if argsLen == 6 && destType == "string" {
							stringOffset = args[4]
							stringSize = args[5]
						}
						err := mw.AddEDNS0Map(args[0], args[1], dataType, destType, stringOffset, stringSize)
						if err != nil {
							return nil, fmt.Errorf("Could not add EDNS0 map for %s: %s", args[0], err)
						}
					} else {
						return nil, fmt.Errorf("Invalid edns0 directive")
					}
				}
			}
			return mw, nil
		}
	}
	return nil, fmt.Errorf("Policy setup called without keyword 'policy' in Corefile")
}
