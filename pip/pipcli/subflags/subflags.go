package subflags

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/infobloxopen/themis/pip/pipcli/global"
)

// Command defines CLI command.
type Command struct {
	// Name is command name and command line identifier.
	Name string
	// Desc is a description of the command.
	Desc string
	// Parser gets rest of command line arguments and returns a function which
	// executes the command.
	Parser func(args []string) (CommandFunc, error)
}

// CommandFunc is a prototype of a function which executes command.
type CommandFunc func(*global.Config) error

// MakeUsage creates usage description function for given set of commands.
func MakeUsage(cmds ...*Command) func() {
	return func() {
		base := path.Base(os.Args[0])
		fmt.Fprintf(os.Stderr,
			"Usage of %s:\n\n"+
				"  %s [GLOBAL OPTIONS] command [OPTIONS]\n\n"+
				"GLOBAL OPTIONS:\n",
			base, base,
		)
		flag.PrintDefaults()

		if len(cmds) > 0 {
			s := make([]string, len(cmds))
			for i, cmd := range cmds {
				s[i] = fmt.Sprintf("%s - %s", cmd.Name, cmd.Desc)
			}

			fmt.Fprintf(os.Stderr, "\nCOMMANDS:\n  %s\n", strings.Join(s, "\n  "))
		}
	}
}

// MakeCommandFlagSet creates flag set for given command.
func MakeCommandFlagSet(cmd *Command) *flag.FlagSet {
	f := flag.NewFlagSet(cmd.Name, flag.ExitOnError)
	f.Usage = func() {
		base := path.Base(os.Args[0])

		count := 0
		f.VisitAll(func(*flag.Flag) { count++ })
		if count > 0 {
			fmt.Fprintf(os.Stderr,
				"Usage of %s.%s:\n\n"+
					"  %s [GLOBAL OPTIONS] %s [%s OPTIONS]\n\n"+
					"GLOBAL OPTIONS:\n"+
					"  See %s -h\n\n"+
					"%s OPTIONS:\n",
				base, cmd.Name,
				base, cmd.Name, strings.ToUpper(cmd.Name),
				base,
				strings.ToUpper(cmd.Name),
			)
		} else {
			fmt.Fprintf(os.Stderr,
				"Usage of %s.%s:\n\n"+
					"  %s [GLOBAL OPTIONS] %s\n\n"+
					"GLOBAL OPTIONS:\n"+
					"  See %s -h\n",
				base, cmd.Name,
				base, cmd.Name,
				base,
			)
		}
		f.PrintDefaults()
	}

	return f
}

// Parse dispatches and run parser for a command from given set.
func Parse(cmds ...*Command) CommandFunc {
	m := make(map[string]*Command, len(cmds))
	for _, cmd := range cmds {
		m[strings.ToLower(cmd.Name)] = cmd
	}

	f, err := parse(m)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		flag.Usage()
		os.Exit(2)
	}

	return f
}

func parse(cmds map[string]*Command) (CommandFunc, error) {
	count := flag.NArg()
	if count < 1 {
		return nil, fmt.Errorf("no command provided")
	}

	name := flag.Arg(0)
	cmd, ok := cmds[strings.ToLower(name)]
	if !ok {
		return nil, fmt.Errorf("command provided but not defined: %s", name)
	}

	var args []string
	if count > 1 {
		args = flag.Args()[1:]
	}

	return cmd.Parser(args)
}
