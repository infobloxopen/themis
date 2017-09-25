package policy

import (
	"fmt"

	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/dnstap"

	"github.com/mholt/caddy"
)

func init() {
	caddy.RegisterPlugin("policy", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	policyPlugin, err := policyParse(c)

	if err != nil {
		return plugin.Error("policy", err)
	}

	c.OnStartup(func() error {
		if taph := dnsserver.GetConfig(c).Handler("dnstap"); taph != nil {
			if tapPlugin, ok := taph.(dnstap.Dnstap); ok {
				policyPlugin.TapWriter = tapPlugin.Out
			}
		}

		policyPlugin.Trace = dnsserver.GetConfig(c).Handler("trace")
		err := policyPlugin.Connect()
		if err != nil {
			return plugin.Error("policy", err)
		}
		return nil
	})

	c.OnRestart(func() error {
		policyPlugin.Close()
		return nil
	})

	c.OnFinalShutdown(func() error {
		policyPlugin.Close()
		return nil
	})

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		policyPlugin.Next = next
		return policyPlugin
	})

	return nil
}

func policyParse(c *caddy.Controller) (*PolicyPlugin, error) {
	policyPlugin := &PolicyPlugin{}

	for c.Next() {
		if c.Val() == "policy" {
			c.RemainingArgs()
			for c.NextBlock() {
				switch c.Val() {
				case "endpoint":
					args := c.RemainingArgs()
					if len(args) > 0 {
						policyPlugin.Endpoints = args
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
						err := policyPlugin.AddEDNS0Map(args[0], args[1], dataType, destType, stringOffset, stringSize)
						if err != nil {
							return nil, fmt.Errorf("Could not add EDNS0 map for %s: %s", args[0], err)
						}
					} else {
						return nil, fmt.Errorf("Invalid edns0 directive")
					}
				case "debug_query_suffix":
					args := c.RemainingArgs()
					if len(args) == 1 {
						policyPlugin.DebugSuffix = args[0]
						continue
					}
					return nil, c.ArgErr()
				}
			}
			return policyPlugin, nil
		}
	}
	return nil, fmt.Errorf("Policy setup called without keyword 'policy' in Corefile")
}
