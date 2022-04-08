package marmot

/*
  albumId     anonomly
  564         no id3 tags

  http://skink.lan/music/jam_and_spoon__tripomatic_fairytales_2001/11_tripomatic_fairytales_2001.mp3

*/
import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/dhowden/tag"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type HttpBitch struct {
	genreButler *GenreButler
}

func NewHttpBitch(genreButler *GenreButler) *HttpBitch {
	// genreMap := make(map[string]string, 10)
	// genreMap[`CLASSIC`] = `24)`
	// genreMap[`SOUNDTRACK`] = `34`
	// genreMap[`WORLD`] = `30,32,33`
	// genreMap[`JAZZ`] = `11,29,20,1`
	// genreMap[`POP`] = `16`
	// genreMap[`ROCKS`] = `27`
	// genreMap[`BEATS`] = `18,40,26,31,41`
	// genreMap[`BLEEPS`] = `44,42,17`
	// genreMap[`RANDOM`] = `24,34,30,32,33,11,29,20,1,16,27,18,40,26,31,41,44,42,17`

	return &HttpBitch{genreButler}
}

func (dbAwareHandler *HttpBitch) HandlePlaylist(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
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

func (dbAwareHandler *HttpBitch) HandleTextSearch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	keys, ok := r.URL.Query()["q"]
	if ok {
		key := keys[0]
		albums, err := dbAwareHandler.getAlbumsForText(key)
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

func (dbAwareHandler *HttpBitch) HandleGenreSearch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	keys, ok := r.URL.Query()["q"]
	if ok {
		key := keys[0]
		albums, err := dbAwareHandler.getAlbumsForGenre(key)
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

func (dbAwareHandler *HttpBitch) getAlbumByAlbumId(albumId string) (*AlbumMetadata, error) {
	sql :=
		`SELECT al.Location, al.Name, GROUP_CONCAT(ar.Name)
		 FROM Album AS al 
         LEFT JOIN AlbumArtist AS aa ON aa.AlbumID=al.ID 
	     LEFT JOIN Artist AS ar ON aa.ArtistID=ar.ID 
		 WHERE al.ID = ?
		 GROUP BY al.Location, al.Name`

	results, err := dbAwareHandler.db.Query(sql, albumId)
	if err == nil {
		var location string
		var title string
		var artists string

		for results.Next() {
			err = results.Scan(&location, &title, &artists)
			if err == nil {
				return &AlbumMetadata{albumId, location, title, artists}, nil
			}
		}
	}

	log.Println(fmt.Sprintf("fail %s", err))
	return nil, err
}

func (dbAwareHandler *DbAwareHandler) getAlbumsForText(text string) ([]*AlbumMetadata, error) {
	result := make([]*AlbumMetadata, 0)

	if len(strings.TrimSpace(text)) > 0 {

		// regex for start of line or start of word (i.e. prefixed by a non-word)
		pattern := `(^|[:blank:])+` + text

		sql :=
			`SELECT al.ID, al.Location, al.Name, GROUP_CONCAT(ar.Name)
			FROM Album AS al 
			LEFT JOIN AlbumArtist AS aa ON aa.AlbumID=al.ID 
			LEFT JOIN Artist AS ar ON aa.ArtistID=ar.ID 
			WHERE al.Name REGEXP ? OR ar.Name REGEXP ?
			GROUP BY al.ID, al.Location, al.Name`

		results, err := dbAwareHandler.db.Query(sql, pattern, pattern)
		if err != nil {
			log.Println(fmt.Sprintf("fail %s", err))

			return result, err
		}

		var albumId string
		var location string
		var title string
		var artists string

		for results.Next() {
			//log.Println("row!")

			err = results.Scan(&albumId, &location, &title, &artists)
			if err != nil {
				return result, err
			}

			result = append(result, &AlbumMetadata{albumId, location, title, artists})
		}
	}

	return result, nil
}

func (dbAwareHandler *DbAwareHandler) getAlbumsForGenre(genre string) ([]*AlbumMetadata, error) {
	result := make([]*AlbumMetadata, 0)

	innerSql := ``

	if genre == `NEW` {
		// MySQl (circa 2021) doesn't support LIMIT in an inner query, hence fudge
		innerSql = `SELECT ID FROM Album WHERE ID IN (SELECT * FROM (SELECT ID FROM Album ORDER BY ID DESC LIMIT 50) AS fudge)`
	} else {
		innerSql = `SELECT DISTINCT(al.Id) FROM Album AS al 
			   LEFT JOIN AlbumGenre AS ag ON ag.AlbumID = al.Id
		       WHERE ag.GenreId IN (` + (*dbAwareHandler.genreMap)[genre] + `)`
	}

	// GROUP_CONCAT collapses the artists, giving us a single row, other fields need to be GROUP BY

	outerSql :=
		`SELECT al.ID, al.Location, al.Name, GROUP_CONCAT(ar.Name)
		 FROM Album AS al
         LEFT JOIN AlbumArtist AS aa ON aa.AlbumID=al.ID
	     LEFT JOIN Artist AS ar ON aa.ArtistID=ar.ID
	     WHERE al.ID  IN (` + innerSql + `)
		 GROUP BY al.ID, al.Location, al.Name
		 ORDER BY RAND()`

	
	//log.Println(fmt.Sprintf("%v", outerSql))

	results, err := dbAwareHandler.db.Query(outerSql)
	if err != nil {
		log.Println(fmt.Sprintf("fail %s", err))
		return result, err
	}

	var albumId string
	var location string
	var title string
	var artists string

	for results.Next() {
		err = results.Scan(&albumId, &location, &title, &artists)
		if err != nil {
			return result, err
		}

		result = append(result, &AlbumMetadata{albumId, location, title, artists})
	}

	log.Println(fmt.Sprintf("%d rows for genre %s", len(result), genre))

	return result, nil
}

func (dbAwareHandler *DbAwareHandler) albumIdToPlaylist(albumId string) (*Playlist, error) {
	albumMetadata, err := dbAwareHandler.getAlbumByAlbumId(albumId)
	fmt.Printf("checking: %v\n", albumMetadata)
	if err == nil {
		albumBaseDir := filepath.Join(settings.rootPath, albumMetadata.Location)
		outputDirRead, err := os.Open(albumBaseDir)
		if err == nil {
			trackFileInfos, err := outputDirRead.Readdir(0)
			if err == nil {
				fmt.Printf("scanning tracks %v\n", trackFileInfos)
				tracks := []*Track{}
				for index, track := range trackFileInfos {
					if strings.HasSuffix(strings.ToLower(track.Name()), ".mp3") {
						filePath := filepath.Join(albumBaseDir, track.Name())
						handle, err := os.Open(filePath)
						if err == nil {
							tags, err := tag.ReadFrom(handle)
							if err == nil {
								number := index + 1
								url := filepath.Join(`music`,albumMetadata.Location,track.Name())
								title := resolveTitle(tags.Title(), number)
								artist := tags.Artist()
								tracks = append(tracks, &Track{number, title, artist, url})
							}
						}
					}
				}
				fmt.Printf("found: %v\n", tracks)
				return NewPlaylist(albumMetadata, tracks), nil
			}
		}
	}
	log.Println(fmt.Sprintf("fail %s", err))
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
