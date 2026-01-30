package main

import (
	"fmt"
	"log"
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
	case "query":
		cmd.RunQuery(flags)
	case "serve":
		cmd.RunServe(flags)
	default:
		log.Fatalf("unknown command %s", subcommand)
	}
}
