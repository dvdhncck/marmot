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

func Accept(db *sql.DB, path string, collection Collection) *Fails {

	fails := Validate(db, path)

	if fails.IsGood() {
		Ingest(db, path, collection)
		return FailsOf()
	} else {
		return fails
	}
}
	
func Validate(db *sql.DB, path string) *Fails {

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
					count := countSubDirsIn(contentsFileInfos) // shallow search
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
	// should contain at least the metadata file, and 0 directories
	if err == nil {
		if len(pathFileInfo) == 0 {
			fails.Add(EMPTY_PATH)
		} else {
			count := countSubDirsIn(pathFileInfo)
			if count > 0 {
				// all directories
				fails.Add(PATH_SHOULD_NOT_CONTAIN_DIRECTORIES)
			}
			// if len(pathFileInfo) < 2 {
			// 	fails.Add(PATH_SHOULD_CONTAIN_TWO_FILES_OR_MORE)
			// }
		}
	}

}

func validateName(fails *Fails, path string) {
	if strings.Count(path, `__`) != 1 {
		fails.Add(PATH_SHOULD_CONTAIN_EXACTLY_ONE_DELIMITER)
	}

	matched, _ := regexp.Match(`^(\/)?([0-9a-z_.]+\/)+[0-9a-z_.]+__[0-9a-z_.]+$`, []byte(path))
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

type metadata struct {
	Title   string
	Genres  []string
	Artists []string
}

func validateMetadata(fails *Fails, path string) {
	expectedPath := filepath.Join(path, "meta.json")

	file, err := ioutil.ReadFile(expectedPath)

	if err != nil {
		fails.Add(METADATA_MISSING)
		return
	}

	metadata := metadata{}
	err = json.Unmarshal([]byte(file), &metadata)

	if err != nil {
		fails.Add(COULD_NOT_PARSE_METADATA)
		// no need to continue
		return
	}

	if len(metadata.Title) == 0 {
		fails.Add(MISSING_TITLE_FIELD)
	}
	
	if metadata.Artists == nil {
		fails.Add(MISSING_ARTISTS_FIELD)
	} else {
		if len(metadata.Artists) == 0 {
			fails.Add(AT_LEAST_ONE_ARTIST_REQUIRED)
		}
		for _, artist := range metadata.Artists {
			validateArtist(fails, artist)
		}
	}

	if metadata.Genres == nil {
		fails.Add(MISSING_GENRES_FIELD)
	} else {
		if len(metadata.Genres) == 0 {
			fails.Add(AT_LEAST_ONE_GENRE_REQUIRED)
		}
		for _, genre := range metadata.Genres {
			validateGenre(fails, genre)
		}
	}
}

func validateArtist(fails *Fails, artist string) {
	//log.Printf("validating artist: %s", artist)
}

func validateGenre(fails *Fails, genre string) {	
	//log.Printf("validating genre: %s", genre)
}
