package marmot

import (
	"flag"
)

/*


*/

type Settings struct {
	dryRun  bool
	verbose bool
	command []string
	acceptPath string
	rootPath string
}

var settings = Settings{}

func ParseArguments() {

	flag.BoolVar(&settings.verbose, "verbose", false, "be verbose")

	flag.BoolVar(&settings.dryRun, "dryRun", false, "don't affect anything")

	flag.StringVar(&settings.rootPath, "rootPath", "/home/dave/projects/html/marmot", "the path of root")

	flag.Parse()

	settings.command = flag.Args()
}

func (s *Settings) indexOf(command string) int {
	
	for index, item := range settings.command {
		if item == command {
			return index
		}
	}
	return -1
}

func (s *Settings) ExportFileName() string {
	return `marmot.xlsx`
}

func (s *Settings) ImportFileName() string {
	return `marmot.xlsx`
}

func (s *Settings) AcceptPath() string {
	return s.acceptPath
}

func (s *Settings) DoImportFromDatabase() bool {
	return s.indexOf(`db_import`) >= 0
}

func (s *Settings) DoExportToDatabase() bool {
	return s.indexOf(`db_export`) >= 0
}

func (s *Settings) DoImportFromExcel() bool {
	return s.indexOf(`excel_import`) >= 0
}

func (s *Settings) DoExportToExcel() bool {
	return s.indexOf(`excel_export`) >= 0
}

func (s *Settings) DoRemapLocations() bool {
	return s.indexOf(`remap_locations`) >= 0
}

func (s *Settings) DoTranslocate() bool {
	return s.indexOf(`translocate`) >= 0
}

func (s *Settings) DoIngest() bool {
	index := s.indexOf(`ingest`)
	if index >= 0 {
		settings.acceptPath = s.command[index + 1]
		return true
	}
	return false
}
