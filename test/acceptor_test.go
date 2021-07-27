package main

import (
	m "davidhancock.com/marmot/marmot"
	"log"
	"os"
	"testing"
)

func expectPass(t *testing.T, path string,) {
	expectFail(t, path, m.FailsOf())
}

func expectFail(t *testing.T, path string, expected *m.Fails) {
	actual := m.Validate(nil, path)
	failed := false

	expected.ForEach(func(f m.Failure) {
		if !actual.Includes(f) {
			log.Printf("%s: Expected %s, not found", path, f.String())
			failed = true
		} else {
			log.Printf("%s: Expected %s, got it", path, f.String())
		}
	})

	actual.ForEach(func(f m.Failure) {
		if !expected.Includes(f) {
			log.Printf("%s: Found %s, which was not expected", path, f.String())
			failed = true
		}
	})

	if(failed) {
		t.Fatal(`Fail on path:`, path)
	}
}

func Test_shouldDetectWonkyPaths(t *testing.T) {

	log.SetOutput(os.Stdout)

	// layout good
	expectPass(t, `data/good1`)

	// layout good, 2 albums in one folder
	expectPass(t, `data/good2`)

	// doesnt exist
	expectFail(t, `data/fitional`, m.FailsOf(m.CANNOT_OPEN))

	// contains nothing
	expectFail(t, `data/empty`, m.FailsOf(m.EMPTY_PATH))

	// contain illegally name subdirs:
	expectFail(t, `data/bad1`, m.FailsOf(m.EMPTY_PATH, m.PATH_SHOULD_CONTAIN_EXACTLY_ONE_DELIMITER, m.PATH_CONTAINS_ILLEGAL_CHARACTERS))
	expectFail(t, `data/bad2`, m.FailsOf(m.EMPTY_PATH, m.PATH_CONTAINS_ILLEGAL_CHARACTERS))
	expectFail(t, `data/bad3`, m.FailsOf(m.PATH_SHOULD_NOT_CONTAIN_DIRECTORIES))
	expectFail(t, `data/bad4`, m.FailsOf(m.COULD_NOT_PARSE_METADATA))
	expectFail(t, `data/bad5`, m.FailsOf(m.EMPTY_PATH))

	expectFail(t, `data/bad6__metadata_has_bad_fields`, m.FailsOf(m.COULD_NOT_PARSE_METADATA))

	expectFail(t, `data/bad7__metadata_missing_artists`, m.FailsOf(m.MISSING_ARTISTS_FIELD))

	expectFail(t, `data/bad8__metadata_wrong_type_for_artists`,	m.FailsOf(m.COULD_NOT_PARSE_METADATA))

	expectFail(t, `data/bad9__metadata_missing_genres`,	m.FailsOf(m.MISSING_GENRES_FIELD))
}
