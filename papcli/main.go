package main

import (
	"fmt"
	"os"

	"github.com/infobloxopen/themis/pdpctrl-client"

	log "github.com/Sirupsen/logrus"
)

func main() {
	log.SetLevel(log.InfoLevel)

	f, policy := openFile()
	defer f.Close()

	hosts := []*pdpcc.Client{}

	for _, addr := range config.Addresses {
		h := pdpcc.NewClient(addr, config.ChunkSize)
		if err := h.Connect(config.Timeout); err != nil {
			panic(err)
		}

		hosts = append(hosts, h)
		defer h.Close()
	}

	log.Infof("Requesting data upload to PDP servers...")

	uids := make([]int32, len(hosts))
	errors := 0
	for i, h := range hosts {
		var (
			ID  int32
			err error
		)
		if policy {
			ID, err = h.RequestPoliciesUpload(config.FromTag, config.ToTag)
		} else {
			ID, err = h.RequestContentUpload(config.ContentID, config.FromTag, config.ToTag)
		}

		if err != nil {
			log.Errorf("Failed to upload data: %v", err)
			uids[i] = -1
			errors++
		} else {
			uids[i] = ID
		}
	}

	if errors >= len(hosts) {
		panic(fmt.Errorf("No hosts accepted upload requests"))
	}

	log.Infof("Uploading data to PDP servers...")

	rem := 0
	for _, id := range uids {
		if id == -1 {
			continue
		}

		rem++
	}

	errors = 0
	for i, h := range hosts {
		id := uids[i]
		if id == -1 {
			continue
		}

		f.Seek(0, 0)
		nid, err := h.Upload(id, f)
		if err != nil {
			uids[i] = -1
			errors++
			log.Errorf("Failed to upload data: %v", err)
		} else {
			uids[i] = nid
		}
	}

	if errors >= rem {
		panic(fmt.Errorf("No hosts got data"))
	}

	for i, h := range hosts {
		id := uids[i]
		if id == -1 {
			continue
		}

		if err := h.Apply(id); err != nil {
			log.Errorf("Failed to apply: %v", err)
		}
	}
}

func openFile() (*os.File, bool) {
	pOk := len(config.Policy) > 0
	cOk := len(config.Content) > 0

	if pOk && cOk {
		panic(fmt.Errorf("Both policy and content are specified. Please choose only one"))
	}

	if !pOk && !cOk {
		panic(fmt.Errorf("Neither policy nor content are specified. Please secifiy any"))
	}

	path := config.Content
	if pOk {
		path = config.Policy
	}

	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	return f, pOk
}
