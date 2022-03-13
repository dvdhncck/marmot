package marmot

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

func HttpHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, "Hello, there\n")
}

func GoForIt() {
	db, err := sql.Open("mysql", "dave:dave@tcp(127.0.0.1:3306)/marmot")

	log.SetOutput(os.Stdout) // daemons should log to stdout

	if err != nil {
		panic(err.Error())
	}

	defer db.Close()

	if settings.server {
		http.HandleFunc("/", HttpHandler)
		fmt.Println("Server started at port 8088")
		log.Fatal(http.ListenAndServe(":8088", nil))
		return
	}

	if !settings.HasToken() {
		usage()
	}

	for settings.HasToken() {

		token := settings.NextToken()

		switch token {
		case `genre`:
			if settings.HasToken() {
				action := settings.NextToken()
				switch action {
				case `list`:
					ListGenres(db)
				case `add`:
					log.Fatal(`Genre action not implemented`)
				default:
					log.Fatal(`Unknown genre action: `, action)
				}
			} else {
				log.Fatal("Expected an action after 'genre'")
			}

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
}
