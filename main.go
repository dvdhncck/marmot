package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"strings"
)

type MediaFolder struct {
	archiveName string
	mountPoint  string
	rootPath    string
	folderPath  string
}

func (mf *MediaFolder) toJson() string {
	return fmt.Sprintf("[\"%s/%s/%s\"]",
		mf.mountPoint, mf.rootPath, mf.folderPath)
}

type Album struct {
	id          string
	name        string
	mediaFolder *MediaFolder
	location    string          // this is where we will migrate it to
	sortAs      string
	artists     []*Artist
	genres      []string
}

func (album *Album) toJson() string {
	artists := []string{}
	for _, artist := range album.artists {
		artists = append(artists, artist.toJson())
	}
	genres := []string{}
	for _, genre := range album.genres {
		genres = append(genres, fmt.Sprintf("\"%s\"", genre))
	}
	return fmt.Sprintf("{ id: \"%s\",\n  name: \"%s\",\n  location: %s,\n  sortAs: \"%s\",\n  genres: [%s]\n  artists: [%s]\n}",
		album.id, album.name, album.mediaFolder.toJson(), album.sortAs, strings.Join(genres, `,`), strings.Join(artists, `,`))
}

type Artist struct {
	id     string
	name   string
	sortAs string
}

func (artist *Artist) toJson() string {
	return fmt.Sprintf("\n  { id: \"%s\",\n    name: \"%s\",\n    sortAs: \"%s\"\n  }",
		artist.id, artist.name, artist.sortAs)
}

type Albums struct {
	lookup map[string]*Album
}

func (albums *Albums) getById(id string) *Album {
	return albums.lookup[id]
}

func (albums *Albums) retrieve(db *sql.DB, filter string) int {
	//results, err := db.Query("SELECT ID, Name, SortAs FROM Album WHERE ID=336")
	//results, err := db.Query("SELECT ID, Name, SortAs FROM Album LIMIT 10")
	results, err := db.Query("SELECT ID, Name, SortAs FROM Album")
	if err != nil {
		panic(err.Error())
	}
	count := 0
	if albums.lookup == nil {
		albums.lookup = make(map[string]*Album)
	}
	for results.Next() {
		var album Album
		album.artists = []*Artist{}
		err := results.Scan(&album.id, &album.name, &album.sortAs)
		if err != nil {
			panic(err.Error())
		}
		count++
		albums.lookup[album.id] = &album
	}
	return count
}

func (albums *Albums) validate() {
	for _, album := range albums.lookup {
		if len(album.artists) == 0 {
		 	log.Printf("Album %s missing artist(s)", album.id)
		}
		if album.mediaFolder == nil {
			log.Printf("Album %s missing mediaFolder", album.id)
		}
	}
}

func (albums *Albums) decideNewLocation() {
	correct := 0
	problematic := 0

	for _, album := range albums.lookup {
		
		oldLocation := album.mediaFolder.folderPath

		/*
		 things to fix....

		  leading /

		  upper case

		  subdirectory

		*/

		
	}
}

func (albums *Albums) makeIdList() string {
	keys := make([]string, len(albums.lookup))
	i := 0
	for k, _ := range albums.lookup {
		keys[i] = k
		i++
	}
	return `(` + strings.Join(keys, ",") + `)`
}

func (albums *Albums) getMediaFolders(db *sql.DB) int {
	albumIdList := albums.makeIdList()
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

		album := albums.getById(albumId)
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

func (albums *Albums) getArtists(db *sql.DB) int {
	albumIdList := albums.makeIdList()
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

			album := albums.getById(albumId)
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

func (albums *Albums) getGenres(db *sql.DB) int {
	albumIdList := albums.makeIdList()
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
			album := albums.getById(albumId)
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

func (albums *Albums) writeToJson(filename string) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic(err.Error())
	}

	defer file.Close()
	delimiter := "{\n"
	for _, album := range albums.lookup {
		fmt.Fprintf(file, "%s%s", delimiter, album.toJson())
		delimiter = ",\n"
	}
	fmt.Fprintf(file, "} /*x*/\n")
}

func main() {
	db, err := sql.Open("mysql", "dave:dave@tcp(127.0.0.1:3306)/marmot")

	if err != nil {
		panic(err.Error())
	}

	defer db.Close()

	albums := Albums{}

	count := albums.retrieve(db, `WHERE NAME LIKE '%Hotel%'`)
	albums.getMediaFolders(db)
	albums.getArtists(db)
	albums.getGenres(db)
	albums.validate()
	albums.writeToJson(`test.json`)

	log.Printf("Scanned %d albums", count)
}
