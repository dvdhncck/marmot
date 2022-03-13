package marmot

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/dhowden/tag"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type DbAwareHandler struct {
	db          *sql.DB
	randomState string
}

func (dbAwareHandler *DbAwareHandler) HandlePlaylist(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["q"]
	if ok {
		playlist, err := dbAwareHandler.albumIdToPlaylist(keys[0])
		if err == nil {
			json, err := json.Marshal(playlist)
			if err == nil {
				fmt.Fprintf(w, "%v\n", string(json))
				return
			}
		} else {
			log.Println(fmt.Printf("Bad playlist request: %v", err))
		}
	}
	w.WriteHeader(http.StatusBadRequest)
}

func (dbAwareHandler *DbAwareHandler) HandleTextSearch(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["q"]
	if ok {
		key := keys[0]
		fmt.Fprintf(w, "Hello %s!\n", key)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (dbAwareHandler *DbAwareHandler) HandleGenreSearch(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["q"]
	if ok {
		key := keys[0]		
		albums, err := dbAwareHandler.getAlbumIdsForGenre(key)
		if err == nil {
			json, err := json.Marshal(albums)
			if err == nil {
				fmt.Fprintf(w, "%v\n", string(json))
				return
			}
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (dbAwareHandler *DbAwareHandler) getAlbumIdsForGenre(genre string) ([]*MinimalAlbum, error) {
	result := make([]*MinimalAlbum, 0)

	sql := ``
	if genre == `NEW` {
		sql = `SELECT ID,Location FROM Album WHERE ID IN (SELECT * FROM (SELECT ID FROM Album ORDER BY ID DESC LIMIT 50) AS fudge) ORDER BY rand()`
	} else {
		sql = ``
	}

	results, err := dbAwareHandler.db.Query(sql)
	if err != nil {
		return result, err
	}

	var albumId string
	var location string

	for results.Next() {
		err = results.Scan(&albumId, &location)
		if err != nil {
			return result, err
		}
		result = append(result, &MinimalAlbum{albumId, location})
	}

	return result, nil
}

func (dbAwareHandler *DbAwareHandler) albumIdToLocation(albumId string) (string, error) {
	var location string

	if err := dbAwareHandler.db.QueryRow("SELECT Location FROM Album WHERE ID=?", albumId).Scan(&location); err != nil {
		if err == sql.ErrNoRows {
			return ``, fmt.Errorf("unknown albumId '%v'", albumId)
		}
		return ``, fmt.Errorf("fail %v", err)
	}
	return filepath.Join(settings.rootPath, location), nil
}

func (dbAwareHandler *DbAwareHandler) albumIdToPlaylist(albumId string) (*Playlist, error) {
	location, err := dbAwareHandler.albumIdToLocation(albumId)
	//fmt.Printf("checking: %v\n", location)
	tracks := []*Track{}
	if err == nil {
		outputDirRead, err := os.Open(location)
		if err == nil {
			trackFileInfos, err := outputDirRead.Readdir(0)
			if err == nil {
				for _, track := range trackFileInfos {
					if strings.HasSuffix(strings.ToLower(track.Name()), ".mp3") {
						filePath := filepath.Join(location, track.Name())
						handle, err := os.Open(filePath)
						if err == nil {
							tags, err := tag.ReadFrom(handle)
							if err == nil {
								url := url.PathEscape(filePath)
								tracks = append(tracks, &Track{tags.Title(), tags.Artist(), track.Name(), url})
								// fmt.Printf("track : %v\n", track)
								// fmt.Printf("artist=%v\n", tags.Artist())
								// fmt.Printf("title=%v\n", tags.Title())
							}
						}
					}
				}
				return NewPlaylist(albumId, `thing`, tracks), nil
			}
		}
	}
	return &Playlist{}, err
}
