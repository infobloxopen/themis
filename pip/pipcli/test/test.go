package test

import (
	"flag"

	"github.com/infobloxopen/themis/pip/pipcli/global"
	"github.com/infobloxopen/themis/pip/pipcli/subflags"
)

const (
	testName = "test"
	testDesc = "sends information requests to PIP"
)

var (
	TestCommand *subflags.Command
	testFlagSet *flag.FlagSet
)

func init() {
	TestCommand = &subflags.Command{
		Name:   testName,
		Desc:   testDesc,
		Parser: parser,
	}
	testFlagSet = subflags.MakeCommandFlagSet(TestCommand)
}

func parser(args []string) (subflags.CommandFunc, error) {
	testFlagSet.Parse(args)

	return command, nil
}

func command(conf *global.Config) error {
	return nil
}
