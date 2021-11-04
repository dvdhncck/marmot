package marmot

import (
	"database/sql"
	"encoding/json"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/disintegration/imaging"

	_ "github.com/go-sql-driver/mysql"
)

var genreCache = make(map[string]int64)
var artistCache = make(map[string]int64)

func Validate(db *sql.DB, path string) *Fails {

	fails := NewFails(path)

	log.Printf("Validating %s", path)

	PopulateGenreCache(db)
	populateArtistCache(db)

	// path should either be a valid-album, or a directory containing 1+ valid-albums

	// each valid-album should:
	//  1. not contain sub-dirs
	//  2. have a valid name
	//  3. contain a valid metadata json file
	//  4. contain a plausible image file
	//  5. contain 1 or more .mp3 files

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

func countSubDirsIn(fileInfos []fs.FileInfo) int {
	count := 0
	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			count += 1
		}
	}
	return count
}

func validateAlbum(fails *Fails, path string) {
	log.Printf("Validating album: %s", path)

	fails.Examine(validateName, path)

	fails.Examine(validateAlbumFolder, path)

	if fails.IsGood() {
		fails.Examine(validateMetadata, path)
		fails.Examine(validateCoverArt, path)
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

	// must have at least zero or more parent directories and a sensible end bit (i.e. aaa__xyz123)

	matched, _ := regexp.Match(`^(\/)?([0-9a-z_.]+\/)*[0-9a-z_.]+__[0-9a-z_.]+(\/)?$`, []byte(path))
	if !matched {
		fails.Add(PATH_CONTAINS_ILLEGAL_CHARACTERS)
	}
}

/*
	meta.json:
	{
		"id": 1234,
		"title": "this is the title",
		"artists": ["artist one", "artist two"]
		"genres": ["one","two","three"]
	}
*/

type metadata struct {
	ID 		int64     `json:"id"`
	Title   string    `json:"title"`
	Genres  []string  `json:"genres"`
	Artists []string  `json:"artists"`
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
	
	if metadata.ID == 0 {
		fails.Add(MISSING_ID_FIELD)
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

func validateCoverArt(fails *Fails, path string) {
	inputPath := filepath.Join(path, "cover.jpg")
	_, err := imaging.Open(inputPath)
	if err != nil {
		fails.Add(CANNOT_OPEN_COVER_ART)
	}

}

func validateArtist(fails *Fails, artist string) {
	if _, ok := artistCache[artist]; ok {
		log.Println(artist + ` exists`)
	} else {
		log.Println(artist + ` is NEW`)
	}
}

func validateGenre(fails *Fails, genre string) {	
	if len(strings.TrimSpace(genre)) == 0 {
		fails.Add(EMPTY_GENRE)
		return
	}

	if _, ok := genreCache[genre]; !ok {
		fails.Add(UNKNOWN_GENRE)
		return
	}
}

func PopulateGenreCache(db *sql.DB) {
	rows, err := db.Query("SELECT ID, Name FROM Genre")
	if err != nil {
		log.Printf("Error querying genres: %s", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var genre string
		var id int64
		err = rows.Scan(&id, &genre)
		if err != nil {
			log.Printf("Error scanning row: %s", err)
			return
		}
		genreCache[genre] = id
	}
}

func populateArtistCache(db *sql.DB) {
	rows, err := db.Query("SELECT ID, Name FROM Artist")
	if err != nil {
		log.Printf("Error querying artists: %s", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var artist string
		var id int64
		err = rows.Scan(&id, &artist)
		if err != nil {
			log.Printf("Error scanning row: %s", err)
			return
		}
		artistCache[artist] = id
	}
}