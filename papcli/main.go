package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	pdpctrl "github.com/infobloxopen/themis/pdpctrl-client"
)

func main() {
	log.SetLevel(log.InfoLevel)

	policies, includes, err := read(config.Policy, config.Includes)
	if err != nil {
		panic(err)
	}

	hosts := []*pdpctrl.Host{}

	for _, addr := range config.Addresses {
		h := pdpctrl.NewHost(addr, config.Chunk, log.StandardLogger())
		if err := h.Connect(config.Timeout); err != nil {
			panic(err)
		}

		hosts = append(hosts, h)
		defer h.Close()
	}

	bids := make([]int32, len(hosts))
	for i, h := range hosts {
		b := &pdpctrl.DataBucket{
			Policies: policies,
			Includes: includes,
		}

		if err := h.Upload(b); err != nil {
			log.Errorf("Failed to upload PDP data: %v", err)
			bids[i] = -1
		} else {
			bids[i] = b.ID
		}
	}

	for i, h := range hosts {
		if bids[i] != -1 {
			if err := h.Apply(bids[i]); err != nil {
				log.Errorf("Failed to apply PDP data: %v", err)
			}
		}
	}
}

func read(policy string, includes StringSet) ([]byte, map[string][]byte, error) {
	m := make(map[string][]byte)

	for _, name := range includes {
		id, b, err := readInclude(name)
		if err != nil {
			return nil, nil, fmt.Errorf("Error on reading content from \"%s\": %s", name, err)
		}

		m[id] = b
		log.Infof("Loaded content from \"%s\" as \"%s\" (%d byte(s)", name, id, len(b))
	}

	b, err := ioutil.ReadFile(policy)
	if err != nil {
		return nil, nil, fmt.Errorf("Error on reading policy from \"%s\": %s", policy, err)
	}
	log.Infof("Loaded policy from \"%s\" (%d byte(s)", policy, len(b))

	return b, m, nil
}

func getIncludeId(name string) string {
	base := filepath.Base(name)
	return base[0 : len(base)-len(filepath.Ext(base))]
}

func readInclude(name string) (string, []byte, error) {
	b, err := ioutil.ReadFile(name)
	return getIncludeId(name), b, err
}
