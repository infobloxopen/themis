package perf

import (
	"flag"
	"fmt"
	"os"

	"github.com/infobloxopen/themis/pip/pipcli/subflags"
)

const (
	perfName = "perf"
	perfDesc = "measures request roundtrip timings"

	defWorkers = 100
)

var (
	// PerfCommand is a command "perf" of PIP CLI.
	PerfCommand = &subflags.Command{
		Name:   perfName,
		Desc:   perfDesc,
		Parser: parser,
	}

	perfFlagSet *flag.FlagSet
)

func init() {
	perfFlagSet = subflags.MakeCommandFlagSet(PerfCommand)
}

func parser(args []string) (subflags.CommandFunc, error) {
	perfFlagSet.IntVar(&workers, "w", defWorkers, "number of workers to make requests in parallel")
	if err := perfFlagSet.Parse(args); err != nil {
		return nil, err
	}

	validateWorkers()

	return command, nil
}

func validateWorkers() {
	if workers <= 0 {
		fmt.Fprintf(os.Stderr, "%d is too small for a number of workers. using default...\n", workers)
		workers = defWorkers
	}
}
