package perf

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"
)

type config struct {
	parallel bool
}

var perfFlagSet = flag.NewFlagSet(Name, flag.ExitOnError)

func FlagsParser(args []string) interface{} {
	conf := config{}

	perfFlagSet.Usage = usage
	perfFlagSet.BoolVar(&conf.parallel, "p", false, "make requests in parallel")
	perfFlagSet.Parse(args)

	count := perfFlagSet.NArg()
	if count > 1 {
		tail := strings.Join(perfFlagSet.Args()[1:count], "\", \"")
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
	perfFlagSet.PrintDefaults()
}
