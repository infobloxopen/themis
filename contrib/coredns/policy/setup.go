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
				policyPlugin.TapIO = tapPlugin.IO
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
	p := &PolicyPlugin{options: make(map[uint16][]edns0Map)}

	for c.Next() {
		if c.Val() == "policy" {
			c.RemainingArgs()
			for c.NextBlock() {
				if err := p.parseOption(c); err != nil {
					return nil, err
				}
			}
			return p, nil
		}
	}
	return nil, fmt.Errorf("Policy setup called without keyword 'policy' in Corefile")
}
