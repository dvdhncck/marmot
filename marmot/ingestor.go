package marmot

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/disintegration/imaging"
	"io/ioutil"
	"log"
	"path/filepath"
)

// this assumes that the path has been validated by the Acceptor
func Ingest(db *sql.DB, path string, collection Collection) {
	contentsFileInfos, err := ioutil.ReadDir(path) // this is the contents of 'path' - it should either be a set of files or a set of directories, but not a mixture
	if err == nil {
		count := countSubDirsIn(contentsFileInfos) // shallow search
		if count == len(contentsFileInfos) {
			// all directories
			for _, fileInfo := range contentsFileInfos {
				IngestPath(db, filepath.Join(path, fileInfo.Name()), collection)
			}
		} else {
			// all files
			IngestPath(db, path, collection)
		}
	}

}

// this assumes that the path has been validated by the Acceptor
func IngestPath(db *sql.DB, path string, collection Collection) {
	expectedPath := filepath.Join(path, "meta.json")

	file, _ := ioutil.ReadFile(expectedPath)
	metadata := metadata{}
	_ = json.Unmarshal([]byte(file), &metadata)

	album := NewAlbumFromFilesystem(path, metadata.Title, metadata.Artists, metadata.Genres)
	collection.Add(db, album)

	settings.dryRun = true

	collection.WriteToDatabase(db)
}

// this assumes that the album's path has been validated by the Acceptor (i.e. there is an image there)
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
