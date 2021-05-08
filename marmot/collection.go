package marmot

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"strings"
)

type Collection struct {
	lookup map[string]*Album
}

func (collection *Collection) getById(id string) *Album {
	return collection.lookup[id]
}

func (collection *Collection) retrieve(db *sql.DB, filter string) int {
	//results, err := db.Query("SELECT ID, Name, SortAs FROM Album WHERE ID=336")
	//results, err := db.Query("SELECT ID, Name, SortAs FROM Album LIMIT 10")
	results, err := db.Query("SELECT ID, Name, SortAs FROM Album")
	if err != nil {
		panic(err.Error())
	}
	count := 0
	if collection.lookup == nil {
		collection.lookup = make(map[string]*Album)
	}
	for results.Next() {
		var album Album
		album.artists = []*Artist{}
		err := results.Scan(&album.id, &album.name, &album.sortAs)
		if err != nil {
			panic(err.Error())
		}
		count++
		collection.lookup[album.id] = &album
	}
	return count
}

func (collection *Collection) validate() {
	for _, album := range collection.lookup {
		if len(album.artists) == 0 {
		 	log.Printf("Album %s missing artist(s)", album.id)
		}
		if album.mediaFolder == nil {
			log.Printf("Album %s missing mediaFolder", album.id)
		}
	}
}

func (collection *Collection) makeIdList() string {
	keys := make([]string, len(collection.lookup))
	i := 0
	for k := range collection.lookup {
		keys[i] = k
		i++
	}
	return `(` + strings.Join(keys, ",") + `)`
}

func (collection *Collection) getMediaFolders(db *sql.DB) int {
	albumIdList := collection.makeIdList()
	results, err := db.Query("SELECT AlbumID, ArchiveName, MountPoint, RootPath, FolderPath FROM MediaFolder WHERE MediaType=1 AND AlbumID IN " + albumIdList)
	if err != nil {
		panic(err.Error())
	}
	count := 0
	albumId := ""
	for results.Next() {
		mediaFolder := MediaFolder{}

		var archiveName sql.NullString // archiveName can be null

		err := results.Scan(&albumId, &archiveName, &mediaFolder.mountPoint, &mediaFolder.rootPath, &mediaFolder.folderPath)

		if archiveName.Valid {
			mediaFolder.archiveName = archiveName.String
		}

		if err != nil {
			panic(err.Error())
		}

		album := collection.getById(albumId)
		if album == nil {
			panic(err.Error())
		}

		if album.mediaFolder != nil {
			log.Printf(`Multiple MediaFolder for Album ID %s`, albumId)
		} else {
			album.mediaFolder = &mediaFolder
		}

		count++
	}
	return count
}

func (collection *Collection) getArtists(db *sql.DB) int {
	albumIdList := collection.makeIdList()
	results, err := db.Query("SELECT AlbumID, ID, Name, SortAs FROM AlbumArtist aa JOIN Artist a ON a.ID=aa.ArtistID AND aa.AlbumID IN " + albumIdList)
	if err != nil {
		panic(err.Error())
	}

	count := 0
	albumId := ""

	var artistId sql.NullString
	var artistName sql.NullString
	var sortAs sql.NullString

	artists := make(map[string]*Artist)

	for results.Next() {

		err = results.Scan(&albumId, &artistId, &artistName, &sortAs)
		if err != nil {
			panic(err.Error())
		}

		if artistId.Valid {
			artist := artists[artistId.String] // have we seen this one already?

			if artist == nil {
				// nope, create it and keep track of it
				artist = &Artist{id: artistId.String, name: artistName.String, sortAs: sortAs.String}
				artists[artistId.String] = artist
			}

			album := collection.getById(albumId)
			if album == nil {
				//panic(err.Error())
				log.Printf("Album %s doesnt exist, but has an artist", albumId)
			} else {
				// add it to the list. all instances of the same artist will point to the same Artist struct
				album.artists = append(album.artists, artist)
			}
		} else {
			log.Printf("Artist with NULL ID for Album %s\n", albumId)
		}
		count++

	}
	return count
}

func (collection *Collection) getGenres(db *sql.DB) int {
	albumIdList := collection.makeIdList()
	results, err := db.Query("SELECT AlbumID, GenreID, Name FROM AlbumGenre ag JOIN Genre g ON g.ID=ag.GenreID AND ag.AlbumID IN " + albumIdList)
	if err != nil {
		panic(err.Error())
	}

	count := 0
	albumId := ""

	var genreId sql.NullString
	var genreName sql.NullString

	for results.Next() {
		err = results.Scan(&albumId, &genreId, &genreName)
		if err != nil {
			panic(err.Error())
		}

		if genreId.Valid {
			album := collection.getById(albumId)
			if album == nil {
				//panic(err.Error())
				log.Printf("Album %s doesnt exist, but has a genre", albumId)
			} else {
				if album.genres == nil {
					album.genres = []string{}
				}
				album.genres = append(album.genres, genreName.String)
			}
		} else {
			log.Printf("Genre with NULL ID for Album %s\n", albumId)
		}
		count++

	}
	return count
}

func (collection *Collection) writeToJson(filename string) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic(err.Error())
	}

	defer file.Close()
	delimiter := "{\n"
	for _, album := range collection.lookup {
		fmt.Fprintf(file, "%s%s", delimiter, album.toJson())
		delimiter = ",\n"
	}
	fmt.Fprintf(file, "} /*x*/\n")
}
