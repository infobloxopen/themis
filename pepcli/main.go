package main

import (
	"fmt"
	"os"
)

func main() {
	err := conf.cmd(conf.server, conf.input, conf.output, conf.count, conf.streams, conf.cmdConf)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
