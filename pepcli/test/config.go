package test

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"
)

type config struct {
	input  string
	count  int
	output string
}

var testFlagSet = flag.NewFlagSet(Name, flag.ExitOnError)

func FlagsParser(args []string) interface{} {
	conf := config{}

	testFlagSet.Usage = usage
	testFlagSet.StringVar(&conf.input, "i", "requests.yaml", "file with YAML formatted list of requests to send to PDP")
	testFlagSet.IntVar(&conf.count, "n", 0, "number or requests to send "+
		"(default and value less than one means all requests from file)")
	testFlagSet.StringVar(&conf.output, "o", "", "file to write YAML formatted list of responses from PDP "+
		"(default stdout)")

	testFlagSet.Parse(args)

	count := testFlagSet.NArg()
	if count > 1 {
		tail := strings.Join(testFlagSet.Args()[1:count], "\", \"")
		fmt.Fprintf(os.Stderr, "trailing arguments after cluster name: \"%s\"\n", tail)
		usage()
		os.Exit(2)
	}

	return conf
}

func usage() {
	base := path.Base(os.Args[0])
	fmt.Fprintf(os.Stderr,
		"Usage of %s.%s:\n\n"+
			"  %s [GLOBAL OPTIONS] %s [%s OPTIONS]\n\n"+
			"GLOBAL OPTIONS:\n"+
			"  See %s -h\n\n"+
			"%s OPTIONS:\n", base, Name, base, Name, strings.ToUpper(Name), base, strings.ToUpper(Name))
	testFlagSet.PrintDefaults()
}
