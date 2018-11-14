package test

import (
	"flag"

	"github.com/infobloxopen/themis/pip/pipcli/subflags"
)

const (
	testName = "test"
	testDesc = "sends information requests to PIP"
)

var (
	// TestCommand is a command "test of PIP CLI.
	TestCommand = &subflags.Command{
		Name:   testName,
		Desc:   testDesc,
		Parser: parser,
	}

	testFlagSet *flag.FlagSet
)

func init() {
	testFlagSet = subflags.MakeCommandFlagSet(TestCommand)
}

func parser(args []string) (subflags.CommandFunc, error) {
	if err := testFlagSet.Parse(args); err != nil {
		return nil, err
	}

	return command, nil
}
