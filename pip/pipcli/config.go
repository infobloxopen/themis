package main

import (
	"flag"

	"github.com/infobloxopen/themis/pip/pipcli/global"
	"github.com/infobloxopen/themis/pip/pipcli/subflags"
	"github.com/infobloxopen/themis/pip/pipcli/test"
)

var (
	conf *global.Config
	cmd  subflags.CommandFunc
)

func init() {
	flag.Usage = subflags.MakeUsage(test.TestCommand)

	conf = global.NewConfigFromCommandLine()
	cmd = subflags.Parse(test.TestCommand)
}
