package marmot

import (
	"flag"
)

type Settings struct {
	server  bool
	dryRun  bool
	verbose bool
	command []string
	rootPath string
}

var settings = Settings{}

func ParseArguments() {

	flag.BoolVar(&settings.verbose, "verbose", false, "be verbose")

	flag.BoolVar(&settings.server, "server", false, "run as server")

	flag.BoolVar(&settings.dryRun, "dryRun", false, "don't affect anything")

	flag.StringVar(&settings.rootPath, "rootPath", "/library/music/", "the path of root")

	flag.Parse()

	settings.command = flag.Args()
}

func (s *Settings) NextToken() string {
	result := s.command[0]
	s.command = s.command[1:]
	return result
}

func (s *Settings) HasToken() bool {
	return len(s.command) > 0
}

