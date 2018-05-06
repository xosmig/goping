package main

import (
	"github.com/xosmig/goping"
	"os"
)

func main() {
	params, err := goping.ParseCommandLine(os.Args[1:], os.Stderr)
	if err != nil {
		os.Exit(3)
	}

	goping.UrlReachable(params, os.Stdout)
}
