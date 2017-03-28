package main

import (
	"fmt"
	"os"
)

func main() {
	client := NewClient()
	err := client.Connect(config.Server)
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't connect to \"%s\": %s\n", config.Server, err)
		os.Exit(1)
	}
	defer client.Close()
}
