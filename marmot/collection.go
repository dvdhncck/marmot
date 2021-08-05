package marmot

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"strings"
)

type Collection struct {
	inDatabase map[string]*Album
	inFlight   []*Album
}

func (collection *Collection) Add(db *sql.DB, album *Album) {
	collection.enrich(db, album)
	collection.inFlight = append(collection.inFlight, album)
}

func (collection *Collection) getById(id string) *Album {
	return collection.inDatabase[id]
}

func (collection *Collection) LoadFromDatabase(db *sql.DB, filter string) int {
	//results, err := db.Query("SELECT ID, Name, SortAs FROM Album WHERE ID=336")
	//results, err := db.Query("SELECT ID, Name, SortAs FROM Album LIMIT 10")
	results, err := db.Query("SELECT ID, Name, SortAs, Location FROM Album")
	if err != nil {
		panic(err.Error())
	}
	count := 0
	if collection.inDatabase == nil {
		collection.inDatabase = make(map[string]*Album)
	}
	for results.Next() {
		var album Album
		var location sql.NullString // archiveName can be null
		album.artists = []*Artist{}
		err := results.Scan(&album.id, &album.name, &album.sortAs, &location)
		if err != nil {
			panic(err.Error())
		}
		if location.Valid {
			album.location = location.String
		}
		count++
		collection.inDatabase[album.id] = &album
	}
	return count
}

func (collection *Collection) validate() {
	for _, album := range collection.inDatabase {
		if len(album.artists) == 0 {
			log.Printf("Album %s missing artist(s)", album.id)
		}
		if album.mediaFolder == nil {
			log.Printf("Album %s missing mediaFolder", album.id)
		}
	}
}

func (collection *Collection) enrich(db *sql.DB, album *Album) {
	collection.enrichArtistsForAlbum(db, album)
	collection.enrichGenresForAlbum(db, album)
}

func (collection *Collection) makeIdList() string {
	keys := make([]string, len(collection.inDatabase))
	i := 0
	for k := range collection.inDatabase {
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

func (collection *Collection) enrichArtistsForAlbum(db *sql.DB, album *Album) {
	for _, artist := range album.artists {
		results, err := db.Query("SELECT ID FROM Artist WHERE Name=?", artist.name)
		if err != nil {
			panic(err.Error())
		} else {
			for results.Next() {
				results.Scan(&artist.id)
			}
		}
	}
}

func (collection *Collection) enrichGenresForAlbum(db *sql.DB, album *Album) {
	for _, genre := range album.genres {
		results, err := db.Query("SELECT ID FROM Genre WHERE Name=?", genre.name)
		if err != nil {
			panic(err.Error())
		} else {
			for results.Next() {
				results.Scan(&genre.id)
			}
		}
	}
}

func (collection *Collection) getArtistsForCollection(db *sql.DB) int {
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

func (collection *Collection) getGenresForCollection(db *sql.DB) int {
	albumIdList := collection.makeIdList()
	results, err := db.Query("SELECT AlbumID, g.ID as GenreID, g.Name AS GenreName FROM AlbumGenre ag JOIN Genre g ON g.ID=ag.GenreID AND ag.AlbumID IN " + albumIdList)
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
					album.genres = []*Genre{}
				}
				album.genres = append(album.genres, &Genre{genreId.String, genreName.String})
			}
		} else {
			log.Printf("Genre with NULL ID for Album %s\n", albumId)
		}
		count++

	}
	return count
}

func (collection *Collection) WriteToJson(filename string) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic(err.Error())
	}

	defer file.Close()
	delimiter := "{\n"
	for _, album := range collection.inDatabase {
		fmt.Fprintf(file, "%s%s", delimiter, album.ToJson())
		delimiter = ",\n"
	}
	fmt.Fprintf(file, "} /*x*/\n")
}

func (collection *Collection) RemapLocations() {
	for _, album := range collection.inDatabase {
		album.location, album.mapState = MapLocation(album)
		album.clean = false
	}
}

func (collection *Collection) Size() int {
	if collection.inDatabase == nil {
		return 0
	} else {
		return len(collection.inDatabase)
	}
}

func (collection *Collection) UpdateNewLocation(albumId string, newLocation string) error {
	if collection.inDatabase == nil {
		return errors.New(`Collection is empty`)
	}
	album := collection.inDatabase[albumId]
	if album == nil {
		return errors.New(`Unable to find album ` + albumId)
	}
	album.location = newLocation
	album.clean = false
	return nil // no error
}

type DryRunResult struct{}

var magicId = int64(1000000)

func (d DryRunResult) LastInsertId() (int64, error) {
	magicId += 1
	return magicId, nil
}
func (d DryRunResult) RowsAffected() (int64, error) {
	return 0, nil
}

func MaybeExecute(db *sql.DB, query string, args ...interface{}) (sql.Result, error) {
	if settings.dryRun {
		message := query + " ("
		for _, a := range args {
			message += fmt.Sprintf("%v,", a)
		}
		message += ")"
		log.Print(message)
		return DryRunResult{}, nil
	} else {
		return db.Exec(query, args...)
	}
}

func (collection *Collection) addAlbumToDatabase(db *sql.DB, album *Album) {

	for _, artist := range album.artists {
		if artist.id == `` {
			result, err := MaybeExecute(db, "INSERT INTO Artist (ID, Name) VALUES (NULL, ?)", artist.name)
			if err == nil {
				id, err := result.LastInsertId()
				if err != nil {
					log.Fatal(err.Error())
				} else {
					artist.id = fmt.Sprintf(`%d`, id)
				}
			} else {
				log.Fatal(err.Error())
			}
		}
	}

	result, err := MaybeExecute(db, "INSERT INTO Album (ID, Name, Location) VALUES (NULL, ?, ?)", album.name, album.location)
	if err == nil {
		id, err := result.LastInsertId()
		if err != nil {
			log.Fatal(err.Error())
		} else {
			album.id = fmt.Sprintf(`%d`, id)
		}
	} else {
		log.Fatal(err.Error())
	}

	_, err = MaybeExecute(db, "DELETE FROM AlbumArtist WHERE AlbumID=?", album.id)
	if err != nil {
		log.Fatal(err.Error())
	}

	for _, artist := range album.artists {
		_, err := MaybeExecute(db, "INSERT INTO AlbumArtist (AlbumID, ArtistID) VALUES (?, ?)", album.id, artist.id)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	_, err = MaybeExecute(db, "DELETE FROM AlbumGenre WHERE AlbumID=?", album.id)
	if err != nil {
		log.Fatal(err.Error())
	}

	for _, genre := range album.genres {
		_, err := MaybeExecute(db, "INSERT INTO AlbumGenre (AlbumID, GenreID) VALUES (?, ?)", album.id, genre.id)
		if err != nil {
			log.Fatal(err.Error())
		}
	}
}

func (collection *Collection) WriteToDatabase(db *sql.DB) {

	count := 0
	for _, album := range collection.inDatabase {
		if !album.clean {
			_, err := MaybeExecute(db, "UPDATE Album SET Location=? WHERE ID=?", album.location, album.id)
			if err != nil {
				log.Fatal(err.Error())
			}
			count += 1
			album.clean = true
		}
	}

	log.Printf("Updated %d location entries in database", count)

	for _, album := range collection.inFlight {
		collection.addAlbumToDatabase(db, album)
		InstallCoverArt(album)
	}
	

	log.Printf("Added %d new entries to database", len(collection.inFlight))
}
