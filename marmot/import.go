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

	count := collection.retrieve(db, `WHERE NAME LIKE '%Hotel%'`)
	collection.getMediaFolders(db)
	collection.getArtists(db)
	collection.getGenres(db)
	collection.validate()
	collection.writeToJson(`test.json`)

	log.Printf("Scanned %d albums", count)
}
