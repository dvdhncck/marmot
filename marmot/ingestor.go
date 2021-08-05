package marmot

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/disintegration/imaging"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

// this assumes that the path has been validated successfully
func Ingest(db *sql.DB, path string) {

	// the path must be rooted inside the music directory structure
	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Println(err.Error())
		log.Fatal(`Unable to determine absolute path of ` + path)
	}
	if !strings.HasPrefix(absPath, settings.rootPath) {
		log.Fatal(path + ` is not rooted under ` + settings.rootPath)
	}

	contentsFileInfos, err := ioutil.ReadDir(path) // this is the contents of 'path' - it should either be a set of files or a set of directories, but not a mixture
	if err == nil {
		count := countSubDirsIn(contentsFileInfos) // shallow search
		if count == len(contentsFileInfos) {
			// all directories
			for _, fileInfo := range contentsFileInfos {
				IngestPath(db, filepath.Join(path, fileInfo.Name()))
			}
		} else {
			// all files
			IngestPath(db, path)
		}
	}

}

// this assumes that the path has been validated successfully
func IngestPath(db *sql.DB, path string) {
	expectedPath := filepath.Join(path, "meta.json")

	file, _ := ioutil.ReadFile(expectedPath)
	metadata := metadata{}
	_ = json.Unmarshal([]byte(file), &metadata)

	// we've alfready checked that its rooted correctly inder the music directory structure
	
	actualLocation := filepath.Base(path)

	album := NewAlbumFromFilesystem(metadata.ID, actualLocation, metadata.Title, metadata.Artists, metadata.Genres)

	WriteToDatabase(db, album)
}

// this assumes that the album's path has been validated successfully (i.e. there is an image there)
func InstallCoverArt(album *Album) {
	inputPath := filepath.Join(album.location, "cover.jpg")
	log.Printf("Seeking cover art from %s", inputPath)
	img, _ := imaging.Open(inputPath)
	log.Printf("Using image from %s", inputPath)
		
	image640 := imaging.Resize(img, 640, 640, imaging.Lanczos)
	path640 := filepath.Join(album.location, fmt.Sprintf("%s_640.jpg", album.id));

	image160 := imaging.Resize(img, 160, 160, imaging.Lanczos)
	path160 := filepath.Join(album.location, fmt.Sprintf("%s_160.jpg", album.id));

	if settings.dryRun {
		log.Printf("Dry run: would save to %s, %s", path640, path160)
	} else {
		log.Printf("Writing to %s", path640)
		err := imaging.Save(image640, path640)
		if err != nil {
			log.Fatal(fmt.Sprintf("Error saving image: %s", err))
		}
		log.Printf("Writing to %s", path160)
		err = imaging.Save(image160, path160)
		if err != nil {
			log.Fatal(fmt.Sprintf("Error saving image: %s", err))
		}
	}
}

func WriteToDatabase(db *sql.DB, album *Album) {

	 _, err := MaybeExecute(db, 
		`INSERT INTO Album (ID, Name, Location) VALUES (?, ?, ?)`, 
		album.id, album.name, album.location)
	
	if err != nil {
		log.Fatal(fmt.Sprintf("Error inserting album: %s", err))
	}

	for _, genre := range album.genres {
		if genreId, ok := genreCache[genre.name]; ok {
			_, err := MaybeExecute(db, 
				`INSERT INTO AlbumGenre (AlbumID, GenreID) VALUES (?, ?)`, 
				album.id, genreId)
			
			if err != nil {
				log.Fatal(fmt.Sprintf("Error linking genre to album: %s", err))
			}
		}
	}

	for _, artist := range album.artists {
		artistId, ok := artistCache[artist.name]
		if !ok {
			artistId, err = createArtist(db, artist)
			if err != nil {
				log.Fatal(fmt.Sprintf("Error creating artist: %s", err))
			} 
		}
		
		_, err := MaybeExecute(db, 
				`INSERT INTO AlbumArtist (AlbumID, ArtistID) VALUES (?, ?)`, 
				album.id, artistId)
			
		if err != nil {
			log.Fatal(fmt.Sprintf("Error linking artist to album: %s", err))
		} 
		}
}

func createArtist(db *sql.DB, artist *Artist) (int64, error) {
	result, err := MaybeExecute(db, 
		`INSERT INTO Artist (Name, SortAs) VALUES (?,?)`, 
		artist.name, artist.name)
	
	if err != nil {
		log.Fatal(fmt.Sprintf("Error inserting artist: %s", err))
	}
	
	return result.LastInsertId()
}