package marmot

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
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
	
	collection.ExportToDatabase(db)
}
