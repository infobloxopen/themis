package pdpctrl_client

import (
	"fmt"
	"time"
)

type Logger interface {
	Infof(format string, a ...interface{})
	Errorf(format string, a ...interface{})
}

type Cluster struct {
	hosts []*Host
	log   Logger
}

func NewCluster(addresses []string, chunk int, log Logger) Cluster {
	c := Cluster{[]*Host{}, log}
	for _, a := range addresses {
		c.hosts = append(c.hosts, NewHost(a, chunk))
	}

	return c
}

func (c Cluster) Connect(timeout time.Duration) error {
	errors := []error{}
	for _, h := range c.hosts {
		err := h.connect(timeout, c.log)
		if err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) >= len(c.hosts) {
		return fmt.Errorf("No connections")
	}

	c.log.Infof("Established %d connection(s).", len(c.hosts)-len(errors))
	return nil
}

func (c Cluster) Close() {
	for _, h := range c.hosts {
		h.close()
	}
}

func (c Cluster) Process(includes map[string][]byte, policy []byte) {
	for _, h := range c.hosts {
		h.upload(includes, policy, c.log)
	}

	for _, h := range c.hosts {
		h.apply(c.log)
	}
}
