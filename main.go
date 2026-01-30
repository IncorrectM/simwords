package main

import (
	"fmt"
	"os"
	"yggdrasil/sim-words/cmd"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("expected subcommand")
		os.Exit(1)
	}

	subcommand := os.Args[1]
	flags := []string{}
	if len(os.Args) >= 3 {
		flags = os.Args[2:]
	}

	switch subcommand {
	case "load":
		cmd.RunLoad(flags)
	}
}
