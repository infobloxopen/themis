package global

import "github.com/infobloxopen/themis/pip/client"

// NewClient creates PIP client according global options.
func (conf *Config) NewClient(f client.ConnErrHandler) client.Client {
	opts := []client.Option{
		client.WithNetwork(conf.Network),
		client.WithAddress(conf.Address),
		client.WithMaxRequestSize(conf.MaxRequestSize),
		client.WithMaxQueue(conf.MaxQueue),
		client.WithBufferSize(conf.BufferSize),
		client.WithConnTimeout(conf.ConnTimeout),
		client.WithWriteInterval(conf.WriteInterval),
		client.WithResponseTimeout(conf.ResponseTimeout),
		client.WithResponseCheckInterval(conf.ResponseCheckInterval),
	}

	if conf.RoundRobinBalancer {
		opts = append(opts, client.WithRoundRobinBalancer(conf.Servers...))
	} else if conf.HotSpotBalancer {
		opts = append(opts, client.WithHotSpotBalancer(conf.Servers...))
	}

	if conf.DNSRadar {
		opts = append(opts, client.WithDNSRadar())
	} else if conf.K8sRadar {
		opts = append(opts, client.WithK8sRadar())
	}

	if f != nil {
		opts = append(opts, client.WithConnErrHandler(f))
	}

	conf.Client = client.NewClient(opts...)
	return conf.Client
}
