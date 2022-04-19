package marmot

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
)

func resolvePath(unresolvedPath string) string {
	resolvedPath, err := filepath.Abs(unresolvedPath)
	if err != nil {
		log.Fatal(fmt.Sprintf("Error resolving path '%s' (%s)", unresolvedPath, err))
	}
	return resolvedPath
}

func usage() {
	log.Printf("Welcome to the marmot.\nCommands:\n  search  {query}\nprepare {path}\n  validate {path}\n  ingest {path}\n  genre list\n  genre add {genre}\n")
	os.Exit(0)
}

func GoForIt() {
	db, err := sql.Open("mysql", "dave:dave@tcp(127.0.0.1:3306)/marmot")

	log.SetOutput(os.Stdout) // daemons should log to stdout

	if err != nil {
		panic(err.Error())
	}

	defer db.Close()

	genreButler := NewGenreButler()
	genreButler.ScanLibrary()
	
	// if !settings.HasToken() {
	// 	usage()
	// }

	for settings.HasToken() {

		token := settings.NextToken()

		switch token {
		case `genre`:
			if settings.HasToken() {
				genrePath := settings.NextToken()
				genreButler.ListAlbumsByGenre(genrePath)
				
			} else {
				log.Fatal("Expected value for 'genre' parameter")
			}

		case `genres`:
			genreButler.ListGenreForest()
	

		case `album`:
			genreButler.ListAllAlbums()
			

		case `ingest`:
			if settings.HasToken() {
				path := resolvePath(settings.NextToken())
				fails := Validate(db, path)
				if fails.IsGood() {
					Ingest(db, path)
				} else {
					fails.Write()
				}
			} else {
				log.Fatal("Expected path after 'ingest'")
			}

		case `prepare`:
			if settings.HasToken() {
				path := resolvePath(settings.NextToken())
				Prepare(db, path)
			} else {
				log.Fatal("Expected path after 'prepare'")
			}

		case `validate`:
			if settings.HasToken() {
				path := resolvePath(settings.NextToken())
				fails := Validate(db, path)
				if fails.IsGood() {
					log.Printf("%s is valid", path)
				} else {
					fails.Write()
				}

			} else {
				log.Fatal("Expected path after 'validate'")
			}

		case `help`:
			usage()
		}

	}

	if settings.server {
		httpBitch := NewHttpBitch(genreButler);
		http.HandleFunc("/playlist", httpBitch.HandleGetPlaylist)
		http.HandleFunc("/search", httpBitch.HandleSearchByText)
		http.HandleFunc("/genre", httpBitch.HandleSearchByGenre)

		fmt.Println("Server started at port 8088")
		log.Fatal(http.ListenAndServe(":8088", nil))
	}

}
