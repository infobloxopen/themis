package main

import (
	"fmt"
	"os"
)

func main() {
	err := conf.cmd(
		conf.servers[0],
		conf.input,
		conf.output,
		conf.count,
		conf.cmdConf,
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
