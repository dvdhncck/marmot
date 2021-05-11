package marmot

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

func GoForIt() {
	db, err := sql.Open("mysql", "dave:dave@tcp(127.0.0.1:3306)/marmot")

	if err != nil {
		panic(err.Error())
	}

	defer db.Close()

	collection := Collection{}

	if settings.DoImportFromDatabase() {

		count := collection.retrieve(db, ``)
		collection.getMediaFolders(db)
		collection.getArtists(db)
		collection.getGenres(db)
		collection.validate()

		log.Printf("Imported %d albums from database", count)
	}

	if settings.DoRemapLocations() {
		collection.RemapLocations()
	}

	if settings.DoExportToExcel() {
		ExcelWrite(settings.ExportFileName(), collection)
		log.Printf("Wrote collection to %s", settings.ExportFileName())
	}

	if settings.DoImportFromExcel() {
		ExcelRead(settings.ImportFileName(), collection)
		log.Printf("Read collection with %d items from %s", collection.Size(), settings.ImportFileName())
	}
	
	if settings.DoExportToDatabase() {
		collection.ExportToDatabase(db)
	}

}
