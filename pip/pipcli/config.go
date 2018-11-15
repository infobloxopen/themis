package main

import (
	"github.com/infobloxopen/themis/pip/pipcli/global"
	"github.com/infobloxopen/themis/pip/pipcli/perf"
	"github.com/infobloxopen/themis/pip/pipcli/subflags"
	"github.com/infobloxopen/themis/pip/pipcli/test"
)

var (
	cmds = []*subflags.Command{
		test.TestCommand,
		perf.PerfCommand,
	}

	conf *global.Config
	cmd  subflags.CommandFunc
)

func parseCommandLine() {
	conf = global.NewConfigFromCommandLine(subflags.MakeUsage(cmds...))
	cmd = subflags.Parse(cmds...)
}
