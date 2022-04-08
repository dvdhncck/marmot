package marmot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type GenrePath struct {
	path string
}

type GenreButler struct {
	albumCount       	int
	counts           	*map[string]int
	genrePathToAlbumsIds *map[string][]string	
}

/*

album X has genre string:
  search for "pop.modern" find


  e.g.
   won_karwai__in_the_mood_for_love
    ["Sunday Afternoon","Classical","Soundtrack","Compilation"]


      "X.Classical" : "X.Classical",
      "X.Soundtrack" : "X.Soundtrack",
	  "World.Brazilian" : "World.Brazilian",
	  "World.African" : "World.African",
	  "World.Latino" : "World.Latino",
	  "World.Putumayo" : "World.Putumayo",
	  "World.Hawaiian" : "World.Hawaiian",
      "Jazz.Crooning" : "Jazz.Crooning",
	  "Jazz.Modern" : "Jazz.Modern",
	  "Jazz.40-50s" : "Jazz.40-50s",
	  "Jazz.Easy Listening" : "Jazz.Easy Listening",
	  "Jazz.Cocktail" : "Jazz.Cocktail",
	  "Pop.60s" : "Pop.60s",
	  "Pop.Modern" : "Pop.Modern",
	  "Pop.Beck" : "Pop.Beck",
      "Rock.Chart" : "Rock.Chart",
	  "Rock.Hard" : "Rock.Hard",
	  "Rock.US" : "Rock.US",
	  "Beats.Techno" : "Beats.Techno",
	  "Beats.Downtempo" : "Beats.Downtempo",
	  "Beats.Drum & Bass" : "Beats.Drum & Bass",
	  "Beats.Electro" : "Beats.Electro",
	  "Beats.Turntablism" : "Beats.Turntablism",
	  "Bleeps.Ambient" : "Bleeps.Ambient",
	  "Bleeps.LoFi" : "Bleeps.LoFi",
	  "Bleeps.Warp" : "Bleeps.Warp",
	  "Bleeps.Electronica" : "Bleeps.Electronica",
	  "Bleeps.Glitch" : "Bleeps.Glitch",
	  "Dub.Reggae" : "Dub.Reggae",
	  "Dub.On-u-Sound" : "Dub.On-u-Sound",
	  "Dub.Perry" : "Dub.Perry",

*/

var genreMap = map[string]string{
	"Classical":  "Classical",
	"Soundtrack": "Soundtrack",
	"Brazilian":  "World.Brazilian",
	"African":    "World.African",
	"Latino":     "World.Latino",
	"Putumayo":   "World.Putumayo",
	"Hawaiian":   "World.Hawaiian",
	"Crooning":   "Jazz.Crooning",
	//	"Modern":         "Jazz.Modern",
	"40-50s":         "Jazz.40-50s",
	"Easy Listening": "Jazz.Easy Listening",
	"Cocktail":       "Jazz.Cocktail",
	"Pop":            "Pop",
	"World":          "World",

	"60s": "Pop.60s",
	//	"Modern":          "Pop.Modern",
	"Beck":        "Pop.Beck",
	"Chart":       "Rock.Chart",
	"Techno":      "Beats.Techno",
	"Downtempo":   "Beats.Downtempo",
	"Drum & Bass": "Beats.Drum & Bass",
	"Rap":         "Beats.Rap",
	"Electro":     "Beats.Electro",
	"Hip Hop":     "Beats.Hip Hop",
	"Trip Hop":    "Beats.Trip-Hop",
	"Turntablism": "Beats.Turntablism",
	"Bleeps":      "Bleeps",
	"Ambient":     "Bleeps.Ambient",
	"LoFi":        "Bleeps.LoFi",
	"Warp":        "Bleeps.Warp",
	"Electronica": "Bleeps.Electronica",
	"Glitch":      "Bleeps.Glitch",
	"Dub":         "Dub",
	"Reggae":      "Dub.Reggae",
	"On-u-Sound":  "Dub.On-u-Sound",
	"Perry":       "Dub.Perry",
}

