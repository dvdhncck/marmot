package main

import (
	"flag"
)

type Settings struct {
	dryRun               bool
	verbose              bool
	command				 []string
}

var settings = Settings{}

func ParseArguments() {

	flag.BoolVar(&settings.verbose, "verbose", false, "be verbose")
	flag.BoolVar(&settings.dryRun, "dryRun", false, "don't affect anything")
	
	flag.Parse()

	settings.command = flag.Args()
}

