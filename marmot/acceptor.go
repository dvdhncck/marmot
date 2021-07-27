package marmot

import (
	"database/sql"
	"encoding/json"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func countSubDirsIn(fileInfos []fs.FileInfo) int {
	count := 0
	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			count += 1
		}
	}
	return count
}

func Accept(db *sql.DB, path string) *Fails {

	// rootPath should either be a valid-album, or a directory containing 1+ valid-albums

	// each valid-album should:
	//  1. not contain sub-dirs
	//  2. have a valid name
	//  3. contain a valid metadata json file
	//  4. contain a plausible image file
	//  5. contain 1 or more .mp3 files

	fails := NewFails(path)

	rootFileInfo, err := os.Stat(path)
	if err == nil {
		if rootFileInfo.IsDir() {
			contentsFileInfos, err := ioutil.ReadDir(path) // this is the contents of 'path' - it should either be a set of files or a set of directories, but not a mixture
			if err == nil {
				if len(contentsFileInfos) == 0 {
					fails.Add(EMPTY_PATH)
				} else {
					count := countSubDirsIn(contentsFileInfos)  // shallow search
					if count == len(contentsFileInfos) {
						// all directories
						for _, fileInfo := range contentsFileInfos {
							if fileInfo.IsDir() {
								fails.Examine(validateAlbum, filepath.Join(path, fileInfo.Name()))
							} else {
								fails.Add(UNEXPECTED_FILE)
							}
						}
					} else if count == 0 {
						// all files
						fails.Examine(validateAlbum, path)
					} else {
						// mixture of files and directories
						fails.Add(CONTENTS_OF_ROOT_PATH_ARE_MIXED)
					}
				}
			} else {
				fails.Add(CANNOT_READ)
			}
		} else {
			fails.Add(ROOT_SHOULD_BE_A_DIRECTORY)
		}
	} else {
		fails.Add(CANNOT_OPEN)
	}
	return fails
}

func validateAlbum(fails *Fails, path string) {
	fails.Examine(validateName, path)

	fails.Examine(validateAlbumFolder, path)

	if fails.IsGood() {
		fails.Examine(validateMetadata, path)
	}
}

func validateAlbumFolder(fails *Fails, path string) {
	pathFileInfo, err := ioutil.ReadDir(path)
	// should contain at least 2 files, and 0 directories
	if err == nil {
		if len(pathFileInfo) == 0 {
			fails.Add(EMPTY_PATH)
		} else {
			count := countSubDirsIn(pathFileInfo)
			if count > 0 {
				// all directories
				fails.Add(PATH_SHOULD_NOT_CONTAIN_DIRECTORIES)
			}
			if len(pathFileInfo) < 2 {
				fails.Add(PATH_SHOULD_CONTAIN_TWO_FILES_OR_MORE)
			}
		}
	}

}

func validateName(fails *Fails, path string) {
	if strings.Count(path, `__`) != 1 {
		fails.Add(PATH_SHOULD_CONTAIN_EXACTLY_ONE_DELIMITER)
	}

	matched, _ := regexp.Match(`^([0-9a-z_]+\/)+[0-9a-z_]+__[0-9a-z_]+$`, []byte(path))
	if !matched {
		fails.Add(PATH_CONTAINS_ILLEGAL_CHARACTERS)
	}
}

/*
	meta.json:
	{
		title: 'this is the title',
		artists: ['artist one', 'artist two']
		genres: ['one','two','three']
	}
*/

func validateMetadata(fails *Fails, path string) {
	expectedPath := filepath.Join(path, "meta.json")

	file, err := ioutil.ReadFile(expectedPath)

	if err != nil {
		fails.Add(METADATA_MISSING)
		return
	}

	var data map[string]interface{}

	err = json.Unmarshal([]byte(file), &data)

	if err != nil {
		fails.Add(COULD_NOT_PARSE_METADATA)
	}

	if data[`title`] == nil {
		fails.Add(MISSING_TITLE_FIELD)
	} else {
		_, goodType := data[`title`].(string)
		if !goodType {
			fails.Add(TITLE_FIELD_SHOULD_BE_STRING)
		}
	}

	if data[`artists`] == nil {
		fails.Add(MISSING_ARTISTS_FIELD)
	} else {
		_, goodType := data[`artists`].([]string)
		if !goodType {
			fails.Add(ARTISTS_FIELD_SHOULD_BE_LIST_OF_STRINGS)
		}
	}

	if data[`genres`] == nil {
		fails.Add(MISSING_GENRES_FIELD)
	} else {
		_, goodType := data[`artists`].([]string)
		if !goodType {
			fails.Add(GENRES_FIELD_SHOULD_BE_LIST_OF_STRINGS)
		}
	}
}