func (genreButler *GenreButler) decodeGenres(genres []string) (string, error) {
	mapped_genres := []string{}
	for _, original_genre := range genres {
		(*genreButler.counts)[original_genre] += 1

		mapped_genre, exists := genreMap[original_genre]

		if exists {
			mapped_genres = append(mapped_genres, mapped_genre)
		} else {
			mapped_genres = append(mapped_genres, original_genre)
		}
	}
	return strings.Join(mapped_genres, `,`), nil
}

func (genreButler *GenreButler) transformMetaFile(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if !info.IsDir() && filepath.Base(path) == `meta.json` {
		//log.Println(path)

		data, err := ioutil.ReadFile(path)

		if err != nil {
			return err
		}

		metadata := Metadata{}
		err = json.Unmarshal([]byte(data), &metadata)
		if err != nil {
			return err
		}

		metadata.GenrePaths, err = genreButler.decodeGenres(metadata.Genres)

		if err != nil {
			return err
		}

		jsonRep, err := json.Marshal(&metadata)
		if err != nil {
			return err
		}

		err = ioutil.WriteFile(path, jsonRep, 0644)

		log.Println(fmt.Printf("Meta2: %v", jsonRep))

		if err != nil {
			return err
		}

		genreButler.albumCount += 1

	}
	return err
}

func (genreButler *GenreButler) assimilate(albumPath string, genrePath string)  {

	existingAlbumPaths, exists := (*genreButler.genrePathToAlbumsIds)[genrePath]
	if !exists {
		existingAlbumPaths = []string{}
	}
	(*genreButler.genrePathToAlbumsIds)[genrePath] = append(existingAlbumPaths, albumPath)

	log.Println(fmt.Sprintf(">> %v %v", albumPath, genrePath))
}

func (genreButler *GenreButler) gatherMetaData(albumPath string, info os.FileInfo, err error) error {
	if err == nil {
		if !info.IsDir() && filepath.Base(albumPath) == `meta.json` {
			//log.Println(path)

			data, err := ioutil.ReadFile(albumPath)

			if err == nil {

				metadata := Metadata{}
				err = json.Unmarshal([]byte(data), &metadata)
				
				if err == nil {
					// get root folder for file
					rootAlbumPath := filepath.Dir(albumPath)	

					genrePathList := strings.Split(metadata.GenrePaths, ",")

					for _, genrePath := range genrePathList {
						genreButler.assimilate(rootAlbumPath, genrePath)
					}

				}

				genreButler.albumCount += 1
			}
		}
	}
	return err
}

func (genreButler *GenreButler) ScanLibrary() {
	start := time.Now()

	err := filepath.Walk(settings.rootPath, genreButler.gatherMetaData)

	if err != nil {
		log.Fatal(err)
	}

	log.Println(fmt.Sprintf("Visited %v albums in %v", genreButler.albumCount, time.Since(start)))
}


func (genreButler *GenreButler) ListAllGenres() {
	log.Println(fmt.Sprintf("Genre list:\n"))
	
	for genrePath, _ := range *genreButler.genrePathToAlbumsIds {
		fmt.Printf("%v\n", genrePath)
	}	
}

func (genreButler *GenreButler) ListAlbumsByGenre(targetGenrePath string) {
	for genrePath, albumPaths := range *genreButler.genrePathToAlbumsIds {
		 // if prefix
		if strings.HasPrefix(genrePath, targetGenrePath) {
			fmt.Printf("%v\n", genrePath)
			for _, albumPath := range albumPaths {
				fmt.Printf("\t%v\n", albumPath)
			}
		}
	}	
}

func NewGenreButler() *GenreButler {
	doings := make(map[string][]string, 10)
	counts := make(map[string]int)
	genreButler := &GenreButler{0, &counts, &doings}
	return genreButler
}
