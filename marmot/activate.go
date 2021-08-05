package marmot

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func GoForIt() {
	db, err := sql.Open("mysql", "dave:dave@tcp(127.0.0.1:3306)/marmot")

	if err != nil {
		panic(err.Error())
	}

	defer db.Close()

	collection := Collection{}

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

		case `ingest`:
			if settings.HasToken() {
				path := settings.NextToken()
				fails := Validate(db, path)
				if fails.IsGood() {
					Ingest(db, path)
				} else {
					fails.Write()
				}
			} else {
				log.Fatal("Expected path after 'prepare'")
			}

		case `prepare`:
			if settings.HasToken() {
				path := settings.NextToken()
				Prepare(db, path)
			} else {
				log.Fatal("Expected path after 'prepare'")
			}
		}
	}
}
