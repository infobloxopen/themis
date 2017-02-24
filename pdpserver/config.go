package main

import (
	"flag"
	"os"

	log "github.com/Sirupsen/logrus"
)

type Config struct {
	Verbose   int
	CWD       string
	Policy    string
	ServiceEP string
	ControlEP string
}

var config Config

func init() {
	flag.IntVar(&config.Verbose, "v", 1, "log verbosity (0 - error, 1 - warn (default), 2 - info, 3 - debug)")
	flag.StringVar(&config.CWD, "d", getWd(), "directory of config files")
	flag.StringVar(&config.Policy, "p", "", "policy file to start with")
	flag.StringVar(&config.ServiceEP, "l", "0.0.0.0:5555", "listen for decision requests on this address:port")
	flag.StringVar(&config.ControlEP, "c", "0.0.0.0:5554", "listen for policies on this address:port")

	flag.Parse()
}

func getWd() string {
	dir, err := os.Getwd()
	if err != nil {
		log.WithField("error", err).Warn("Can't get current directory. Using \".\" instead")
		return "."
	}

	return dir
}
