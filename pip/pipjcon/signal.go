package main

import (
	"os"
	"os/signal"
	"syscall"
)

func waitForInterrupt() {
	ch := make(chan os.Signal, 1)
	defer close(ch)

	signal.Notify(ch, syscall.SIGHUP, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(ch)

	<-ch
}
