package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/infobloxopen/themis/pepcli/perf"
	"github.com/infobloxopen/themis/pepcli/test"
)

type config struct {
	server string
	input  string
	count  int
	output string

	cmdConf interface{}
	cmd     cmdExec
}

var conf = config{}

type (
	cmdExec       func(addr, in, out string, n int, conf interface{}) error
	cmdFlagParser func(args []string) interface{}

	command struct {
		exec   cmdExec
		parser cmdFlagParser
	}

	cmdDesc struct {
		name string
		desc string
	}
)

var (
	cmds = map[string]command{
		test.Name: {
			exec:   test.Exec,
			parser: test.FlagsParser,
		},
		perf.Name: {
			exec:   perf.Exec,
			parser: perf.FlagsParser,
		},
	}

	descs = []cmdDesc{
		{
			name: test.Name,
			desc: test.Description,
		},
		{
			name: perf.Name,
			desc: perf.Description,
		},
	}
)

func init() {
	flag.Usage = usage

	flag.StringVar(&conf.server, "s", "127.0.0.1:5555", "PDP server to work with")
	flag.StringVar(&conf.input, "i", "requests.yaml", "file with YAML formatted list of requests to send to PDP")
	flag.IntVar(&conf.count, "n", 0, "number or requests to send "+
		"(default and value less than one means all requests from file)")
	flag.StringVar(&conf.output, "o", "", "file to write YAML formatted list of responses from PDP "+
		"(default stdout)")

	flag.Parse()

	count := flag.NArg()
	if count < 1 {
		fmt.Fprint(os.Stderr, "no command provided\n")
		flag.Usage()
		os.Exit(2)
	}

	name := flag.Arg(0)
	cmd, ok := cmds[name]
	if !ok {
		fmt.Fprintf(os.Stderr, "command provided but not defined: %s\n", name)
		flag.Usage()
		os.Exit(2)
	}

	var args []string
	if count > 1 {
		args = flag.Args()[1:count]
	}

	conf.cmdConf = cmd.parser(args)
	conf.cmd = cmd.exec
}

func usage() {
	base := path.Base(os.Args[0])
	fmt.Fprintf(os.Stderr,
		"Usage of %s:\n\n"+
			"  %s [GLOBAL OPTIONS] command [OPTIONS]\n\n"+
			"GLOBAL OPTIONS:\n", base, base)
	flag.PrintDefaults()

	s := make([]string, len(descs))
	for i, desc := range descs {
		s[i] = fmt.Sprintf("%s - %s", desc.name, desc.desc)
	}

	fmt.Fprintf(os.Stderr, "\nCOMMANDS:\n  %s\n", strings.Join(s, "\n  "))
}
