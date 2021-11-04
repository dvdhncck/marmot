package marmot

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"log"
    "os"
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
	log.Printf("Welcome to the marmot.\nCommands:\n  prepare {path}\n  validate {path}\n  ingest {path}\n  genre list\n  genre add {genre}\n")
	os.Exit(0)
}

func GoForIt() {
	db, err := sql.Open("mysql", "dave:dave@tcp(127.0.0.1:3306)/marmot")

	if err != nil {
		panic(err.Error())
	}

	defer db.Close()

	collection := Collection{}

	if ! settings.HasToken() {
		usage()
	}

	for settings.HasToken() {

		token := settings.NextToken()

		switch token {
		case `db_import`:
			count := collection.LoadFromDatabase(db, ``)
			collection.getMediaFolders(db)
			collection.getArtistsForCollection(db)
			collection.getGenresForCollection(db)
			collection.validate()
			log.Printf("Imported %d albums from database", count)

		case `db_export`:
			collection.WriteToDatabase(db)

		case `excel_import`:
			ExcelRead(settings.ImportFileName(), collection)
			log.Printf("Read collection with %d items from %s", collection.Size(), settings.ImportFileName())

		case `excel_export`:
			ExcelWrite(settings.ExportFileName(), collection)
			log.Printf("Wrote collection to %s", settings.ExportFileName())

		case `remap_locations`:
			collection.RemapLocations()

		case `sanitise`:
			collection.Sanitise(db)

		case `translocate`:
			collection.Translocate(db)

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
