package marmot

import (
	"fmt"
)

type Failure int

const (
	UNEXPECTED_FILE Failure = iota
	EMPTY_PATH
	CANNOT_OPEN
	CANNOT_READ
	CONTENTS_OF_ROOT_PATH_ARE_MIXED
	ROOT_SHOULD_BE_A_DIRECTORY
	PATH_CONTAINS_ILLEGAL_CHARACTERS
	PATH_SHOULD_CONTAIN_EXACTLY_ONE_DELIMITER
	PATH_SHOULD_NOT_CONTAIN_DIRECTORIES
	PATH_SHOULD_CONTAIN_TWO_FILES_OR_MORE
	METADATA_MISSING
	COULD_NOT_PARSE_METADATA
	MISSING_TITLE_FIELD
	TITLE_FIELD_SHOULD_BE_STRING
	MISSING_ARTISTS_FIELD
	AT_LEAST_ONE_ARTIST_REQUIRED
	MISSING_GENRES_FIELD
	AT_LEAST_ONE_GENRE_REQUIRED
	GENRES_FIELD_SHOULD_BE_LIST_OF_STRINGS
)

func (f Failure) String() string {
	switch f {
	case UNEXPECTED_FILE:
		return `UNEXPECTED_FILE`
	case EMPTY_PATH:
		return `EMPTY_PATH`
	case CANNOT_OPEN:
		return `CANNOT_OPEN`
	case CANNOT_READ:
		return `CANNOT_READ`
	case CONTENTS_OF_ROOT_PATH_ARE_MIXED:
		return `CONTENT_OF_ROOT_PATH_ARE_MIXED`
	case ROOT_SHOULD_BE_A_DIRECTORY:
		return `ROOT_SHOULD_BE_A_DIRECTORY`
	case PATH_CONTAINS_ILLEGAL_CHARACTERS:
		return `PATH_CONTAINS_ILLEGAL_CHARACTERS`
	case PATH_SHOULD_CONTAIN_EXACTLY_ONE_DELIMITER:
		return `PATH_SHOULD_CONTAIN_EXACTLY_ONE_DELIMITER`
	case PATH_SHOULD_NOT_CONTAIN_DIRECTORIES:
		return `PATH_SHOULD_NOT_CONTAIN_DIRECTORIES`
	case PATH_SHOULD_CONTAIN_TWO_FILES_OR_MORE:
		return `PATH_SHOULD_CONTAIN_TWO_FILES_OR_MORE`
	case METADATA_MISSING:
		return `METADATA_MISSING`
	case COULD_NOT_PARSE_METADATA:
		return `COULD_NOT_PARSE_METADATA`
	case MISSING_TITLE_FIELD:
		return `MISSING_TITLE_FIELD`
	case TITLE_FIELD_SHOULD_BE_STRING:
		return `TITLE_FIELD_SHOULD_BE_STRING`
	case MISSING_ARTISTS_FIELD:
		return `MISSING_ARTISTS_FIELD`
	case AT_LEAST_ONE_ARTIST_REQUIRED:
		return `AT_LEAST_ONE_ARTIST_REQUIRED`
	case MISSING_GENRES_FIELD:
		return `MISSING_GENRES_FIELD`
	case AT_LEAST_ONE_GENRE_REQUIRED:
		return `AT_LEAST_ONE_GENRE_REQUIRED`
	case GENRES_FIELD_SHOULD_BE_LIST_OF_STRINGS :
		return `GENRES_FIELD_SHOULD_BE_LIST_OF_STRINGS`
	default:
		return `Unknown fail value`
	}
}

type Fails struct {
	path string
	failures []Failure
}

type ValidationFunc func(*Fails, string)

func NewFails(path string) *Fails {
	return &Fails{path,[]Failure{}}
}

func FailsOf(failures ...Failure) *Fails {
	return &Fails{`no path`,failures}
}

func (fails *Fails) IsGood() bool {
	return len(fails.failures) == 0
}

func (fails *Fails) Includes(failure Failure) bool {
	for _,f := range fails.failures {
		if f == failure {
			return true
		}
	}
	return false
}

func (fails *Fails) ForEach(predicate func(Failure)) {
	for _,f := range fails.failures {
		predicate(f)
	}
}

func (fails *Fails) Examine(validationFunc ValidationFunc, path string) {
	validationFunc(fails, path)
}

func (fails *Fails) Add(failure Failure) {
	fails.failures = append(fails.failures, failure)
}

func (fails *Fails) Write() {
	for _, failure := range fails.failures {
		fmt.Println(failure)
	}
}
