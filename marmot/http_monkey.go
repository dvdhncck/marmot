package marmot

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dhowden/tag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type HttpMonkey struct {
	genreButler *GenreButler
}

func NewHttpMonkey(genreButler *GenreButler) *HttpMonkey {
	return &HttpMonkey{genreButler}
}

func (httpMonkey *HttpMonkey) HandleGetGenres(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	json, err := json.Marshal(httpMonkey.genreButler.genreForest)

	if err == nil {
		fmt.Fprintf(w, "%v\n", string(json))
		return
	} else {
		log.Println(fmt.Printf("Bad playlist request. %v", err))
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (httpMonkey *HttpMonkey) HandleGetPlaylist(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	keys, ok := r.URL.Query()["q"]
	if ok {
		playlist, err := httpMonkey.albumIdToPlaylist(keys[0]) // albumId as parameter
		if err == nil {
			json, err := json.Marshal(playlist)
			if err == nil {
				fmt.Fprintf(w, "%v\n", string(json))
				return
			}
		} else {
			log.Println(fmt.Printf("Bad playlist request. %v", err))
		}
	}
	w.WriteHeader(http.StatusBadRequest)
}

func (httpMonkey *HttpMonkey) HandleSearchByText(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	keys, ok := r.URL.Query()["q"]
	if ok {
		key := keys[0]
		albums, err := httpMonkey.genreButler.GetAlbumsForText(key)
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

func (httpMonkey *HttpMonkey) HandleSearchByGenre(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	keys, ok := r.URL.Query()["q"]
	if ok {
		key := keys[0]
		albums, err := httpMonkey.genreButler.GetAlbumsForGenrePath(key)
		if err == nil {
			json, err := json.Marshal(albums)
			if err == nil {
				fmt.Fprintf(w, "%v\n", string(json))
				return
			}
		}
		log.Println(fmt.Printf("Bad genre search. %v", err))
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}

}

func (httpMonkey *HttpMonkey) albumIdToPlaylist(albumIdString string) (*Playlist, error) {
	albumId, err := strconv.Atoi(albumIdString)
	if err != nil {
		return nil, errors.New("bad albumId")
	}

	metadata, err := httpMonkey.genreButler.GetMetadataById(int64(albumId))
	if err == nil {
		outputDirRead, err := os.Open(metadata.Location)
		if err == nil {
			// TODO: confirm this is alphabetical
			trackFileInfos, err := outputDirRead.Readdir(0)
			if err == nil {
				//fmt.Printf("scanning tracks %v\n", trackFileInfos)
				tracks := []*Track{}
				for index, track := range trackFileInfos {
					if strings.HasSuffix(strings.ToLower(track.Name()), ".mp3") {
						filePath := filepath.Join(metadata.Location, track.Name())
						handle, err := os.Open(filePath)
						if err == nil {
							tags, err := tag.ReadFrom(handle)
							if err == nil {
								number := index + 1
								url := filepath.Join(metadata.UrlBase, track.Name())
								title := resolveTitle(tags.Title(), number)
								artist := tags.Artist()
								tracks = append(tracks, &Track{number, title, artist, url})
							}
						}
					}
				}
				log.Println(fmt.Printf("Scan. %v, %d", metadata.Location, len(tracks)))
				//fmt.Printf("found: %v\n", tracks)
				return NewPlaylist(metadata, tracks), nil
			}
		}
	}

	log.Println(fmt.Sprintf("FAIL Scan. %v", err))
	return &Playlist{}, err
}

func resolveTitle(id3tag string, trackNumber int) string {
	trimmed := strings.TrimSpace(id3tag)
	if len(trimmed) > 0 {
		return trimmed
	} else {
		return fmt.Sprintf(`Track %d`, trackNumber)
	}
}
